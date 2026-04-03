FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/api cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./api"]
