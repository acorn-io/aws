FROM alpine:latest

RUN apk add curl zip postgresql15-client

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install

CMD [ "tail", "-f", "/dev/null" ]
