# --- Stage 1: Frontend Build ---
FROM node:20 AS frontend

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY vite.config.ts tsconfig.json ./
COPY scripts ./scripts
RUN npm run build

# --- Stage 2: Final runtime image ---
FROM alpine:latest

WORKDIR /root

# Create db dir (ignored by Git, needed at runtime)
RUN mkdir -p ./db

# Copy statically linked Go binary (uses modernc.org/sqlite)
ARG TARGETARCH
ARG BINARYCTX=binary-${TARGETARCH}
# Each arch gets its own context
COPY --from=${BINARYCTX} /main ./main
RUN chmod +x ./main

RUN apk add --no-cache file
RUN file ./main

# Copy frontend assets
COPY --from=frontend /app/assets ./assets

# Copy runtime data
COPY ./data ./data

EXPOSE 42069
CMD ["./main"]

