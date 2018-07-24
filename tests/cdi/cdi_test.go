package cdi_test

import (
	"flag"
	"os"

	. "github.com/onsi/ginkgo"
	"kubevirt.io/kubevirt-ansible/tests"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("Importing and starting a VM using CDI", func() {
	flag.Parse()
	ktests.SkipIfNoOc()
	dstPVCFilePath := "/tmp/test-pvc.json"
	dstPVCFilePath1 := "/tmp/test-pvc1.json"
	dstVMFilePath := "/tmp/test-vm.json"

	url, ok := os.LookupEnv("STREAM_IMAGE_URL")
	if !ok {
		url = pvcEPHTTPNOAUTHURL
	}

	Context("PVC with valid image url will succeed", func() {
		It("should create the PVC successfully", func() {
			tests.ProcessTemplateWithParameters(rawPVCFilePath, dstPVCFilePath, "PVC_NAME="+pvcName, "EP_URL="+url)
			tests.CreateResourceWithFilePathTestNamespace(dstPVCFilePath)
		})

		Specify("the PVC should become bound", func() {
			tests.WaitUntilResourceReadyByNameTestNamespace("pvc", pvcName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes")
		})

		Specify("the importer-pod should become completed", func() {
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Succeeded")
		})

		It("should delete the importer-pod", func() {
			tests.DeleteResourceWithLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR)
		})

		It("the VM should be created successfully", func() {
			tests.ProcessTemplateWithParameters(rawVMFilePath, dstVMFilePath, "VM_NAME="+vmName, "PVC_NAME="+pvcName, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace(dstVMFilePath)
		})

		Specify("the VM should be running", func() {
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running")
		})
	})

	Context("PVC with invalid image url will be failed", func() {
		It("should create the PVC successfully", func() {
			tests.ProcessTemplateWithParameters(rawPVCFilePath, dstPVCFilePath1, "PVC_NAME="+pvcName1, "EP_URL="+invalidPVCURL)
			tests.CreateResourceWithFilePathTestNamespace(dstPVCFilePath1)
		})

		Specify("the PVC should become failed", func() {
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Failed")
		})
	})


})
