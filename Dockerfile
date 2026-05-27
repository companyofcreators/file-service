FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/server \
    ./cmd/api/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /build/server /usr/local/bin/server
COPY --from=builder /build/migrations /migrations

USER app

EXPOSE 8086

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
    CMD curl -f http://localhost:8086/internal/health || exit 1

ENTRYPOINT ["/usr/local/bin/server"]
