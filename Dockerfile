FROM node:24-alpine AS client-builder

WORKDIR /app/client

COPY client/package*.json ./
RUN npm ci

COPY client/ ./
RUN npm run build

FROM golang:1.25.2-alpine AS server-builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY server/ ./server/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.22.1

LABEL org.opencontainers.image.source="https://github.com/dazkins/webscape" \
      org.opencontainers.image.licenses="AGPL-3.0-only"

RUN adduser -D -s /bin/sh appuser
WORKDIR /app

COPY --from=server-builder /app/main ./main
COPY --from=client-builder /app/client/dist ./client/dist
COPY config.json ./config.json
COPY game-project/ ./game-project/
COPY LICENSE THIRD_PARTY_LICENSES.txt ./
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["./main"]
