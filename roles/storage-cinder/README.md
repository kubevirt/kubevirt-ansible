# Ansible Meta Role to Deploy Standalone Cinder

This role aggregates Cinder, RabbitMQ and MariaDB to deploy Standalone
Cinder with no authentication (noauth). 

MariaDB is deployed on a node with label=mariadb. MariaDB uses hostPath
for storage. 

## Configuration parameters
* **namespace**: The namespace into which the Cinder components should be
  installed. Defaults to 'kube-system'
* **cluster**: The type of cluster (openshift or kubernetes). Defaults to 'openshift'
* **action**: The action to perform. Defaults to 'provision'
* **cinder_template_dir**: The location of the deployment template file.
  There is no need to change this value.
* **service_account**: Defaults to cinder
* **privileged_service_account**: Defaults to cinder_privileged
* **mariadb_root_password**: Defaults to 'weakpassword'
* **mariadb_user**: Defaults to 'root'
* **cinder_password**: Defaults to 'cinderpassword'
* **cinder_user**: Defaults to 'cinder'
* **rabbitmq_password**: Defaults to 'rabbitmqpassword'
* **rabbitmq_user**: Defaults to 'guest'

## Backend Specific Parameters
One of more of the backends have to be enabled.
* **cinder_enable_xtremio_backend**: Default is 'false'
* **cinder_enable_rbd_backend**: Default is 'false'
* **cinder_enable_netapp_backend**: Default is 'false'

### Xtremio
Required when xtremio backend is enabled.
* **xtremio.max_over_subscription_ratio**: Default is '40'
* **xtremio.use_multipath_for_image_xfer**: Default is 'true'
* **xtremio.san_ip**:
* **xtremio.xtremio_cluster_name**:
* **xtremio.san_login**:
* **xtremio.san_password**:
* **xtremio.image_volume_cache_enabled**:

## Ceph
* **ceph.cinder_rbd_pool_name**: Default is 'cinder_volumes'
* **ceph.cinder_rbd_user_name**: Default is 'cinder'
* **ceph.ceph_authentication_type**: Default is 'cephx'
* **ceph.client_key**:
* **ceph.ceph_mon_host**:


  
