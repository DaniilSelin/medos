# этап компиляции
FROM golang:1.22 AS builder

WORKDIR /app

COPY ../go.mod ../go.sum ./
RUN go mod download

COPY ../ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-system ./cmd/server/main.go

# этап сборки
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/auth-system .

COPY ../config/config.yml ./config/config.yml

COPY internal/domain/migrations ./internal/domain/migrations

EXPOSE 8080

CMD ["./auth-system"]
