#!/bin/bash

set -e

source hack/common.sh

prefix=${DOCKER_PREFIX:-kubevirt}
tag=${DOCKER_TAG:-v0.9.6}
kubeconfig=${KUBECONFIG:-~/.kube/config}
webdriver=${WEBDRIVER:-chromedriver}
[ -z "$OC_PATH" ] && OC_PATH=$(command -v oc)
[ -z "$KUBECTL_PATH" ] && KUBECTL_PATH=$(which kubectl)
[ -z "$VIRTCTL_PATH" ] && VIRTCTL_PATH=$(which virtctl)

yum install chromium -y
oc get route -n kubevirt-web-ui

# turn off exit on error, so below test can be run.
set +e

${TESTS_OUT_DIR}/tests.test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -oc-path=${OC_PATH} -kubectl-path=${KUBECTL_PATH} -virtctl-path=${VIRTCTL_PATH} ${FUNC_TEST_ARGS}

${TESTS_OUT_DIR}/ui.test -webDriver=$webdriver

rpm -qa | grep chrom

yum install chromium -y
oc get route -n kubevirt-web-ui
ping kubevirt-web-ui.cloudapps.example.com
nslookup kubevirt-web-ui.cloudapps.example.com
ifconfig

find . | grep png
mv *.png $ARTIFACTS_PATH/
