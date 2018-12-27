#!/bin/bash -ex

export OPENSHIFT_VERSION="3.11"
export ANSIBLE_MODULES_VERSION="openshift-ansible-3.11.86-1"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/deploy_cluster.yml"
export OPENSHIFT_ANSIBLE_URL="https://github.com/openshift/openshift-ansible.git"
"${0%/*}/check-patch.sh" "$@"
