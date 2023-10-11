FROM cgr.dev/chainguard/go as build

WORKDIR /src/iam/role
COPY --from=common . ../../libs/
COPY . .

RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o role .

FROM ghcr.io/acorn-io/aws/utils/cdk-runner:v0.6.0 as cdk-runner
FROM cgr.dev/chainguard/wolfi-base
RUN apk add -U --no-cache nodejs bash busybox jq && \
    apk del --no-cache wolfi-base apk-tools
RUN npm install -g aws-cdk
WORKDIR /app
COPY ./cdk.json ./
COPY ./scripts ./scripts
COPY --from=utils ./scripts/ ./scripts/
COPY --from=cdk-runner /cdk-runner .
COPY --from=build /src/iam/role/role .

ENV GOGC="25"
ENV NODE_OPTIONS="--max-old-space-size=256"

CMD [ "/app/cdk-runner" ]