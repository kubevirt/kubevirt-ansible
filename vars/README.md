# Top Level Variables

List of top level variables.

| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| cluster| openshift|<ul><li>openshift</li><li>kubernetes</li></ul> |Cluster type to deploy KubeVirt on.|
| namespace|kube-system | |Namespace to create resources.|
| kubevirt_openshift_version | 3.7| <ul><li>3.7</li><li>3.9</li></ul>|OpenShift cluster version.|
|manifest_version | release|<ul><li>release</li><li>dev</li></ul>|KubeVirt manifest version.|
| docker_tag|latest| | Container image tag.|
| storage_role|storage-none|<ul><li>storage-none</li><li>storage-demo</li><li>storage-glusterfs</li></ul>| Storage role  to install with KubeVirt.|
