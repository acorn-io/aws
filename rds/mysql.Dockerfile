FROM cgr.dev/chainguard/go as build
ARG MAIN
WORKDIR /src/rds
COPY --from=common . ../libs/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o rds ./aurora/mysql/${MAIN}

FROM cgr.dev/chainguard/wolfi-base as dependencies

RUN apk add -U --no-cache mariadb-connector-c

FROM cgr.dev/chainguard/mariadb as user

WORKDIR /app
COPY ./scripts ./scripts
COPY --from=dependencies /usr/lib/mariadb/plugin/caching_sha2_password.so /usr/lib/mariadb-10.11/plugin/caching_sha2_password.so
ENTRYPOINT ["/app/scripts/create_and_grant_users.sh"]

FROM ghcr.io/acorn-io/aws/utils/cdk-runner:v0.7.1 as cdk-runner
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
COPY ./hooks ./hooks
COPY --from=cdk-runner /cdk-runner .
COPY --from=build /src/rds/rds .

ENV GOGC="25"
ENV NODE_OPTIONS="--max-old-space-size=256"

CMD [ "/app/cdk-runner" ]
