#!/bin/bash

set -e

source hack/common.sh

prefix=${DOCKER_PREFIX:-kubevirt}
tag=${DOCKER_TAG:-v0.17.0}
kubeconfig=${KUBECONFIG:-~/.kube/config}
oc_in_framework="oc"
virtctl_in_framework="virtctl"
[ -z "$OC_PATH" ] && OC_PATH=$(command -v oc)
[ -z "$KUBECTL_PATH" ] && KUBECTL_PATH=$(which kubectl)
[ -z "$VIRTCTL_PATH" ] && VIRTCTL_PATH=$(which virtctl)

OC_IN_FRAMEWORK=$oc_in_framework VIRTCTL_IN_FRAMEWORK=$virtctl_in_framework ${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -container-tag=$tag -container-prefix=$prefix -oc-path=${OC_PATH} -kubectl-path=${KUBECTL_PATH} -virtctl-path=${VIRTCTL_PATH} ${FUNC_TEST_ARGS}
