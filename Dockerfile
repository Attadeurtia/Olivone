# syntax=docker/dockerfile:1

# ── 1) Build du frontend (Svelte -> web/dist) ────────────────────────────────
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

# ── 2) Build du serveur Go (binaire statique, sans CGO) ──────────────────────
FROM golang:1.27-alpine AS server
WORKDIR /src
COPY server/go.mod server/go.sum* ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/olivone ./cmd/olivone

# ── 3) Image finale minimale ─────────────────────────────────────────────────
# NB : Typst (rendu PDF) sera ajouté au jalon M2, quand on générera les lettres.
FROM alpine:3.20
RUN adduser -D -u 1000 olivone
WORKDIR /app
COPY --from=server /out/olivone /app/olivone
COPY --from=web /web/dist /app/web
COPY server/prompt/prompt_app.default.md /app/prompt/prompt_app.default.md
ENV OLIVONE_ENV=prod \
    OLIVONE_BIND=:8080 \
    OLIVONE_DATA_DIR=/data \
    OLIVONE_WEB_DIR=/app/web
VOLUME /data
EXPOSE 8080
USER olivone
ENTRYPOINT ["/app/olivone"]
