FROM golang:1.22.8-alpine3.20 AS builder

# Build arguments
ARG MONGODB_TOOLS_VERSION=100.11.0
ARG TARGETOS
ARG TARGETARCH

# Set environment variables for Go build
ENV GOCACHE="/go-cache" \
    GOTMPDIR="/tmp" \
    GOMODCACHE="/go/pkg/mod"

# Install dependencies and clone the repository
RUN apk add --no-cache git build-base krb5-dev && \
    git clone https://github.com/mongodb/mongo-tools.git /src/mongo-tools --branch $MONGODB_TOOLS_VERSION --depth 1

WORKDIR /src/mongo-tools

# Build MongoDB tools
RUN mkdir -p bin && \
    for bin in bsondump mongodump mongoexport mongofiles mongoimport mongorestore mongostat mongotop; do \
    OOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/$bin \
    -ldflags="-X main.VersionStr=$MONGODB_TOOLS_VERSION -X main.GitCommit=alpine" \
    -tags "ssl sasl gssapi failpoints" \
    $bin/main/$bin.go; \
    done

# You might want to use a smaller image for the final stage
FROM alpine:3.20

# Copy the built binaries from the builder
COPY --from=builder /src/mongo-tools/bin/* /usr/local/bin/

# Ensure the binaries are executable
RUN chmod +x /usr/local/bin/*
