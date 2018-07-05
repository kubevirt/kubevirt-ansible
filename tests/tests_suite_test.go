package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	
	"kubevirt.io/kubevirt/tests"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	tests.BeforeTestSuitSetup()
})

var _ = AfterSuite(func() {
	tests.AfterTestSuitCleanup()
})
