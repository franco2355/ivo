FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o search-api ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates netcat-openbsd

WORKDIR /root/

COPY --from=builder /app/search-api .

# Copy wait script and entrypoint
COPY --from=builder /app/wait-for-rabbitmq.sh /wait-for-rabbitmq.sh
COPY entrypoint.sh .
RUN chmod +x /wait-for-rabbitmq.sh entrypoint.sh

EXPOSE 8084

ENTRYPOINT ["./entrypoint.sh"]
