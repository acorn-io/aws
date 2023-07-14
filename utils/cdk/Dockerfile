FROM cgr.dev/chainguard/go as build
WORKDIR /src
COPY . ./
RUN go build -o cdk-context-render

FROM cgr.dev/chainguard/glibc-dynamic as app
COPY --from=build /src/cdk-context-render /app/
ENV CONFIG_FILE "/app/config.json"
ENV SERVICE_OUTFILE "/run/secrets/output"
ENV CDK_CONTEXT_OUTFILE "/dev/null"
USER 1000
WORKDIR "/app"
CMD ["/app/cdk-context-render"]
