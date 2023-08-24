FROM cgr.dev/chainguard/go as build

WORKDIR /src/s3
COPY --from=common . ../libs/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    ls ../ &&\
    go build -o s3 .

FROM ghcr.io/acorn-io/aws/utils/cdk-runner:v0.5.0 as cdk-runner
FROM cgr.dev/chainguard/wolfi-base
RUN apk add -U --no-cache nodejs bash busybox jq curl zip && \
    apk del --no-cache wolfi-base apk-tools
RUN npm install -g aws-cdk
WORKDIR /app
COPY ./cdk.json ./
COPY ./scripts ./scripts
COPY --from=utils ./scripts/ ./scripts/
COPY --from=cdk-runner /cdk-runner .
COPY --from=build /src/s3/s3 .
CMD [ "/app/cdk-runner" ]
