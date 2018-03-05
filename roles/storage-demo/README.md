# Ansible role: storage-demo

**Note: This role is not for production use.  All storage is erased any time**
**the storage-demo pod is stopped.**

This role deploys a self-contained storage environment suitable for development
and testing.  A single-instance ephemeral ceph cluster is created inside a pod.
Cinder is deployed and connected to ceph.  A dynamic provisioner and related
kubernetes resources are created to interface with the cluster.

## Configuration parameters
* **namespace**: The namespace into which the storage-demo components should be
  installed
* **cluster**: The type of cluster (openshift or kubernetes)
* **cinder_provisioner_repo**: The repository containing the cinder provisioner
* **cinder_provisioner_release**: The docker image tag to use for the cinder
  provisioner
* **action**: The action to perform.  Currently only 'provision' is supported
* **storage_demo_template_dir**: The location of the deployment template file.
  There is no need to change this value.
  
