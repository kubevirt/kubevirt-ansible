package tests_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	f "kubevirt.io/kubevirt-ansible/tests/framework"
	kubev1 "kubevirt.io/kubevirt/pkg/api/v1"
	ktests "kubevirt.io/kubevirt/tests"
)

const (
	cirrosURL              = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	invalidURL             = "https://noneexist.com"
	emptyURL               = ""
	rawDataVolumePath      = "tests/manifests/datavolume.yml"
	rawPVCPath             = "tests/manifests/golden-pvc.yml"
	rawBlankDiskPVCPath    = "tests/manifests/blankdisk-pvc.yml"
	rawBlankDiskDataVolumePath = "tests/manifests/blankdisk-datavolume.yml"
	rawVMPath              = "tests/manifests/test-vm.yml"
	rawDataVolumeClonePath = "tests/manifests/datavolume-clone.yml"
	rawPVCClonePath        = "tests/manifests/target-pvc.yml"
)

var _ = Describe("Importing and starting a VMI using CDI", func() {
	prepareCDIResource := func(manifest, url, sourceName, sourceNS string) string {
		t, err := f.NewTestRandom()
		Expect(err).ToNot(HaveOccurred())
		defer t.CleanUp()
		switch manifest {
		case rawDataVolumePath, rawPVCPath:
			f.ProcessTemplateWithParameters(manifest, t.ABSPath(), "RESOURCE_NAME="+t.Name(), "EP_URL="+f.ReplaceImageURL(url))
		case rawDataVolumeClonePath:
			f.ProcessTemplateWithParameters(manifest, t.ABSPath(), "RESOURCE_NAME="+t.Name(), "SOURCE_NAME="+sourceName, "SOURCE_NS="+sourceNS)
		case rawPVCClonePath:
			f.ProcessTemplateWithParameters(manifest, t.ABSPath(), "RESOURCE_NAME="+t.Name(), "SOURCE_NAME="+sourceName, "SOURCE_NS="+sourceNS)
		}
		f.CreateResourceWithFilePath(t.ABSPath())
		return t.Name()
	}
	waitForImporterPodWriteImg := func(phase, resourceName string) {
		switch phase {
		case "Succeeded":
			f.WaitUntilResourceReadyByNameTestNamespace("pvc", resourceName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes")
			f.WaitUntilResourceReadyByLabelTestNamespace("pod", f.CDI_LABEL_SELECTOR, "-o=jsonpath='{.items[*].status.phase}'", "Succeeded")
		case "Failed":
			f.WaitUntilResourceReadyByLabelTestNamespace("pod", f.CDI_LABEL_SELECTOR, "", "CrashLoopBackOff")
		case "NOExist":
			Consistently(func() bool {
				args := []string{"get", "pods", "-l", f.CDI_LABEL_SELECTOR, "-n", f.NamespaceTestDefault}
				out, _, err := ktests.RunCommand("oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return out == "No resources found.\n"
			}, time.Duration(10)*time.Second).Should(BeTrue())
		}
	}

	startVMIConnectToCDIStorage := func(resourceName string) {
		t, err := f.NewTestRandom()
		Expect(err).ToNot(HaveOccurred())
		f.ProcessTemplateWithParameters(rawVMPath, t.ABSPath(), "VM_NAME="+t.Name(), "PVC_NAME="+resourceName, "VM_APIVERSION="+kubev1.GroupVersion.String())
		f.CreateResourceWithFilePath(t.ABSPath())
		f.WaitUntilResourceReadyByNameTestNamespace("vmi", t.Name(), "-o=jsonpath='{.status.phase}'", "Running")
		defer t.CleanUp()
	}

	AfterEach(func() {
		By("Deleting all pvc with the oc-delete command")
		args := []string{"delete", "pvc", "-l", f.CDI_TEST_LABEL_SELECTOR, "-n", f.NamespaceTestDefault}
		_, _, err := ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		By("Deleting all datavolumes with the oc-delete command")
		args = []string{"delete", "datavolumes.cdi.kubevirt.io", "-l", f.CDI_TEST_LABEL_SELECTOR, "-n", f.NamespaceTestDefault}
		_, _, err = ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())

		By("Deleting all vmi with the oc-delete command")
		args = []string{"delete", "vmi", "-l", f.CDI_TEST_LABEL_SELECTOR, "-n", f.NamespaceTestDefault}
		_, _, err = ktests.RunCommand("oc", args...)
		Expect(err).ToNot(HaveOccurred())
	})

	table.DescribeTable("with different cases:", func(manifest, url, cloneManifest string) {
		resourceName := prepareCDIResource(manifest, url, "", "")
		switch url {
		case invalidURL:
			waitForImporterPodWriteImg("Failed", resourceName)
		case emptyURL:
			waitForImporterPodWriteImg("NoExist", resourceName)
		default:
			startVMIConnectToCDIStorage(resourceName)
		}
		if cloneManifest != "" {
			cloneName := prepareCDIResource(cloneManifest, "", resourceName, f.NamespaceTestDefault)
			startVMIConnectToCDIStorage(cloneName)
		}
	},
		table.Entry("PVC with valid image url will succeed", rawPVCPath, cirrosURL, ""),
		table.Entry("PVC with invalid image url will be failed", rawPVCPath, invalidURL, ""),
		table.Entry("PVC with empty image url will be failed", rawPVCPath, emptyURL, ""),
		table.Entry("PVC with blank disk will succeed", rawBlankDiskPVCPath, cirrosURL, ""),
		table.Entry("DataVolume with valid image url will succeed", rawDataVolumePath, cirrosURL, ""),
		table.Entry("DataVolume with invalid image url will be failed", rawDataVolumePath, invalidURL, ""),
		table.Entry("DataVolume with empty image url will be failed", rawDataVolumePath, emptyURL, ""),
		table.Entry("DataVolume with blank disk will succeed", rawBlankDiskDataVolumePath, cirrosURL, ""),
		table.Entry("DataVolume cloning with source datavolume generated pvc", rawDataVolumePath, cirrosURL, rawDataVolumeClonePath),
		table.Entry("DataVolume cloning with source pvc", rawPVCPath, cirrosURL, rawDataVolumeClonePath),
		table.Entry("PVC cloning with source pvc", rawPVCPath, cirrosURL, rawPVCClonePath),
		table.Entry("PVC cloning with source datavolume generated pvc", rawDataVolumePath, cirrosURL, rawPVCClonePath),
	)
})
