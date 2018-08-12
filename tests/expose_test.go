package tests_test

import (
	"flag"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt-ansible/tests"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

const (
	vmName       = "vm-cirros"
	testPod      = "testpod"
	vmCirrosPath = "tests/manifests/vm-cirros.yml"
)

var _ = Describe("VM/VMI system test", func() {

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	Context("Expose service via ssh and console VNC", func() {

		It("the VM should be created successfully", func() {
			args := []string{"create", "-f", vmCirrosPath, "-n", tests.NamespaceTestDefault}
			_, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		It("the VM should be started by virtctl", func() {
			args := []string{"start", vmName, "-n", tests.NamespaceTestDefault}
			_, err := ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		Specify("the VMI should be running", func() {
			By("Wait until vmi ready")
			Eventually(func() bool {
				args := []string{"get", "vmi", vmName, "-o=jsonpath='{.status.phase}'", "-n", tests.NamespaceTestDefault}
				out, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return strings.Contains(out, "Running")
			}, time.Duration(2)*time.Minute).Should(BeTrue(), "Timed out waiting for vmi to appear")
		})

		It("the VM should be exposed by virtctl via ClusterIP", func() {
			args := []string{"expose", "vmi", vmName, "-n", tests.NamespaceTestDefault, "--name=myvm-ssh", "--port=22", "--type=ClusterIP"}
			_, err := ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
		})

		It("the VMI should be able to connect via ssh", func() {
			By("Create a pod as ssh client to connect to VMI")
			args := []string{"run", testPod, "-n", tests.NamespaceTestDefault, "--image=shiywang/test-ssh", "--", "sleep", "3600"}
			_, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Wait until test pod ready")
			Eventually(func() bool {
				args := []string{"get", "pod", "-l", "deploymentconfig=" + testPod, "-o=jsonpath='{.items[*].status.phase}'", "-n", tests.NamespaceTestDefault}
				out, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return strings.Contains(out, "Running")
			}, time.Duration(2)*time.Minute).Should(BeTrue(), "Timed out waiting for vmi to appear")

			By("Fetch the name of pod")
			args = []string{"get", "pod", "-l", "deploymentconfig=" + testPod, "-n", tests.NamespaceTestDefault, "-o=jsonpath='{.items[*].metadata.name}'"}
			podName, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Fetch the clusterIP of service")
			args = []string{"get", "svc", "myvm-ssh", "-o=jsonpath='{.spec.clusterIP}'", "-n", tests.NamespaceTestDefault}
			clusterIP, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Run exec ssh command inside the testpod")
			args = []string{"exec", podName[1 : len(podName)-1], "-n", tests.NamespaceTestDefault, "-i", "-t", "--", "sshlogin", clusterIP[1 : len(clusterIP)-1], "cirros", "gocubsgo"}
			out, err := ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).Should(ContainSubstring("vm-cirros"))
		})

		It("the VMI should be able to connect via virtctl console", func() {
			vmi, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(vmName, &metav1.GetOptions{})
			expecter, err := ktests.LoggedInCirrosExpecter(vmi)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
		})

		It("the VM should be stop by virtctl", func() {
			args := []string{"stop", vmName, "-n", tests.NamespaceTestDefault}
			_, err := ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
