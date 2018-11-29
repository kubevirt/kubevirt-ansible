#!/bin/bash -ex

export OPENSHIFT_VERSION="3.10"
export ANSIBLE_MODULES_VERSION="openshift-ansible-3.10.41-1"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/deploy_cluster.yml"
"${0%/*}/check-patch.sh" "$@"
