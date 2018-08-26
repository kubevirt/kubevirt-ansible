# Network Multus cni plugin

This role deploys the network attachment CRD multus daemonset, cni plugins and a L2 plugin.

### Kubernetes Usage
For a kubernetes cluster if you are using a network plugin different than flannel you need to edit the `kubernetes_cni_config` variable
 
```
defaults/main.yml
```

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|namespace | kube-system |  |Namespace into which the multus and cni plugins components should be installed.|
|cluster|openshift|<ul><li>openshift</li><li>kubernetes</li></ul> |Cluster type.|


### Usage

```
ansible-playbook -i inventory -e apb_action=provision -e network_role=network-multus playbooks/network.yml
```