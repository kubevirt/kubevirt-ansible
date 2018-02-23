# Automation with oVirt CI

In order to keep playbooks in this repository operational,
it is being integrated with [oVirt CI System][ovirt-ci-system-doc].
Everything located under `automation` directory is related to integration
to [oVirt CI System][ovirt-ci-system-doc].

There is a list of events which can trigger a job on [oVirt CI System][ovirt-ci-system-doc],
The KubeVirt Ansible repository uses the `check-patch` event at the moment,
to verify incoming patches.

Once you submit a PR for this repository, the [oVirt CI System][ovirt-ci-system-doc]
triggers [Jenkins job](http://jenkins.ovirt.org/blue/organizations/jenkins/kubevirt_kubevirt-ansible_standard-check-pr/activity)
which executes the [`automation/check-patch.sh`](./check-patch.sh) script.

This script will execute the Ansible playbook [`./playbooks/automation/check-patch.yml`](../playbooks/automation/check-patch.yml)
which wraps entire testing flow.


## `check-patch.yml` Ansible playbook

This Ansible playbook is a wrapper for the entire testing flow which is composed from
following steps (partial playbooks)
* [provision testing environment](#testing-environment)
* [deploy OpenShift](#deploy-openshift)
* [install KubeVirt](#install-kubevirt)

This playbook is executed inside of CentOS 7.4 mock.
There is additional software installed, please read
[automation/check-patch.packages](./check-patch.packages) for complete list.

Parameters and usage of this playbook is described at
[Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago](../README.md#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago).

This is list of playbooks which are executed.
* [check-patch.yml](../playbooks/automation/check-patch.yml)
  * [playbooks/provider/lago/config.yml](../playbooks/provider/lago/config.yml)
  * [playbooks/cluster/openshift/config.yml](../playbooks/cluster/openshift/config.yml)
  * [playbooks/components/install-kubevirt-release.yml](../playbooks/components/install-kubevirt-release.yml)

### Testing environment

To provision testing environment the [`playbooks/provider/lago/config.yml`](../playbooks/provider/lago/config.yml) playbook is used.

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

To deploy OpenShift the [`./playbooks/cluster/openshift/config.yml`](../playbooks/cluster/openshift/config.yml) playbook is used.

Test flow deploys OpenShift in two versions:

* OpenShift 3.7
* OpenShift 3.9

These two deployments are running simultaneously, each in it's own separated environment.
The flow maybe be extended with upcoming OpenShift releases in future.

### Install KubeVirt

To install KubeVirt on top of  OpenShift the [`./playbooks/components/install-kubevirt-release.yml`](../playbooks/components/install-kubevirt-release.yml) playbook is used.

Test flow installs KubeVirt v0.2.0 .


# oVirt CI Integration

The [oVirt CI][ovirt-ci-system-doc] project is open to integrating additional
KubeVirt related projects or integrations, to run per each submitted PR.

[ovirt-ci-system-doc]: http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Build_and_test_standards/index.html
