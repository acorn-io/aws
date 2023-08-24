FROM cgr.dev/chainguard/go as build

WORKDIR /src/s3-cleanup
COPY . ./
RUN go build -o s3-cleanup

FROM cgr.dev/chainguard/wolfi-base

WORKDIR /app
COPY --from=build /src/s3-cleanup/s3-cleanup .

CMD ["/app/s3-cleanup"]
