FROM golang:1.19-alpine3.15 as mongo-tools-builder

ARG MONGODB_TOOLS_VERSION=100.8.0

RUN apk add --no-cache git build-base krb5-dev && \
    git clone https://github.com/mongodb/mongo-tools.git --depth 1 -b $MONGODB_TOOLS_VERSION

WORKDIR mongo-tools
RUN ./make build
