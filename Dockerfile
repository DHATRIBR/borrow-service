# Use official Golang image
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire service directory
COPY . .

# Build the binary from the cmd folder with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o borrow-service ./cmd/main.go


# Use minimal image for production
FROM alpine:latest

WORKDIR /root/

# Copy the compiled binary from builder stage
COPY --from=builder /app/borrow-service .

# Expose the service port
EXPOSE 8080

# Set environment variables for DB connection
ENV DB_HOST=borrow-db
ENV DB_USER=user
ENV DB_PASSWORD=password
ENV DB_NAME=borrow_service

# Start the service
CMD ["./borrow-service"]