# Ansible Meta Role to Deploy Standalone Cinder

This role aggregates Cinder, RabbitMQ and MariaDB to deploy Standalone
Cinder with no authentication (noauth). 

MariaDB is deployed on a node with label=mariadb. MariaDB uses hostPath
for storage. 

### Role Variables
| variable | default |
|:-------------|:-------------|
| action | provision |
| infra_node_label | mariadb |
| namespace | openstack |
| service_account | cinder |
| privileged_service_account | cinder_privileged |
| mariadb_root_password | weakpassword |
| mariadb_user | root |
| cinder_password | cinderpassword |
| cinder_user | cinder |
| rabbitmq_password | rabbitmqpassword |
| rabbitmq_user | guest |

### Backend Specific Variables

#### Xtremio
| variable | default |
|:-------------|:-------------|
| xtremio.enabled | false |
| xtremio.max_over_subscription_ratio | 40 |
| xtremio.use_multipath_for_image_xfer | true |
| xtremio.san_ip | |
| xtremio.xtremio_cluster_name | |
| xtremio.san_login | | 
| xtremio.san_password | |
| xtremio.image_volume_cache_enabled | |

#### Ceph
| variable | default |
|:-------------|:-------------|
| ceph.cinder_rbd_pool_name | cinder_volumes |
| ceph.cinder_rbd_user_name | cinder |
| ceph.ceph_authentication_type | cephx |
| ceph.client_key | |
| ceph.ceph_mon_host | |



