package tests_test

import (
	"fmt"
	"flag"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Sanity", func() {

	flag.Parse()

	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	var vmi *v1.VirtualMachineInstance

	BeforeEach(func() {
		tests.BeforeTestCleanup()
		vmi = tests.NewRandomVMIWithEphemeralDisk(tests.RegistryDiskFor(tests.RegistryDiskAlpine))
	})

	Describe("Creating a VM", func() {
		It("should success", func() {
			_, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
			Expect(err).To(BeNil())
		})

		It("should start it", func() {
			vmi, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
			Expect(err).To(BeNil())
			
			Eventually(func () bool {
				vmis, err := virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).List(&metav1.ListOptions{})
				if err != nil {
					return false
				}
				return len(vmis.Items) > 0
			})
			
			vmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(vmi.Name, &metav1.GetOptions{})
			Expect(err).To(BeNil())

			for _, c := range vmi.Status.Conditions {
				fmt.Println(c.Reason)
				fmt.Println(c.Message)
			}
			
			tests.WaitForSuccessfulVMIStart(vmi)
		})
	})
})
