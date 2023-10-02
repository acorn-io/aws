# Text Based K/V Storage

Just a simple test service using the DynamoDB service acorn.

## Usage

1) Run the Acorn `acorn run .`
2) Put in some text `curl ${ACORN_URL}/put\?key\=test -d "hello world"`
3) Retrieve the text `curl ${ACORN_URL}/get\?key\=test`
