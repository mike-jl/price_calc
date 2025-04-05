# Build stage
FROM golang:1.24 as builder

# Set working directory inside the container
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application (source code, static files, etc.)
COPY . .

# Assume sqlc and templ-generated files are already present and committed
# Build the Go app
RUN go build -o main .

# Final, minimal runtime image
FROM alpine:latest

# Install CA certificates (for HTTPS if needed)
RUN apk add --no-cache ca-certificates

# Set working directory in the runtime image
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy static assets (needed at runtime)
COPY --from=builder /app/assets ./assets

# Expose port (make sure your app listens on this port)
EXPOSE 8080

# Run the binary
CMD ["./main"]

