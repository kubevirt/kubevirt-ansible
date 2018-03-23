# Storage CNS

This role deploys cluster resources necessary for KubeVirt to interface with
CNS (glusterfs) storage.  This role assumes that CNS itself has already been
installed.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|admin_user |   | _optional_ |User with cluster-admin permissions.|
|admin_password| |_optional_|Password for **admin_user**.|
|cluster |openshift |<ul><li>openshift</li><li>kubernetes</li></ul>|Cluster type.| 
|action |provision| <ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
