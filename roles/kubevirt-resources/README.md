# KubeVirt Resources Role

Deploy KubeVirt resources onto a cluster.

### Variables
| Variable        | Default Value           | Description  |
|:------------- |:-------------|:----- |
| admin_user | _optional_ | User with cluster-admin permissions. Used in the APB |
| admin_password | _optional_ | Password for cluster-admin user. Used in the APB |
| namespace | kube-system | Namespace to create resources |
| manifest_version | release | KubeVirt manifest version |
| docker_prefix | kubevirt | Container image organization |
| dev_template_resources | [ "rbac.authorization.k8s", "replicase-resource", "virt-controller", "virt-handler", "vm-resource" ] | Individual resource templates |
| docker_tag | latest | Container image tag |
| action | provision | The action of provisioning or deprovisioning KubeVirt |
| cluster | openshift | The cluster we're running on |
