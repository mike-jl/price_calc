# --- Stage 1: Frontend Build ---
FROM node:20 AS frontend

WORKDIR /app

COPY package*.json vite.config.ts tsconfig.json ./
COPY scripts ./scripts
RUN npm install
RUN npm run build

# --- Stage 2: Final runtime image ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y sqlite3 libsqlite3-0 && rm -rf /var/lib/apt/lists/*

WORKDIR /root

# Ensure db directory exists (since it's .gitignored and not copied)
RUN mkdir -p ./db

# Copy precompiled Go binary
ARG TARGETARCH
COPY dist/${TARGETARCH}/main .

# Copy static assets (built JS, CSS)
COPY --from=frontend /app/assets ./assets

# Copy runtime data (db etc.)
COPY ./data ./data

EXPOSE 42069
CMD ["./main"]

