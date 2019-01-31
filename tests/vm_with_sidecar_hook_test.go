package tests_test

import (
	"flag"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
)

const (
	vmiName               = "vmi-with-sidecar-hook"
	vmiFilePath           = "tests/manifests/vm_with_sidecar_hook.yml"
	checkDmidecodePackage = "sudo dmidecode -s baseboard-manufacturer | grep 'Radical Edward' | wc -l\n"
)

var _ = Describe("[rfe_id:839][crit:medium][vendor:cnv-qe@redhat.com][level:component]VMI with sidecar hook test", func() {

	flag.Parse()

	ktests.BeforeAll(func() {
		ktests.BeforeTestCleanup()
	})

	Context("Check package installtion with hook", func() {
		It("[test_id:840]will succeed", func() {
			By("Create VM with hook and install dmidecode")
			tests.CreateResourceWithFilePathTestNamespace(vmiFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmiName, "-o=jsonpath='{.status.phase}'", "Running")

			By("Expecting console")
			expecter, err := tests.LoggedInFedoraExpecter(vmiName, tests.NamespaceTestDefault, 720)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()

			By("Checking dmidecode manufacturer")
			_, err = expecter.ExpectBatch([]expect.Batcher{
				&expect.BSnd{S: checkDmidecodePackage},
				&expect.BExp{R: "1"},
			}, 720*time.Second)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
