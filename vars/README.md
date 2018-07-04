# Top Level Variables

List of top level variables.

| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| cluster| openshift|<ul><li>openshift</li><li>kubernetes</li></ul> |Cluster type to deploy KubeVirt on.|
| namespace|kube-system | |Namespace to create resources.|
| kubevirt_openshift_version | 3.7| <ul><li>3.7</li><li>3.9</li></ul>|OpenShift cluster version.|
| version |0.7.0|<ul><li>0.7.0</li><li>0.6.0</li><li>0.5.0</li><li>0.4.1</li><li>0.4.0</li><li>0.3.0</li><li>0.2.0</li><li>0.1.0</li></ul>|KubeVirt release version.|
| storage_role | storage-none | <ul><li>storage-none</li><li>storage-demo</li><li>storage-glusterfs</li></ul> | Storage role to install with KubeVirt.|
