FROM golang:1.23.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git


COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.22.1

RUN adduser -D -s /bin/sh appuser
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/client ./client
RUN chown -R appuser:appuser /root/

USER appuser

EXPOSE 8080

CMD ["./main"]
