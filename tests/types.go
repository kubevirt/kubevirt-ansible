package tests

import (
	ktests "kubevirt.io/kubevirt/tests"
)

// A VMManifest contains a name, a manifest and an optional template of
// a virtual machine or virtual machine instance.
type VMManifest struct {
	Name     string   // vm or vmi name
	Manifest string   // vm or vmi manifest file
	Template string   // vm template name or vm template file
	Params   []string // vm template params
}

// ProcessClusterTemplate process vm template in the cluster.
func (vm VMManifest) ProcessClusterTemplate() (string, error) {
	args := []string{"process", "-n", TemplateNS, vm.Template}
	args = append(args, vm.Params...)
	return ktests.RunCommand("oc", args...)
}

// ProcessFileTemplate process vm template from a file.
func (vm VMManifest) ProcessFileTemplate() (string, error) {
	args := []string{"process", "-f", vm.Template}
	args = append(args, vm.Params...)
	return ktests.RunCommand("oc", args...)
}
