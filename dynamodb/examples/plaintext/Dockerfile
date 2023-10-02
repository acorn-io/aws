FROM cgr.dev/chainguard/go as build

WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o plaintext-ddb .

FROM cgr.dev/chainguard/wolfi-base

RUN apk add curl zip

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install

WORKDIR /app
COPY --from=build /src/plaintext-ddb .

CMD [ "/app/plaintext-ddb" ]
