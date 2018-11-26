package framework

import (
	"fmt"
	"strings"

	ktests "kubevirt.io/kubevirt/tests"
)

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
