package framework

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	ktests "kubevirt.io/kubevirt/tests"
)

const (
	NamespaceTestDefault  = "kubevirt-test-default"
	NamespaceTestTemplate = "openshift"
	UsernameTestUser      = "kubevirt-test-user"
	UsernameAdminUser     = "test_admin"
	PasswordAdminUser     = "123456"

	CDI_LABEL_KEY      = "app"
	CDI_LABEL_VALUE    = "containerized-data-importer"
	CDI_LABEL_SELECTOR = CDI_LABEL_KEY + "=" + CDI_LABEL_VALUE
	paramFlag          = "-p"

	ShortTimeout    = time.Duration(2) * time.Minute
	LongTimeout     = time.Duration(4) * time.Minute
	LongLongTimeout = time.Duration(10) * time.Minute
)

// VirtualMachine can be a vm, vmi, vmirs, vmiPreset.
type VirtualMachine struct {
	Name              string
	Type              string
	UID               string
	Manifest          string
	TemplateInCluster string
	TemplateFromFile  string
	TemplateParams    []string
	Namespace         string
}

func (vm VirtualMachine) Create() (string, string, error) {
	args := []string{"create", "-f", vm.Manifest}
	return ktests.RunCommandWithNS(vm.Namespace, ktests.KubeVirtOcPath, args...)
}

func (vm VirtualMachine) Start() (string, string, error) {
	args := []string{"start", vm.Name}
	return ktests.RunCommandWithNS(vm.Namespace, ktests.KubeVirtVirtctlPath, args...)
}

func (vm VirtualMachine) Stop() (string, string, error) {
	args := []string{"stop", vm.Name}
	return ktests.RunCommandWithNS(vm.Namespace, ktests.KubeVirtVirtctlPath, args...)
}

func (vm VirtualMachine) Delete() (string, string, error) {
	args := []string{"delete", vm.Type, vm.Name}
	return ktests.RunCommandWithNS(vm.Namespace, ktests.KubeVirtVirtctlPath, args...)
}

func (vm VirtualMachine) IsRunning() (bool, error) {
	output, cmderr, err := vm.GetVMInfo("{{.status.phase}}")
	if err != nil {
		return false, err
	}
	if cmderr != "" {
		return false, errors.New(cmderr)
	}
	if output == "Running" {
		return true, nil
	}
	return false, nil
}

func (vm VirtualMachine) GetVMInfo(spec string) (string, string, error) {
	args := []string{"get", vm.Type, vm.Name, "--template", spec}
	return ktests.RunCommandWithNS(vm.Namespace, ktests.KubeVirtOcPath, args...)
}

func (vm VirtualMachine) GetVMUID() (string, error) {
	output, cmderr, err := vm.GetVMInfo("{{.metadata.uid}}")
	if err != nil {
		return "", err
	}
	if cmderr != "" {
		return "", errors.New(cmderr)
	}
	vm.UID = output
	return output, nil
}

func (vm VirtualMachine) ProcessTemplate() (string, error) {
	var args []string
	if vm.TemplateInCluster != "" {
		args = append(args, []string{"process", vm.TemplateInCluster}...)
	}

	if vm.TemplateFromFile != "" {
		args = append(args, []string{"process", "-f", vm.TemplateFromFile}...)
	}

	args = append(args, vm.TemplateParams...)

	output, cmderr, err := ktests.RunCommandWithNS(NamespaceTestTemplate, ktests.KubeVirtOcPath, args...)
	if err != nil {
		return "", err
	}
	if cmderr != "" {
		return "", errors.New(cmderr)
	}
	// TODO：if the image is pullable naturally, will remove the string replacement.
	if strings.Contains(output, "registry:5000/") {
		output = strings.Replace(output, "registry:5000/", "", -1)
	}
	if strings.Contains(output, ":devel") {
		output = strings.Replace(output, ":devel", ":latest", -1)
	}
	err = ioutil.WriteFile(vm.Manifest, []byte(output), 0644)

	return "", nil
}
