#!/bin/bash

KUBEVIRT_ANSIBLE_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../"
    pwd
)"
OUT_DIR=$KUBEVIRT_ANSIBLE_DIR/_out
VENDOR_DIR=$KUBEVIRT_ANSIBLE_DIR/vendor
TESTS_OUT_DIR=$OUT_DIR/tests
FUNC_TEST_ARGS="${FUNC_TEST_ARGS:--test.timeout 60m --junit-output=exported-artifacts/tests.junit.xml -ginkgo.seed 1541336879}"

function build_func_tests() {
    mkdir -p ${TESTS_OUT_DIR}/
    ginkgo build ${KUBEVIRT_ANSIBLE_DIR}/tests
    mv ${KUBEVIRT_ANSIBLE_DIR}/tests/tests.test ${TESTS_OUT_DIR}/
}
