FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY openapi.yaml ./openapi.yaml

COPY --from=frontend-builder /app/frontend/build ./cmd/server/web/dist

RUN go build -o /out/foosball-server ./cmd/server

FROM alpine:3.21
WORKDIR /app

COPY --from=go-builder /out/foosball-server /app/foosball-server

ENV PORT=8080
ENV SQLITE_PATH=/data/foosball.db

VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["/app/foosball-server"]
