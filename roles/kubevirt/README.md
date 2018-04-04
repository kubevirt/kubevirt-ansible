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
|manifest_version| release |<ul><li>release</li><li>dev</li></ul>| KubeVirt manifest version. |
|kubevirt_manifest_url|https://raw.githubusercontent.com/kubevirt/kubevirt/master/manifests|||
|kubevirt_template_dir|./templates||Location of the deployment template file.|
|docker_prefix| kubevirt | |Container image organization.|
|dev_template_resources| |<ul><li>rbac.authorization.k8s</li><li>replicase-resource</li><li>virt-controller</li><li>virt-handler</li><li>vm-resource</li><li>offline-vm</li><li>vmpreset-resource</li></ul>| Individual resource templates.|
| docker_tag|latest| | Container image tag.|
|storage_role|storage-none|<ul><li>storage-none</li><li>storage-demo</li><li>storage-glusterfs</li></ul>| Storage role  to install with KubeVirt.|
