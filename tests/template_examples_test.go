package tests_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"kubevirt.io/kubevirt-ansible/tests"
	ktests "kubevirt.io/kubevirt/tests"
)

const (
	fedoraCloudImageDev  string = "registry:5000/kubevirt/fedora-cloud-registry-disk-demo:devel"
	fedoraCloudImageTest string = "kubevirt/fedora-cloud-registry-disk-demo:latest"
)

var _ = Describe("Virtual machine template tests", func() {
	ns := ktests.NamespaceTestDefault

	var vm tests.VMManifest

	BeforeEach(func() {
		ktests.BeforeTestCleanup()
	})

	AfterEach(func() {
		err := os.RemoveAll(filepath.Dir(vm.Manifest))
		Expect(err).ToNot(HaveOccurred(), "failed to remove %s", vm.Manifest)
	})

	Describe("Test fedora template", func() {
		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "vm-template-")
			Expect(err).ToNot(HaveOccurred(), "failed to create a temporary directory in %s", os.TempDir())

			vm.Manifest = filepath.Join(dir, "vm-template-fedora.yaml")
			vm.Name = "fedora"
			vm.Template = "vm-template-fedora"
			vm.Params = []string{"-p", "NAME=" + vm.Name, "-p", "CPU_CORES=2", "-p", "MEMORY=2048Mi"}

			output, _ := vm.ProcessClusterTemplate()
			if strings.Contains(output, fmt.Sprintf("template \"%s\" could not be found", vm.Template)) {
				Skip(fmt.Sprintf("template \"%s\" could not be found", vm.Template))
			}
		})
		It("Create VM from fedora template successfully", func() {
			By("Process fedora template")
			output, err := vm.ProcessClusterTemplate()
			Expect(err).ToNot(HaveOccurred(), "failed to process template %s in namespace %s", vm.Manifest, tests.TemplateNS)

			By("Write the produced string to vm manifest")
			output = strings.Replace(output, fedoraCloudImageDev, fedoraCloudImageTest, -1)
			err = ioutil.WriteFile(vm.Manifest, []byte(output), 0644)
			Expect(err).ToNot(HaveOccurred(), "failed to write output to %s", vm.Manifest)

			By("Create VM from manifest")
			args := []string{"create", "-n", ns, "-f", vm.Manifest}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred(), "failed to run command: oc %s", strings.Join(args, " "))

			By("Start the virtual machine via virtctl")
			args = []string{"-n", ns, "start", vm.Name}
			_, err = ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred(), "failed to run command: virtctl %s", strings.Join(args, " "))

			By("Check that the virtualmachine instance is running")
			args = []string{"get", "-n", ns, "vmi", vm.Name, "--template", "{{.status.phase}}"}
			Eventually(func() string {
				output, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred(), "failed to run command: oc %s", strings.Join(args, " "))
				return output
			}, tests.DefaultTimeoutForVMReady, tests.DefaultPollInterval).Should(Equal("Running"))

			By("Check that the number of virtualmachine cores is matching")
			args = []string{"get", "-n", ns, "vmi", vm.Name, "--template", "{{.spec.domain.cpu.cores}}"}
			output, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred(), "failed to run command: oc %s", strings.Join(args, " "))
			Expect(output).To(Equal("2"))

			By("Check that the size of virtualmachine memory is matching")
			args = []string{"get", "-n", ns, "vmi", vm.Name, "--template", "{{.spec.domain.resources.requests.memory}}"}
			output, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred(), "failed to run command: oc %s", strings.Join(args, " "))
			Expect(output).To(Equal("2Gi"))
		})
	})
})
