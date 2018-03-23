# Storage Demo NodeConfig

>**Note:** This role is not for production use.  All storage will be erased if the storage-demo pod is stopped.

This role configures the nodes in a cluster to work with the storage-demo role.
Currently this means adjusting the firewall to allow Ceph traffic to flow
between nodes.  As such, this will only work on a cluster that is running
iptables.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
|action|provision|<ul><li>provision</li><li>deprovision</li></ul> |Action to perform.|
