#!/bin/sh

# Wait for IAM to propagate
sleep 30

while {
    sleep 4
    /src/consume-message -q ${QUEUE_NAME}
}; do :; done