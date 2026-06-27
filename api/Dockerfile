# Stage 1: Build the binary
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy dependency definition files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Stage 2: Run the binary in a clean minimal container
FROM alpine:latest

# Install ca-certificates in case we connect to external secure services (like SMTP or external databases)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the builder stage
COPY --from=builder /app/main .

# Copy env file if present (optional, usually overridden by docker-compose environment vars)
COPY --from=builder /app/.env .

# Export port
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
