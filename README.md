# KubeVirt Ansible

This repository provides collection of playbooks to
- [x] [install KubeVirt on existing OpenShift cluster](#install-kubevirt-on-existing-openshift-cluster)
- [ ] deploy Kubernetes cluster on given machines and install KubeVirt
- [x] [deploy OpenShift cluster on given machines and install KubeVirt](#deploy-kubernetes-or-openshift-and-kubevirt)
- [ ] deploy Kubernetes cluster with KubeVirt with Lago
- [x] [deploy Kubernetes cluster with KubeVirt with Lago](#deploy-new-kubernetes-or-openshift-cluster-and-kubevirt-with-lago)

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

## Install KubeVirt on existing OpenShift cluster

When you have your cluster up and running you can use
[install-kubevirt-on-openshift.yml](./install-kubevirt-on-openshift.yml)
playbook to install KubeVirt from given manifest.

| Name             |  Value        | Description                            |
| ---------------- | ------------- | -------------------------------------- |
| kubevirt_mf      | string | Path to KubeVirt manifest, you can get one from [releases](https://github.com/kubevirt/kubevirt/releases) or build it from sources |
| openshift_ansible_dir | string | Path to [OpenShift Ansible][openshift-ansible-project] repository |
| kconfig | string | Path to kubeconfig |

```bash
$ wget https://github.com/kubevirt/kubevirt/releases/download/v0.2.0/kubevirt.yaml
$ git clone https://github.com/openshift/openshift-ansible.git
$ ansible-playbook -i localhost, --connection=local \
        -e "openshift_ansible_dir=openshift-ansible/ \
        kconfig=$HOME/.kube/config \
        kubevirt_mf=kubevirt.yaml" \
        install-kubevirt-on-openshift.yml
```

[![asciicast](https://asciinema.org/a/161278.png)](https://asciinema.org/a/161278)


## Deploy Kubernetes or OpenShift and KubeVirt

In order to deploy new cluster with KubeVirt on top of it, you need to perform 2 steps

1. Run one of following playbooks to deploy new cluster

  * [deploy-kubernetes.yml](./deploy-kubernetes.yml) for Kubernetes cluster
  * [deploy-openshift.yml](./deploy-openshift.yml) for OpenShift cluster

2. Run [install-kubevirt-on-openshift.yml](./install-kubevirt-on-openshift.yml) playbook to install KubeVirt

  Follow [Install KubeVirt on existing OpenShift cluster](#install-kubevirt-on-existing-openshift-cluster) section

### Kubernetes

- Add your master and nodes to `inventory` file.

```bash
$ # 1) deploy Kubernetes
$ ansible-playbook -i inventory deploy-kubernetes.yml
$ # 2) install KubeVirt, unfortunately we don't have playbook to install KubeVirt on Kubernetes
$ # ansible-playbook ... install-kubevirt-on-kubernetes.yml
```

### OpenShift

- Be sure that you have enough space on your hosts for docker storage and
you can modify [defaults values](docker-storage-setup-defaults) accordingly.
Follow [docker-storage-setup] documentation for more details.
- Add your master and nodes to `inventory` file.

```bash
$ # 1) deploy OpenShift
$ ansible-playbook -i inventory \
    -e "openshift_ansible_dir=openshift-ansible/" deploy-openshift.yml
$ # 2) install KubeVirt
$ ansible-playbook ... install-kubevirt-on-openshift.yml
```

## Deploy new Kubernetes or OpenShift cluster and KubeVirt with Lago

This section describes how to deploy KubeVirt or OpenShift cluster with KubeVirt installed on it using Lago.

### Parameters

| Name             |  Value        | Description                            |
| ---------------- | ------------- | -------------------------------------- |
| cluster\_type    | `kubernetes`, `openshift` | Desired cluster            |
| mode             | `release`, `dev` | If `dev` it will build KubeVirt from sources |
| openshift\_ansible\_dir | string | Path to OpenShift Ansible repository   |
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
