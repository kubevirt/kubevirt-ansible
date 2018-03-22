# OpenShift Cluster Playbook

Create an OpenShift Cluster

### Variables
| variable       | default           |choices           | comments  |
|:------------- |:-------------|:----- |:----- |
| openshift_version| 3.7| <ul><li>3.7</li><li>3.9</li></ul>|OpenShift cluster version.|
| openshift_ansible_dir | openshift-ansible | |Path to the openshift-ansible directory.|
| openshift_playbook_path | playbooks/byo/config.yml |<ul><li>playbooks/byo/config.yml</li><li>playbooks/deploy_cluster.yml</li></ul>|Path to the OpenShift deploy playbook. <br>3.7: **playbooks/byo/config.yml**<br>3.9: **playbooks/deploy_cluster.yml**|
