# Build the admin assets from the current source instead of shipping stale bundles.
FROM node:22-alpine AS assets
WORKDIR /app/assets
COPY assets/package.json assets/package-lock.json ./
RUN npm ci
COPY assets/ ./
COPY public/ /app/public/
RUN npm run build

# Stage 1: Build a static Go binary
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install git for module fetching (no C-Go code)
RUN apk add --no-cache git                                    

# Cache modules, then compile with CGO disabled for a static binary
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/app ./src

# Stage 2: Runtime with only the binary
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/app /app/app
# Copy the public folders
COPY --from=builder /app/views ./views
COPY --from=assets /app/public ./public

# Use an unprivileged user (65532 is “nobody” on many base images)
USER 65532

ENTRYPOINT ["/app/app"]
