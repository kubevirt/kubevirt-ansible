package tests_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/src-d/go-git.v4"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
	ginkgo_reporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"
)

func CustomFailHandler(message string, callerSkip ...int) {
	tests.CollectObjDescUsingTestDesc(CurrentGinkgoTestDescription())
	ktests.KubevirtFailHandler(message, callerSkip...)
}

const (
	venvPath = "venv/"
	venvBin  = venvPath + "bin/"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(CustomFailHandler)
	reporters := make([]Reporter, 0)
	if ginkgo_reporters.Polarion.Run {
		reporters = append(reporters, &ginkgo_reporters.Polarion)
	}
	if ginkgo_reporters.JunitOutput != "" {
		reporters = append(reporters, ginkgo_reporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Tests Suite", reporters)
}

var _ = BeforeSuite(func() {
	testNamespaces := []string{ktests.NamespaceTestDefault, ktests.NamespaceTestAlternative}
	for _, v := range testNamespaces {
		err := tests.CreateNamespaceWithParameter(v)
		Expect(err).ToNot(HaveOccurred())
	}

	// Create a directory and clone the ansible-kubevirt-module in it
	os.Mkdir(ansibleModuleClonePath, 0755)
	_, err := git.PlainClone(ansibleModuleClonePath, false, &git.CloneOptions{
		URL:      ansibleModuleURL,
		Progress: os.Stdout,
	})
	Expect(err).ToNot(HaveOccurred())

	// create a virtualenv for ansible
	venv := exec.Command("virtualenv", venvPath)
	venv.Stdout = os.Stdout
	venv.Stderr = os.Stderr
	venvErr := venv.Run()
	if venvErr != nil {
		fmt.Println(venvErr)
		Expect(venvErr).ToNot(HaveOccurred())
	} else {
		fmt.Printf("\nCreated VirtualEnv\n")
	}

	// install Ansible from Github devel
	ansible := exec.Command(venvBin+"pip", "install", "git+https://github.com/ansible/ansible.git", "--upgrade")
	ansible.Stdout = os.Stdout
	ansible.Stderr = os.Stderr
	ansible_err := ansible.Run()
	if ansible_err != nil {
		Expect(ansible_err).ToNot(HaveOccurred())
	} else {
		fmt.Printf("\nInstalled ansible\n")
	}

	// print Ansible version info
	ansible_info := exec.Command(venvBin+"ansible", "--version")
	ansible_info.Stdout = os.Stdout
	ansible_info.Stderr = os.Stderr
	ansible_info_err := ansible_info.Run()
	if ansible_info_err != nil {
		fmt.Println(ansible_info_err)
	}

	// install module dependencies
	pip := exec.Command(venvBin+"pip", "install", "-r", ansibleModuleClonePath+"/requirements-dev.txt", "--upgrade")
	pip.Stdout = os.Stdout
	pip.Stderr = os.Stderr
	pipErr := pip.Run()
	if pipErr != nil {
		Expect(pipErr).ToNot(HaveOccurred())
	} else {
		fmt.Printf("\nInstalled kubevirt module dependencies\n")
	}
})

var _ = AfterSuite(func() {
	testNamespaces := []string{ktests.NamespaceTestDefault, ktests.NamespaceTestAlternative}
	for _, v := range testNamespaces {
		err := tests.RemoveNamespaceWithParameter(v)
		Expect(err).ToNot(HaveOccurred())
	}

	// clean up for cloned module
	os.RemoveAll(ansibleModuleClonePath)
})
