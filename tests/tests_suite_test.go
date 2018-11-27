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

var cdiTestNamespaces = []string{"cdi-test-1", "cdi-test-2", "cdi-test-3", "cdi-test-4", "cdi-test-5", "cdi-test-6", "cdi-test-7", "cdi-test-8", "cdi-test-9", "cdi-test-10"}

var _ = BeforeSuite(func() {
	testNamespaces := []string{ktests.NamespaceTestDefault, ktests.NamespaceTestAlternative}
	testNamespaces = append(testNamespaces, cdiTestNamespaces...)
	for _, v := range testNamespaces {
		err := tests.CreateNamespaceWithParameter(v)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	testNamespaces := []string{ktests.NamespaceTestDefault, ktests.NamespaceTestAlternative}
	testNamespaces = append(testNamespaces, cdiTestNamespaces...)
	for _, v := range testNamespaces {
		err := tests.RemoveNamespaceWithParameter(v)
		Expect(err).ToNot(HaveOccurred())
	}
})
