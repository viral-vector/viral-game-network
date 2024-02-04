# Stage 1: Build the Go binary
FROM golang:1.22rc2-alpine3.19 AS build

ENV PROJECT_DIR=/app \
    GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=0

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Install the Air tool
RUN go get -u  github.com/cosmtrek/air
RUN go install github.com/cosmtrek/air

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./src

# Expose the port that the application listens on
EXPOSE 8080

# 
CMD ["air"]