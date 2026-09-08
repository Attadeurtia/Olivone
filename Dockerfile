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

# ── 3) Récupération du binaire Typst (rendu PDF des lettres) ─────────────────
FROM alpine:3.20 AS typst
ARG TYPST_VERSION=v0.15.1
RUN apk add --no-cache curl tar xz \
 && curl -fsSL https://github.com/typst/typst/releases/download/${TYPST_VERSION}/typst-x86_64-unknown-linux-musl.tar.xz -o /tmp/typst.tar.xz \
 && mkdir -p /tmp/typst && tar -xJf /tmp/typst.tar.xz -C /tmp/typst --strip-components=1 \
 && install -m 0755 /tmp/typst/typst /usr/local/bin/typst

# ── 4) Image finale minimale ─────────────────────────────────────────────────
FROM alpine:3.20
RUN adduser -D -u 1000 olivone
WORKDIR /app
COPY --from=server /out/olivone /app/olivone
COPY --from=web /web/dist /app/web
COPY --from=typst /usr/local/bin/typst /usr/local/bin/typst
ENV OLIVONE_ENV=prod \
    OLIVONE_BIND=:8080 \
    OLIVONE_DATA_DIR=/data \
    OLIVONE_WEB_DIR=/app/web
# Le volume /data doit appartenir à l'utilisateur non-root, sinon la base
# SQLite ne peut pas être créée (SQLITE_CANTOPEN). Un volume nommé vide hérite
# de l'appartenance de ce dossier dans l'image.
RUN mkdir -p /data && chown olivone:olivone /data
VOLUME /data
EXPOSE 8080
USER olivone
ENTRYPOINT ["/app/olivone"]
