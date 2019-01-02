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

const (
	invalidURL            = "https://noneexist.com"
	dataVolumeName        = "golden-data-volume"
	dstDataVolumeFilePath = "/tmp/test-data-volume.json"
	rawDataVolumeFilePath = "tests/manifests/golden-data-volume.yml"
	apiVersion            = "cdi.kubevirt.io/v1alpha1"
	timeout               = 1 * 60 * time.Second
)

var _ = Describe("Importing image using CDI with invalid URL", func() {

	flag.Parse()

	ktests.BeforeAll(func() {
		ktests.BeforeTestCleanup()
		By("Create a Data Volume with invalid image URL")
		tests.ProcessTemplateWithParameters(rawDataVolumeFilePath, dstDataVolumeFilePath, "DV_NAME="+dataVolumeName, "EP_URL="+invalidURL, "API_VERSION="+apiVersion)
		tests.CreateResourceWithFilePathTestNamespace(dstDataVolumeFilePath)
	})

	Context("Data Volume with invalid image url", func() {
		It("Importer pod should be removed once data volume is deleted", func() {
			By("Importer pod should be in CrashLoopBackOff status")
			tests.WaitUntilResourceReadyByLabelTestNamespace("pod", tests.CDI_LABEL_SELECTOR, "", "CrashLoopBackOff")

			By("Data Volume deleted successfully")
			importer_pod := GetImporterPodName(dataVolumeName)
			Expect(importer_pod).ToNot(BeEmpty())
			tests.RemoveDataVolume(dataVolumeName, tests.NamespaceTestDefault)

			By("Importer pod should be removed")
			Eventually(func() string {
				importer_pod := GetImporterPodName(dataVolumeName)
				return importer_pod
			}, timeout, 1*time.Second).Should(BeEmpty())
		})
	})
})

func GetImporterPodName(dataVolumeName string) string {
	var importer_pod string
	args := []string{"get", "pods", "-o=custom-columns=NAME:.metadata.name"}
	output, _, err := ktests.RunCommand("oc", args...)
	Expect(err).ToNot(HaveOccurred())
	Expect(output).ToNot(BeEmpty())
	pods := strings.Split(output, "\n")
	for _, pod := range pods {
		if strings.Contains(pod, dataVolumeName) && strings.Contains(pod, "importer") {
			importer_pod = pod
			break
		}
	}
	return importer_pod
}
