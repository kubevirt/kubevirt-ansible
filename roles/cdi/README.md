# CDI

This role deploys the CDI controller.

### Role Variables
| variable       | default           |choices           | comments  |
|:-------------|:-------------|:----------|:----------|
| cdi_image_namespace | kube-system | |Namespace into which the CDI components should be installed. |
| cdi_kubevirt_storageclass | kubevirt | |Storageclass that CDI will use to create PersistentVolumes. |
| apb_action | provision |<ul><li>provision</li><li>deprovision</li></ul>|Action to perform.|
| cdi_repo_tag | kubevirt | |CDI docker hub repo tag.|
| cdi_release_tag | v1.3.0 | |CDI docker hub release tag.|
| cdi_uploadproxy_image | cdi-uploadproxy | Name of CDI uploadproxy docker image. |
| cdi_apiserver_image | cdi-apiserver | Name of CDI apiserver docker image. |
| cdi_uploadserver_image | cdi-uploadserver | Name of CDI uploadserver docker image. |
| cdi_controller_image | cdi-controller | Name of CDI controller docker image. |
| cdi_importer_image | cdi-importer | Name of CDI importer docker image. |
| cdi_cloner_image | cdi-cloner | Name of CDI cloner docker image. |
| cdi_template_dir | ./templates||Location of the deployment template file.|

### Usage

```
ansible-playbook -i inventory -e apb_action=provision -e cdi_image_namespace=golden playbooks/cdi.yml
```
