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

This playbook is executed inside of CentOS 7.5 mock.
There is additional software installed, please read
[automation/check-patch.packages](./check-patch.packages) for complete list.

Parameters and usage of this playbook is described at
[Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago](../playbooks/README.md#deploy-a-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago).

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

It will provision three virtual machines with CentOS 7.5
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

## oVirt CI Integration

The [oVirt CI][ovirt-ci-system-doc] project is open to integrating additional
KubeVirt related projects or integrations, to run per each submitted PR.

## How to run the CI locally

It's possible to reproduce the CI process on your local machine.
This is useful when there is a need to debug errors, or even if
you just want to experiment with KubeVirt.

### System requirements

|               |         |
|---------------|---------|
| RAM memory    | 8GB
| Free Disk space| 7GB
| OS            | Centos 7+ or Fedora 26+

* Libvirt is installed and running.
* Nested virtualization is enabled.
* `/var/lib/lago` directory exists (will be used by Lago for caching VM images).
* Docker is installed and running.

### Preparing the environment

In order to avoid dependencies issues, the entire flow should be run
inside a mock environment, that will be populated with the dependencies automatically.

We are going to use `mock_runner`, which is a wrapper for [mock].
Please follow the steps in [Setting up mock_runner].

### Running the CI flow

Use the following step in order to trigger the CI flow:

#### Create the mock environment

From the root of kubevirt-ansible repository,
run the following command (replace the following variables):

* `jenkins` - the directory that contains the jenkins repository
  (as explained in [Setting up mock_runner])

* `cluster_version` - The CI flow that you want to run, for example `openshift_3-9.sh`

```bash

  $ ${jenkins}/mock_configs/mock_runner.sh \
      --mock-confs-dir ${jenkins}/mock_configs/ \
      -e automation/check-patch.${cluster_version}.sh \
      -s \
      el7

```

#### Cd into kubevirt-ansible repository

The mock_runner tool bind mount the the current directory into
the `$HOME` directory inside the mock env.

```bash
  $ cd
```

#### Provision and deploy the environment

```bash
    ./automation/check-patch.${cluster_version}.sh
```


The default is to remove the env when script finishes.
In order to keep the env, add the `--skip-cleanup` flag.
When finished, you can remove it by calling the same script with
`--only-cleanup`.

[ovirt-ci-system-doc]: http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Build_and_test_standards/index.html
[Setting up mock_runner]:
http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Using_mock_runner/index.html

[mock]:
https://github.com/rpm-software-management/mock/wiki
