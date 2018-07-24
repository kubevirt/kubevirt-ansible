package cdi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt-ansible/tests"
)

// template parameters
const (
	pvcEPHTTPNOAUTHURL = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	invalidPVCURL      = "https://noneexist.com"
	pvcName            = "golden-pvc"
	pvcName1            = "golden-pvc1"
	vmName             = "test-vm"
	vmAPIVersion       = "kubevirt.io/v1alpha2"
	rawPVCFilePath     = "tests/manifests/golden-pvc.yml"
	rawVMFilePath      = "tests/manifests/test-vm.yml"
)

func TestCDI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CDI Suite")
}

var _ = BeforeSuite(func() {
	tests.CreateNamespaces()
})

var _ = AfterSuite(func() {
	tests.RemoveNamespaces()
})
