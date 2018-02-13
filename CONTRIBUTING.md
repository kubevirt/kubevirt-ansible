## Contributing to KubeVirt-Ansible

### Intro

In addition for allowing deployment of KubeVirt via Ansible,
the KubeVirt-Ansible project is open for any contributions from
Additional projects which aims to integrate with KubeVirt, such 
as specific storage or network project.

The integration will be in the form of Ansible roles per component
which might later be integrated into a main Ansible Playbook Bundle.o

Information on how to contribute will follow in the next sections.

### Contributing new Ansible roles  

Contributing to KubeVirt-Ansible should be as simple as possible. 
Have a question? Want to discuss something? Want to contribute something? Just open an
[Issue](https://github.com/kubevirt/kubevirt-ansible/issues) or a [Pull
Request](https://github.com/kubevirt/kubevirt-ansible/pulls).

## Roles Structure

[TODO]: 
* Add info on role template and how to role for each supported flow ( K8S,OpenShift, W/ VMs )
* Add example from existing role for KubeVirt

### Running deployment and tests locally on your laptop

[TODO]:
* Add info on how to run lago locally to test deployments and run tests
* Add info on how to use Vagrant to test deployments locally
* Add link to supported platform versions

### Contribute functional tests

In addition for contributing roles for deploying a new component,
Its important also to verify the deployment actually works and 
run various tests on the deployed environment to verify its functional.

[TODO]:
* Add info on existing tests location and how to run them 
* Add info on how to use common libraries in order to write new test
* Add info on how to verify the new tests using CI

### Getting your code reviewed/merged

In order to merge your code, it needs to pass a maintainer review and CI.
The automation in CI will deploy your new code on currently
Supported platform matix (OS,OpenShift,Ansible,etc.. ) and potentially
run addition functional tests on the deployed components.

Maintainers are:

 * @lukas-bednar
 * @gbenhaim

### Integrating into existing KubeVirt APB

Currently, the deployment of KubeVirt ( and possibly other projects )
is done via Ansible roles and playbook. 
However, we also want to build a main APB which will deploy all projects.
This section contains info on how to integrate your project into this APB.

[TODO]:
* Add info on current APB and KubeVirt deployment
* Add info on how to add new projects

### Related Projects & Communities

# [KubeVirt](https://github.com/kubevirt/)

### Additional Topics

* Golang
  * [Documentation - The Go Programming Language](https://golang.org/doc/)
  * [Getting Started - The Go Programming Language](https://golang.org/doc/install)
* Patterns
  * [Introducing Operators: Putting Operational Knowledge into Software](https://coreos.com/blog/introducing-operators.html)
  * [Microservices](https://martinfowler.com/articles/microservices.html) nice
    content by Martin Fowler
* Testing
  * [Ginkgo - A Golang BDD Testing Framework](https://onsi.github.io/ginkgo/)
