#!/bin/bash -ex

export OPENSHIFT_VER="3.7"
export ANSIBLE_MODULES_VERSION="openshift-ansible-3.7.29-1"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/byo/config.yml"
"${0%/*}/check-patch.sh"
