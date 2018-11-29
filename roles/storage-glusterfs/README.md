# Ansible role: storage-glusterfs

This role deploys cluster resources necessary for kubevirt to interface with
GlusterFS storage.  This role assumes that a GlusterFS cluster managed by heketi
has already been installed and also that these have been deployed or configured
using either [openshift-ansible](https://github.com/openshift/openshift-ansible)
or [gk-deploy](https://github.com/gluster/gluster-kubernetes/).

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|admin_user |   | _optional_ |User with cluster-admin permissions.|
|admin_password| |_optional_|Password for **admin_user**.|
|cluster |openshift |<ul><li>openshift</li><li>kubernetes</li></ul>|Cluster type.| 
|apb_action |provision| <ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
| glusterfs_namespace | glusterfs | | The namespace where GlusterFS and Heketi resources are deployed. |
| glusterfs_name | storage | | The name of the GlusterFS installation. |
| heketi_url | | | (Optional) The URL to the Heketi service. Auto-detected if unspecified. |
| heketi_admin_key | | | (Optional) The key for calling Heketi as the user 'admin'. |
| external_provisioner | false | <ul><li>true</li><li>false</li></ul> | Whether to use the external GlusterFS provisioner. Enables additional features. |

### Usage

```
ansible-playbook -i inventory -e action=provision -e glusterfs_namespace=glusterfs -e storage_role=storage-glusterfs playbooks/storage.yml
```