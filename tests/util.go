package tests

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

type Result struct {
	verb          string
	resourceType  string
	resourceName  string
	resourceLabel string
	filePath      string
	nameSpace     string
	query         string
	expectOut     string
	actualOut     string
}

var KubeVirtOcPath = ""

const (
	CDI_LABEL_KEY          = "app"
	CDI_LABEL_VALUE        = "containerized-data-importer"
	CDI_LABEL_SELECTOR     = CDI_LABEL_KEY + "=" + CDI_LABEL_VALUE
	NamespaceTestDefault = "kubevirt-test-default"
	paramFlag            = "-p"
)

func CreateNamespaces() {
	virtCli, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)

	testNamespaces := []string{ktests.NamespaceTestDefault}
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
	testNamespaces := []string{ktests.NamespaceTestDefault}

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

func ProcessTemplateWithParameters(srcFilePath, dstFilePath string, params ...string) string {
	By(fmt.Sprintf("Overriding the template from %s to %s", srcFilePath, dstFilePath))
	args := []string{"process", "-f", srcFilePath}
	for _, v := range params {
		args = append(args, paramFlag)
		args = append(args, v)
	}
	out, err := ktests.RunOcCommand(args...)
	Expect(err).ToNot(HaveOccurred())
	filePath, err := writeJson(dstFilePath, out)
	Expect(err).ToNot(HaveOccurred())
	return filePath
}

func CreateResourceWithFilePathTestNamespace(filePath string) {
	exec(Result{verb: "create", filePath: filePath, nameSpace: NamespaceTestDefault})
}

func DeleteResourceWithLabelTestNamespace(resourceType, resourceLabel string) {
	By(fmt.Sprintf("Deleting %s:%s from the json file with the oc-delete command", resourceType, resourceLabel))
	exec(Result{verb: "delete", resourceType: resourceType, resourceLabel: resourceLabel, nameSpace: NamespaceTestDefault})
}

func WaitUntilResourceReadyByNameTestNamespace(resourceType, resourceName, query, expectOut string) {
	By(fmt.Sprintf("Wait until %s with name %s ready", resourceType, resourceName))
	exec(Result{verb: "get", resourceType: resourceType, resourceName: resourceName, query: query, expectOut: expectOut, nameSpace: NamespaceTestDefault})
}

func WaitUntilResourceReadyByLabelTestNamespace(resourceType, label, query, expectOut string) {
	By(fmt.Sprintf("Wait until resource %s with label=%s ready",resourceType, label))
	exec(Result{verb: "get", resourceType: resourceType, resourceLabel: label, query: query, expectOut: expectOut, nameSpace: NamespaceTestDefault})
}

func exec(r Result) {
	var err error
	if r.verb == "" {
		Expect(fmt.Errorf("verb can not be empty"))
	}
	cmd := []string{r.verb}
	if r.filePath == "" {
		if r.resourceType == "" {
			Expect(fmt.Errorf("resourceType can not be empty"))
		}
		cmd = append(cmd, r.resourceType)
	}
	if r.resourceName != "" {
		cmd = append(cmd, r.resourceName)
	}
	if r.filePath != "" {
		cmd = append(cmd, "-f", r.filePath)
	}
	if r.resourceLabel != "" {
		cmd = append(cmd, "-l", r.resourceLabel)
	}
	if r.query != "" {
		cmd = append(cmd, r.query)
	}
	if r.nameSpace != "" {
		cmd = append(cmd, "-n", r.nameSpace)
	}

	if r.expectOut != "" {
		Eventually(func() bool {
			r.actualOut, err = ktests.RunOcCommand(cmd...)
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(r.actualOut, r.expectOut)
		}, time.Duration(2)*time.Minute).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s to appear", r.resourceType))
	} else {
		r.actualOut, err = ktests.RunOcCommand(cmd...)
		Expect(err).ToNot(HaveOccurred())
	}
}

func writeJson(jsonFile string, json string) (string, error) {
	err := ioutil.WriteFile(jsonFile, []byte(json), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write the json file %s", jsonFile)
	}
	return jsonFile, nil
}
