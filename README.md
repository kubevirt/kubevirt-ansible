# KubeVirt Ansible

This repository provides collection of playbooks to
- [x] [Install KubeVirt on existing OpenShift cluster](#install-kubevirt-on-existing-cluster)
- [ ] Deploy Kubernetes cluster on given machines and install KubeVirt
- [x] [Deploy OpenShift cluster on given machines and install KubeVirt](#deploy-kubernetes-or-openshift-and-kubevirt)
- [x] [Provision resources, deploy cluster and install KubeVirt](#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago)

*NOTE: Checked box means that playbook is working and supported, unchecked box means that playbook needs stabilization.*

**Tested on CentOS Linux release 7.4 (Core), OpenShift 3.7, OpenShift 3.9 and Ansible 2.4.2**

## List of playbooks

| Playbook name | Description |
| ------------- | ----------- |
| install-kubevirt-on-openshift.yml | This playbook installs KubeVirt on existing OpenShift cluster. |
| deploy-kubernetes.yml | This playbook deploys Kubernetes cluster on given machines. |
| deploy-openshift.yml | This playbook deploys OpenShift cluster on given machines. |
| control.yml | This is top level playbook which encapsulates entire flow, provision resources, deploying cluster, and installing KubeVirt on it. |

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

* **master**

  The master is the host or hosts that contain the master components,
  including the API server, controller manager server, and etcd.
  The master manages nodes in its Kubernetes cluster and schedules pods
  to run on nodes.

* **etcd**

  **etcd** stores the persistent master state while other components watch
  **etcd** for changes to bring themselves into the desired state.

* **node**

  A node provides the runtime environments for containers.
  Each node in a Kubernetes cluster has the required services
  to be managed by the master. Nodes also have the required services
  to run pods, including the Docker service, a kubelet, and a service proxy.

Beside these components you might want to specify
[persistent storage for Ansible Broker](https://docs.openshift.org/latest/install_config/install/advanced_install.html#configuring-openshift-ansible-broker) .
Without persistent storage you can not use Ansible Broker.
Following example shows usage of NFS storage but there are
[other options available](https://docs.openshift.org/latest/install_config/persistent_storage/index.html#install-config-persistent-storage-index).

* **nfs**

  A node which will be the NFS host

There are two basic environment scenarios on how a cluster can be deployed.
If you need more information please read
[documentation](https://docs.openshift.org/latest/install_config/install/planning.html).

* All-in-one
  * all components on single machine
* Single master and multiple nodes
  * master & etcd & nfs is deployed on single machine
  * nodes are separated machines

For minimal hardware requirements please follow
[documentation](https://docs.openshift.org/latest/install_config/install/prerequisites.html) .

If you take a look at [inventory](./inventory) file, you can find four
host groups which you need fill in with your machines according to prefered
topology.

For example if you choose `all-in-one` topology you can read the following inventory file.

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
| kubeconfig | string | Path to the [`~/.kube/config`](https://docs.openshift.com/container-platform/3.7/cli_reference/manage_cli_profiles.html#switching-between-cli-profiles) file containing authentication information for the desired cluster |

```bash
$ # Get KubeVirt manifest
$ wget https://github.com/kubevirt/kubevirt/releases/download/v0.2.0/kubevirt.yaml
$ ansible-playbook -i localhost, --connection=local \
        -e "openshift_ansible_dir=openshift-ansible/ \
        kubeconfig=$HOME/.kube/config \
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
| openshift\_ansible\_dir | string | Path to [OpenShift Ansible project][openshift-ansible-project] |
| provider         | `lago` | So far `lago` is only supported provider |
| inventory\_file  | string | Path to inventory file, lago is adding VMs into inventory file once they are created, so it needs to know where it is located. |

### Example how to run playbooks

Following example is executing top level flow `control.yml` which is executed
by tests.

```bash
$ ansible-playbook -i inventory \
    -e "cluster_type=openshift \
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
