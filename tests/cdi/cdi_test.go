package cdi_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	"kubevirt.io/kubevirt-ansible/tests"
)

var _ = Describe("Importing and starting a VM using CDI", func() {
	var dstPVCFilePath, dstVMFilePath, newPVCName, url string

	BeforeEach(func() {
		var ok bool
		dstPVCFilePath = "/tmp/test-pvc.json"
		dstVMFilePath = "/tmp/test-vm.json"
		newPVCName = pvcName
		url, ok = os.LookupEnv("STREAM_IMAGE_URL")
		if !ok {
			url = pvcEPHTTPNOAUTHURL
		}
	})

	JustBeforeEach(func() {
		tests.ProcessTemplateWithParameters(rawPVCFilePath, dstPVCFilePath, "PVC_NAME="+newPVCName, "EP_URL="+url)
		tests.CreateResourceWithFilePathTestNamespace(dstPVCFilePath)
	})

	Context("PVC with valid image url", func() {

		It("will succeed", func() {
			tests.WaitUntilResourceReadyByNameTestNamespace("pvc", pvcName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes")
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Succeeded")
			tests.DeleteResourceWithLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR)
			tests.ProcessTemplateWithParameters(rawVMFilePath, dstVMFilePath, "VM_NAME="+vmName, "PVC_NAME="+pvcName, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(dstVMFilePath)
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running")
		})
	})

	Context("PVC with invalid image url", func() {
		BeforeEach(func() {
			newPVCName = pvcName1
			url = invalidPVCURL
		})

		It("will be failed because the PVC should become failed", func() {
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Failed")
		})
	})

})
