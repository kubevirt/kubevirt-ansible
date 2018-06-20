package tests_test

import (
	"flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
			err := virtClient.RestClient().Post().Resource(tests.VMIResource).Namespace(tests.NamespaceTestDefault).Body(vmi).Do().Error()
			Expect(err).To(BeNil())
		})

		It("should start it", func(done Done) {
			obj, err := virtClient.RestClient().Post().Resource(tests.VMIResource).Namespace(tests.NamespaceTestDefault).Body(vmi).Do().Get()
			Expect(err).To(BeNil())
			tests.WaitForSuccessfulVMIStart(obj)

			close(done)
		}, 45)
	})
})
