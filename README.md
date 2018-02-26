# KubeVirt Ansible

This repository provides a collection of playbooks to
- [x] [Deploy an OpenShift cluster on given machines](#deploy-kubernetes-or-openshift-and-kubevirt): `playbooks/cluster/openshift/config.yml`
- [x] [Install KubeVirt on an existing OpenShift cluster](#install-kubevirt-on-existing-cluster): `playbooks/components/install-kubevirt-on-openshift.yml`
- [ ] Deploy a Kubernetes cluster on given machines and install KubeVirt: `playbooks/cluster/kubernetes/config.yml`
- [x] [Provision resources, deploy a cluster and install KubeVirt](#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago): `control.yml`

> **NOTE:** Checked box means that playbook is working and supported, unchecked box means that playbook needs stabilization.

**Tested on CentOS Linux release 7.4 (Core), OpenShift 3.7, OpenShift 3.9 and Ansible 2.4.2**

## Requirements

Install depending roles, and export `ANSIBLE_ROLES_PATH`

```bash
$ git clone https://github.com/kubevirt/kubevirt-ansible
$ cd kubevirt-ansible
$ mkdir $HOME/galaxy-roles
$ ansible-galaxy install -p $HOME/galaxy-roles -r requirements.yml
$ export ANSIBLE_ROLES_PATH=$HOME/galaxy-roles
```

For OpenShift deployment clone [**OpenShift Ansible**][openshift-ansible-project]

```bash
$ git clone -b release-3.7 https://github.com/openshift/openshift-ansible
```

## Cluster configuration
This section describes how to set up a new cluster on given machines. [Skip](#install-kubevirt-on-an-existing-cluster) this part if you already have a cluster.

There are three key cluster components which need to be deployed: master, etcd and node.

* The **master** is the host or hosts that contain the master components,
  including the API server, controller manager server, and etcd.
  The master manages nodes in its Kubernetes cluster and schedules pods
  to run on nodes.

* **etcd** stores the persistent master state while other components watch
  etcd for changes to bring themselves into the desired state.

* A **node** provides the runtime environments for containers.
  Each node in a Kubernetes cluster has the required services
  to be managed by the master. Nodes also have the required services
  to run pods, including the Docker service, a kubelet, and a service proxy.

Beside these three key components you might want to specify [persistent storage](https://docs.openshift.org/latest/install_config/persistent_storage/index.html#install-config-persistent-storage-index) component required for using [Ansible Broker](https://docs.openshift.org/latest/install_config/install/advanced_install.html#configuring-openshift-ansible-broker). Below cluster configuration sample uses **nfs** component as an example of NFS persistent storage.

There are two supported cluster configurations:

* All-in-one, where all cluster components are deployed on a single machine
* Single master and multiple nodes, where master, etcd and persistant storage are assiged to a single machine, and nodes are on separated machines.

For more inforamation on cluster configuration please see [Planning](https://docs.openshift.org/latest/install_config/install/planning.html) and [Requirements](https://docs.openshift.org/latest/install_config/install/prerequisites.html) documentation.

Edit the [inventory](./inventory) file according to chosen cluster configuration. For example, for an `all-in-one` case it will look like:

```ini
[masters]
master.example.com
[etcd]
master.example.com
[nodes]
master.example.com openshift_node_labels="{'region': 'infra','zone': 'default'}" openshift_schedulable=true
[nfs]
master.example.com
```

Using this inventory file one can deploy a Kubernetes or OpenShift cluster.

### Kubernetes cluster


```bash
$ ansible-playbook -i inventory playbooks/cluster/kubernetes/config.yml
```

### OpenShift cluster


Be sure that you have enough space on your machines for docker storage and
modify [defaults values](docker-storage-setup-defaults) accordingly.
Follow [docker-storage-setup] documentation for more details.


```bash
$ ansible-playbook -i inventory \
    -e "openshift_ansible_dir=openshift-ansible/ \
    openshift_playbook_path=playbooks/byo/config.yml \
    openshift_ver=3.7" playbooks/cluster/openshift/config.yml
```
where
* `openshift_ver` specifies what version of OpenShift one wants to deploy. Choose from
  * 3.7
  * 3.9
* `openshift_ansible_dir` is a path to a cloned [OpenShift Ansible][openshift-ansible-project] git repository.
* `openshift_playbook_path` is a path to OpenShift deploy playbook in [OpenShift Ansible][openshift-ansible-project]. Choose from
  * `playbooks/byo/config.yml` for OpenShift 3.7 (default)
  * `playbooks/deploy_cluster.yml` for OpenShift 3.9

## Install KubeVirt on an existing cluster

### Kubernetes cluster

Currently we don't have a playbook which installs KubeVirt on a Kubernetes cluster.

### OpenShift cluster

When you have your cluster up and running you can install KubeVirt from a given manifest.

```bash
$ # Get KubeVirt manifest
$ wget https://github.com/kubevirt/kubevirt/releases/download/v0.2.0/kubevirt.yaml
$ ansible-playbook -i localhost, --connection=local \
        -e "openshift_ansible_dir=openshift-ansible/ \
        kubeconfig=$HOME/.kube/config \
        kubevirt_mf=kubevirt.yaml" \
        playbooks/components/install-kubevirt-on-openshift.yml
```
where
* `kubevirt_mf` is a path to KubeVirt manifest. See [releases](https://github.com/kubevirt/kubevirt/releases) or build it from sources,
* `openshift_ansible_dir` is a path to a cloned [OpenShift Ansible][openshift-ansible-project] git repository,
* `kubeconfig` is a path to the [`~/.kube/config`](https://docs.openshift.com/container-platform/3.7/cli_reference/manage_cli_profiles.html#switching-between-cli-profiles) file containing authentication information for the desired cluster.

[![asciicast](https://asciinema.org/a/161278.png)](https://asciinema.org/a/161278)


## Deploy a new Kubernetes or OpenShift cluster and KubeVirt with Lago

This section describes how to deploy a KubeVirt or OpenShift cluster with KubeVirt installed on it using Lago.

Following example is executing top level playbook `control.yml`.

```bash
$ ansible-playbook -i inventory \
    -e "cluster_type=openshift \
    provider=lago \
    inventory_file=inventory \
    openshift_ver=3.7 \
    openshift_playbook_path=playbooks/byo/config.yml \
    openshift_ansible_dir=openshift-ansible/" \
    control.yml
```

where
* `cluster_type=kubernetes|openshift` defines a desired cluster type,
* `openshift_ver` specifies what version of OpenShift one wants to deploy. Choose from
  * 3.7
  * 3.9
* `openshift_ansible_dir` is a path to a cloned [OpenShift Ansible][openshift-ansible-project] git repository,
* `openshift_playbook_path` is a path to OpenShift deploy playbook in [OpenShift Ansible][openshift-ansible-project]. Choose from
  * `playbooks/byo/config.yml` for OpenShift 3.7 (default)
  * `playbooks/deploy_cluster.yml` for OpenShift 3.9
* `provider` defines provider. At the moment, `lago` is the only supported provider,
* `inventory_file` is a path to the inventory file. Lago adds VMs into the inventory file once they are created, so it needs to know where it is located.

## Questions ? Help ? Ideas ?

Stop by the [#kubevirt](https://webchat.freenode.net/?channels=kubevirt) chat channel on freenode IRC

## Contributing

Please see the [contributing guidelines](./CONTRIBUTING.md) for information regarding the contribution process.

# Useful Links
- [**KubeVirt**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible**][openshift-ansible-project]
- [**Golang Ansible playbook**](https://github.com/jlund/ansible-go)

[docker-storage-setup]: https://docs.openshift.org/latest/install_config/install/host_preparation.html#configuring-docker-storage
[docker-storage-setup-defaults]: https://github.com/openshift/openshift-ansible-contrib/blob/master/roles/docker-storage-setup/defaults/main.yaml
[openshift-ansible-project]: https://github.com/openshift/openshift-ansible
