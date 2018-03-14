# Top Level Variables

List of top level variables.

### Variables
| Variable        | Default Value           | Description  |
|:------------- |:-------------|:----- |
| cluster | openshift | The cluster we're running on |
| namespace | kube-system | Namespace to create resources |
| openshift_version | 3.7 | OpenShift cluster version. Either 3.7 or 3.9 |
| manifest_version | release | KubeVirt manifest version |
| docker_tag | latest | Container image tag |
| storage_role | None | Storage role  to install with KubeVirt |


### Storage Roles

**storage-demo**
All-in-one ephemeral storage.

**CNS**
```openshift-ansible``` will provide Gluster storage and the CNS role will
create the storage class.

**Cinder**
