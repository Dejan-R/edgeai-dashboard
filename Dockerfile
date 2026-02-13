
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o edgeai-backend


FROM alpine:latest

# CA certificates (PostgreSQL, MQTT)
RUN apk --no-cache add ca-certificates

WORKDIR /app


COPY --from=builder /app/edgeai-backend .


COPY backend/templates ./templates
COPY backend/static ./static

EXPOSE 8080

CMD ["./edgeai-backend"]