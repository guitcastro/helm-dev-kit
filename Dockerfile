# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o helm-dev-kit \
    ./cmd/helm-dev-kit

# Final stage
FROM scratch

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy SSL certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/helm-dev-kit /usr/local/bin/helm-dev-kit

# Create non-root user
USER 65534:65534

# Set entrypoint
ENTRYPOINT ["helm-dev-kit"]

# Default command
CMD ["--help"]

# Metadata
LABEL maintainer="Guilherme Castro <guitcastro@example.com>"
LABEL description="Convert HCL configurations to Helm charts"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/guitcastro/helm-dev-kit"
LABEL org.opencontainers.image.documentation="https://github.com/guitcastro/helm-dev-kit/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"