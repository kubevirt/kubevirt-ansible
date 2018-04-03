# Ansible role: storage-glusterfs

This role deploys cluster resources necessary for kubevirt to interface with
GlusterFS storage.  This role assumes that a GlusterFS cluster managed by heketi
has already been installed.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|admin_user |   | _optional_ |User with cluster-admin permissions.|
|admin_password| |_optional_|Password for **admin_user**.|
|cluster |openshift |<ul><li>openshift</li><li>kubernetes</li></ul>|Cluster type.| 
|action |provision| <ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
| heketi_url | | The URL to the Heketi service. Ignored if heketi_route_name is specified. |
| heketi_route_name | heketi-storage | | (OpenShift only) The OpenShift route that exposes the Heketi service. |
| heketi_namespace | app-storage | | The namespace where Heketi is deployed. |
| heketi_secret_name | heketi-storage-admin-secret | | The name of the secret which contains Heketi credentials. |
