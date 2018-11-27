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
	ktests "kubevirt.io/kubevirt/tests"
)

const (
	ddCommand      = "dd count=10 bs=1024 if=/dev/%s of=/tmp/%s.txt\n"
	checkRNGDevice = "cat /sys/devices/virtual/misc/hw_random/%s\n"
	checkFileSize  = "ls /tmp/%s.txt | wc -l'\n"
	outputDevice   = "virtio_rng.0"
)

var rngAvailable string
var rngCurrent string
var ddWithRandom string
var ddWithHWRandom string
var checkFileRandom string
var checkFileHWRng string

var _ = Describe("VIRT RNG test", func() {
	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	setCommands()

	ktests.BeforeAll(func() {
		ktests.BeforeTestCleanup()
	})

	Context("With VirtIO RNG device", func() {
		var withRngVmi *v1.VirtualMachineInstance
		withRngVmi = ktests.NewRandomVMIWithEphemeralDisk(ktests.RegistryDiskFor(ktests.RegistryDiskAlpine))
		It("Virtio rng device should be present", func() {
			withRngVmi.Spec.Domain.Devices.Rng = &v1.Rng{}

			By("Starting a VMI")
			withRngVmi, err = virtClient.VirtualMachineInstance(ktests.NamespaceTestDefault).Create(withRngVmi)
			Expect(err).ToNot(HaveOccurred())
			ktests.WaitForSuccessfulVMIStart(withRngVmi)

			By("Expecting console")
			expecter, err := ktests.LoggedInAlpineExpecter(withRngVmi)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()

			By("Checking the virtio RNG")
			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: rngAvailable},
				&expect.BExp{R: outputDevice},
				&expect.BSnd{S: rngCurrent},
				&expect.BExp{R: outputDevice},
				&expect.BSnd{S: ddWithRandom},
				&expect.BSnd{S: ddWithHWRandom},
				&expect.BSnd{S: checkFileRandom},
				&expect.BExp{R: "1"},
				&expect.BSnd{S: checkFileHWRng},
				&expect.BExp{R: "1"},
			}, 60*time.Second)
			Expect(err).ToNot(HaveOccurred())
		})
	})

})

func setCommands() {
	rngAvailable = fmt.Sprintf(checkRNGDevice, "rng_available")
	rngCurrent = fmt.Sprintf(checkRNGDevice, "rng_current")
	ddWithRandom = fmt.Sprintf(ddCommand, "random", "random")
	ddWithHWRandom = fmt.Sprintf(ddCommand, "hwrng", "hwrng")
	checkFileRandom = fmt.Sprintf(checkFileSize, "random")
	checkFileHWRng = fmt.Sprintf(checkFileSize, "hwrng")
}
