# KubeVirt Ansible

__Tools to provision resources, deploy clusters and install KubeVirt.__

## Overview

KubeVirt Ansible consists of a set of Ansible playbooks that deploy fully functional virtual machine management add-on for Kubernetes - KubeVirt. Optionally, a KubeVirt or OpenShift cluster can also be configured.

## Contents

* `atomation/`: CI scripts to verify the functionality of playbooks.
* `playbooks/`: Ansible playbooks to provision resources, deploy a cluster and install KubeVirt for various scenarios.
* `roles/`: Roles to use in playbooks.
* `vars/`: Variables to use in playbooks.
* `inventory`: A template for the cluster and nodes configuration.
* `requirements.yml`: A list of required Ansible-Galaxy roles to use in playbooks.
* `stdci.yaml`: A configuration file for CI system.

## Usage

See the [playbooks readme file](playbooks/README.md) for details.

## Questions ? Help ? Ideas ?

Stop by the [#kubevirt](https://webchat.freenode.net/?channels=kubevirt) chat channel on freenode IRC

## Contributing

Please see the [contributing guidelines](./CONTRIBUTING.md) for information regarding the contribution process.

## Automation & Testing

Please check the [CI automation guidelines](./automation/README.md) for information on playbooks verification.

# Useful Links
- [**KubeVirt**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible**][openshift-ansible-project]
- [**Golang Ansible playbook**](https://github.com/jlund/ansible-go)

[openshift-ansible-project]: https://github.com/openshift/openshift-ansible
