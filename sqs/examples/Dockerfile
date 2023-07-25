FROM cgr.dev/chainguard/go as build

WORKDIR /src/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    ls ../ &&\
    go build -o send-message

CMD ["/send-message", "-q", "${QUEUE_NAME}"]