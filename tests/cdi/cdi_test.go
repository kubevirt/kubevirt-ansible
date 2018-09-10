package cdi_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/ginkgo/extensions/table"

	"kubevirt.io/kubevirt-ansible/tests"
	ktests "kubevirt.io/kubevirt/tests"
)

var _ = Describe("Importing and starting a VMI using CDI", func() {
	prepareCDIResource := func(manifest, url string) string {
		t := new(tests.TestRandom)
		defer t.CleanUp()
		Expect(t.Generate()).ToNot(HaveOccurred())
		envURL, ok := os.LookupEnv("STREAM_IMAGE_URL")
		if ok {
			url = envURL
		}
		tests.ProcessTemplateWithParameters(manifest, t.ABSPath(), "RESOURCE_NAME="+t.Name(), "EP_URL="+url)
		tests.CreateResourceWithFilePath(t.ABSPath(), "")
		return t.Name()
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
		t := new(tests.TestRandom)
		defer t.CleanUp()
		Expect(t.Generate()).ToNot(HaveOccurred())
		tests.ProcessTemplateWithParameters(rawVMFilePath, t.ABSPath(), "VM_NAME="+t.Name(), "PVC_NAME="+resourceName, "VM_APIVERSION="+vmAPIVersion)
		tests.CreateResourceWithFilePath(t.ABSPath(), "")
		tests.WaitUntilResourceReadyByName("vmi", t.Name(), "-o=jsonpath='{.status.phase}'", "Running", "")
	}

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
