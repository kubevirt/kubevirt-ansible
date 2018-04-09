# KubeVirt Resources Role

Deploy KubeVirt resources onto a cluster.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|admin_user|   | _optional_ |User with cluster-admin permissions.|
|admin_password| |_optional_|Password for **admin_user**.|
|cluster|openshift |<ul><li>openshift</li><li>kubernetes</li></ul>|Cluster type.| 
|namespace|kube-system | |Namespace to create resources.|
|action|provision| <ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
|release_manifest_url|https://github.com/kubevirt/kubevirt/releases/download|||
|kubevirt_template_dir|./templates||Location of the deployment template file.|
|docker_prefix| kubevirt | |Container image organization.|
|storage_role|storage-none|<ul><li>storage-none</li><li>storage-demo</li><li>storage-glusterfs</li></ul>| Storage role  to install with KubeVirt.|
|version|v0.4.0|<ul><li>v0.4.0</li><li>v0.3.0</li><li>v0.2.0</li><li>v0.1.0</li></ul>|KubeVirt release version.|
