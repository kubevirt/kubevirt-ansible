package tests

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ktests "kubevirt.io/kubevirt/tests"
	"github.com/davecgh/go-spew/spew"
)

type CmdBuilder struct {
	cmd           string
	verb          string
	resourceType  string
	resourceName  string
	resourceLabel string
	filePath      string
	namespace     string
	query         string
	expectOut     string
	actualOut     string
	cmdArgs       []string
}

const (
	CDI_LABEL_KEY        = "app"
	CDI_LABEL_VALUE      = "containerized-data-importer"
	CDI_LABEL_SELECTOR   = CDI_LABEL_KEY + "=" + CDI_LABEL_VALUE
	NamespaceTestDefault = "kubevirt-test-default"
)

func Kubectl(cmdVerb string) *CmdBuilder {
	return newInitCmd("kubectl", cmdVerb)
}

func Oc(cmdVerb string) *CmdBuilder {
	return newInitCmd("oc", cmdVerb)
}

func Virtctl(cmdVerb string) *CmdBuilder {
	return newInitCmd("virtctl", cmdVerb)
}

func newInitCmd(cmd, cmdVerb string) *CmdBuilder {
	b := new(CmdBuilder)
	b.cmdArgs = []string{cmd, cmdVerb}
	return b
}

func (b *CmdBuilder) Param(params ...string) *CmdBuilder {
	for _, v := range params {
		b.cmdArgs = append(b.cmdArgs, v)
	}
	return b
}

func (b *CmdBuilder) Query(query string) *CmdBuilder {
	b.query = query
	b.cmdArgs = append(b.cmdArgs, query)
	return b
}

func (b *CmdBuilder) Namespace(namespace string) *CmdBuilder {
	b.namespace = namespace
	b.cmdArgs = append(b.cmdArgs, "-n", namespace)
	return b
}

func (b *CmdBuilder) TestNamespace() *CmdBuilder {
	b.namespace = NamespaceTestDefault
	b.cmdArgs = append(b.cmdArgs, "-n", NamespaceTestDefault)
	return b
}

func (b *CmdBuilder) FilePath(filePath string) *CmdBuilder {
	if b.cmd == "virtctl" {
		ktests.PanicOnError(fmt.Errorf("virtctl doesn't support -f currently"))
		return nil
	}
	b.filePath = filePath
	b.cmdArgs = append(b.cmdArgs, "-f", filePath)
	return b
}

func (b *CmdBuilder) ResourceType(resourceType string) *CmdBuilder {
	b.resourceType = resourceType
	b.cmdArgs = append(b.cmdArgs, resourceType)
	return b
}

func (b *CmdBuilder) ResourceName(resourceName string) *CmdBuilder {
	b.resourceName = resourceName
	b.cmdArgs = append(b.cmdArgs, b.resourceName)
	return b
}

func (b *CmdBuilder) ResourceLabel(resourceLabel string) *CmdBuilder {
	b.resourceLabel = resourceLabel
	b.cmdArgs = append(b.cmdArgs, "-l", resourceLabel)
	return b
}

func (b *CmdBuilder) WaitUntil(expectOut string) {
	Eventually(func() bool {
		spew.Dump(b.cmdArgs)
		out, err := ktests.RunCommand(b.cmdArgs[0], b.cmdArgs[1:]...)
		Expect(err).ToNot(HaveOccurred())
		return strings.Contains(out, expectOut)
	}, time.Duration(2)*time.Minute).Should(BeTrue(), fmt.Sprintf("Timed out waiting for %s to appear", b.resourceType))
}

func (b *CmdBuilder) WriteJson(jsonFile string) {
	var err error
	b.actualOut, err = ktests.RunCommand(b.cmdArgs[0], b.cmdArgs[1:]...)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(jsonFile, []byte(b.actualOut), 0644)
	if err != nil {
		ktests.PanicOnError(fmt.Errorf("failed to write the json file %s", jsonFile))
	}
}

func (b *CmdBuilder) Run() string {
	var err error
	b.actualOut, err = ktests.RunCommand(b.cmdArgs[0], b.cmdArgs[1:]...)
	Expect(err).ToNot(HaveOccurred())
	return b.actualOut
}