package tests

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

	var vms [2]*v1.VirtualMachineInstance
	var vmsIP [2]string

	tests.BeforeAll(func() {
		for i := 0; i < 2; i++ {
			vms[i] = tests.NewRandomVMIWithEphemeralDiskAndUserdata(tests.RegistryDiskFor(tests.RegistryDiskCirros), "#!/bin/bash\necho 'hello'\n")
			vms[i], err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vms[i])
			Expect(err).ToNot(HaveOccurred())
			tests.WaitForSuccessfulVMIStart(vms[i])
			vms[i], err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Get(vms[i].Name, &metav1.GetOptions{})
			vmsIP[i] = vms[i].Status.Interfaces[0].IP
		}
	})

	Context("Connectivity between VMs", func() {
			It("two vms ping each other should be successful", func() {
				for i := 0; i < 2; i++ {
					expecter, err := tests.LoggedInCirrosExpecter(vms[i])
					Expect(err).ToNot(HaveOccurred())
					defer expecter.Close()
					cmdCheck := fmt.Sprintf("ping -w 3 %s \n", vmsIP[1-i])
					err = tests.CheckForTextExpecter(vms[i], []expect.Batcher{
						&expect.BSnd{S: cmdCheck},
						&expect.BExp{R: "3 packets transmitted"},
					}, 60)
					Expect(err).ToNot(HaveOccurred())
				}
			})
	})
})
