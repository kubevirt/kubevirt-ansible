# Automation with oVirt CI 
Any code changes in this repository** are verified using
[oVirt CI System][ovirt-ci-system-doc].

All configurtion and script files required
by [oVirt CI][ovirt-ci-system-doc] are located in the [automation](../automation)
directory. In addition, there is the [`stdci.yaml`](../stdci.yaml) configuration file.

According to [oVirt CI][ovirt-ci-system-doc] specification all incoming patches
are verified at the `check-patch` stage. Once a new PR is submitted to this repository,
the [oVirt CI System][ovirt-ci-system-doc] triggers
[Jenkins job](http://jenkins.ovirt.org/blue/organizations/jenkins/kubevirt_kubevirt-ansible_standard-check-pr/activity) which executes the [`automation/check-patch.sh`](./check-patch.sh) script.
This script runs the Ansible playbook
[`./playbooks/automation/check-patch.yml`](../playbooks/automation/check-patch.yml)
which wraps the entire testing flow.

>**Note:** Any KubeVirt component or related project can be automated using [oVirt-CI][ovirt-ci-system-doc]
and [oVirt CI Team](ovirt-ci-team) will be happy to help.

## About `check-patch.yml`

The [`./playbooks/automation/check-patch.yml`](../playbooks/automation/check-patch.yml)
Ansible playbook is composed of the following steps:
* [provision testing environment](#testing-environment): [playbooks/provider/lago/config.yml](../playbooks/provider/lago/config.yml)
* [deploy OpenShift](../README.md#cluster-configuration): [playbooks/cluster/openshift/config.yml](../playbooks/cluster/openshift/config.yml)
* [install KubeVirt](../README.md#install-kubevirt-on-an-existing-cluster): [playbooks/kubevirt.yml](../playbooks/kubevirt.yml)

This playbook is executed inside of CentOS 7.5 mock. The required packages
specified in [automation/check-patch.packages](./check-patch.packages)
are installed.

The usage and parameters of this playbook are described at
[Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago](../playbooks/README.md#deploy-a-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago).

### Testing environment

The Testing environment is populated by
[Lago](https://github.com/lago-project/lago) project according to requirements defined
in the Lago configuration file [`LagoInitFile.yml`](../playbooks/provider/lago/LagoInitFile.yml).

The testing environment consists of three virtual machines with CentOS 7.5, namely:
* 1x master node
* 2x compute nodes

For any details regarding these virtual machines, for example memory,
number of CPUs, disks alocation and et cetera, see
[`LagoInitFile.yml`](../playbooks/provider/lago/LagoInitFile.yml).

### Testing matrix

The testing environment can be populated with different versions of an OpenShift cluster.
It defines a testing matrix. The `substage` section of the [`stdci.yaml`](../stdci.yaml)
configuration file contains the currenly used matrix or matrices.

Files describing various matrices and labeled by the OpenShift version are located in
the [`automation`](../automation) directory. 

The KubeVirt version is constant and not changed for a testing matrix.

## How to run the CI locally

It's possible to reproduce the CI process on a local machine,
for example to experiment with KubeVirt or to debug errors.

### System requirements

* 8GB RAM memory
* 7GB Free Disk space
* Centos 7+ or Fedora 26+ OS
* Libvirt installed and running
* Nested virtualization enabled
* the `/var/lib/lago` directory, which is used by Lago to cache VM images
* Docker installed and running

### Preparing the environment

In order to avoid dependencies issues, the entire flow should run
inside a mock environment with automatically installed dependencies.

Follow the steps in [Setting up mock_runner] to configure a wrapper for [mock].

### Running the CI flow

Use the following steps in order to trigger the CI flow:

1. Create the mock environment

    From the root of kubevirt-ansible repository
    run the following command and replace the following variables:

    * `jenkins` - the directory containing the jenkins repository
      (see [Setting up mock_runner])

    * `cluster_version` - the CI flow to run, for example `openshift_3-10.sh`

    ```bash

      $ ${jenkins}/mock_configs/mock_runner.sh \
          --mock-confs-dir ${jenkins}/mock_configs/ \
          -e automation/check-patch.${cluster_version}.sh \
          -s \
          el7

    ```

2. Navigate into kubevirt-ansible repository

    The `mock_runner` tool bind mounts the current directory into
    the `$HOME` directory inside the mock environment.

    ```bash
      $ cd
    ```

3. Provision and deploy the environment

    ```bash
      $ ./automation/check-patch.${cluster_version}.sh
    ```


    By default the environment is removed when the script finishes.
    In order to keep it, use the `--skip-cleanup` flag.
    The environment can be removed by calling the same script with the
    `--only-cleanup` flag.

[ovirt-ci-system-doc]: http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Build_and_test_standards/index.html
[ovirt-ci-team]: https://ovirt-infra-docs.readthedocs.io/en/latest/General/Communication/index.html
[Setting up mock_runner]:
http://ovirt-infra-docs.readthedocs.io/en/latest/CI/Using_mock_runner/index.html
[mock]:
https://github.com/rpm-software-management/mock/wiki
