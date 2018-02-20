# Automation with oVirt CI

In order to keep playbooks in this repository operational,
it is being integrated with [oVirt CI System][ovirt-ci-system-doc].
Everything located under `automation` directory is related to integration
to [oVirt CI System][ovirt-ci-system-doc].

There is list of events which can trigger job on [oVirt CI System][ovirt-ci-system-doc],
The KubeVirt Ansible repository is using `check-patch` event at the moment,
to verify incoming patches.

Once you submit PR for this repository, the [oVirt CI System][ovirt-ci-system-doc]
triggers [Jenkins job](http://jenkins.ovirt.org/blue/organizations/jenkins/kubevirt_kubevirt-ansible_standard-check-pr/activity)
which executes [`automation/check-patch.sh`](./check-patch.sh) script.

This script will execute Ansible playbook [`./control.yml`](../control.yml)
which wraps entire testing flow.


## `control.yml` Ansible playbook

This Ansible playbook is wrapper for entire testing flow which is composed from
following steps (partial playbooks)
* [provision testing environment](#testing-environment)
* [deploy OpenShift](#deploy-openshift)
* [install KubeVirt](#install-kubevirt)
* [TODO](https://github.com/kubevirt/kubevirt-ansible/issues/47) run KubeVirt test suite

This playbook is executed by Ansible 2.4.2.

Parameters and usage of this playbook is described at
[Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago](../README.md#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago).

### Testing environment

To provision testing environment the [`./deploy-with-lago.yml`](../deploy-with-lago.yml) playbook is used.

Testing environment is populated by
[Lago](https://github.com/lago-project/lago) project, it provisions desired
resources to match requirements described in Lago configuration file.
This configuration file is located in root of repository
[`LagoInitFile.yml`](../LagoInitFile.yml).

It will provision three virtual machines with CentOS 7.4
* 1x master node
* 2x compute node


| Node type | Memory | CPUs | Disks | Ansible Groups |
| ---- | ---- | ---- | ---- | ---- |
| master | 4Gb | ? | * root<br> * docker\_storage (10Gb)<br> * docker\_lib (10Gb)<br> * main\_nfs (101Gb) | * masters<br> * nodes<br> * nfs<br> * etcd |
| node   | 2Gb | ? | * root<br> * docker\_storage (10Gb)<br> * docker\_lib (10Gb) | * nodes |


### Deploy OpenShift

To deploy OpenShift the [`./deploy-openshift.yml`](../deploy-openshift.yml) playbook is used.

Test flow deploys OpenShift v3.7 .

### Install KubeVirt

To install KubeVirt on top of  OpenShift the [`./install-kubevirt-release.yml`](../install-kubevirt-release.yml) playbook is used.

Test flow installs KubeVirt v0.2.0 .

[ovirt-ci-system-doc]: http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Build_and_test_standards/index.html
