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
			tests.Oc("process").TestNamespace().FilePath(rawPVCFilePath).Param("-p", "PVC_NAME="+pvcName, "-p", "EP_URL="+url).WriteJson(dstPVCFilePath)
			tests.Oc("create").TestNamespace().FilePath(dstPVCFilePath).Run()
		})

		Specify("the PVC should become bound", func() {
			tests.Oc("get").TestNamespace().ResourceType("pvc").ResourceName(pvcName).Query("-o=jsonpath='{.metadata.annotations}'").WaitUntil("pv.kubernetes.io/bind-completed:yes")
		})

		Specify("the importer-pod should become completed", func() {
			tests.Oc("get").ResourceType("pod").ResourceLabel(tests.CDI_LABEL_SELECTOR).Query("-o=jsonpath='{.items[*].status.phase}'").WaitUntil("Succeeded")

		})

		It("should delete the importer-pod", func() {
			tests.Oc("delete").TestNamespace().ResourceType("pod").ResourceLabel(tests.CDI_LABEL_SELECTOR).Run()
		})

		It("the VM should be created successfully", func() {
			tests.Oc("process").TestNamespace().FilePath(rawVMFilePath).Param("-p", "VM_NAME="+vmName, "-p", "PVC_NAME="+pvcName, "-p", "VM_APIVERSION="+vmAPIVersion).WriteJson(dstVMFilePath)
			tests.Oc("create").TestNamespace().FilePath(dstVMFilePath).Run()
		})

		It("the VM should be started by virtctl", func() {
			tests.Oc("start").ResourceName(vmName).Run()
		})

		It("the VM should be stop by virtctl", func() {
			tests.Virtctl("stop").ResourceName(vmName).Run()
		})

		It("the VM should be started by virtctl", func() {
			tests.Virtctl("start").ResourceName(vmName).Run()
		})

		Specify("the VMI should be running", func() {
			tests.Oc("get").ResourceType("vmi").ResourceName(vmName).Query("-o=jsonpath='{.status.phase}'").WaitUntil("Running")
		})
	})

	Context("PVC with invalid image url will be failed", func() {
		It("should create the PVC successfully", func() {
			tests.Oc("process").TestNamespace().FilePath(rawPVCFilePath).Param("-p", "PVC_NAME="+pvcName1, "-p", "EP_URL="+invalidPVCURL).WriteJson(dstPVCFilePath1)
			tests.Oc("create").TestNamespace().FilePath(dstPVCFilePath1)
		})

		Specify("the PVC should become failed", func() {
			tests.Oc("get").ResourceType("pod").ResourceLabel(tests.CDI_LABEL_SELECTOR).Query("-o=jsonpath='{.items[*].status.phase}'").WaitUntil("Failed")
		})
	})

})
