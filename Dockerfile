FROM golang:1.25-alpine AS build

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o auth-service ./cmd/server

FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=build /app/auth-service .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 9001

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9001/health || exit 1

ENTRYPOINT ["./auth-service"]
