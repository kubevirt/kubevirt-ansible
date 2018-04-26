# OpenShift Cluster Playbook

Create an OpenShift Cluster

### Variables
| variable       | default           |choices           | comments  |
|:------------- |:-------------|:----- |:----- |
| kubevirt_openshift_version| 3.9 | <ul><li>3.7</li><li>3.9</li></ul>|OpenShift cluster version.|
| openshift_ansible_dir | openshift-ansible | |Path to the openshift-ansible directory.|
| openshift_playbook_path | playbooks/deploy_cluster.yml |<ul><li>playbooks/byo/config.yml</li><li>playbooks/deploy_cluster.yml</li></ul>|Path to the OpenShift deploy playbook. <br>3.7: **playbooks/byo/config.yml**<br>3.9: **playbooks/deploy_cluster.yml**|
| storage_role | | <ul><li>storage-glusterfs</li></ul> | Storage flavor to deploy in the cluster. Don't forget to add gluster nodes in the [inventory file](https://github.com/kubevirt/kubevirt-ansible/blob/master/inventory).|
