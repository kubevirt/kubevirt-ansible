package network_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"kubevirt.io/kubevirt-ansible/tests"

)

func TestNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Suite")
}


var _ = BeforeSuite(func() {
	tests.CreateNamespaces()
})

var _ = AfterSuite(func() {
	tests.RemoveNamespaces()
})
