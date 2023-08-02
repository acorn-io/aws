FROM cgr.dev/chainguard/go as build

WORKDIR /src/s3
COPY --from=common . ../libs/
COPY --from=utils ./cfn-events ../cfn-events/
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    ls ../ &&\
    go build -o s3 . && \
    cd ../cfn-events && \
    go build -o cfn-events .

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
COPY --from=utils ./scripts/ ./scripts/
COPY --from=build /src/s3/s3 .
COPY --from=build /src/cfn-events/cfn-events .
CMD [ "/app/scripts/apply.sh" ]
