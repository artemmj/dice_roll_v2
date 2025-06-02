FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /dice_roll ./cmd/dice_roll
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrator ./cmd/migrator

# Создание финального образа
FROM alpine:latest
RUN apk add --no-cache postgresql-client
COPY --from=builder /dice_roll /app/dice_roll
COPY --from=builder /migrator /app/migrator
COPY migrations /app/migrations
COPY config /app/config
WORKDIR /app
# Команда запуска (будет переопределена в docker-compose)
CMD ["/app/dice_roll"]
