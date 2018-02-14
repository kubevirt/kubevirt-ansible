# KubeVirt Ansible

This repository provides collection of playbooks to
- [x] [install KubeVirt on existing OpenShift cluster](#install-kubevirt-on-existing-cluster)
- [ ] deploy Kubernetes cluster on given machines and install KubeVirt
- [x] [deploy OpenShift cluster on given machines and install KubeVirt](#deploy-kubernetes-or-openshift-and-kubevirt)
- [ ] deploy Kubernetes cluster with KubeVirt with Lago
- [x] [deploy OpenShift cluster with KubeVirt with Lago](#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago)

**Tested on CentOS Linux release 7.4 (Core), OpenShift 3.7 and Ansible 2.4.2**

## List of playbooks

| Playbook name | Description |
| ------------- | ----------- |
| install-kubevirt-on-openshift.yml | This playbook installs KubeVirt on existing OpenShift cluster. |
| deploy-kubernetes.yml | This playbook deploys Kubernetes cluster on given machines. |
| deploy-openshift.yml | This playbook deploys OpenShift cluster on given machines. |
| control.yml | This is top level playbook which deploys Kubernetes or OpenShift cluster using Lago provider and installs KubeVirt on top of it. |

## Requirements

Install depending roles, and export `ANSIBLE_ROLES_PATH`

```bash
$ ansible-galaxy install -p $HOME/galaxy-roles -r requirements.yml
$ export ANSIBLE_ROLES_PATH=$HOME/galaxy-roles
```

For OpenShift deployment clone [**OpenShift Ansible project**][openshift-ansible-project]

```bash
$ git clone -b release-3.7 https://github.com/openshift/openshift-ansible
```

## (Optional) Setup a cluster

This section describes how to setup cluster on given machines. There are two
playbook which you can use depends on what cluster you want to bring up.


### General resources for cluster

There are three key cluster components which needs to be deployed.

* master - is component which is brain of cluster
* etcd - is database of cluster
* node - is compute component of cluster

There are two basic environment scenarios how cluster can be deployed.
If you need more information please read
[documentation](https://docs.openshift.org/latest/install_config/install/planning.html).

* All-in-one
  * all components on single machine
* Single master and multiple nodes
  * master & etcd is deployed on single machine
  * nodes are separated machines

For minimal hardware requirements please follow
[documentation](https://docs.openshift.org/latest/install_config/install/prerequisites.html) .

When you take a look at [inventory](./inventory) file, you can find three
host groups which you need fill in with your machines according to prefered
topology.

For example if you choose `all-in-one` topology you can read following inventory file.

```ini
[masters]
master.example.com
[etcd]
master.example.com
[nodes]
master.example.com openshift_node_labels="{'region': 'infra','zone': 'default'}" openshift_schedulable=true
```

### Kubernetes cluster

- Add your master and nodes to `inventory` file.
- Run [deploy-kubernetes.yml](./deploy-kubernetes.yml) playbook

```bash
$ ansible-playbook -i inventory deploy-kubernetes.yml
```

### OpenShift cluster


- Be sure that you have enough space on your machines for docker storage and
you can modify [defaults values](docker-storage-setup-defaults) accordingly.
Follow [docker-storage-setup] documentation for more details.
- Add your master and nodes to `inventory` file.
- Run [deploy-openshift.yml](./deploy-openshift.yml) playbook

| Parameter name | Value | Description |
| -------------- | ----- | ----------- |
| openshift\_ansible\_dir | string | Path to [OpenShift Ansible project][openshift-ansible-project] |

```bash
$ ansible-playbook -i inventory \
    -e "openshift_ansible_dir=openshift-ansible/" deploy-openshift.yml
```

## Install KubeVirt on existing cluster

### Kubernetes cluster

**TODO:** Currently we don't have a playbook which installs KubeVirt on Kubernetes cluster.

### OpenShift cluster

When you have your cluster up and running you can use
[install-kubevirt-on-openshift.yml](./install-kubevirt-on-openshift.yml)
playbook to install KubeVirt from given manifest.

| Name             |  Value        | Description                            |
| ---------------- | ------------- | -------------------------------------- |
| kubevirt\_mf     | string | Path to KubeVirt manifest, you can get one from [releases](https://github.com/kubevirt/kubevirt/releases) or build it from sources |
| openshift\_ansible\_dir | string | Path to [OpenShift Ansible project][openshift-ansible-project] |
| kconfig | string | Path to kubeconfig |

```bash
$ # Get KubeVirt manifest
$ wget https://github.com/kubevirt/kubevirt/releases/download/v0.2.0/kubevirt.yaml
$ ansible-playbook -i localhost, --connection=local \
        -e "openshift_ansible_dir=openshift-ansible/ \
        kconfig=$HOME/.kube/config \
        kubevirt_mf=kubevirt.yaml" \
        install-kubevirt-on-openshift.yml
```

[![asciicast](https://asciinema.org/a/161278.png)](https://asciinema.org/a/161278)


## Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago

This section describes how to deploy KubeVirt or OpenShift cluster with KubeVirt installed on it using Lago.

### Parameters

| Name             |  Value        | Description                            |
| ---------------- | ------------- | -------------------------------------- |
| cluster\_type    | `kubernetes`, `openshift` | Desired cluster            |
| mode             | `release`, `dev` | If `dev` it will build KubeVirt from sources |
| openshift\_ansible\_dir | string | Path to [OpenShift Ansible project][openshift-ansible-project] |
| provider         | `lago` | So far `lago` is only supported provider |
| inventory\_file  | string | Path to inventory file, lago is adding VMs into inventory file once they are created, so it needs to know where it is located. |

### Example how to run playbooks

Following example is executing top level flow `control.yml` which is executed
by tests.

```bash
$ ansible-playbook -i inventory \
    -e "cluster_type=openshift \
    mode=release \
    provider=lago \
    inventory_file=inventory \
    openshift_ansible_dir=openshift-ansible/" \
    control.yml
```

# Useful Links
- [**KubeVirt project**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible project**][openshift-ansible-project]
- [**Golang Ansible playbook project**](https://github.com/jlund/ansible-go)

[docker-storage-setup]: https://docs.openshift.org/latest/install_config/install/host_preparation.html#configuring-docker-storage
[docker-storage-setup-defaults]: https://github.com/openshift/openshift-ansible-contrib/blob/master/roles/docker-storage-setup/defaults/main.yaml
[openshift-ansible-project]: https://github.com/openshift/openshift-ansible
