# Storage Cinder Role

This role aggregates Cinder, RabbitMQ and MariaDB to deploy Standalone
Cinder with no authentication (noauth).

MariaDB is deployed on a node with label=mariadb. MariaDB uses hostPath
for storage.

### Requirements


### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| action |  provision | <ul><li>provision</li><li>deprovision</li></ul>| Action to perform on the role|
| infra_node_label | controller | | Label a node to allow MariaDB to utilize its hostpath
| namespace | openstack | | A namespace where cinder and its dependencies will be deployed | 
| service_account | cinder | | A service account with at least anyuid capability for use in OpenShift |
| privileged_service_account | cinder_privileged | | A privileged service account for elevated privileges in OpenShift |
| mariadb_root_password | weakpassword | | |
| mariadb_user | root | | |
| cinder_user | cinder | | |
| cinder_password | cinderpassword | | |
| rabbitmq_user | guest | | |
| rabbitmq_password | rabbitmqpassword | | |

### Backend-specific Variables

#### Ceph
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| enabled | false  | <ul><li>true</li><li>false</li></ul>| Enables Ceph backend |
| cinder_rbd_pool_name | cinder_volumes  | | |
| cinder_rbd_user_name | cinder  | | |
| client_key | | | |
| ceph_authentication_type | cephx  | | |
| ceph_mon_host | | | IP address of Ceph Monitors. Comma-separated list of IPs |

#### Xtremio
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| enabled | false  | <ul><li>true</li><li>false</li></ul>| Enables Xtremio backend |
| max_over_subscription_ratio | 40  | | |
| use_multipath_for_image_xfer | | | |
| volume_backend_name | xtremio | | |
| san_ip | | | |
| san_login | | | |
| san_password | | | |
| image_volume_cache_enabled | | | |

#### NetApp
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| enabled | false  | <ul><li>true</li><li>false</li></ul>| Enables Netapp backend |
| netapp_storage_family | ontap_cluster  | | |
| netapp_storage_protocol: | | | |
| nfs_shares_config | | | |
| nfs_shares | | | |
| netapp_server_hostname | | | |
| netapp_server_port | | | |
| netapp_login | | | |
| netapp_password | | | |

### Usage

```
- name: storage-cinder
  hosts: localhost
  gather_facts: false
  connection: local
  vars:
    action: provision
    namespace: openstack
    ceph:
        enabled: true
        cinder_rbd_pool_name: cinder_volumes
        cinder_rbd_user_name: cinder
        client_key: keykeykeykeykeykeykeykeykeykeykeykeykeykeykeykeykey
        authentication_type: cephx
        ceph_mon_host: 10.10.10.10

  roles:
    - role: storage-cinder
      playbook_debug: false

```

