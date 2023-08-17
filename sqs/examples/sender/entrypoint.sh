#!/bin/sh

# Wait for IAM to propagate
sleep 10

while {
    sleep 5
    /src/send-message -q ${QUEUE_NAME}
}; do :; done