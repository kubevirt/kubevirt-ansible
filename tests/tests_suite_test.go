package tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/tests"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tests Suite")
}

var _ = BeforeSuite(func() {
	virtClient, err := kubecli.GetKubevirtClient()
	tests.PanicOnError(err)

	cfgMap := &k8sv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceSystem,
			Name:      "kubevirt-config",
		},
		Data: map[string]string{"debug.useEmulation": "true"},
	}
	_, err = virtClient.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(cfgMap)
	tests.PanicOnError(err)

	tests.BeforeTestSuitSetup()
})

var _ = AfterSuite(func() {
	tests.AfterTestSuitCleanup()
})
