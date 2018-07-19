package cdi_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt-ansible/tests"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

// template parameters
const (
	pvcEPHTTPNOAUTHURL = "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
	pvcName            = "golden-pvc"
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
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	testNamespaces := []string{tests.NamespaceTestDefault}
	// Create a Test Namespaces
	for _, namespace := range testNamespaces {
		ns := &k8sv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		_, err = virtCli.CoreV1().Namespaces().Create(ns)
		if !errors.IsAlreadyExists(err) {
			ktests.PanicOnError(err)
		}
	}
})

var _ = AfterSuite(func() {
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	testNamespaces := []string{tests.NamespaceTestDefault}

	// First send an initial delete to every namespace
	for _, namespace := range testNamespaces {
		err := virtCli.CoreV1().Namespaces().Delete(namespace, nil)
		if !errors.IsNotFound(err) {
			ktests.PanicOnError(err)
		}
	}
	// Wait until the namespaces are terminated
	fmt.Println("")
	for _, namespace := range testNamespaces {
		fmt.Printf("Removing the %s namespace. It can take some time...\n", namespace)
		Eventually(func() bool { return errors.IsNotFound(virtCli.CoreV1().Namespaces().Delete(namespace, nil)) }, 180*time.Second, 1*time.Second).
			Should(BeTrue())
	}
})
