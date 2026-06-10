FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/app
RUN go build -o notification_service ./cmd/notification_service


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app ./app
COPY --from=builder /app/notification_service ./notification_service
COPY --from=builder /app/locales ./locales
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./app"]
