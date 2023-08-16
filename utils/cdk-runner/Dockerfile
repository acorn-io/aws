FROM cgr.dev/chainguard/go as build
WORKDIR /src/cdk-runner
COPY . ./
RUN go build -o cdk-runner

## Small binary delivery image
FROM scratch as bin
COPY --from=build /src/cdk-runner/cdk-runner /cdk-runner
CMD ["/cdk-runner"]
