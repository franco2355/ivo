# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Regenerate go.sum and tidy dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o activities-api ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/activities-api .

# Expose port
EXPOSE 8082

# Run
CMD ["./activities-api"]
