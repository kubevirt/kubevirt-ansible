# CDI

This role deploys the CDI controller.

### Usage

```
ansible-playbook -i inventory -e action=provision -e cdi_image_namespace=golden playbooks/cdi.yml
```

### Variables
| Variable        | Default Value           | Description  |
|:------------- |:-------------|:----- |
| cdi_image_namespace | golden-images | The namespace into which the CDI components should be installed |
| action | provision | The action of provisioning or deprovisioning CDI |
| cdi_repo_tag | jcoperh | CDI docker hub repo tag |
| cdi_release_tag | latest | CDI docker hub release tag |


