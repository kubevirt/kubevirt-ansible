# kubevirt-ansible
The purpose of this project to create Kubernetes or OpenShift cluster and deploy KubeVirt environment on it.
Tested on CentOS Linux release 7.3.1611 (Core).

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

### 
# Useful Links
- [**KubeVirt project**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible project**](https://github.com/openshift/openshift-ansible)
- [**Golang Ansible playbook project**](https://github.com/jlund/ansible-go)
