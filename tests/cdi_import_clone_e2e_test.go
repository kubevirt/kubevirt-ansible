package tests_test

import (
	"strconv"
	"strings"
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
	emptyURL               = ""
	rawDataVolumePath      = "tests/manifests/datavolume.yml"
	rawPVCPath             = "tests/manifests/golden-pvc.yml"
	rawVMPath              = "tests/manifests/test-vm.yml"
	rawDataVolumeClonePath = "tests/manifests/datavolume-clone.yml"
	rawPVCClonePath        = "tests/manifests/target-pvc.yml"
)

var _ = Describe("Importing and starting a VMI using CDI", func() {
	delNum := 0
	genManifests := func(manifest, url, sourceName, sourceNS string) string {
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
		// datavolume admission controller will throw out errors when creating empty url pvc
		if url == emptyURL && manifest == rawDataVolumePath {
			args := []string{"create", "-f", t.ABSPath(), "-n", sourceNS}
			_, outErr, err := ktests.RunCommandWithNS(sourceNS, "oc", args...)
			Expect(err).To(HaveOccurred())
		} else {
			f.CreateResourceWithFilePathWithNamespace(t.ABSPath(), sourceNS)
		}
		return t.Name()
	}
	checkPVC := func(phase, pvcName, namespace string) bool {
		switch phase {
		case cirrosURL:
			f.WaitUntilResourceReadyByNameWithNamespace("pvc", pvcName, "-o=jsonpath='{.metadata.annotations}'", "pv.kubernetes.io/bind-completed:yes", namespace)
			return true
		case invalidURL:
			Eventually(func() bool {
				args := []string{"get", "pods", "-l", f.CDI_LABEL_SELECTOR, "-n", namespace}
				out, _, err := ktests.RunCommandWithNS(namespace, "oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return strings.Contains(out, "CrashLoopBackOff")
			}, time.Duration(10)*time.Minute).Should(BeTrue())
			return false
		case emptyURL:
			Eventually(func() bool {
				args := []string{"get", "pods", "-l", f.CDI_LABEL_SELECTOR, "-n", namespace}
				_, out, err := ktests.RunCommandWithNS(namespace, "oc", args...)
				Expect(err).ToNot(HaveOccurred())
				return strings.Contains(out, "No resources found")
			}, time.Duration(1)*time.Minute).Should(BeTrue())
			return false
		}
		return false
	}

	startVMIConnectToCDIStorage := func(resourceName, namespace string) {
		t, err := f.NewTestRandom()
		Expect(err).ToNot(HaveOccurred())
		f.ProcessTemplateWithParameters(rawVMPath, t.ABSPath(), "VM_NAME="+t.Name(), "PVC_NAME="+resourceName, "VM_APIVERSION="+kubev1.GroupVersion.String())
		f.CreateResourceWithFilePathWithNamespace(t.ABSPath(), namespace)
		f.WaitUntilResourceReadyByNameWithNamespace("vmi", t.Name(), "-o=jsonpath='{.status.phase}'", "Running", namespace)
		defer t.CleanUp()
	}

	AfterEach(func() {
		delNum = delNum + 1
		args := []string{"delete", "project", "cdi-test-" + strconv.Itoa(delNum)}
		_, _, err := ktests.RunCommandWithNS("", "oc", args...)
		Expect(err).ToNot(HaveOccurred())
	})

	table.DescribeTable("with different cases:", func(manifest, url, cloneManifest, namespace string) {
		resourceName := genManifests(manifest, url, "", namespace)
		if checkPVC(url, resourceName, namespace) == true {
			startVMIConnectToCDIStorage(resourceName, namespace)
			if cloneManifest != "" {
				cloneName := genManifests(cloneManifest, "", resourceName, namespace)
				startVMIConnectToCDIStorage(cloneName, namespace)
			}
		}
	},
		table.Entry("PVC with valid image url will succeed", rawPVCPath, cirrosURL, "", "cdi-test-1"),
		table.Entry("PVC with invalid image url will be failed", rawPVCPath, invalidURL, "", "cdi-test-2"),
		table.Entry("PVC with empty image url will be failed", rawPVCPath, emptyURL, "", "cdi-test-3"),
		table.Entry("DataVolume with valid image url will succeed", rawDataVolumePath, cirrosURL, "", "cdi-test-4"),
		table.Entry("DataVolume with invalid image url will be failed", rawDataVolumePath, invalidURL, "", "cdi-test-5"),
		table.Entry("DataVolume with empty image url will be failed", rawDataVolumePath, emptyURL, "", "cdi-test-6"),
		table.Entry("DataVolume cloning with source datavolume generated pvc", rawDataVolumePath, cirrosURL, rawDataVolumeClonePath, "cdi-test-7"),
		table.Entry("DataVolume cloning with source pvc", rawPVCPath, cirrosURL, rawDataVolumeClonePath, "cdi-test-8"),
		table.Entry("PVC cloning with source pvc", rawPVCPath, cirrosURL, rawPVCClonePath, "cdi-test-9"),
		table.Entry("PVC cloning with source datavolume generated pvc", rawDataVolumePath, cirrosURL, rawPVCClonePath, "cdi-test-10"),
	)
})
