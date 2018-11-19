package framework

import (
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

func writeJson(jsonFile string, json string) (string, error) {
	err := ioutil.WriteFile(jsonFile, []byte(json), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write the json file %s", jsonFile)
	}
	return jsonFile, nil
}

func RunOcDescribeCommand(resourceType, resourceName string) string {
	fmt.Printf("Getting 'oc describe' with: %s ", resourceName)
	return execute(Result{cmd: "oc", verb: "describe", resourceType: resourceType, resourceName: resourceName, nameSpace: NamespaceTestDefault})
}
