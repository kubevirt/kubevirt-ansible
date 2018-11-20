#!/usr/bin/env bash

if [[ -n "$(gofmt -l ./tests)" ]] ; then
    echo "ERROR: It seems like you need to run 'make generate'. Please run it and commit the changes."
    gofmt -l ./tests
    exit 1
fi
