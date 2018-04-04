#!/bin/bash -ex

export OPENSHIFT_VERSION="3.9"
export ANSIBLE_MODULES_VERSION="release-3.9"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/deploy_cluster.yml"
"${0%/*}/check-patch.sh"
