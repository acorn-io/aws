#!/bin/sh

# Wait for IAM to propagate
sleep 10

while {
    sleep 8
    /src/consume-message -q ${QUEUE_NAME}
}; do :; done