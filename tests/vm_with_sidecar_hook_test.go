package tests_test

import (
	"flag"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"kubevirt.io/kubevirt-ansible/tests"
)

const (
	vmiName                      = "vmi-with-sidecar-hook"
	vmiFilePath                  = "tests/manifests/vm_with_sidecar_hook.yml"
	checkDmidecodePackage        = "sudo dmidecode -s baseboard-manufacturer | grep 'Radical Edward' | wc -l\n"
)

var _ = Describe("VMI with sidecar hook test", func() {
	
	flag.Parse()

	Context("Check package installtion with hook", func() {
		It("will succeed", func() {
			By("Create VM with hook and install dmidecode")
			tests.CreateResourceWithFilePathTestNamespace(vmiFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmiName, "-o=jsonpath='{.status.phase}'", "Running")
		
			By("Expecting console")
			expecter, err := tests.LoggedInFedoraExpecter(vmiName, tests.NamespaceTestDefault, 240)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()

			By("Checking dmidecode manufacturer")
				_, err = expecter.ExpectBatch([]expect.Batcher{
					&expect.BSnd{S: checkDmidecodePackage},
					&expect.BExp{R: "1"},
				}, 180*time.Second)
				Expect(err).ToNot(HaveOccurred())
		})
	})
})