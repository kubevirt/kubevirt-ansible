# kubevirt-ansible

This repository provides collection of playbooks to
* install KubeVirt on existing OpenShift cluster
* deploy K8S on given machines and install KubeVirt
* deploy OpenShift on given machines and install KubeVirt

**Tested on CentOS Linux release 7.3.1611 (Core), OpenShift 3.7 and Ansible 2.3.1**


## Install KubeVirt on existing OpenShift cluster

When you have your cluster up and running you can use
[install-kubevirt-on-openshift.yml](./install-kubevirt-on-openshift.yml)
playbook to install KubeVirt from given manifest.

You need three things to do that
* **kubevirt_mf** - Path to KubeVirt manifest, you can get one from
  [releases](https://github.com/kubevirt/kubevirt/releases) or build it from sources
* **openshift_ansible_dir** - Path to
  [OpenShift Ansible](https://github.com/openshift/openshift-ansible) repository
* **kconfig** - Path to kubeconfig

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


## Deploy new cluster + KubeVirt

### Requirements

Make sure that you have Ansible 2.3.1.

Install depending roles, and export `ANSIBLE_ROLES_PATH`

```bash
$ ansible-galaxy install -p $HOME/galaxy-roles -r requirements.yml
$ export ANSIBLE_ROLES_PATH=$HOME/galaxy-roles
```

For OpenShift deployment clone [**OpenShift Ansible project**](https://github.com/openshift/openshift-ansible)

```bash
$ git clone -b release-3.7 https://github.com/openshift/openshift-ansible
```

### Parameters

| Name             |  Value        | Description                            |
| ---------------- | ------------- | -------------------------------------- |
| cluster\_type    | `kubernetes`, `openshift` | Desired cluster            |
| mode             | `release`, `dev` | If `dev` it will build KubeVirt from sources |
| openshift\_ansible\_dir | string | Path to OpenShift Ansible repository   |

### Kubernetes

Preparing Kubernetes cluster and deploy KubeVirt on it.
- Add your master and nodes to `inventory` file.

### OpenShift
Preparing OpenShift cluster and deploy KubeVirt on it.
- Be sure that you have enough space on your hosts for docker storage and
edit openshift/roles/docker-setup/defaults/main.yaml accordingly.
You can read more about docker-storage-setup [**here**](https://docs.openshift.org/1.5/install_config/install/host_preparation.html#configuring-docker-storage).
- Add your master and nodes to `inventory` file.


### Example how to run playbook

```bash
$ ansible-playbook -i inventory \
    -e "cluster_type=openshift \
    mode=release \
    openshift_ansible_dir=openshift-ansible/" \
    control.yml
```

# Useful Links
- [**KubeVirt project**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible project**](https://github.com/openshift/openshift-ansible)
- [**Golang Ansible playbook project**](https://github.com/jlund/ansible-go)
