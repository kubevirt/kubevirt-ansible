package tests

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ktests "kubevirt.io/kubevirt/tests"
)

var KubeVirtOcPath = ""

const (
	NamespaceTestDefault = "kubevirt-test-default"
	paramFlag            = "-p"
)

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

func CreateResourceWithFilePathTestNamespace(resourceType, resourceName, filePath string) {
	createResourceWithFilePath(resourceType, resourceName, filePath, NamespaceTestDefault)
}

func createResourceWithFilePath(resourceType, resourceName, filePath, nameSpace string) {
	By(fmt.Sprintf("Creating %s:%s from the json file with the oc-create command", resourceType, resourceName))
	_, err := ktests.RunOcCommand("create", "-f", filePath, "-n", nameSpace)
	Expect(err).ToNot(HaveOccurred())
	Eventually(func() bool {
		out, err := ktests.RunOcCommand("get", resourceType, "-n", nameSpace)
		Expect(err).ToNot(HaveOccurred())
		return strings.Contains(out, resourceName)
	}, time.Duration(2)*time.Minute).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s to appear", resourceType))
}

func WaitUntilResourceReadyByNameTestNamespace(resourceType, resourceName, query, expectOut string) {
	By(fmt.Sprintf("Wait until %s:%s ready", resourceType, resourceName))
	waitUntilResourceReady(resourceType, "", resourceName, query, expectOut, NamespaceTestDefault)
}

func WaitUntilResourceReadyByLabelTestNamespace(resourceType, label, query, expectOut string) {
	By(fmt.Sprintf("Wait until label=%s ready", label))
	waitUntilResourceReady(resourceType, label, "", query, expectOut, NamespaceTestDefault)
}

func waitUntilResourceReady(resourceType, label, resourceName, query, expectOut, nameSpace string) {
	cmd := []string{"get"}
	if label != "" {
		cmd = append(cmd, resourceType, "-l", label)
	} else {
		cmd = append(cmd, resourceType, resourceName)
	}
	cmd = append(cmd, query, "-n", nameSpace)
	Eventually(func() bool {
		out, err := ktests.RunOcCommand(cmd...)
		Expect(err).ToNot(HaveOccurred())
		return strings.Contains(out, expectOut)
	}, time.Duration(2)*time.Minute).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s to appear", resourceType))
}

func writeJson(jsonFile string, json string) (string, error) {
	err := ioutil.WriteFile(jsonFile, []byte(json), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write the json file %s", jsonFile)
	}
	return jsonFile, nil
}
