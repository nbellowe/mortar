# ── Stage 1: build the Expo web frontend ──────────────────────────────────────
FROM node:20-alpine AS web-builder
WORKDIR /app
COPY package.json package-lock.json app.json tsconfig.json ./
RUN npm ci --prefer-offline
COPY src/frontend/ ./src/frontend/
RUN npm run build:web

# ── Stage 2: build the Go backend ─────────────────────────────────────────────
FROM golang:1.25-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src/backend/ ./src/backend/
COPY src/db/ ./src/db/
# Bake the web export into the embed path before compiling.
COPY --from=web-builder /app/dist ./src/backend/cmd/server/web
RUN CGO_ENABLED=1 GOOS=linux go build -o /mortar ./src/backend/cmd/server

# ── Stage 3: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /mortar .
VOLUME /data
EXPOSE 3000
ENTRYPOINT ["./mortar"]
CMD ["--config", "/data/config.yaml", "--db", "/data/mortar.db"]
