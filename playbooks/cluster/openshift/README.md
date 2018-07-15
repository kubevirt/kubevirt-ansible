# OpenShift Cluster Playbook

Create an OpenShift Cluster

### Variables
| variable       | default           |choices           | comments  |
|:------------- |:-------------|:----- |:----- |
| kubevirt_openshift_version| 3.10| <li>3.10</li></ul>|OpenShift cluster version.|
| openshift_ansible_dir | openshift-ansible | |Path to the openshift-ansible directory.|
| openshift_playbook_path | playbooks/byo/config.yml |<ul><li>playbooks/byo/config.yml</li><li>playbooks/deploy_cluster.yml</li></ul>|Path to the OpenShift deploy playbook. 3.10: **playbooks/deploy_cluster.yml**|
| storage_role | | <ul><li>storage-glusterfs</li></ul> | Storage flavor to deploy in the cluster. Don't forget to add gluster nodes in the [inventory file](https://github.com/kubevirt/kubevirt-ansible/blob/master/inventory). For OpenShift, you will also need to have DNS configured such that `heketi-{{ glusterfs_name }}-{{ glusterfs_namespace }}.{{ openshift_master_default_subdomain }}` resolves to the IP address of a node running an OpenShift router.|
