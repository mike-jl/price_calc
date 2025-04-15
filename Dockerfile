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

# Use ARG to select platform
ARG TARGETPLATFORM

# Set default to amd64 to avoid empty FROM in some cases
COPY --from=import-amd64 /main /tmp/main_amd64
COPY --from=import-arm64 /main /tmp/main_arm64

# Use shell command to pick the correct binary
# RUN case "$TARGETPLATFORM" in \
#         "linux/amd64") cp /tmp/main_amd64 ./main ;; \
#         "linux/arm64") cp /tmp/main_arm64 ./main ;; \
#         *) echo "Unsupported TARGETPLATFORM: $TARGETPLATFORM" && exit 1 ;; \
#     esac && \
#     chmod +x ./main \
#     rm -f /tmp/main_amd64 /tmp/main_arm64

ARG TARGETARCH
COPY --from=import-${TARGETARCH} /main ./main
RUN chmod +x ./main

RUN apk add --no-cache file
RUN file ./main

# Copy frontend assets
COPY --from=frontend /app/assets ./assets

# Copy runtime data
COPY ./data ./data

EXPOSE 42069
CMD ["./main"]

