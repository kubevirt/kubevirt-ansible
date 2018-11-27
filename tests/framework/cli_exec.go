package framework

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/gomega"

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
