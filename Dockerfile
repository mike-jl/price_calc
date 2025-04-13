# --- Stage 1: Frontend Build ---
FROM node:20 AS frontend

WORKDIR /app

COPY package*.json vite.config.ts tsconfig.json ./
COPY scripts ./scripts
RUN npm install
RUN npm run build

# --- Stage 2: Go Backend Build ---
FROM golang:1.24 AS backend

# Accept build args
ARG TARGETARCH

# Map TARGETARCH to values Zig expects
# Add a second ARG that we override in the build step
ARG ZIGTARGET

# Set up zig
RUN apt-get update && apt-get install -y wget xz-utils && \
    wget https://ziglang.org/download/0.14.0/zig-linux-x86_64-0.14.0.tar.xz && \
    tar -xf zig-linux-x86_64-0.14.0.tar.xz && \
    mv zig-linux-x86_64-0.14.0 /zig

ENV PATH="/zig:${PATH}"

# Tell Go what we’re building for
ENV CGO_ENABLED=1
ENV CC="zig cc"
ENV CFLAGS="--target=${ZIGTARGET}"

# Build app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -v -o main .

# --- Stage 3: Final Image ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y sqlite3 libsqlite3-0 && rm -rf /var/lib/apt/lists/*

WORKDIR /root

# Ensure db directory exists (since it's .gitignored and not copied)
RUN mkdir -p ./db

# Copy Go binary
COPY --from=backend /app/main .

# Copy static assets (built JS, CSS)
COPY --from=frontend /app/assets ./assets

# Copy runtime data (db etc.)
COPY --from=backend /app/data ./data

EXPOSE 42069
CMD ["./main"]

