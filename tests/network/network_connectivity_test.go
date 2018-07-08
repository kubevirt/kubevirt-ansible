package network_test

import (
	"flag"
	"fmt"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

var _ = Describe("Network Connectivity", func() {

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	var vma *v1.VirtualMachineInstance
	var vmb *v1.VirtualMachineInstance
	var vmaIP string
	var vmbIP string

	tests.BeforeAll(func() {
		vma = tests.NewRandomVMIWithEphemeralDiskAndUserdata(tests.RegistryDiskFor(tests.RegistryDiskCirros), "#!/bin/bash\necho 'hello'\n")
		vma, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vma)
		Expect(err).ToNot(HaveOccurred())
		tests.WaitForSuccessfulVMIStart(vma)
		vmb = tests.NewRandomVMIWithEphemeralDiskAndUserdata(tests.RegistryDiskFor(tests.RegistryDiskCirros), "#!/bin/bash\necho 'hello'\n")
		_, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmb)
		Expect(err).ToNot(HaveOccurred())
		tests.WaitForSuccessfulVMIStart(vmb)
		vma, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(vma.Name, &metav1.GetOptions{})
		vmb, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(vmb.Name, &metav1.GetOptions{})
		vmaIP = vma.Status.Interfaces[0].IP
		vmbIP = vmb.Status.Interfaces[0].IP
	})

	Context("Connectivity between VMs", func() {

		It("in vmb ping vma should be successful", func() {
			expecter, err := tests.LoggedInCirrosExpecter(vmb)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
			cmdCheck := fmt.Sprintf("ping -w 3 %s \n", vmaIP)
			err = tests.CheckForTextExpecter(vmb, []expect.Batcher{
				&expect.BSnd{S: cmdCheck},
				&expect.BExp{R: "0% packet loss"},
			}, 60)
			Expect(err).ToNot(HaveOccurred())
		})

		It("in vma ping vmb should be successful", func() {
			expecter, err := tests.LoggedInCirrosExpecter(vma)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()
			cmdCheck := fmt.Sprintf("ping -w 3 %s \n", vmbIP)
			err = tests.CheckForTextExpecter(vma, []expect.Batcher{
				&expect.BSnd{S: cmdCheck},
				&expect.BExp{R: "0% packet loss"},
			}, 60)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
