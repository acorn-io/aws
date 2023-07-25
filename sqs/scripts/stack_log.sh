#!/bin/bash

SEEN=""
while sleep 5; do
    while read TS REST ; do
        if ! grep -q $TS; then
            SEEN="$SEEN $TS"
            echo $TS $REST
        fi <<< "$SEEN"
    done < <(aws cloudformation describe-stack-events --stack-name $1 | jq -r '.StackEvents[] | [ .Timestamp, .ResourceStatus, .ResourceType, .ResourceStatusReason, .LogicalResourceId, .PhysicalResourceId] | join(" ")' | sort)
done
