package framework

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/gomega"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

func CreateNamespaces() {
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	testNamespaces := []string{ktests.NamespaceTestDefault, ktests.NamespaceTestAlternative}
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
}

func RemoveNamespaces() {
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	testNamespaces := []string{ktests.NamespaceTestDefault, ktests.NamespaceTestAlternative}

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
}

func CreateNamespaceWithParameter(namespace string) error {
	output, stderr, err := ktests.RunCommandWithNS("", "oc", "new-project", namespace)
	if err != nil {
		if strings.Contains(stderr, fmt.Sprintf("Error from server (AlreadyExists): project.project.openshift.io \"%s\" already exists", namespace)) {
			err = nil
		} else {
			err = fmt.Errorf("create user: command os new-project %s: output: %s, stderr: %s: %v", namespace, output, stderr, err)
		}
	}
	return err
}
func RemoveNamespaceWithParameter(namespace string) error {
	output, stderr, err := ktests.RunCommandWithNS("", "oc", "delete", "project", namespace)
	if err != nil {
		if strings.Contains(stderr, fmt.Sprintf("Error from server (NotFound): namespaces \"%s\" not found", namespace)) {
			err = nil
		} else {
			err = fmt.Errorf("delete project: command os delete project %s: output: %s, stderr: %s: %v", namespace, output, stderr, err)
		}
	}
	return err
}
