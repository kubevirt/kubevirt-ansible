#!/bin/bash -ex

export OPENSHIFT_VERSION="3.9"
export ANSIBLE_MODULES_VERSION="openshift-ansible-3.9.30-1"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/deploy_cluster.yml"
"${0%/*}/check-patch.sh" "$@"
