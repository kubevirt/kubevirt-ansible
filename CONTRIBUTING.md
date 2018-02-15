# Contributing to KubeVirt-Ansible

## Intro

In addition for allowing deployment of KubeVirt via Ansible,
the KubeVirt-Ansible project is open for any contributions from
Additional projects which aims to integrate with KubeVirt, such 
as a specific storage or network project.

The integration will be in the form of Ansible roles per component
which might later be integrated into a main Ansible Playbook Bundle.

Information on how to contribute will follow in the next sections.

## KubeVirt Ansible repository structure and roles

This section will describe the current repository structure and
give a short example of existing role for KubeVirt.

### Roles Structure

[TODO]: 
* Add info on role template and how to role for each supported flow ( K8S,OpenShift, W/ VMs )
* Add example from existing role for KubeVirt

### Contributing new Ansible roles  

Contributing to KubeVirt-Ansible should be as simple as possible. 
Have a question? Want to discuss something? Want to contribute something? Just open an
[Issue](https://github.com/kubevirt/kubevirt-ansible/issues) or a [Pull
Request](https://github.com/kubevirt/kubevirt-ansible/pulls).

## Functional Tests

So far we've been focused mostly on using Ansible to deploy various
projects such as KubeVirt on OpenShift or K8S. 

However, to really verify if the deployment actually worked, we have to run
misc functional tests which will verify the various features the deployed
project brought. 

Currently, the tests are part of the KubeVirt repo under [tests](https://github.com/kubevirt/kubevirt/tree/master/tests).

[TODO]:
* Add links or short examples of existing functional tests   

### Contribute functional tests

[TODO]:
* Add info on how to write new tests (or best practices)
* Add info on how to verify the new tests using CI
* Add link to developer guide showing how to run tests locally

## Getting your code reviewed/merged

In order to merge your code, it needs to pass a maintainer review and CI.
The automation in CI will deploy your new code on currently
Supported platform matix (OS,OpenShift,Ansible,etc.. ) and potentially
run addition functional tests on the deployed components.

KubeVirt-Ansible maintainers are:

 * @lukas-bednar
 * @gbenhaim
 * @cynepco3hahue

In addition, the new code should be merged with basic documentation
on the new project integrated in the form of:

* Project links
* Short description and purpose
* Maintainer or POC

[TODO]:
* Add info on how to contribute documentation and where

## Integrating into existing KubeVirt APB

Currently, the deployment of KubeVirt ( and possibly other projects )
is done via Ansible roles and playbook. 
However, we also want to build a main APB which will deploy all projects.
This section contains info on how to integrate your project into this APB.

The current APB code can be found here [KubeVirt-APB](https://github.com/ansibleplaybookbundle/kubevirt-apb)

Current focal point for the APB:

 * @rthallisey


## Additional Links

* Ansible
  * [Official Product Page](https://ansible.com/)
  * [Best Practices](http://docs.ansible.com/ansible/latest/playbooks_best_practices.html)
* OpenShift
  * [Official Product Page](https://openshift.org/)
  * [OpenShift Ansible](https://github.com/openshift/openshift-ansible)
  * [Ansible Playbook Bundle](https://docs.openshift.org/latest/apb_devel/writing/reference.html)
* KubeVirt
  * [KubeVirt GitHub](https://github.com/kubevirt/)
