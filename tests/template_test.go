package tests_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

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

	Describe("Test fedora template", func() {
		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "vm-template-")
			Expect(err).ToNot(HaveOccurred())

			vm.Manifest = filepath.Join(dir, "vm-template-fedora.yaml")
			vm.Name = "fedora"
			vm.Template = "vm-template-fedora"
			vm.Params = []string{"-p", "NAME=" + vm.Name, "-p", "CPU_CORES=2", "-p", "MEMORY=2048Mi"}
		})
		It("Create VM from fedora template successfully", func() {
			By("Process fedora template")
			output, err := vm.ProcessClusterTemplate()
			Expect(err).ToNot(HaveOccurred())

			By("Write the produced string to vm manifest")
			output = strings.Replace(output, fedoraCloudImageDev, fedoraCloudImageTest, -1)
			err = ioutil.WriteFile(vm.Manifest, []byte(output), 0644)
			Expect(err).ToNot(HaveOccurred())

			By("Create VM from manifest")
			args := []string{"create", "-n", ns, "-f", vm.Manifest}
			_, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Start the virtual machine via virtctl")
			args = []string{"-n", ns, "start", vm.Name}
			_, err = ktests.RunCommand("virtctl", args...)
			Expect(err).ToNot(HaveOccurred())

			By("Check that the virtualmachine instance is running")
			args = []string{"get", "-n", ns, "vmi", vm.Name, "--template", "{{.status.phase}}"}
			Eventually(func() string {
				output, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return output
			}, time.Minute*10).Should(Equal("Running"))

			By("Check that the virtualmachine cores")
			args = []string{"get", "-n", ns, "vmi", vm.Name, "--template", "{{.spec.domain.cpu.cores}}"}
			output, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("2"))

			By("Check that the virtualmachine memory")
			args = []string{"get", "-n", ns, "vmi", vm.Name, "--template", "{{.spec.domain.resources.requests.memory}}"}
			output, err = ktests.RunCommand("oc", args...)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("2Gi"))
		})
	})
})
