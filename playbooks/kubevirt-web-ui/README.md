# Kubevirt-web-ui deployment
Used for deployment of the [Kubevirt Web UI](https://github.com/kubevirt/web-ui) application into running OpenShift cluster.

The playbook is based on [opensift-ansible](https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-console).

### Required Variables
- `cluster`
  - To install Kubevirt Web UI, please set `cluster=openshift`
- `kubevirt_web_ui_image_name`
  - example: quay.io/kubevirt/kubevirt-web-ui:latest
  - The docker image with the kubevirt-web-ui application

### Optional Variables:
- `openshift_master_default_subdomain`
  - example: `router.default.svc.cluster.local`
  - Used for composition of web-ui's public URL
  - If not set, the default is retrieved from openshift-ansible deployment
- `public_master_hostname`
  - example: `master:8443`
  - Public URL of your first master node, used for composition of public `console` URL for redirects
  - If not set, the default is retrieved from openshift-ansible deployment

## How To Run
### Prerequisities
To run the playbook, an ansible's inventory file including the required variables as stated above is required.

From the hosts, just the master node is required.

Please check `playbooks/kubevirt-web-ui/inventory_example.ini` for an example.

### Invocation
```
ansible-playbook -i your_inventory_file.ini playbooks/kubevirt-web-ui/config.yml -e "apb_action=provision cluster=openshift"
```

