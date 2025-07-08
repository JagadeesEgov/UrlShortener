diff --git a/core-services/egov-url-shortening-go/Makefile b/core-services/egov-url-shortening-go/Makefile
--- a/core-services/egov-url-shortening-go/Makefile
+++ b/core-services/egov-url-shortening-go/Makefile
@@ -0,0 +1,62 @@
+.PHONY: build run test clean docker-build docker-run docker-stop deps
+
+# Build the application
+build:
+	go build -o bin/url-shortener main.go
+
+# Run the application locally
+run:
+	go run main.go
+
+# Run tests
+test:
+	go test ./...
+
+# Clean build artifacts
+clean:
+	rm -rf bin/
+
+# Download dependencies
+deps:
+	go mod download
+	go mod tidy
+
+# Build Docker image
+docker-build:
+	docker build -t egov-url-shortening-go .
+
+# Run with Docker Compose
+docker-run:
+	docker-compose up --build
+
+# Stop Docker Compose
+docker-stop:
+	docker-compose down
+
+# Format code
+fmt:
+	go fmt ./...
+
+# Lint code (requires golangci-lint)
+lint:
+	golangci-lint run
+
+# Development setup
+dev-setup: deps
+	@echo "Development environment setup complete"
+
+# Help
+help:
+	@echo "Available commands:"
+	@echo "  build       - Build the application"
+	@echo "  run         - Run the application locally"
+	@echo "  test        - Run tests"
+	@echo "  clean       - Clean build artifacts"
+	@echo "  deps        - Download and tidy dependencies"
+	@echo "  docker-build - Build Docker image"
+	@echo "  docker-run  - Run with Docker Compose"
+	@echo "  docker-stop - Stop Docker Compose"
+	@echo "  fmt         - Format code"
+	@echo "  lint        - Lint code"
+	@echo "  dev-setup   - Set up development environment"
+	@echo "  help        - Show this help message"