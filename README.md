# Url Shortener

diff --git a/core-services/egov-url-shortening-go/README.md b/core-services/egov-url-shortening-go/README.md
--- a/core-services/egov-url-shortening-go/README.md
+++ b/core-services/egov-url-shortening-go/README.md
@@ -0,0 +1,116 @@
+# egov-url-shortening-go

- +A Go implementation of the eGov URL shortening service, converted from the original Java Spring Boot version.
- +## Features
- +- URL shortening using HashIDs
  +- Redis and PostgreSQL storage options
  +- URL validation with regex pattern matching
  +- Multi-tenant support with state-level tenant extraction
  +- Redirect functionality with expiry validation
  +- REST API endpoints compatible with the Java version
  +- Docker support with multi-stage builds
  +- Structured logging with JSON format
- +## API Endpoints
- +- `POST /shortener` - Shorten a URL
  +- `GET /{id}` - Redirect to original URL
  +- `GET /health` - Health check endpoint
- +## Quick Start
- +### Using Docker Compose (Recommended)
- +```bash
  +# Clone and navigate to the directory
  +cd core-services/egov-url-shortening-go
- +# Start all services (app, PostgreSQL, Redis)
  +make docker-run
- +# The service will be available at http://localhost:8091
  +```
- +### Local Development
- +```bash
  +# Install dependencies
  +make deps
- +# Set up environment variables (copy from .env.example)
  +cp .env.example .env
- +# Run the service
  +make run
  +```
- +## Configuration
- +The service uses environment variables for configuration. See `.env.example` for all available options:
- +### Key Configuration Options
- +- `SERVER_PORT`: Server port (default: 8091)
  +- `SERVER_CONTEXT_PATH`: API context path (default: /egov-url-shortening)
  +- `DATABASE_ENABLED`: Use PostgreSQL if true, Redis if false
  +- `HASHIDS_SALT`: Salt for HashID generation
  +- `APP_IS_MULTI_INSTANCE`: Enable multi-tenant mode
- +## API Usage
- +### Shorten a URL
- +```bash
  +curl -X POST http://localhost:8091/egov-url-shortening/shortener \
- -H "Content-Type: application/json" \
- -d '{"url": "https://www.example.com"}'
  +```
- +### Access shortened URL
- +`bash
+curl -L http://localhost:8091/egov-url-shortening/{shortened_id}
+`
- +## Development
- +```bash
  +# Format code
  +make fmt
- +# Run tests
  +make test
- +# Build binary
  +make build
- +# Clean build artifacts
  +make clean
  +```
- +## Database Setup
- +The service uses the same PostgreSQL schema as the Java version:
- +```sql
  +CREATE TABLE "eg_url_shortener" (
- "id" VARCHAR(128) NOT NULL,
- "validform" bigint,
- "validto" bigint,
- "url" VARCHAR(1024) NOT NULL,
- PRIMARY KEY ("id")
  +);
  +```
- +## Migration from Java
- +This Go implementation maintains:
  +- Same API endpoints and request/response formats
  +- Same database schema and data compatibility
  +- Same HashID generation algorithm
  +- Same multi-tenant logic
  +- Same URL validation patterns
- +You can switch from the Java version to this Go version without data migration.
