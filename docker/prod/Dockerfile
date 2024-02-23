# Stage 1: Build the Go binary
FROM golang:1.22rc2-alpine3.19 AS Base

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
COPY go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Tidy up the go.mod file
RUN go mod tidy

# Build the Go binary with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./src

# Stage 2: Create the final lightweight image
FROM alpine:3.19

# Update the package lists and install required packages
RUN apk update && \
    apk add --no-cache curl sudo ca-certificates && \
    rm -rf /var/cache/apk/*

# Install kubectl
ENV KUBECTL_VERSION=1.23.3
RUN curl -LO "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \ 
    && chmod +x ./kubectl \ 
    && mv ./kubectl /usr/local/bin/kubectl

# Set the working directory
WORKDIR /var/www/html

# Copy the built Go binary from the previous stage
COPY --from=Base /app/app .

# Expose the port that the application listens on
EXPOSE 3000

# Command to run the executable
CMD ["/var/www/html/app"]
