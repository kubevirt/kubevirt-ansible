#!/bin/bash

set -e

source hack/common.sh

prefix=docker.io/kubevirt
tag=v0.5.0-alpha.1
kubeconfig=admin.kubeconfig

${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -test.timeout 60m
