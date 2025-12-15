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

COPY --from=client-builder /app/client/dist ./client/dist

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.22.1

RUN adduser -D -s /bin/sh appuser
WORKDIR /root/

COPY --from=server-builder /app/main .
RUN chown -R appuser:appuser /root/

USER appuser

EXPOSE 8080

CMD ["./main"]
