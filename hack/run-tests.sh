#!/bin/bash

set -e

source hack/common.sh

prefix=kubevirt
tag=v0.7.0
kubeconfig=~/.kube/config

${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -test.timeout 60m ${FUNC_TEST_ARGS}
${TESTS_OUT_DIR}/cdi.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -test.timeout 60m ${FUNC_TEST_ARGS}
