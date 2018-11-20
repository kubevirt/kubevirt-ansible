package tests_test

import (
	"flag"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"kubevirt.io/kubevirt-ansible/tests"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("High performance vm test", func() {
	/*
	* This test includes the features:
	* 1. Headless
	* 2. Support memory over commitment
	 */

	const (
		virtRawVMFilePath          = "tests/manifests/virt-testing-vm.yml"
		graphicDeviceOffStr        = "Autoattach Graphics Device:  false"
		vncErr                     = "Can't connect to websocket (400): No graphics devices are present"
		overcommitGuestOverheadStr = "Overcommit Guest Overhead:  true"
		memoryOvercommit           = "true"
		headless                   = "false"
		vmAPIVersion               = "kubevirt.io/v1alpha3"
	)

	flag.Parse()
	virtClient, err := kubecli.GetKubevirtClient()
	registryDisk := ktests.RegistryDiskFor(ktests.RegistryDiskCirros)
	ktests.PanicOnError(err)

	Context("Headless vm test", func() {
		headlessDstVMFilePath := "/tmp/headlesstest-vm.json"
		headlesstestVMName := "headlesstest"

		It("Create headless VM", func() {
			tests.ProcessTemplateWithParameters(virtRawVMFilePath, headlessDstVMFilePath, "VM_NAME="+headlesstestVMName, "AUTO_GRAPHIC_DEVICE="+headless, "IMAGE_NAME="+registryDisk, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(headlessDstVMFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", headlesstestVMName, "-o=jsonpath='{.status.phase}'", "Running")
		})

		It("Check VM settings with 'oc describe'", func() {
			res := tests.RunOcDescribeCommand("vmis", headlesstestVMName)
			Expect(strings.Contains(res, graphicDeviceOffStr)).To(BeTrue())
		})
		It("[Negative] Check console VNC is disable", func() {
			_, _, err := tests.OpenConsole(virtClient, headlesstestVMName, tests.NamespaceTestDefault, 20*time.Second, "vnc")
			Expect(strings.Contains(string(err.Error()), vncErr)).To(BeTrue())
		})
	})

	Context("Support memory over commitment test", func() {
		memoryOvercommitDstVMFilePath := "/tmp/memoryOvercommit-vm.json"
		memoryOvercommitVMName := "memoryovercommit"

		It("Create memoryOvercommit VM", func() {
			tests.ProcessTemplateWithParameters(virtRawVMFilePath, memoryOvercommitDstVMFilePath, "VM_NAME="+memoryOvercommitVMName, "OVER_COMMIT_GUEST_OVERLOAD="+memoryOvercommit, "IMAGE_NAME="+registryDisk, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(memoryOvercommitDstVMFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", memoryOvercommitVMName, "-o=jsonpath='{.status.phase}'", "Running")
		})
		It("Check VM settings with 'oc describe'", func() {
			res := tests.RunOcDescribeCommand("vmis", memoryOvercommitVMName)
			Expect(strings.Contains(res, overcommitGuestOverheadStr)).To(BeTrue())
		})
	})

	Context("Headless and Support memory over commitment VM test", func() {
		memoryOvercommitDstVMFilePath := "/tmp/headlessAndMemoryOvercommit-vm.json"
		memoryOvercommitVMName := "headlessandmemoryovercommit"

		It("Create headless and memory over commit VM", func() {
			tests.ProcessTemplateWithParameters(virtRawVMFilePath, memoryOvercommitDstVMFilePath, "VM_NAME="+memoryOvercommitVMName, "OVER_COMMIT_GUEST_OVERLOAD="+memoryOvercommit, "AUTO_GRAPHIC_DEVICE="+headless, "IMAGE_NAME="+registryDisk, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(memoryOvercommitDstVMFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", memoryOvercommitVMName, "-o=jsonpath='{.status.phase}'", "Running")
		})
		It("Check VM settings with 'oc describe'", func() {
			res := tests.RunOcDescribeCommand("vmis", memoryOvercommitVMName)
			Expect(strings.Contains(res, overcommitGuestOverheadStr)).To(BeTrue())
			Expect(strings.Contains(res, graphicDeviceOffStr)).To(BeTrue())
		})
		It("[Negative] Check console VNC is disable", func() {
			_, _, err := tests.OpenConsole(virtClient, memoryOvercommitVMName, tests.NamespaceTestDefault, 20*time.Second, "vnc")
			Expect(strings.Contains(string(err.Error()), vncErr)).To(BeTrue())
		})
	})
})
