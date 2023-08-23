FROM cgr.dev/chainguard/go as build

WORKDIR /src/sqs
COPY --from=common . ../libs/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    ls ../ &&\
    go build -o sqs

FROM ghcr.io/acorn-io/aws/utils/cdk-runner:v0.4.0 as cdk-runner
FROM cgr.dev/chainguard/wolfi-base
RUN apk add -U --no-cache nodejs bash busybox aws-cli jq curl zip && \
    apk del --no-cache wolfi-base apk-tools
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install
RUN npm install -g aws-cdk
WORKDIR /app
COPY ./cdk.json ./
COPY ./scripts ./scripts
COPY --from=build /src/sqs/sqs .
COPY --from=utils . ./scripts/ ./scripts/
COPY --from=cdk-runner /cdk-runner .
CMD [ "/app/cdk-runner" ]
