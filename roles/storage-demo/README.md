# Storage Demo

> **Note:** This role is not for production use.  All storage will be erased if the storage-demo pod is stopped.

This role deploys a self-contained storage environment suitable for development
and testing.  A single-instance ephemeral Ceph cluster is created inside a pod.
Cinder is deployed and connected to Ceph.  A dynamic provisioner and related
kubernetes resources are created to interface with the cluster.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|storage_namespace | kube-system |  |Namespace into which the storage-demo components should be installed.|
|cluster|openshift|<ul><li>openshift</li><li>kubernetes</li></ul> |Cluster type.|
|cinder_provisioner_repo|quay.io/aglitke| |Repository containing the Cinder provisioner.| 
|cinder_provisioner_release|sprint4| |Docker image tag to use for the Cinder provisioner.|
|apb_action|provision| |Action to perform.  Currently only **provision** is supported.|
|storage_demo_template_dir| ./templates| |Location of the deployment template file.|

### Usage

```
ansible-playbook -i inventory -e apb_action=provision -e storage_role=storage-demo playbooks/storage.yml
```