PROJECT="default"
oc create sa kubevirt-controller
oc adm policy add-cluster-role-to-user cluster-admin -z kubevirt-controller
oc patch deployment/virt-controller --patch '{"spec":{"template":{"spec":{"serviceAccountName": "kubevirt-controller"}}}}'
oc create sa kubevirt-iscsi
oc adm policy add-scc-to-user hostmount-anyuid -z kubevirt-iscsi
oc patch deployment/iscsi-demo-target-tgtd --patch '{"spec":{"template":{"spec":{"serviceAccountName": "kubevirt-iscsi"}}}}'
oc create sa kubevirt-privileged
oc adm policy add-scc-to-user privileged -z kubevirt-privileged
oc adm policy add-cluster-role-to-user cluster-admin -z kubevirt-privileged
oc patch deployment/virt-manifest --patch '{"spec":{"template":{"spec":{"serviceAccountName": "kubevirt-privileged"}}}}'
oc patch daemonset/virt-handler --patch '{"spec":{"template":{"spec":{"serviceAccountName": "kubevirt-privileged"}}}}'
oc patch daemonset/libvirt --patch '{"spec":{"template":{"spec":{"serviceAccountName": "kubevirt-privileged"}}}}'kubevirt-privileged
