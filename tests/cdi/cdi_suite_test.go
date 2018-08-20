package cdi_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirt-ansible/tests"
	"kubevirt.io/qe-tools/pkg/ginkgo-reporters"
)

// template parameters
const (
	cirrosURL    = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	invalidURL   = "https://noneexist.com"
	emptyURL     = ""
	vmAPIVersion = "kubevirt.io/v1alpha2"
)

const (
	rawDataVolumePath = "tests/manifests/datavolume.yml"
	rawPVCFilePath    = "tests/manifests/golden-pvc.yml"
	rawVMFilePath     = "tests/manifests/test-vm.yml"
)

const (
	tmpTestDir = "/tmp/kubevirt-ansible-test/"
)

func TestCDI(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := make([]Reporter, 0)
	if ginkgo_reporters.Polarion.Run {
		reporters = append(reporters, &ginkgo_reporters.Polarion)
	}
	if ginkgo_reporters.JunitOutput != "" {
		reporters = append(reporters, ginkgo_reporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "CDI Suite", reporters)
}

var _ = BeforeSuite(func() {
	if !flag.Parsed() {
		flag.Parse()
	}
	tests.CreateNamespaces()
})

var _ = AfterSuite(func() {
	tests.RemoveNamespaces()
})
