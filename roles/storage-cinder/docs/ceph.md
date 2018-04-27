## Ceph Cluster-side Administrative Tasks

### Create cinder_volumes pool
```
ceph osd pool create cinder_volumes 128
```

### Create 'cinder' user and give permission to use the 'cinder_volumes' pool
```
ceph-authtool -C /tmp/ceph.client.cinder.keyring -n client.cinder --cap osd 'allow rwx pool=cinder_volumes' --cap mon 'allow r' --cap mds 'allow rw' --gen-key
ceph auth add client.cinder -i /tmp/ceph.client.cinder.keyring 
```

### Gather the 'cinder' user's key
```
ceph auth list 2>/dev/null | grep -A1  client.cinder | grep key | awk '{print $2}'
```
