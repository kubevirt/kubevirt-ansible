# Ansible role: storage-demo-nodeconfig

**Note: This role is not for production use.  All storage is erased any time**
**the storage-demo pod is stopped.**

This role configures the nodes in a cluster to work with the storage-demo role.
Currently this means adjusting the firewall to allow ceph traffic to flow
between nodes.  As such, this will only work on a cluster that is running
iptables.
