FROM cgr.dev/chainguard/go as build
WORKDIR /src/cdk-runner
COPY . ./
RUN go build -o cdk-runner
CMD ["/app/cdk-runner"]
