# --- Stage 1: Frontend Build ---
FROM node:20 AS frontend

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY vite.config.ts tsconfig.json ./
COPY scripts ./scripts
RUN npm run build

# Stage that imports the binary from the appropriate build context
FROM --platform=$BUILDPLATFORM binary-amd64 AS import-amd64
FROM --platform=$BUILDPLATFORM binary-arm64 AS import-arm64

# --- Stage 2: Final runtime image ---
FROM alpine:latest

WORKDIR /root

# Create db dir (ignored by Git, needed at runtime)
RUN mkdir -p ./db

ARG TARGETPLATFORM
# Copy the correct binary depending on TARGETPLATFORM
# Use a conditional COPY — only one will be valid depending on the target platform
COPY --from=import-amd64 /main ./main
COPY --from=import-arm64 /main ./main

RUN apk add --no-cache file
RUN file ./main

# Copy frontend assets
COPY --from=frontend /app/assets ./assets

# Copy runtime data
COPY ./data ./data

EXPOSE 42069
CMD ["./main"]

