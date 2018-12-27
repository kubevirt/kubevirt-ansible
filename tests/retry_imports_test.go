package tests_test

import (
	"flag"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
)

// template parameters
const (
	invalidURL     = "https://noneexist.com"
	dstPVCFilePath = "/tmp/test-pvc.json"
	dstVMFilePath  = "/tmp/test-vm.json"
)

var _ = FDescribe("Importing image using CDI with invalid URL", func() {

	flag.Parse()

	ktests.BeforeAll(func() {
		ktests.BeforeTestCleanup()
		By("Create a PVC with invalid image URL")
		tests.ProcessTemplateWithParameters(rawPVCFilePath, dstPVCFilePath, "PVC_NAME="+pvcName, "EP_URL="+invalidURL)
		tests.CreateResourceWithFilePathTestNamespace(dstPVCFilePath)

	})

	Context("PVC with invalid image url", func() {
		It("Importer pod should be in CrashLoopBackOff status", func() {
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR, "", "CrashLoopBackOff")
		})
		It("PVC deleted successfully", func() {
			importer_pod := GetImporterPodName(pvcName)
			Expect(importer_pod).ToNot(BeEmpty())
			timeout := 5 * 60 * time.Second
			tests.RemovePvcWithTimeout(pvcName, timeout)

		})
		It("Importer pod should be removed", func() {
			importer_pod := GetImporterPodName(pvcName)
			Expect(importer_pod).To(BeEmpty())
		})
	})
})

func GetImporterPodName(pvcName string) string {
	var importer_pod string
	args := []string{"get", "pods", "-o=custom-columns=NAME:.metadata.name"}
	output, _, err := ktests.RunCommand("oc", args...)
	Expect(err).ToNot(HaveOccurred())
	Expect(output).ToNot(BeEmpty())
	pods := strings.Split(output, "\n")
	for _, pod := range pods {
		if strings.Contains(pod, pvcName) && strings.Contains(pod, "importer") {
			importer_pod = pod
			break
		}
	}
	return importer_pod
}
