package cdi_test

import (
	"flag"
	"os"

	. "github.com/onsi/ginkgo"
	"kubevirt.io/kubevirt-ansible/tests"
	ktests "kubevirt.io/kubevirt/tests"
	"github.com/davecgh/go-spew/spew"
)

var _ = Describe("Importing and starting a VM using CDI", func() {
	flag.Parse()
	ktests.SkipIfNoOc()
	dstPVCFilePath := "/tmp/test-pvc.json"
	dstVMFilePath := "/tmp/test-vm.json"

	url, ok := os.LookupEnv("STREAM_IMAGE_URL")
	if !ok {
		url = pvcEPHTTPNOAUTHURL
	}

	Context("PVC with valid image url will succeed", func() {
		It("should create the PVC successfully", func() {
			tests.ProcessTemplateWithParameters(rawPVCFilePath, dstPVCFilePath, "PVC_NAME="+pvcName, "EP_URL="+url)
			tests.CreateResourceWithFilePathTestNamespace("pvc", pvcName, dstPVCFilePath)
		})

		Specify("the PVC should become bound", func() {
			tests.WaitUntilResourceReadyByNameTestNamespace("pvc", pvcName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes pv.kubernetes.io/bound-by-controller:yes")


			out, _ := ktests.RunOcCommand("describe", pvcName, "-n", tests.NamespaceTestDefault)
			spew.Dump("================================================")
			spew.Dump(out)
			spew.Dump("================================================")

		})

		Specify("the importer-pod should become completed", func() {
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", "app=containerized-data-importer", "-o=jsonpath='{.items[0].status.phase}'", "Succeeded")
		})

		It("the VM should be created successfully", func() {
			tests.ProcessTemplateWithParameters(rawVMFilePath, dstVMFilePath, "VM_NAME="+vmName, "PVC_NAME="+pvcName, "VM_APIVERSION="+vmAPIVersion)
			tests.CreateResourceWithFilePathTestNamespace("vmi", vmName, dstVMFilePath)
		})

		Specify("the VM should be running", func() {
			tests.WaitUntilResourceReadyByNameTestNamespace("vmi", vmName, "-o=jsonpath='{.status.phase}'", "Running")

			out, _ := ktests.RunOcCommand("describe", vmName, "-n", tests.NamespaceTestDefault)
			spew.Dump("================================================")
			spew.Dump(out)
			spew.Dump("================================================")
		})
	})
})
