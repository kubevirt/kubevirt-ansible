package cdi_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/rand"
	"kubevirt.io/kubevirt-ansible/tests"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("Importing and starting a VMI using CDI", func() {
	prepareCDIResource := func(manifest, url string) string {
		dstFilePath := filepath.Join(tmpTestDir, "test-prepare-file-"+rand.String(5)+".json")
		resourceName := "test-resrouce-" + rand.String(10)

		envURL, ok := os.LookupEnv("STREAM_IMAGE_URL")
		if ok {
			url = envURL
		}
		tests.ProcessTemplateWithParameters(manifest, dstFilePath, "RESOURCE_NAME="+resourceName, "EP_URL="+url)
		tests.CreateResourceWithFilePath(dstFilePath, "")
		return resourceName
	}

	waitForImporterPodWriteImg := func(phase, resourceName string) {
		switch phase {
		case "Succeeded":
			tests.WaitUntilResourceReadyByName("pvc", resourceName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes", "")
			tests.WaitUntilResourceReadyByLabel("pod", tests.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Succeeded", "")
		case "Failed":
			tests.WaitUntilResourceReadyByLabel("pod", tests.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Failed", "")
		case "NOExist":
			Consistently(func() bool {
				args := []string{"get", "pods", "-l", tests.CDI_LABEL_SELECTOR, "-n", tests.NamespaceTestDefault}
				out, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return out == "No resources found.\n"
			}, time.Duration(10)*time.Second).Should(BeTrue())
		}
	}

	startVMIConnectToCDIStorage := func(resourceName string) {
		vmiName := "testvmi" + rand.String(10)
		dstVMFilePath := tmpTestDir + "test-vm-" + rand.String(5) + ".json"
		tests.ProcessTemplateWithParameters(rawVMFilePath, dstVMFilePath, "VM_NAME="+vmiName, "PVC_NAME="+resourceName, "VM_APIVERSION="+vmAPIVersion)
		tests.CreateResourceWithFilePath(dstVMFilePath, "")
		tests.WaitUntilResourceReadyByName("vmi", vmiName, "-o=jsonpath='{.status.phase}'", "Running", "")
	}

	BeforeEach(func() {
		err := os.MkdirAll(tmpTestDir, 0755)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting all pvc with the oc-delete command")
		args := []string{"delete", "pvc", "-l", tests.CDI_TEST_LABEL_SELECTOR, "-n", tests.NamespaceTestDefault}
		_, err := ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		By("Deleting all datavolumes with the oc-delete command")
		args = []string{"delete", "datavolumes.cdi.kubevirt.io", "-l", tests.CDI_TEST_LABEL_SELECTOR, "-n", tests.NamespaceTestDefault}
		_, err = ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		By("Deleting all vmi with the oc-delete command")
		args = []string{"delete", "vmi", "-l", tests.CDI_TEST_LABEL_SELECTOR, "-n", tests.NamespaceTestDefault}
		_, err = ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		//clean up resources after each Entry
		tests.DeleteResourceWithLabel("pod", tests.CDI_LABEL_SELECTOR, "")

		By("Deleting tmp test dir: " + tmpTestDir)
		err = os.RemoveAll(tmpTestDir)
		Expect(err).ToNot(HaveOccurred())
	})

	table.DescribeTable("with different cases:", func(manifest, url string) {
		resourceName := prepareCDIResource(manifest, url)
		switch url {
		case invalidURL:
			waitForImporterPodWriteImg("Failed", resourceName)
		case emptyURL:
			waitForImporterPodWriteImg("NoExist", resourceName)
		default:
			waitForImporterPodWriteImg("Succeeded", resourceName)
			startVMIConnectToCDIStorage(resourceName)
		}
	},
		table.Entry("PVC with valid image url will succeed", rawPVCFilePath, cirrosURL),
		table.Entry("PVC with invalid image url will be failed", rawPVCFilePath, invalidURL),
		table.Entry("PVC with empty image url will be failed", rawPVCFilePath, emptyURL),
		table.Entry("DataVolume with valid image url will succeed", rawDataVolumePath, cirrosURL),
		table.Entry("DataVolume with invalid image url will be failed", rawDataVolumePath, invalidURL),
		table.Entry("DataVolume with empty image url will be failed", rawDataVolumePath, emptyURL),
	)
})
