package framework

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/goexpect"
	. "github.com/onsi/ginkgo"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
	ktests "kubevirt.io/kubevirt/tests"
)

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

func LoggedInFedoraExpecter(vmiName string, vmiNamespace string, timeout int64, vmNameInPromt bool) (expect.Expecter, error) {
	virtClient, err := kubecli.GetKubevirtClient()
	ktests.PanicOnError(err)
	vmi, err := virtClient.VirtualMachineInstance(vmiNamespace).Get(vmiName, &metav1.GetOptions{})
	ktests.PanicOnError(err)
	expecter, _, err := ktests.NewConsoleExpecter(virtClient, vmi, 30*time.Second)
	if err != nil {
		return nil, err
	}

	loginPromt := ""

	if vmNameInPromt {
		loginPromt = vmiName + " " + "login:"
	} else {
		loginPromt = "login:"
	}

	b := append([]expect.Batcher{
		&expect.BSnd{S: "\n"},
		&expect.BSnd{S: "\n"},
		&expect.BExp{R: loginPromt},
		&expect.BSnd{S: "fedora\n"},
		&expect.BExp{R: "Password:"},
		&expect.BSnd{S: "fedora\n"},
		&expect.BExp{R: "$"}})
	res, err := expecter.ExpectBatch(b, time.Duration(timeout)*time.Second)
	if err != nil {
		log.DefaultLogger().Object(vmi).Infof("Login: %v", res)
		By(fmt.Sprintf("Login: %v", res))
		expecter.Close()
		return nil, err
	}
	return expecter, err
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
