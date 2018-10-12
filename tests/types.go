package tests

import (
	"io/ioutil"
	"strings"

	ktests "kubevirt.io/kubevirt/tests"
)

// A VMManifest contains a name and a manifest of
// a virtual machine or virtual machine instance.
type VMManifest struct {
	Name     string
	Manifest string
}

// VirtualMachine can be a vm, vmi, vmirs, vmiPreset.
type VirtualMachine struct {
	Name              string
	Type              string
	Manifest          string
	TemplateInCluster string
	TemplateFromFile  string
	TemplateParams    []string
}

func (vm VirtualMachine) Create() (string, error) {
	args := []string{"create", "-n", ktests.NamespaceTestDefault, "-f", vm.Manifest}
	return ktests.RunCommand("oc", args...)
}

func (vm VirtualMachine) Start() (string, error) {
	args := []string{"start", "-n", ktests.NamespaceTestDefault, vm.Name}
	return ktests.RunCommand("virtctl", args...)
}

func (vm VirtualMachine) Stop() (string, error) {
	args := []string{"stop", "-n", ktests.NamespaceTestDefault, vm.Name}
	return ktests.RunCommand("virtctl", args...)
}

func (vm VirtualMachine) Delete() (string, error) {
	args := []string{"delete", "-n", ktests.NamespaceTestDefault, vm.Type, vm.Name}
	return ktests.RunCommand("oc", args...)
}

func (vm VirtualMachine) IsRunning() (bool, error) {
	output, err := vm.GetVMInfo("{{.status.phase}}")
	if err != nil {
		return false, err
	}
	if output == "Running" {
		return true, nil
	}

	return false, nil
}

func (vm VirtualMachine) GetVMInfo(spec string) (string, error) {
	args := []string{"get", "-n", ktests.NamespaceTestDefault, vm.Type, vm.Name, "--template", spec}
	return ktests.RunCommand("oc", args...)
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

	output, err := ktests.RunCommand("oc", args...)
	if err != nil {
		return output, err
	}
	// TODO：if the image is pullable naturally, will remove the string replacement.
	if strings.Contains(output, "registry:5000/") {
		output = strings.Replace(output, "registry:5000/", "", -1)
	}
	if strings.Contains(output, ":devel") {
		output = strings.Replace(output, ":devel", ":latest", -1)
	}
	err = ioutil.WriteFile(vm.Manifest, []byte(output), 0644)

	return "", err
}
