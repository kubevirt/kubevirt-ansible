# Contributing to KubeVirt Ansible

Thank you for interest in contributing to KubeVirt Ansible! :tada::+1:

The following is a set of guidelines for contributing to this project. These are not rules. Use your best judgment, and feel free to propose changes to this document in a pull request. Please make sure you are welcoming and friendly in all of our spaces.

## Questions ? Help ? Ideas ? :bulb:

Please, don't use the issue tracker for support questions. Stop by the [#kubevirt](https://webchat.freenode.net/?channels=kubevirt) chat channel on freenode IRC or contact us on the [KubeVirt Google Group](https://groups.google.com/forum/#!forum/kubevirt-dev).

## Reporting Bugs :bug:

Before creating bug reports, please check a [list of known issues](https://github.com/kubevirt/kubevirt-ansible/issues) to see if the problem has already been reported.

If not, go ahead and [make one](https://github.com/kubevirt/kubevirt-ansible/issues/new)! Be sure to include a **descriptive title and clear description** and please include **as many details as possible** to help maintainers reproduce the issue and resolve it faster. If possible, add a **code sample** or an **executable test case** demonstrating the expected behavior that is not occurring.

> **Note:** If you find a **Closed** issue that seems like it is the same thing that you're experiencing, open a new issue and include a link to the original issue in the body of your new one.

## Suggesting Enhancements :hatched_chick:

Enhancement suggestions are tracked as [GitHub issues](https://github.com/kubevirt/kubevirt-ansible/issues). When you are creating an enhancement issue, **use a clear and descriptive title** and **provide a clear description of the suggested enhancement** in as many details as possible.

## Submitting changes :wrench:

To submit changes, please send a [GitHub Pull Request](https://github.com/kubevirt/kubevirt-ansible/pulls). 
* Before submitting a PR, please **read the [Styleguides](#styleguides)** to know more about coding conventions used in this project.
* Always **fork** [KubeVirt Ansible](https://github.com/kubevirt/kubevirt-ansible) and **create a new branch with a descriptive name for each pull request** to avoid intertwingling different features or fixes on the same branch.
* Always **do "git pull --rebase" and "git rebase"** vs "git pull" or "git merge".
* Ensure the PR description clearly describes the problem and solution. Include the relevant issue number if applicable.

Before merging, a PR needs to pass the review of two [maintainers'](#kubevirt-ansible-contributors) and CI tests. The PR should be up to date with the current master branch and has no requested changes.

### Ansible playbooks changes

In addition for allowing deployment of KubeVirt via Ansible, the KubeVirt-Ansible project is open for any contributions from additional projects which aims to integrate with KubeVirt, such as a specific storage or network project. The integration will be in the form of Ansible roles per component which might later be integrated into a main [Ansible Playbook Bundle (APB)](https://github.com/ansibleplaybookbundle/kubevirt-apb).

### Functional Tests changes

So far we've been focused mostly on using Ansible to deploy various projects such as KubeVirt on OpenShift or Kubernetes (K8S). However, to really verify if the deployment actually worked, we have to run misc functional tests which will verify the various features the deployed project brought. Currently, the tests are part of the KubeVirt repo under [tests](https://github.com/kubevirt/kubevirt/tree/master/tests).

## Styleguides

### Git Commit Messages

  * Always write a clear log message for your commits
  * Use the present tense ("Add feature" not "Added feature")
  * Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
  * Limit the first line to 72 characters or less

      ```
      $ git commit -m "A brief summary of the commit
      >
      > A paragraph describing what changed and its impact."
      ```

  * Reference issues and pull requests as `Fix issue 13`

### Codding guidelines

#### Comments

  * Readability is one of the most important goals for this project
  * Comment any non-trivial code where someone might not know why you are doing something in a particular way
  * Commenting above a line is preferable to commenting at the end of a line

#### Variables

  * Use descriptive variable names instead of variables like 'x, y, a, b, foo, boo'

#### Documentation

  * Please ensure all code changes reflected in documentation accordingly.
  
For me codding guidelines please see [Additional Links](#additional-links).

### KubeVirt Ansible Contributors

To see the list of KubeVirt Ansible Contributors, please

* Check the [Contributors](https://github.com/kubevirt/kubevirt-ansible/graphs/contributors) page or
* Run `git shortlog -sne` in a cloned [KubeVirt Ansible](https://github.com/kubevirt/kubevirt-ansible) git repository.


Thank you! :heart: :heart: :heart:

KubeVirt Team

## Additional Links

* Ansible
  * [Official Product Page](https://ansible.com/)
  * [Best Practices](http://docs.ansible.com/ansible/latest/playbooks_best_practices.html)
  * [Ansible Best Practices: The Essentials](https://www.ansible.com/blog/ansible-best-practices-essentials)
* OpenShift
  * [Official Product Page](https://openshift.org/)
  * [OpenShift Ansible](https://github.com/openshift/openshift-ansible)
  * [Ansible Playbook Bundle](https://docs.openshift.org/latest/apb_devel/writing/reference.html)
* KubeVirt
  * [KubeVirt GitHub](https://github.com/kubevirt/)
