# Logo Provider Playbook

Create VMs with Lago

### Variables
| Variable        | Default Value           | Description  |
|:------------- |:-------------|:----- |
| cluster | openshift | The cluster we're running on |
| provider | lago | VM provider to use |
| inventory_file | inventory | Path to inventory file for ansible to use |
| openshift_version | 3.7 | OpenShift cluster version. Either 3.7 or 3.9 |
| openshift_ansible_dir | openshift-ansible | Path to openshift-ansible directory |
| openshift_playbook_path | playbooks/byo/config.yml | Path to OpenShift deploy playbook (3.7 -> playbooks/byo/config.yml, 3.9 -> playbooks/deploy_cluster.yml |
