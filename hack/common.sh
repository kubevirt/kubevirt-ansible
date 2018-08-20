#!/bin/bash

KUBEVIRT_ANSIBLE_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../"
    pwd
)"
OUT_DIR=$KUBEVIRT_ANSIBLE_DIR/_out
TESTS_OUT_DIR=$OUT_DIR/tests

FUNC_TEST_ARGS="${FUNC_TEST_ARGS:--test.timeout 60m --junit-output=junit-func-tests.xml}"
CDI_TEST_ARGS="${CDI_TEST_ARGS:--test.timeout 60m --junit-output=junit-cdi-tests.xml}"

function build_func_tests() {
    mkdir -p ${TESTS_OUT_DIR}/
    ginkgo build ${KUBEVIRT_ANSIBLE_DIR}/tests
    ginkgo build ${KUBEVIRT_ANSIBLE_DIR}/tests/cdi
    mv ${KUBEVIRT_ANSIBLE_DIR}/tests/tests.test ${TESTS_OUT_DIR}/
    mv ${KUBEVIRT_ANSIBLE_DIR}/tests/cdi/cdi.test ${TESTS_OUT_DIR}/
}
