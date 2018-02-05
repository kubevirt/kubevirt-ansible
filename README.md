# kubevirt-ansible

This repository provides collection of playbooks to
* install KubeVirt on existing OpenShift cluster
* deploy K8S on given machines and install KubeVirt
* deploy OpenShift on given machines and install KubeVirt

Tested on CentOS Linux release 7.3.1611 (Core), OpenShift 3.7 and Ansible 2.3.1


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
        -e "openshift_ansible_dir=openshift-ansible/ kconfig=$HOME/.kube/config kubevirt_mf=kubevirt.yaml"
```

[![asciicast](https://asciinema.org/a/161148.png)](https://asciinema.org/a/161148)


## Deploy new cluster + KubeVirt


### Requirements
Install depending roles, and export `ANSIBLE_ROLES_PATH`
```
$ ansible-galaxy install -p $HOME/galaxy-roles -r requirements.yml
$ export ANSIBLE_ROLES_PATH=$HOME/galaxy-roles
```
For OpenShift deployment clone [**OpenShift Ansible project**](https://github.com/openshift/openshift-ansible)
```
$ git clone https://github.com/openshift/openshift-ansible
```

### Kubernetes
Preparing Kubernetes cluster and deploy KubeVirt on it.
- Add your master and nodes to `inventory` file.
- Run ansible playbook `# ansible-playbook -i inventory deploy-kubernetes.yml`.

### OpenShift
Preparing OpenShift cluster and deploy KubeVirt on it.
- Be sure that you have enough space on your hosts for docker storage and
edit openshift/roles/docker-setup/defaults/main.yaml accordingly.
You can read more about docker-storage-setup [**here**](https://docs.openshift.org/1.5/install_config/install/host_preparation.html#configuring-docker-storage).
- Add your master and nodes to `inventory` file.
- Run ansible playbook `# ansible-playbook -i inventory -e "openshift_ansible_dir=... deploy-openshift.yml`.
You must give directory where you placed `openshift-ansible` to the variable `openshift_ansible_dir`.

# Useful Links
- [**KubeVirt project**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible project**](https://github.com/openshift/openshift-ansible)
- [**Golang Ansible playbook project**](https://github.com/jlund/ansible-go)
