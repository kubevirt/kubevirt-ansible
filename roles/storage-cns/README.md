# Ansible role: storage-cns

This role deploys cluster resources necessary for kubevirt to interface with
CNS (glusterfs) storage.  This role assumes that CNS itself has already been
installed.

### Variables
| Variable        | Default Value           | Description  |
|:------------- |:-------------|:----- |
| admin_user | _optional_ | User with cluster-admin permissions. Used in the APB |
| admin_password | _optional_ | Password for cluster-admin user. Used in the APB |
| cluster | openshift | The cluster we're running on |
| action | provision | The action of provisioning or deprovisioning CNS add-ons |
