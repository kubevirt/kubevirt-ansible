package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
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

const (
	UsernameTestUser     = "kubevirt-test-user"
	NamespaceTestDefault = "kubevirt-test-default"
	UsernameAdminUser    = "test_admin"
)

const (
	CDI_LABEL_KEY      = "app"
	CDI_LABEL_VALUE    = "containerized-data-importer"
	CDI_LABEL_SELECTOR = CDI_LABEL_KEY + "=" + CDI_LABEL_VALUE
	paramFlag          = "-p"
)

const (
	ShortTimeout = time.Duration(2) * time.Minute
	LongTimeout  = time.Duration(4) * time.Minute
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

func ProcessTemplateWithParameters(srcFilePath, dstFilePath string, params ...string) string {
	By(fmt.Sprintf("Overriding the template from %s to %s", srcFilePath, dstFilePath))
	out := execute(Result{cmd: "oc", verb: "process", filePath: srcFilePath, params: params})
	filePath, err := writeJson(dstFilePath, out)
	Expect(err).ToNot(HaveOccurred())
	return filePath
}

func CreateResourceWithFilePathTestNamespace(filePath string) {
	By("Creating resource from the json file with the oc-create command")
	execute(Result{cmd: "oc", verb: "create", filePath: filePath, nameSpace: NamespaceTestDefault})
}

func DeleteResourceWithLabelTestNamespace(resourceType, resourceLabel string) {
	By(fmt.Sprintf("Deleting %s:%s from the json file with the oc-delete command", resourceType, resourceLabel))
	execute(Result{cmd: "oc", verb: "delete", resourceType: resourceType, resourceLabel: resourceLabel, nameSpace: NamespaceTestDefault})
}

func WaitUntilResourceReadyByNameTestNamespace(resourceType, resourceName, query, expectOut string) {
	By(fmt.Sprintf("Wait until %s with name %s ready", resourceType, resourceName))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceName: resourceName, query: query, expectOut: expectOut, nameSpace: NamespaceTestDefault})
}

func WaitUntilResourceReadyByLabelTestNamespace(resourceType, label, query, expectOut string) {
	By(fmt.Sprintf("Wait until resource %s with label=%s ready", resourceType, label))
	execute(Result{cmd: "oc", verb: "get", resourceType: resourceType, resourceLabel: label, query: query, expectOut: expectOut, nameSpace: NamespaceTestDefault})
}

func CreateUser(username string) {
	By(fmt.Sprintf("Wait until user %s is created ", username))
	execute(Result{cmd: "oc", verb: "create", resourceType: "user", resourceName: username})
}

func DeleteUser(username string) {
	By(fmt.Sprintf("Wait until user %s is deleted", username))
	execute(Result{cmd: "oc", verb: "delete", resourceType: "user", resourceName: username})
}

func VNCConnection(namespace, vmname string) (string, error) {
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	pipeInReader, _ := io.Pipe()
	pipeOutReader, pipeOutWriter := io.Pipe()
	k8ResChan := make(chan error)
	readStop := make(chan string)

	go func() {
		GinkgoRecover()
		vnc, err := virtClient.VirtualMachineInstance(namespace).VNC(vmname)
		if err != nil {
			k8ResChan <- err
			return
		}

		k8ResChan <- vnc.Stream(kubecli.StreamOptions{
			In:  pipeInReader,
			Out: pipeOutWriter,
		})
	}()

	go func() {
		GinkgoRecover()
		buf := make([]byte, 1024, 1024)
		n, err := pipeOutReader.Read(buf)
		if err != nil && err != io.EOF {
			return
		}
		if n == 0 && err == io.EOF {
			log.Log.Info("zero bytes read from vnc socket.")
			return
		}
		readStop <- strings.TrimSpace(string(buf[0:n]))
	}()
	response := ""
	select {
	case response = <-readStop:
	case err = <-k8ResChan:
	case <-time.After(45 * time.Second):
		Fail("Timout reached while waiting for valid VNC server response")
	}
	return response, err
}

func execute(r Result) string {
	return executeWithCustomTimeout(r, LongTimeout)
}

func executeWithCustomTimeout(r Result, timeout time.Duration) string {
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
			r.actualOut, _, err = ktests.RunCommand(r.cmd, cmd...)
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(r.actualOut, r.expectOut)
		}, timeout).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s to appear", r.resourceType))
	} else {
		r.actualOut, _, err = ktests.RunCommand(r.cmd, cmd...)
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

func LoggedInFedoraExpecter(vmiName string, vmiNamespace string, timeout int64) (expect.Expecter, error) {
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	vmi, err := virtClient.VirtualMachineInstance(vmiNamespace).Get(vmiName, &metav1.GetOptions{})
	ktests.PanicOnError(err)
	expecter, _, err := ktests.NewConsoleExpecter(virtClient, vmi, 30*time.Second)
	if err != nil {
		return nil, err
	}
	b := append([]expect.Batcher{
		&expect.BSnd{S: "\n"},
		&expect.BSnd{S: "\n"},
		&expect.BExp{R: "login:"},
		&expect.BSnd{S: "fedora\n"},
		&expect.BExp{R: "Password:"},
		&expect.BSnd{S: "fedora\n"},
		&expect.BExp{R: "fedora@" + vmiName + " ~]$"}})
	res, err := expecter.ExpectBatch(b, time.Duration(timeout)*time.Second)
	if err != nil {
		log.DefaultLogger().Object(vmi).Infof("Login: %v", res)
		By(fmt.Sprintf("Login: %v", res))
		expecter.Close()
		return nil, err
	}
	return expecter, err
}

func RunOcDescribeCommand(resourceType, resourceName string) string {
	fmt.Printf("Getting 'oc describe' with: %s ", resourceName)
	return execute(Result{cmd: "oc", verb: "describe", resourceType: resourceType, resourceName: resourceName, nameSpace: NamespaceTestDefault})
}

func OpenConsole(virtCli kubecli.KubevirtClient, vmiName string, vmiNamespace string, timeout time.Duration, consoleType string, opts ...expect.Option) (expect.Expecter, <-chan error, error) {
	vmiReader, vmiWriter := io.Pipe()
	expecterReader, expecterWriter := io.Pipe()
	resCh := make(chan error)
	var con kubecli.StreamInterface
	var err error
	startTime := time.Now()
	if consoleType == "serial" {
		con, err = virtCli.VirtualMachineInstance(vmiNamespace).SerialConsole(vmiName, timeout)
	} else if consoleType == "vnc" {
		con, err = virtCli.VirtualMachineInstance(vmiNamespace).VNC(vmiName)
	}
	if err != nil {
		return nil, nil, err
	}
	timeout = timeout - time.Now().Sub(startTime)

	go func() {
		resCh <- con.Stream(kubecli.StreamOptions{
			In:  vmiReader,
			Out: expecterWriter,
		})
	}()

	return expect.SpawnGeneric(&expect.GenOptions{
		In:  vmiWriter,
		Out: expecterReader,
		Wait: func() error {
			return <-resCh
		},
		Close: func() error {
			expecterWriter.Close()
			vmiReader.Close()
			return nil
		},
		Check: func() bool { return true },
	}, timeout, opts...)
}
