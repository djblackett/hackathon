FROM golang:1.24.5-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o relay-server ./cmd/server/main.go



FROM alpine:latest

COPY --from=builder /app/relay-server /usr/local/bin/relay-server

ENTRYPOINT ["relay-server"]
