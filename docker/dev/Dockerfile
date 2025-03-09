# Stage 1: Build the ENVironment
FROM golang:1.23.0-alpine3.19 AS base

ENV PROJECT_DIR=/app \
    GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=0

# Set the working directory
WORKDIR /tmp
# Update the package lists and install required packages
RUN apk update && \
    apk add --no-cache curl bash git jq sudo docker openssl ca-certificates && \
    rm -rf /var/cache/apk/*
# Install kubectl
ENV KUBECTL_VERSION=1.32.0
RUN curl -LO "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \ 
    && chmod +x ./kubectl \ 
    && mv ./kubectl /usr/local/bin/kubectl
# Install k3d
ENV K3D_VERSION=5.8.3
RUN curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v${K3D_VERSION} bash
# Set up k3d configuration
COPY ./k8/k3d-config.yaml /etc/rancher/k3d/config.yaml
# Copy the build script
COPY ./k8 /etc/rancher/k3d
COPY ./k8 /etc/rancher/k3s
# Make the script executable
#RUN chmod +x /etc/rancher/k3d/k3d-build.sh
#ENV KUBECONFIG=/etc/rancher/k3d/kubeconfig.yaml
# Run the script
#CMD ["/etc/rancher/k3d/k3d-build.sh"]

# Set the working directory
WORKDIR /app

# Generate ssl certificates
RUN mkdir -p ./bin/ssl
RUN openssl genrsa -out ./bin/ssl/vgn.server.key 2048
RUN openssl req -new -x509 -key ./bin/ssl/vgn.server.key -out ./bin/ssl/vgn.server.crt -days 365 -subj "/CN=vgn-server.local"

RUN openssl genrsa -out ./bin/ssl/vgn.client.key 2048
RUN openssl req -new -key ./bin/ssl/vgn.client.key -subj "/CN=my-client" -out ./bin/ssl/vgn.client.csr
RUN openssl x509 -req -in ./bin/ssl/vgn.client.csr -CA ./bin/ssl/vgn.server.crt -CAkey ./bin/ssl/vgn.server.key -CAcreateserial -out ./bin/ssl/vgn.client.crt -days 365

# Copy the Go module files and download dependencies
COPY go.mod ./
# Download the dependencies
RUN go mod tidy
# Download the dependencies
RUN go mod download
# Install the Air tool
RUN go get -u  github.com/air-verse/air
RUN go install github.com/air-verse/air
# Expose the port that the application listens on
EXPOSE 3000

# Copy the rest of the application source code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./src

# 
CMD ["air"]