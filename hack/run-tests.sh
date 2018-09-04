#!/bin/bash

set -e

source hack/common.sh

prefix=${DOCKER_PREFIX:-kubevirt}
tag=${DOCKER_TAG:-v0.8.0-alpha.1}
kubeconfig=${KUBECONFIG:-~/.kube/config}
OC_PATH=$(command -v oc)
KUBECTL_PATH=$(which kubectl)
VIRTCTL_PATH=$(which virtctl)

${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -oc-path=${OC_PATH} -kubectl-path=${KUBECTL_PATH} -virtctl-path=${VIRTCTL_PATH} ${FUNC_TEST_ARGS}
${TESTS_OUT_DIR}/cdi.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -oc-path=${OC_PATH} -kubectl-path=${KUBECTL_PATH} -virtctl-path=${VIRTCTL_PATH} ${CDI_TEST_ARGS}
