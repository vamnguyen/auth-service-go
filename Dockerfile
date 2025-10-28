FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o auth-service ./cmd/app

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/auth-service .
EXPOSE 9001
ENTRYPOINT ["./auth-service"]
