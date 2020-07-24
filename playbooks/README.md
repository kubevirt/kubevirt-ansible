This repository provides a collection of playbooks to

- [x] [Deploy an OpenShift cluster on given machines](#deploy-kubernetes-or-openshift-and-kubevirt): `playbooks/cluster/openshift/config.yml`
- [x] [Install KubeVirt on an existing OpenShift cluster](#install-kubevirt-on-existing-cluster): `playbooks/kubevirt.yml`
- [ ] Deploy a Kubernetes cluster on given machines and install KubeVirt: `playbooks/cluster/kubernetes/config.yml`
- [x] [Provision resources, deploy a cluster and install KubeVirt](#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago): `playbooks/automation/check-patch.yml`

> **NOTE:** Checked box means that playbook is working and supported, unchecked box means that playbook needs stabilization.

**Tested on CentOS Linux release 7.5 (Core), OpenShift 3.10 and Ansible 2.6.1**

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
$ git clone -b release-3.10 https://github.com/openshift/openshift-ansible
```

Once cloned, you should configure `openshift_ansible_dir` to point to the local
repo.

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

Beside these three key components you might want to specify [persistent storage](https://docs.openshift.com/enterprise/latest/install_config/persistent_storage/index.html) component required for using [Ansible Broker](https://docs.openshift.com/container-platform/latest/install_config/oab_broker_configuration.html). Below cluster configuration sample uses **nfs** component as an example of NFS persistent storage.

There are two supported cluster configurations:

* All-in-one, where all cluster components are deployed on a single machine
* Single master and multiple nodes, where master, etcd and persistent storage are assigned to a single machine, and nodes are on separated machines.

For more information on cluster configuration please see [Planning](https://docs.okd.io/latest/install/index.html) and [Requirements](https://docs.okd.io/latest/install/prerequisites.html) documentation.

Edit the [inventory](../inventory) file according to chosen cluster configuration. For example, for an `all-in-one` case it will look like:

```ini
[masters]
master.example.com
[etcd]
master.example.com
[nodes]
master.example.com openshift_node_labels="{'region': 'infra','zone': 'default'}" openshift_schedulable=true openshift_node_group_name='node-config-infra-compute'
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
for **docker storage** and modify [defaults values][container_runtime-defaults]
accordingly. The docker storage is installed using [container_runtime] role from openshift-ansible.
In most of cases you need to set `container_runtime_docker_storage_setup_device` variable
to match the name of your extra disk for docker storage.
Please follow [docker-storage-setup] documentation for more details.


```bash
$ ansible-playbook -i inventory -e@vars/all.yml playbooks/cluster/openshift/config.yml -e "openshift_ansible_dir=$PWD/openshift-ansible/"
```
See [OpenShift parameters documentation](./cluster/openshift/README.md) for more details and update [vars/all.yml](../vars/all.yml) if needed.

## Install KubeVirt on an existing cluster

### Kubernetes cluster

Currently we don't have a playbook which installs KubeVirt on a Kubernetes cluster.

### OpenShift cluster

The playbook expects you have ```oc```, rpm package ```origin-clients```,
installed and a config file in ```$HOME/.kube/config```.  Also make sure you
identified with the cluster using ```oc login```.

Be sure to update the [inventory file](../inventory) according to your OpenShift cluster configuration or use the file you used to deploy the cluster.

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

See [Lago parameters documentation](../playbooks/provider/lago/README.md) for more details and update [vars/all.yml](../vars/all.yml) if needed.

### Storage

**storage-demo**
All-in-one ephemeral storage.

**GlusterFS**
```openshift-ansible``` will provide GlusterFS storage and the storage-glusterfs role will
create the StorageClass.

**Cinder**

### Network

**network-multus**
Deploy additional multus cni plugin.

***Note***  This is a dev preview, not OpenShift officially supported.

***Note***  For a kubernetes cluster if you are using a network plugin different than flannel you need to edit the `kubernetes_cni_config` variable inside the file:         
```
roles/network-multus/defaults/main.yml
```

**network-none**
No additional cni plugin will be deployed.

### SELinux

In case you are experiencing permission or SELinux issues, please consider
creating an issue for [kubevirt](https://github.com/kubevirt/kubevirt/issues/)
and report what is not working for you.

As a temporary workaround, you can disable SELinux by running following playbook.

```bash
$ ansible-playbook -i inventory -e "selinux=permissive" playbooks/selinux.yml
```

Be sure to update the [inventory file](../inventory) according to your OpenShift cluster configuration or use the file you used to deploy the cluster.

### SR-IOV

```kubevirt-ansible``` enables SR-IOV by default, using the kubernetes CNI
plugin ```multus```.

```bash
# vars/all.yml
...
deploy_sriov_plugin: true
```

For SR-IOV to work properly, configure your hosts to enable IOMMU, SR-IOV in
BIOS and kernel command line, and load the ```vfio-pci``` kernel module. Also,
enable the expected number of VFs before starting your cluster.

SR-IOV device plugin uses ```/etc/pcidp/config.json``` file to determine which
PFs should be used to allocate VFs from. By default, ```kubevirt-ansible```
doesn't configure any PFs, meaning that the plugin has zero devices to manage.

You can configure allocatable PFs as follows:

$ ansible-playbook -i inventory -e@vars/all.yml -e "sriov_pci_ids=\"0000:05:00.0\",\"0000:05:00.1\"" playbooks/kubevirt.yaml

Alternatively, you can, after successful deployment, edit SR-IOV ConfigMap
resource where you can specify default as well as per node PF addresses:

$ kubectl -n kubevirt edit configmap sriov-nodes-config

After that, remove corresponding SR-IOV device plugin pods so that new
instances start with the new configuration.

[container_runtime]: https://github.com/openshift/openshift-ansible/tree/master/roles/container_runtime
[docker-storage-setup]: https://docs.openshift.org/latest/install_config/install/host_preparation.html#configuring-docker-storage
[container_runtime-defaults]: https://github.com/openshift/openshift-ansible/blob/master/roles/container_runtime/defaults/main.yml
[openshift-ansible-project]: https://github.com/openshift/openshift-ansible
