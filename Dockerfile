# --- Stage 1: Frontend Build ---
FROM node:20 AS frontend

WORKDIR /app

COPY package*.json vite.config.ts tsconfig.json ./
COPY scripts ./scripts
RUN npm install
RUN npm run build

# --- Stage 2: Go Backend Build ---
FROM golang:1.24 AS backend

RUN apt-get update && apt-get install -y gcc libc6-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1
RUN go build -o main .

# --- Stage 3: Final Image ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y sqlite3 libsqlite3-0 && rm -rf /var/lib/apt/lists/*

WORKDIR /root

# Copy Go binary
COPY --from=backend /app/main .

# Copy static assets (built JS, CSS)
COPY --from=frontend /app/dist ./assets

# Copy runtime data (db etc.)
COPY --from=backend /app/data ./data

EXPOSE 42069
CMD ["./main"]

