# CDI

This role deploys the CDI controller.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| cdi_image_namespace | golden-images | |Namespace into which the CDI components should be installed. |
| cdi_kubevirt_storageclass | kubevirt | |Storageclass that CDI will use to create PersistentVolumes. |
| apb_action | provision |<ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
| cdi_repo_tag | kubevirt | |CDI docker hub repo tag.|
| cdi_release_tag | v1.1.1 | |CDI docker hub release tag.|

### Usage

```
ansible-playbook -i inventory -e apb_action=provision -e cdi_image_namespace=golden playbooks/cdi.yml
```
