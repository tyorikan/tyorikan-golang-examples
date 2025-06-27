
# Use the official Golang image to create a build artifact.
# This is known as a multi-stage build.
FROM golang:1.24-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# -o main: output file name
# CGO_ENABLED=0: disable Cgo
# GOOS=linux: target operating system
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main .

# Start a new stage from scratch for a smaller image
FROM alpine:latest

WORKDIR /app

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Change ownership of the binary
RUN chown appuser:appgroup /app/main

# Switch to the non-root user
USER appuser

# Command to run the executable
CMD ["./main"]
