#!/bin/bash

prefix=docker.io/kubevirt
tag=v0.5.0-alpha.1
kubeconfig=/etc/origin/master/admin.kubeconfig

go test -kubeconfig=$kubeconfig -tag=$tag -prefix=$prefix -test.timeout 60m
