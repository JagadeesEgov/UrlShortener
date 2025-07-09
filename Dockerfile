# ---- Build Stage ----
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener main.go

# ---- Run Stage ----
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/url-shortener .
COPY schema.sql .
COPY .env.example .
EXPOSE 8080
CMD ["./url-shortener"] 