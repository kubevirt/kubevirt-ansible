# Network skydive project

This role deploys the [skydive-project](http://skydive.network/) 

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|skydive_namespace | skydive |  |Namespace into which the skydive components should be installed.|
|cluster|openshift|<ul><li>openshift</li><li>kubernetes</li></ul> |Cluster type.|


### Usage

```
ansible-playbook -i inventory -e apb_action=provision -e deploy_skydive=True playbooks/network.yml
```