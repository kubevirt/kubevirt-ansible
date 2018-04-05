#!/bin/bash -ex

export OPENSHIFT_VERSION="3.7"
export ANSIBLE_MODULES_VERSION="release-3.7"
export OPENSHIFT_PLAYBOOK_PATH="playbooks/byo/config.yml"
"${0%/*}/check-patch.sh"
