package tests_test

import (
	. "github.com/onsi/ginkgo"
	f "kubevirt.io/kubevirt-ansible/tests/framework"
	"time"
)

// template parameters
const (
	cirrosURL = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	invalidPVCURL      = "https://noneexist.com"
	pvcName            = "golden-pvc"
	pvcName1            = "golden-pvc1"
	vmName             = "test-vm"
	vmAPIVersion       = "kubevirt.io/v1alpha2"
	rawPVCFilePath     = "tests/manifests/golden-pvc.yml"
	rawVMFilePath      = "tests/manifests/test-vm.yml"
)

var _ = Describe("Importing and starting a VM using CDI", func() {
	var dstPVCFilePath, dstVMFilePath, newPVCName, url string

	BeforeEach(func() {
		dstPVCFilePath = "/tmp/test-pvc.json"
		dstVMFilePath = "/tmp/test-vm.json"
		newPVCName = pvcName
		url = cirrosURL
	})

	JustBeforeEach(func() {
		f.ProcessTemplateWithParameters(rawPVCFilePath, dstPVCFilePath, "PVC_NAME="+newPVCName, "EP_URL="+f.ReplaceImageURL(url))
		f.CreateResourceWithFilePath(dstPVCFilePath, "")
	})

	Context("PVC with valid image url", func() {

		It("will succeed", func() {
			f.WaitUntilResourceReadyByName("pvc", pvcName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes", "")
			f.WaitUntilResourceReadyByLabelTimeOut("pod", f.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Succeeded", "", 10*time.Minute)
			f.DeleteResourceWithLabel("pod", f.CDI_LABEL_SELECTOR, "")
			f.ProcessTemplateWithParameters(rawVMFilePath, dstVMFilePath, "VM_NAME="+vmName, "PVC_NAME="+pvcName, "VM_APIVERSION="+vmAPIVersion)
			f.CreateResourceWithFilePath(dstVMFilePath, "")
			f.WaitUntilResourceReadyByName("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running", "")
		})
	})

	Context("PVC with invalid image url", func() {
		BeforeEach(func() {
			newPVCName = pvcName1
			url = invalidPVCURL
		})

		It("will be failed because the PVC should become failed", func() {
			f.WaitUntilResourceReadyByLabel("pod", f.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Failed", "")
		})
	})

})
