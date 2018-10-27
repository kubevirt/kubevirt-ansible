package tests_test

import (
	"flag"
	"fmt"
	"time"


	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

const (
	DDCommand       =  "dd count=10 bs=1024 if=/dev/%s of=/tmp/%s.txt\n"
	CheckRNGDevice  =  "cat /sys/devices/virtual/misc/hw_random/%s\n"
	CheckFileSize   =  "ls /tmp/%s.txt | wc -l'\n"
	outputDevice    =  "virtio_rng.0" 
)
var RngAvailable    string
var RngCurrent      string
var DdWithRandom    string
var DdWithHWRandom  string
var CheckFileRandom string 
var CheckFileHWRng  string

var _ = Describe("VIRT RNG test", func() {
	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)
	setCommands() 
	
    Context("With VirtIO RNG device", func() {
		var withRngVmi *v1.VirtualMachineInstance
		withRngVmi = tests.NewRandomVMIWithEphemeralDisk(tests.RegistryDiskFor(tests.RegistryDiskAlpine))
	    It("Virtio rng device should be present", func() {
			withRngVmi.Spec.Domain.Devices.Rng = &v1.Rng{}

			By("Starting a VMI")
			withRngVmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(withRngVmi)
			Expect(err).ToNot(HaveOccurred())
			tests.WaitForSuccessfulVMIStart(withRngVmi)

			By("Expecting console")
			expecter, err := tests.LoggedInAlpineExpecter(withRngVmi)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()

			By("Checking the virtio RNG")
			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: RngAvailable},
				&expect.BExp{R: outputDevice},
				&expect.BSnd{S: RngCurrent},
				&expect.BExp{R: outputDevice},
				&expect.BSnd{S: DdWithRandom},
				&expect.BSnd{S: DdWithHWRandom},
				&expect.BSnd{S: CheckFileRandom},
				&expect.BExp{R: "1"},
				&expect.BSnd{S: CheckFileHWRng},
				&expect.BExp{R: "1"},
			}, 60*time.Second)
			Expect(err).ToNot(HaveOccurred())
    	})
	})
	
})

func setCommands(){
	RngAvailable     = fmt.Sprintf(CheckRNGDevice, "rng_available")
	RngCurrent       = fmt.Sprintf(CheckRNGDevice, "rng_current")
	DdWithRandom   = fmt.Sprintf(DDCommand, "random", "random")
	DdWithHWRandom   = fmt.Sprintf(DDCommand, "hwrng", "hwrng")
	CheckFileRandom  = fmt.Sprintf(CheckFileSize,"random")
	CheckFileHWRng   = fmt.Sprintf(CheckFileSize,"hwrng")
}
