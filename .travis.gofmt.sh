#!/bin/bash

# Script inspired by: https://gist.github.com/ngdinhtoan/8c43fd92c379c4760735ec7186f215cf

# Only check go files (not vendors)
GO_DIR=$(go list -f '{{ .Dir }}' ./... | grep -v /vendor/ | grep -v /test/)

if [ -n "$(gofmt -s -l $GO_DIR )" ]; then
    echo "Go code is not formatted:"
    gofmt -l -s -d -e $GO_DIR
    exit 1
fi