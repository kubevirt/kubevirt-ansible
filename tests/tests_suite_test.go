package tests_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "kubevirt.io/kubevirt-ansible/tests/framework"
	ktests "kubevirt.io/kubevirt/tests"
	"kubevirt.io/qe-tools/pkg/ginkgo-reporters"
)

func CustomFailHandler(message string, callerSkip ...int) {
	tests.CollectObjDescUsingTestDesc(CurrentGinkgoTestDescription())
	ktests.KubevirtFailHandler(message, callerSkip...)
}

func TestTests(t *testing.T) {
	RegisterFailHandler(CustomFailHandler)
	reporters := make([]Reporter, 0)
	if ginkgo_reporters.Polarion.Run {
		reporters = append(reporters, &ginkgo_reporters.Polarion)
	}
	if ginkgo_reporters.JunitOutput != "" {
		reporters = append(reporters, ginkgo_reporters.NewJunitReporter())
	}
	RunSpecsWithDefaultAndCustomReporters(t, "Tests Suite", reporters)
}

var _ = BeforeSuite(func() {
	err := tests.CreateNamespaces()
	Expect(err).ToNot(HaveOccurred())

})

var _ = AfterSuite(func() {
	err := tests.RemoveNamespaces()
	Expect(err).ToNot(HaveOccurred())
})
