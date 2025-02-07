# Define build arguments with default values
ARG MONGODB_TOOLS_VERSION=100.11.0
ARG EN_AWS_CLI=false
ARG AWS_CLI_VERSION=1.29.44
ARG EN_AZURE=false
ARG AZURE_CLI_VERSION=2.67.0
ARG EN_GCLOUD=false
ARG GOOGLE_CLOUD_SDK_VERSION=499.0.0
ARG EN_GPG=true
ARG EN_MINIO=false
ARG EN_RCLONE=false
ARG VERSION

# Stage 1: tools-builder stage for MongoDB tools
FROM maxisam/mongo-tool:${MONGODB_TOOLS_VERSION} AS tools-builder

# Stage 2: mgob-builder stage for the mgob binary
FROM --platform=$BUILDPLATFORM golang:1.21 AS mgob-builder
ARG VERSION
ARG TARGETOS
ARG TARGETARCH

# Set environment variables for Go
ENV GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    CGO_ENABLED=0

# Set working directory
WORKDIR /go/src/github.com/stefanprodan/mgob

# Copy source code
COPY . .

# Build and test the mgob binary
RUN go test ./pkg/... && \
    go build -ldflags "-X main.version=$VERSION" -o mgob ./cmd/mgob

# Stage 3: final image setup with Alpine
FROM alpine:3.18

# Define build arguments
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
ARG MONGODB_TOOLS_VERSION
ARG AWS_CLI_VERSION
ARG AZURE_CLI_VERSION
ARG GOOGLE_CLOUD_SDK_VERSION
ARG EN_AWS_CLI
ARG EN_AZURE
ARG EN_GCLOUD
ARG EN_GPG
ARG EN_MINIO
ARG EN_RCLONE

# Set environment variables
ENV MONGODB_TOOLS_VERSION=${MONGODB_TOOLS_VERSION} \
    GOOGLE_CLOUD_SDK_VERSION=${GOOGLE_CLOUD_SDK_VERSION} \
    AZURE_CLI_VERSION=${AZURE_CLI_VERSION} \
    AWS_CLI_VERSION=${AWS_CLI_VERSION} \
    MGOB_EN_AWS_CLI=${EN_AWS_CLI} \
    MGOB_EN_AZURE=${EN_AZURE} \
    MGOB_EN_GCLOUD=${EN_GCLOUD} \
    MGOB_EN_GPG=${EN_GPG} \
    MGOB_EN_MINIO=${EN_MINIO} \
    MGOB_EN_RCLONE=${EN_RCLONE}

# Set working directory
WORKDIR /

# Copy and run the build script
COPY build.sh /tmp/
RUN chmod +x /tmp/build.sh && /tmp/build.sh

# Set the PATH for Google Cloud SDK
ENV PATH="/google-cloud-sdk/bin:${PATH}"

# Copy the mgob binary from the builder
COPY --from=mgob-builder /go/src/github.com/stefanprodan/mgob/mgob /usr/local/bin/
RUN chmod +x /usr/local/bin/mgob

# Copy MongoDB tools from the tools-builder
COPY --from=tools-builder /usr/local/bin/ /usr/bin/

# Install necessary runtime dependencies for MongoDB tools
RUN apk add --no-cache krb5-libs

# Volumes for storage
VOLUME ["/storage", "/tmp", "/data"]

# Labels for image metadata
LABEL org.label-schema.build-date=${BUILD_DATE} \
    org.label-schema.name="mgob" \
    org.label-schema.description="MongoDB backup automation tool" \
    org.label-schema.url="https://github.com/stefanprodan/mgob" \
    org.label-schema.vcs-ref=${VCS_REF} \
    org.label-schema.vcs-url="https://github.com/stefanprodan/mgob" \
    org.label-schema.vendor="stefanprodan.com,maxisam" \
    org.label-schema.version=${VERSION} \
    org.label-schema.schema-version="1.0"

# Entry point for the mgob application
ENTRYPOINT [ "/usr/local/bin/mgob" ]
