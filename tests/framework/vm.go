package framework

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

func StartVirtualMachine(vmName, namespace string) {
	By("Start VM with virtctl")
	_, _, err := ktests.RunCommandWithNS(namespace, "virtctl", "start", vmName)
	Expect(err).ToNot(HaveOccurred())
	WaitUntilResourceExists("vmi", vmName)
	WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running")
}

func StopVirtualMachine(vmName, namespace string) {
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	By("Stop VM with virtctl")
	_, _, err = ktests.RunCommandWithNS(namespace, "virtctl", "stop", vmName)
	Expect(err).ToNot(HaveOccurred())

	updatedVM, err := virtClient.VirtualMachine(namespace).Get(vmName, &metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())

	// Observe the VirtualMachineInstance deleted
	Eventually(func() bool {
		_, err = virtClient.VirtualMachineInstance(updatedVM.Namespace).Get(updatedVM.Name, &metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true
		}
		return false
	}, 300*time.Second, 1*time.Second).Should(BeTrue(), "The vmi did not disappear")

	// Observe the VirtualMachine has not running condition
	By("VM has not the running condition")
	Eventually(func() bool {
		vm, err := virtClient.VirtualMachine(updatedVM.Namespace).Get(updatedVM.Name, &metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		return vm.Status.Ready
	}, 300*time.Second, 1*time.Second).Should(BeFalse())
}
