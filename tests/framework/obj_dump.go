package framework

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"

	ktests "kubevirt.io/kubevirt/tests"
)

func DescribeObject(name string, namespace string) (string, error) {
	result, _, err := ktests.RunCommandWithNS(
		namespace, "oc", "describe", name,
	)
	if err != nil {
		return "", err
	}

	return result, nil
}

func DescribeObjects(namespace string, names []string) (map[string]string, error) {
	m := make(map[string]string)

	for _, name := range names {
		desc, err := DescribeObject(name, namespace)
		if err != nil {
			return nil, err
		}
		m[name] = desc
	}

	return m, nil
}

func GetObjects(namespace string, objType string) ([]string, error) {
	result, _, err := ktests.RunCommandWithNS(
		namespace, "oc", "get", objType, "-o", "name",
	)
	if err != nil {
		return nil, err
	}

	if result == "" {
		return make([]string, 0), nil
	}

	return strings.Split(strings.Trim(result, "\n"), "\n"), nil
}

func GetNamespaces() ([]string, error) {
	result, _, err := ktests.RunCommandWithNS("", "oc", "projects", "--short")
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.Trim(result, "\n"), "\n"), nil
}

func DumpObjectsByType(namespace string, objType string, dest string) error {
	//var filename string
	//fullDest := filepath.Join(dest, namespace, objType)
	//names, err := GetObjects(namespace, objType)
	//
	//if err != nil {
	//	return err
	//}
	//m, err := DescribeObjects(namespace, names)
	//
	//if err != nil {
	//	return err
	//}
	//
	//if len(m) == 0 {
	//	return nil
	//}
	//
	//os.MkdirAll(fullDest, 0770)
	//
	//for name, desc := range m {
	//	filename = filepath.Join(fullDest, filepath.Base(name))
	//	err = ioutil.WriteFile(filename, []byte(desc), 0644)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func DumpObjects(namespace string, objTypes []string, dest string) error {
	for _, objType := range objTypes {
		err := DumpObjectsByType(namespace, objType, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func CollectObjDesc(dest string, objTypes ...string) error {
	namespaces, err := GetNamespaces()
	if err != nil {
		return err
	}

	for _, namespace := range namespaces {
		err = DumpObjects(namespace, objTypes, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func CollectObjDescUsingTestDesc(td GinkgoTestDescription) error {
	testName := fmt.Sprintf(
		"%s-%d",
		filepath.Base(td.FileName),
		td.LineNumber,
	)
	dest := filepath.Join("exported-artifacts", "obj-desc", "after", testName)

	admin := User{Name: UsernameAdminUser, Password: PasswordAdminUser}
	err := admin.Login()
	if err != nil {
		fmt.Printf(
			"Failed to login as Admin user. Skipping obj desc collection",
		)
		return err
	}

	fmt.Println("Running log collection")
	err = CollectObjDesc(dest, "pod", "pv", "pvc")

	if err != nil {
		fmt.Printf("Failed to collect logs\n%s", err)
	}

	err = DumpObjectsByType("default", "node", dest)
	if err != nil {
		fmt.Printf("Failed to collect nodes description\n %s", err)
	}

	return nil
}
