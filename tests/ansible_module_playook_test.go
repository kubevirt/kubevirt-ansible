package tests_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	ansibleModuleURL          = "https://github.com/kubevirt/ansible-kubevirt-modules/"
	ansibleModuleClonePath    = "ansible-kubevirt-modules"
	ansibleModulePlaybookPath = "ansible-kubevirt-modules/tests/playbooks/"
	ansibleModuleTestCfg      = "ansible-kubevirt-modules/tests/ansible.cfg"
	ansibleModuleLogPath      = ansibleModuleClonePath + "/ansible.log"
)

var _ = Describe("Run test playbooks from KubeVirt-Ansible module", func() {
	Context("Test playbook ", func() {
		DescribeTable("Playbook name:", func(playbookName string) {
			// use Config file of the module
			os.Setenv("ANSIBLE_CONFIG", ansibleModuleTestCfg)
			defer os.Unsetenv("ANSIBLE_CONFIG")

			// Run the parametrized playbook
			cmd := exec.Command(venvBin+"ansible-playbook", "-vvvv", ansibleModulePlaybookPath+playbookName)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			ok := cmd.Run()
			if ok != nil {
				fmt.Fprint(GinkgoWriter, ok)
				ansibleLog, err := ioutil.ReadFile(ansibleModuleLogPath)
				if err == nil {
					fmt.Fprintf(GinkgoWriter, string(ansibleLog))
				} else {
					fmt.Fprint(GinkgoWriter, err)
				}
			}
			// assert Playbook result
			Expect(ok).ToNot(HaveOccurred())
		},
			// Entry list of playbooks to run from ansible-kubevirt-module
			// Entry("e2e", "e2e.yaml"),
			Entry("k8s-facts", "k8s_facts.yml"),
		)
	})
})
