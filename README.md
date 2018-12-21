# KubeVirt Ansible

__Tools to provision resources, deploy clusters and install KubeVirt.__

## Overview

KubeVirt Ansible consists of a set of Ansible playbooks that deploy fully functional virtual machine management add-on for Kubernetes - KubeVirt. Optionally, a Kuberenetes or OpenShift cluster can also be configured.

## Contents

* `automation/`: CI scripts to verify the functionality of playbooks.
* `playbooks/`: Ansible playbooks to provision resources, deploy a cluster and install KubeVirt for various scenarios.
* `roles/`: Roles to use in playbooks.
* `vars/`: Variables to use in playbooks.
* `inventory`: A template for the cluster and nodes configuration.
* `requirements.yml`: A list of required Ansible-Galaxy roles to use in playbooks.
* `stdci.yaml`: A configuration file for CI system.

## Usage

### Deploy
To deploy KubeVirt on an existing OpenShift cluster run the command below. For more information on clusters and other deployment scenarious see [playbooks instructions](./playbooks/README.md).

```
ansible-playbook -i localhost playbooks/kubevirt.yml -e@vars/all.yml
```
>**Note:** Check default variables in [vars/all.yml](./vars/all.yml) and update them if needed.

### E2E Testing

1. Ensure it is possible to login into the cluster

```
oc login
```

2. Compile tests from the [tests](./tests) directory inside the docker container and copy it to the `kubevirt-ansible/_out` directory.
```
make build-tests
```

3. Run all the e2e tests with the `~/.kube/config` file

```
make test
```

If you'd like to run specific tests only, you can leverage `ginkgo`
command line options as follows (run a specified suite):

```
FUNC_TEST_ARGS='-ginkgo.focus=sanity_test -ginkgo.regexScansFilePath' make test
```


or you can pass it to tests via:
```
./_out/tests/<name>.test -kubeconfig=your_kubeconfig -tag=kubevirt_images_tag -prefix=kubevirt -test.timeout 60m
```

>**Note:** To test PVC's `storage.import.endpoint` with other images, use the `STREAM_IMAGE_URL` environment variable:
```
export STREAM_IMAGE_URL=<the_image_url>
```

## Questions ? Help ? Ideas ?

Stop by the [#kubevirt](https://webchat.freenode.net/?channels=kubevirt) chat channel on freenode IRC

## Contributing

Please see the [contributing guidelines](./CONTRIBUTING.md) for information regarding the contribution process.

## Automation & Testing

Please check the [CI automation guidelines](./automation/README.md) for information on playbooks verification.

# Useful Links
- [**KubeVirt**](https://github.com/kubevirt/kubevirt)
- [**OpenShift Ansible**][openshift-ansible-project]
- [**Golang Ansible playbook**](https://github.com/jlund/ansible-go)

[openshift-ansible-project]: https://github.com/openshift/openshift-ansible
