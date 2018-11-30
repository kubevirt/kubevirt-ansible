package tests_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	kubev1 "kubevirt.io/kubevirt/pkg/api/v1"
	ktests "kubevirt.io/kubevirt/tests"
)

// template parameters
const rawDataVolumeVMFilePath = "tests/manifests/template/datavolume-vm.yml"
const rawDataVolumeVMIFilePath = "tests/manifests/template/datavolume-vmi.yml"
const rawDataVolumeFilePath = "tests/manifests/template/datavolume.yml"

var _ = FDescribe("DataVolume Integration Test", func() {
	var dataVolumeName, vmName, dstDataVolumeFilePath, url, dstVMIFilePath string

	Context("Datavolume with VM", func() {
		BeforeEach(func() {
			dataVolumeName = "datavolume1"
			vmName = "test-vm-i"
			dstDataVolumeFilePath = "/tmp/test-datavolume-vm.json"
			url = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
		})

		It("Creating VM and start VMI multiple times will be succeed", func() {
			tests.ProcessTemplateWithParameters(rawDataVolumeVMFilePath, dstDataVolumeFilePath, "VM_APIVERSION="+kubev1.GroupVersion.String(), "VM_NAME="+vmName, "IMG_URL="+tests.ReplaceImageURL(url), "DATAVOLUME_NAME="+dataVolumeName)
			tests.CreateResourceWithFilePathTestNamespace(dstDataVolumeFilePath)
			tests.WaitUntilResourceExists("vm", vmName)

			num := 2
			By("Starting and stopping the VirtualMachine number of times")
			for i := 0; i < num; i++ {
				By(fmt.Sprintf("Doing run: %d", i))
				tests.StartVirtualMachine(vmName, ktests.NamespaceTestDefault)
				// Verify console on last iteration to verify the VirtualMachineInstance is still booting properly
				// after being restarted multiple times
				if i == num {
					By("Checking that the VirtualMachineInstance console has expected output")
					expecter, err := tests.LoggedInCirrosExpecter(vmName, tests.NamespaceTestDefault, 360)
					Expect(err).ToNot(HaveOccurred())
					defer expecter.Close()
				}
				tests.StopVirtualMachine(vmName, ktests.NamespaceTestDefault)
			}
			tests.DeleteResourceByNameTestNamespace("vm", vmName)

			By("Ensuring that the PVC is deleted")
			tests.WaitUntilResourceDoesNotExist("pvc", dataVolumeName)
		})
	})

	Context("Datavolume with VMI", func() {
		BeforeEach(func() {
			dataVolumeName = "datavolume2"
			vmName = "test-vmi"
			dstDataVolumeFilePath = "/tmp/test-datavolume.json"
			dstVMIFilePath = "/tmp/test-datavolume-vmi.json"
			url = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
		})

		It("Pre creating datavolume then create VMI will be success", func() {
			tests.ProcessTemplateWithParameters(rawDataVolumeFilePath, dstDataVolumeFilePath, "DATAVOLUME_NAME="+dataVolumeName, "IMG_URL="+tests.ReplaceImageURL(url))
			tests.CreateResourceWithFilePathTestNamespace(dstDataVolumeFilePath)
			tests.ProcessTemplateWithParameters(rawDataVolumeVMIFilePath, dstVMIFilePath, "VM_APIVERSION="+kubev1.GroupVersion.String(), "VM_NAME="+vmName, "DATAVOLUME_NAME="+dataVolumeName)
			tests.CreateResourceWithFilePathTestNamespace(dstVMIFilePath)
			tests.WaitUntilResourceExists("vmi", vmName)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running")
			By("Checking that the VirtualMachineInstance console has expected output")
			expecter, err := tests.LoggedInCirrosExpecter(vmName, tests.NamespaceTestDefault, 360)
			Expect(err).ToNot(HaveOccurred())
			defer expecter.Close()

			tests.DeleteResourceByNameTestNamespace("vmi", vmName)

			By("Ensuring that the PVC is deleted")
			tests.WaitUntilResourceDoesNotExist("pvc", dataVolumeName)
		})
	})

})
