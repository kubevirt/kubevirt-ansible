package framework

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"

	ktests "kubevirt.io/kubevirt/tests"
)

type User struct {
	Name     string
	Password string
}

func (u *User) Login() error {
	_, _, err := ktests.RunCommandWithNS(
		"", "oc", "login",
		"-u", u.Name,
		"-p", u.Password,
	)

	if err != nil {
		return err
	}

	return nil
}

func CreateUserWithParameter(username string) error {
	By(fmt.Sprintf("Wait until user %s is created ", username))
	output, stderr, err := ktests.RunCommandWithNS("", "oc", "create", "user", username)
	if err != nil {
		if strings.Contains(stderr, fmt.Sprintf("Error from server (AlreadyExists): users.user.openshift.io \"%s\" already exists", username)) {
			err = nil
		} else {
			err = fmt.Errorf("create user: command oc create user %s: output: %s, stderr: %s: %v", username, output, stderr, err)
		}
	}
	return err
}

func DeleteUserWithParameter(username string) error {
	By(fmt.Sprintf("Wait until user %s is deleted", username))
	output, stderr, err := ktests.RunCommandWithNS("", "oc", "delete", "user", username)
	if err != nil {
		if strings.Contains(stderr, fmt.Sprintf("Error from server (NotFound): users.user.openshift.io \"%s\" not found", username)) {
			err = nil
		} else {
			err = fmt.Errorf("delete user: command oc delete user %s: output: %s, stderr: %s: %v", username, output, stderr, err)
		}
	}
	return err
}
