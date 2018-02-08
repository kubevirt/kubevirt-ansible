#!/bin/bash -ex

export OPENSHIFT_IMAGE_TAG="v3.9.0"
export ANSIBLE_MODULES_VERSION="openshift-ansible-3.9.0-0.40.0"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/deploy_cluster.yml"
"${0%/*}/check-patch.sh"
