This repository provides a collection of playbooks to

- [x] [Deploy an OpenShift cluster on given machines](#deploy-kubernetes-or-openshift-and-kubevirt): `playbooks/cluster/openshift/config.yml`
- [x] [Install KubeVirt on an existing OpenShift cluster](#install-kubevirt-on-existing-cluster): `playbooks/kubevirt.yml`
- [ ] Deploy a Kubernetes cluster on given machines and install KubeVirt: `playbooks/cluster/kubernetes/config.yml`
- [x] [Provision resources, deploy a cluster and install KubeVirt](#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago): `playbooks/automation/check-patch.yml`

> **NOTE:** Checked box means that playbook is working and supported, unchecked box means that playbook needs stabilization.

**Tested on CentOS Linux release 7.4 (Core), OpenShift 3.7, OpenShift 3.9 and Ansible 2.4.2, 2.4.3**

## Requirements

Install depending roles, and export `ANSIBLE_ROLES_PATH`

```bash
$ git clone https://github.com/kubevirt/kubevirt-ansible
$ cd kubevirt-ansible
$ mkdir $HOME/galaxy-roles
$ ansible-galaxy install -p $HOME/galaxy-roles -r requirements.yml
$ export ANSIBLE_ROLES_PATH=$HOME/galaxy-roles
```

For OpenShift deployment clone [**OpenShift Ansible**](openshift-ansible-project)

```bash
$ git clone -b release-3.7 https://github.com/openshift/openshift-ansible
```

> **NOTE:** For OpenShift 3.9 use `release-3.9` branch.

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
* Single master and multiple nodes, where master, etcd and persistent storage are assigned to a single machine, and nodes are on separated machines.

For more information on cluster configuration please see [Planning](https://docs.openshift.org/latest/install_config/install/planning.html) and [Requirements](https://docs.openshift.org/latest/install_config/install/prerequisites.html) documentation.

Edit the [inventory](../inventory) file according to chosen cluster configuration. For example, for an `all-in-one` case it will look like:

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
Make sure that you have **passwordless ssh access** to all machines of the cluster.

### Kubernetes cluster


```bash
$ ansible-playbook -i inventory playbooks/cluster/kubernetes/config.yml
```

### OpenShift cluster

Be sure that you have an **extra disk** attached to your machines
for **docker storage** and modify [defaults values][docker-storage-setup-defaults]
accordingly. In most of cases you need to set `docker_dev` variable
to match the name of your extra disk for docker storage.
Please follow [docker-storage-setup] documentation for more details.


```bash
$ ansible-playbook -i inventory -e@vars/all.yml playbooks/cluster/openshift/config.yml
```
See [OpenShift parameters documentation](./cluster/openshift/README.md) for more details and update [vars/all.yml](../vars/all.yml) if needed.

## Install KubeVirt on an existing cluster

### Kubernetes cluster

Currently we don't have a playbook which installs KubeVirt on a Kubernetes cluster.

### OpenShift cluster

The playbook expects you have ```oc```, rpm package ```origin-clients```,
installed and a config file in ```$HOME/.kube/config```.  Also make sure you
indentified with the cluster using ```oc login```.

Before installing KubeVirt on an existing OpenShift cluster, ensure that SELinux is disabled on all hosts:

```bash
$ ansible-playbook -i inventory playbooks/selinux.yml
```

Be sure to update the [inventory file](../inventory) according to your OpenSift cluster configuration or use the file you used to deploy the cluster.

Install KubeVirt on your OpenShift cluster:

```bash
$ ansible-playbook -i localhost playbooks/kubevirt.yml -e@vars/all.yml
```

See [KubeVirt parameters documentation](../roles/kubevirt/README.md) for more details and update [vars/all.yml](../vars/all.yml) if needed.

## Deploy a new Kubernetes or OpenShift cluster and KubeVirt with Lago

This section describes how to deploy a KubeVirt or OpenShift cluster with KubeVirt installed on it using Lago.

Following example is executing top level playbook `playbooks/automation/check-patch.yml`.

```bash
$ ansible-playbook -i inventory -e@vars/all.yml playbooks/automation/check-patch.yml
```

See [Lago parameters documentation](./playbooks/provider/lago/README.md) for more details and update [vars/all.yml](../vars/all.yml) if needed.

### Storage

**storage-demo**
All-in-one ephemeral storage.

**GlusterFS**
```openshift-ansible``` will provide GlusterFS storage and the storage-glusterfs role will
create the StorageClass.

**Cinder**


[docker-storage-setup]: https://docs.openshift.org/latest/install_config/install/host_preparation.html#configuring-docker-storage
[docker-storage-setup-defaults]: https://github.com/openshift/openshift-ansible-contrib/blob/master/roles/docker-storage-setup/defaults/main.yaml
[openshift-ansible-project]: https://github.com/openshift/openshift-ansible
