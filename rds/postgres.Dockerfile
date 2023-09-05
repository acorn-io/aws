FROM cgr.dev/chainguard/go as build
ARG MAIN
WORKDIR /src/rds
COPY --from=common . ../libs/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o rds ./aurora/postgres/${MAIN}

FROM cgr.dev/chainguard/postgres as user

WORKDIR /app
COPY ./scripts ./scripts
ENTRYPOINT ["/app/scripts/create_and_grant_users_psql.sh"]

FROM ghcr.io/acorn-io/aws/utils/cdk-runner:v0.6.0 as cdk-runner
FROM cgr.dev/chainguard/wolfi-base
RUN apk add -U --no-cache nodejs bash busybox jq curl zip && \
    apk del --no-cache wolfi-base apk-tools
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
     unzip awscliv2.zip && \
     ./aws/install
RUN npm install -g aws-cdk
WORKDIR /app
COPY ./cdk.json ./
COPY ./scripts ./scripts
COPY --from=cdk-runner /cdk-runner .
COPY --from=build /src/rds/rds .
CMD [ "/app/cdk-runner" ]
