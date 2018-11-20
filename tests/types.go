package tests

import (
	"errors"
	"io/ioutil"
	"strings"

	ktests "kubevirt.io/kubevirt/tests"
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
}

func (vm VirtualMachine) Create() (string, string, error) {
	args := []string{"create", "-n", ktests.NamespaceTestDefault, "-f", vm.Manifest}
	return ktests.RunCommand(ktests.KubeVirtOcPath, args...)
}

func (vm VirtualMachine) Start() (string, string, error) {
	args := []string{"start", "-n", ktests.NamespaceTestDefault, vm.Name}
	return ktests.RunCommand(ktests.KubeVirtVirtctlPath, args...)
}

func (vm VirtualMachine) Stop() (string, string, error) {
	args := []string{"stop", "-n", ktests.NamespaceTestDefault, vm.Name}
	return ktests.RunCommand(ktests.KubeVirtVirtctlPath, args...)
}

func (vm VirtualMachine) Delete() (string, string, error) {
	args := []string{"delete", "-n", ktests.NamespaceTestDefault, vm.Type, vm.Name}
	return ktests.RunCommand(ktests.KubeVirtOcPath, args...)
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
	args := []string{"get", "-n", ktests.NamespaceTestDefault, vm.Type, vm.Name, "--template", spec}
	return ktests.RunCommand(ktests.KubeVirtOcPath, args...)
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
		args = append(args, []string{"process", "-n", NamespaceTestTemplate, vm.TemplateInCluster}...)
	}

	if vm.TemplateFromFile != "" {
		args = append(args, []string{"process", "-f", vm.TemplateFromFile}...)
	}

	args = append(args, vm.TemplateParams...)

	output, cmderr, err := ktests.RunCommand(ktests.KubeVirtOcPath, args...)
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
