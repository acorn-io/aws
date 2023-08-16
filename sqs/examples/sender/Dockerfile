FROM cgr.dev/chainguard/go as build

WORKDIR /src/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o send-message .

ENTRYPOINT ["/src/entrypoint.sh"]