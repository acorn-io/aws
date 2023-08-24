FROM cgr.dev/chainguard/go as build

WORKDIR /src/cleanup
COPY . ./
RUN go build -o cleanup

FROM cgr.dev/chainguard/wolfi-base

WORKDIR /app
COPY --from=build /src/cleanup/cleanup .

CMD ["/app/cleanup"]
