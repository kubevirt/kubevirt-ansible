#!/bin/bash

set -e

source hack/common.sh

prefix=${DOCKER_PREFIX:-kubevirt}
tag=${DOCKER_TAG:-v0.7.0}
kubeconfig=${KUBECONFIG:-~/.kube/config}

${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix ${FUNC_TEST_ARGS}
${TESTS_OUT_DIR}/cdi.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix ${CDI_TEST_ARGS}
