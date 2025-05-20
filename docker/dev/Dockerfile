# Stage 1: Builder stage
FROM golang:1.24-alpine AS builder

ENV PROJECT_DIR=/app \
    GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=0

WORKDIR /app

# Install only build dependencies
RUN apk update && \
    apk add --no-cache git curl && \
    rm -rf /var/cache/apk/*

# Copy Go module files and download dependencies
COPY go.mod ./
RUN go mod tidy && go mod download

# Copy source code and build the application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./src

# Install Air for development
RUN go install github.com/air-verse/air@latest

# Stage 2: SSL Certificate generation
FROM alpine:3.21 AS ssl-cert

RUN apk add --no-cache openssl

WORKDIR /certs

# Generate SSL certificates in a single layer
RUN mkdir -p ./ssl && \
    openssl genrsa -out ./ssl/tls.key 2048 && \
    openssl req -new -x509 -key ./ssl/tls.key -out ./ssl/tls.crt -days 365 -subj "/CN=vgn-server.local" && \
    openssl genrsa -out ./ssl/tls.client.key 2048 && \
    openssl req -new -key ./ssl/tls.client.key -subj "/CN=vgn-server.local" -out ./ssl/tls.client.csr && \
    openssl x509 -req -in ./ssl/tls.client.csr -CA ./ssl/tls.crt -CAkey ./ssl/tls.key -CAcreateserial -out ./ssl/tls.client.crt -days 365

# Stage 3: Runtime stage
FROM golang:1.24-alpine

# Install runtime dependencies only
RUN apk update && \
    apk add --no-cache bash curl jq ca-certificates && \
    rm -rf /var/cache/apk/*

# Install kubectl for runtime operations
ENV KUBECTL_VERSION=1.32.0
RUN curl -LO "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

# Install k3d for runtime operations
ENV K3D_VERSION=5.8.3
RUN curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v${K3D_VERSION} bash

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app

# Copy K8s configuration files (optimized to avoid duplication)
COPY ./k8/k3d-config.yaml /etc/rancher/k3d/config.yaml
COPY ./k8 /etc/rancher/k3d/

# Copy certificates from SSL stage
COPY --from=ssl-cert /certs/ssl /app/bin/ssl

# Copy compiled application and Air from builder stage
COPY --from=builder /app/app /app/
COPY --from=builder /go/bin/air /usr/local/bin/air

# Add application source for Air hot-reloading in dev
COPY . .

# Set ownership for security
RUN chown -R appuser:appgroup /app

# Switch to non-root user
# USER appuser

# Expose application port
EXPOSE 3000

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
CMD curl -f http://localhost:3000/health || exit 1

# Run Air for development
CMD ["air"]
