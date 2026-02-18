FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/server \
    ./cmd/server

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /app/bin/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config.yaml ./config.yaml

EXPOSE 8080

ENTRYPOINT ["./server"]