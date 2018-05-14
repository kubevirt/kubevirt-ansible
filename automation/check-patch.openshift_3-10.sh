#!/bin/bash -ex

export OPENSHIFT_VERSION="3.10"
export ANSIBLE_MODULES_VERSION="v3.10.0-rc.0"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/deploy_cluster.yml"
"${0%/*}/check-patch.sh" "$@"
