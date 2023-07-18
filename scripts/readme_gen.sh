#!/bin/bash

set -e

cd $(dirname $0)
pushd ../${1}

MD_FILE="README.md"

h1() {
  echo "# $@" > "${MD_FILE}"
}

h2() {
  echo >> "${MD_FILE}"
  echo "## $@" >> "${MD_FILE}"
}

code_content() {
    echo >> "${MD_FILE}"
    echo "\`\`\`" >> "${MD_FILE}"
    echo "$@" >> "${MD_FILE}"
    echo >> "${MD_FILE}"
    echo "\`\`\`" >> "${MD_FILE}"
}

content() {
    echo >> "${MD_FILE}"
    echo "\`\`\`" >> "${MD_FILE}"
    echo "$@" >> "${MD_FILE}"
    echo >> "${MD_FILE}"
    echo "\`\`\`" >> "${MD_FILE}"
}

h1 "$(basename $(pwd) |tr '[:lower:]' '[:upper:]')"

h2 "Args"

code_content "$(acorn run . --help 2>&1)"

h2 "Service Output"

content "$(acorn render --profile docs -o json . | jq '.services | to_entries | map(select(.key | startswith("generated-"))) | map({(.key | sub("generated-"; "")): .value}) | add')"

h2 "permissions"

code_content "$(acorn render --profile docs -o json .| jq 'if .jobs then (.jobs | to_entries[]| {(.key): .value.permissions}) else empty end, if .containers then (.containers | to_entries[]| {(.key): .value.permissions}) else empty end')"
