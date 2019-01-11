# ember-csi-ansible
Ansible Role to deploy Ember CSI via Ember CSI operator

### Role Variables

| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| apb_action | provision |<ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
| ember_csi_namespace | ember-csi | |Namespace where Ember CSI is deployed|
| repo_name | quay.io/kirankt | |Image repo name|
| operator_image | ember-csi-operator | |Ember CSI operator image name|
| release_tag | v0.0.3 | |Ember CSI operator release name|
| cluster_command | oc |<ul><li>oc</li><li>kubectl</li></ul>|Openshift/K8s cluster command|

### Usage

```
ansible-playbook -i inventory -e apb_action=provision playbooks/ember-csi.yml
```
