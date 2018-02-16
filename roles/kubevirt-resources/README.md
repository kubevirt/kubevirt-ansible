# Kubevirt Resources Role

Deploy Kubevirt resources onto a cluster.

### Variables
| Variable        | Default Value           | Description  |
|:------------- |:-------------|:----- |
| admin_user | _optional_ | User with cluster-admin permissions. Used in the APB |
| admin_password | _optional_ | Password for cluster-admin user. Used in the APB |
| namespace | kube-system | Namespace to create resources |
| docker_prefix | kubevirt | Container image organization |
| docker_tag | latest | Container image tag |
