#!/bin/bash

function record_event() {
	# Consume expected environment variables and files
	local sa="/var/run/secrets/kubernetes.io/serviceaccount"
	local token=$(cat "${sa}/token")
	local ca_cert="${sa}/ca.crt"
	local apiserver="https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}"
	local acorn_name="${ACORN_NAME}"
	local project="${ACORN_PROJECT}"

	# Required arguments
	local event_type
	local description

	# Optional arguments
	local details

	# Options
	local observed
	local severity="info"

	function display_usage() {
		echo "Usage: process_event [--observed <argument>] [--severity <argument>] <event-type> <description> [details]"
		echo "Options:"
		echo "  --observed <argument>    Set the observation time."
		echo "  --severity <argument>    Set the severity, one of info|error."
	}

	# Parse options using getopts
	while [[ $# -gt 0 ]]; do
		case "$1" in
		--observed)
			observed="$2"
			shift 2
			;;
		--severity)
			severity="$2"
			shift 2
			;;
		*)
			break # Stop processing options and move to positional arguments
			;;
		esac
	done

	# Ensure and assign required positional arguments
	if [ $# -lt 2 ]; then
		echo "Error: Two positional arguments are required (event-type and description)."
		display_usage
		exit 1
	fi

	event_type="$1"
	description="$2"

	# Assign optional arguments if given
	local details
	if [ $# -gt 2 ]; then
		shift 2
		details=$(cat)
	fi

	# Prep event payload
	local data=$(cat <<EOF
{
  "apiVersion": "api.acorn.io/v1",
  "kind": "Event",
  "metadata": {
    "generateName": "se-"
  },
  "type": "$event_type",
  "appName": "$acorn_name",
  "description": "$description",
  "severity": "$severity",
  "resource": {
    "kind": "app",
    "name": "$acorn_name"
  }
}
EOF
)

	# Add observed if present
	if [[ -n "$observed" ]]; then
		data=$(jq --argjson observed "$observed" '.observed = $observed' <<<"$data")
	fi

	# Add details if present
	if [[ -n "$details" ]]; then
		data=$(jq --argjson details "$details" '.details = $details' <<<"$data")
	fi

	# POST event
	curl -s -o /dev/null --cacert "${ca_cert}" \
		-H "Authorization: Bearer ${token}" \
		-H "Accept: application/json" \
		-H "Content-Type: application/json" \
		--data "${data}" \
		"${apiserver}/apis/api.acorn.io/v1/namespaces/${project}/events"
}
