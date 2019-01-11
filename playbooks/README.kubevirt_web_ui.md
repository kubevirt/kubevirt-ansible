# Kubevirt-web-ui Deployment
Used for deployment of the [Kubevirt Web UI](https://github.com/kubevirt/web-ui) application into running OpenShift cluster.

The playbook deploys:
- [KubeVirt Web UI Operator](https://github.com/kubevirt/web-ui-operator)
- Custom Resource for the Operator to initiate application deployment

Based on content of the Custom Resource (CR), the [KubeVirt Web UI](https://github.com/kubevirt/web-ui) is (un)installed.

## Parameters
Default parameters can be found in `vars/all.yml` or `roles/kubevirt_web_ui/defaults/main.yml`.

### Generic parameters:
- platform: must be set to "openshift"
- apb_action: `provision` or `deprovision`
- registry_url: docker registry
  - example: `quay.io`
- registry_namespace: 
  - example: kubevirt

### Specific parameters:
- kubevirt_web_ui_operator_image_tag: image tag of the operator
  - example: v1.4.0-3
  - list of available tags: [see quay.io](https://quay.io/repository/kubevirt/kubevirt-web-ui-operator?tab=tags)
- kubevirt_web_ui_version:
  - version of the Web UI to be installed by the Operator
  - can be be changed after initial deployment by patching the Custom Resource (see bellow)
  - please note, the preffixed `v` is missing
  - example: 1.4.0-9
  - list of Web UI releases: [https://github.com/kubevirt/web-ui/releases](https://github.com/kubevirt/web-ui/releases)
  - list of docker tags: [https://quay.io/repository/kubevirt/kubevirt-web-ui?tab=tags](https://quay.io/repository/kubevirt/kubevirt-web-ui?tab=tags)
kubevirt_web_ui_branding: either `openshiftvirt` or `okdvirt`

### Optional Variables:
Following parameters _must_ be set if the `openshift-console` project is not present (please note, it is installed by default with openshift).

- `openshift_master_default_subdomain`
  - example: `router.default.svc.cluster.local`
  - Used for composition of web-ui's public URL
  - If not set, the default is retrieved from openshift-ansible deployment
- `public_master_hostname`
  - example: `master:8443`
  - Public URL of your first master node, used for composition of public `console` URL for redirects
  - If not set, the default is retrieved from openshift-ansible deployment

## How To Run
Please see the Parameters section above.

## Prerequisities
Prior running the playbook, the user needs to be logged in OC (see `~/.kube/config`).

### Invocation Examples
For default parameters:
```
$ ansible-playbook -e@vars/all.yml ./playbooks/kubevirt_web_ui.yml -vvv
```

For customization:
```
$ ansible-playbook -e@vars/all.yml -e@vars/my_vars.yml ./playbooks/kubevirt_web_ui.yml -vvv

$ cat vars/my_vars.yml
registry_url: my.registry.com:8888
registry_namespace: my-registry-namespace

kubevirt_web_ui_operator_image_tag: v1.4.0-3
kubevirt_web_ui_version: 1.4.0-9
kubevirt_web_ui_branding: openshiftvirt
```

### Change Web UI version
Deployment of the Web UI is managed by the Operator.

Initial deployment of the Operator via this ansible playbook automatically creates the Custom Resource for Web UI deployment.

To change the Web UI version,
- either `kubevirt_web_ui_version` playbook paramter can be set and the playbook re-executed
- or patch the Custom Resource, e.g.:
```
$ cat <<EOF
apiVersion: kubevirt.io/v1alpha1
kind: KWebUI
metadata:
  name: kubevirt-web-ui
spec:
  version: "1.4.0-9"
  registry_url: "quay.io"
  registry_namespace: "kubevirt"
  branding: "okdvirt"
EOF | oc apply -f -
```

To get Operator's CR:
```
$ oc get KWebUI kubevirt-web-ui -o yaml
```

Please note, the `status` section of the CR contains additional information about progress.
More detailed information cat be retrieved from the operator's pod log.

For further details, like undeploy of Web UI, please refer the [Web UI Operator homepage](https://github.com/kubevirt/web-ui-operator).
