 # Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o url-shortener ./cmd/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/url-shortener .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./url-shortener"]

# # Build stage
# FROM golang:1.23-alpine AS builder
# # Install build dependencies
# RUN apk add --no-cache git gcc musl-dev
# # Set working directory
# WORKDIR /app
# # Copy go mod and sum files
# COPY go.mod go.sum ./
# # Download dependencies
# RUN go mod download
# # Copy source code
# COPY . .
# # Build the application
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o template-config ./cmd/main.go
# # Final stage
# FROM alpine:latest
# # Install runtime dependencies
# RUN apk add --no-cache ca-certificates tzdata
# # Set working directory
# WORKDIR /app
# # Copy the binary from builder
# COPY --from=builder /app/template-config .
# # Expose any necessary ports (if needed)
# EXPOSE 8080
# # Run the application
# CMD ["./template-config"] 