FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache bash aws-cli jq && \
    apk del --no-cache --purge wolfi-base apk-tools 
ADD . /app/
WORKDIR /app

ENTRYPOINT [ "/app/render.sh" ]