package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"kubevirt.io/kubevirt/pkg/kubecli"
	ktests "kubevirt.io/kubevirt/tests"
)

type Result struct {
	cmd           string
	verb          string
	resourceType  string
	resourceName  string
	resourceLabel string
	filePath      string
	nameSpace     string
	query         string
	expectOut     string
	actualOut     string
	params        []string
}

type TestRandom struct {
	testDir         string
	generateName    string
	generateABSPath string
}

func NewTestRandom() (*TestRandom, error) {
	var err error
	t := new(TestRandom)
	t.testDir, err = ioutil.TempDir("", TestDir)
	if err != nil {
		return nil, err
	}
	t.generateName = "generate-" + rand.String(10)
	t.generateABSPath = filepath.Join(t.testDir, t.generateName+".json")
	return t, nil
}

func (t *TestRandom) Name() string {
	return t.generateName
}

func (t *TestRandom) ABSPath() string {
	return t.generateABSPath
}

func (t *TestRandom) CleanUp() error {
	if t.testDir != "" {
		return os.RemoveAll(t.testDir)
	} else {
		return fmt.Errorf("testDir is empty")
	}
}

var KubeVirtOcPath = ""

const (
	CDI_LABEL_KEY           = "app"
	CDI_LABEL_VALUE         = "containerized-data-importer"
	CDI_LABEL_SELECTOR      = CDI_LABEL_KEY + "=" + CDI_LABEL_VALUE
	CDI_TEST_LABEL_KEY      = "test"
	CDI_TEST_LABEL_VALUE    = "cdi-manifests"
	CDI_TEST_LABEL_SELECTOR = CDI_TEST_LABEL_KEY + "=" + CDI_TEST_LABEL_VALUE
	NamespaceTestDefault    = "kubevirt-test-default"
	paramFlag               = "-p"
	TestDir                 = "kubevirt-ansible-test"
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
	out := execute(Result{cmd: "oc", verb: "process", filePath: srcFilePath, params: params})
	filePath, err := writeJson(dstFilePath, out)
	Expect(err).ToNot(HaveOccurred())
	return filePath
}

func CreateResourceWithFilePath(filePath, namespace string) {
	if namespace == "" {
		namespace = NamespaceTestDefault
	}
	By("Creating resource from the json file with the oc-create command")
	execute(Result{cmd: "oc", verb: "create", filePath: filePath, nameSpace: namespace})
}

func DeleteResourceWithLabel(resourceType, resourceLabel, namespace string) {
	if namespace == "" {
		namespace = NamespaceTestDefault
	}
	By(fmt.Sprintf("Deleting %s:%s from the json file with the oc-delete command", resourceType, resourceLabel))
	execute(Result{cmd: "oc", verb: "delete", resourceType: resourceType, resourceLabel: resourceLabel, nameSpace: namespace})
}

func WaitUntilResourceReadyByName(resourceType, resourceName, query, expectOut, namespace string) {
	if namespace == "" {
		namespace = NamespaceTestDefault
	}
	By(fmt.Sprintf("Wait until %s with name %s ready", resourceType, resourceName))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceName: resourceName, query: query, expectOut: expectOut, nameSpace: namespace})
}

func WaitUntilResourceReadyByLabel(resourceType, label, query, expectOut, namespace string) {
	if namespace == "" {
		namespace = NamespaceTestDefault
	}
	By(fmt.Sprintf("Wait until resource %s with label=%s ready", resourceType, label))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceLabel: label, query: query, expectOut: expectOut, nameSpace: namespace})
}

func execute(r Result) string {
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
	if len(r.params) > 0 {
		for _, v := range r.params {
			cmd = append(cmd, paramFlag, v)
		}
	}
	if r.expectOut != "" {
		Eventually(func() bool {
			r.actualOut, err = ktests.RunCommand(r.cmd, cmd...)
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(r.actualOut, r.expectOut)
		}, time.Duration(3)*time.Minute).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s to appear", r.resourceType))
	} else {
		r.actualOut, err = ktests.RunCommand(r.cmd, cmd...)
		Expect(err).ToNot(HaveOccurred())
	}
	return r.actualOut
}

func writeJson(jsonFile string, json string) (string, error) {
	err := ioutil.WriteFile(jsonFile, []byte(json), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write the json file %s", jsonFile)
	}
	return jsonFile, nil
}
