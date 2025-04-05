# Build stage
FROM golang:1.24 AS builder

# Install tools needed to build go-sqlite3 (CGO)
RUN apt-get update && apt-get install -y gcc libc6-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1

RUN go build -o main .

# Runtime stage (Debian)
FROM debian:bookworm-slim

# Install required sqlite shared libraries
RUN apt-get update && apt-get install -y sqlite3 libsqlite3-0 && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/data ./data

EXPOSE 42069

CMD ["./main"]
