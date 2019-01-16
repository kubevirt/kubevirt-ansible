#!/bin/bash

set -e

source hack/common.sh

prefix=${DOCKER_PREFIX:-kubevirt}
tag=${DOCKER_TAG:-v0.13.0}
kubeconfig=${KUBECONFIG:-~/.kube/config}
[ -z "$OC_PATH" ] && OC_PATH=$(command -v oc)
[ -z "$KUBECTL_PATH" ] && KUBECTL_PATH=$(which kubectl)
[ -z "$VIRTCTL_PATH" ] && VIRTCTL_PATH=$(which virtctl)

${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -oc-path=${OC_PATH} -kubectl-path=${KUBECTL_PATH} -virtctl-path=${VIRTCTL_PATH} ${FUNC_TEST_ARGS}
