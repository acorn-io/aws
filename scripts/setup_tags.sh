#!/bin/bash

committed_tag=${GITHUB_REF#refs/*/}

image=$(dirname "${committed_tag}")
tag=$(basename "${committed_tag}")

echo "IMAGE=${image}" >> $GITHUB_ENV
echo "TAG=${tag}" >> $GITHUB_ENV
