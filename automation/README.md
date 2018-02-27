# Automation with oVirt CI

In order to keep playbooks in this repository operational,
it is being integrated with [oVirt CI System][ovirt-ci-system-doc].
Everything located under `automation` directory is related to integration
to [oVirt CI System][ovirt-ci-system-doc].
In addition there is [`stdci.yaml`](../stdci.yaml) a configuration file for
[oVirt CI System][ovirt-ci-system-doc].

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
* [deploy OpenShift](../README.md#cluster-configuration)
* [install KubeVirt](../README.md#install-kubevirt-on-an-existing-cluster)

This playbook is executed inside of CentOS 7.4 mock.
There is additional software installed, please read
[automation/check-patch.packages](./check-patch.packages) for complete list.

Parameters and usage of this playbook is described at
[Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago](../README.md#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago).

This is list of playbooks which are executed.
* [playbooks/automation/check-patch.yml](../playbooks/automation/check-patch.yml)
  * [playbooks/provider/lago/config.yml](../playbooks/provider/lago/config.yml)
  * [playbooks/cluster/openshift/config.yml](../playbooks/cluster/openshift/config.yml)
  * [playbooks/kubevirt.yml](../playbooks/kubevirt.yml)

### Testing environment

To provision testing environment the [`playbooks/provider/lago/config.yml`](../playbooks/provider/lago/config.yml) playbook is used.

Testing environment is populated by
[Lago](https://github.com/lago-project/lago) project, it provisions desired
resources to match requirements described in Lago configuration file.
This configuration file is located in `playbooks/provider/lago` directory
named [`LagoInitFile.yml`](../playbooks/provider/lago/LagoInitFile.yml),
and it contains all details about testing environment.

It will provision three virtual machines with CentOS 7.4
* 1x master node
* 2x compute node

For any details regarding these virtual machines, for example memory,
number of CPUs, disks alocation and etcetera, please read
[`LagoInitFile.yml`](../playbooks/provider/lago/LagoInitFile.yml) !

### Testing matrix

The testing environment is populated with different versions of OpenShift cluster.
Please read [`stdci.yaml`](../stdci.yaml) configuration file,
current matrix is defined under `substage` section.

If you take a look under [`automation`](./automation) you can find files there,
which has a name of substage in name. So then you can understand what files
are related to specific substage.

The KubeVirt version is static for whole matrix.

# oVirt CI Integration

The [oVirt CI][ovirt-ci-system-doc] project is open to integrating additional
KubeVirt related projects or integrations, to run per each submitted PR.

[ovirt-ci-system-doc]: http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Build_and_test_standards/index.html
