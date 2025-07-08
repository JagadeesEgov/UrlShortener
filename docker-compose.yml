diff --git a/core-services/egov-url-shortening-go/docker-compose.yml b/core-services/egov-url-shortening-go/docker-compose.yml
--- a/core-services/egov-url-shortening-go/docker-compose.yml
+++ b/core-services/egov-url-shortening-go/docker-compose.yml
@@ -0,0 +1,128 @@
+version: '3.8'
+
+services:
+  app:
+    build: .
+    ports:
+      - "8091:8091"
+    environment:
+      # Server Configuration
+      - SERVER_PORT=8091
+      - SERVER_CONTEXT_PATH=/egov-url-shortening
+      
+      # Redis Configuration
+      - REDIS_HOST=redis
+      - REDIS_PORT=6379
+      - REDIS_PASSWORD=
+      - REDIS_DB=0
+      
+      # Database Configuration
+      - DATABASE_HOST=postgres
+      - DATABASE_PORT=5432
+      - DATABASE_NAME=devdb
+      - DATABASE_USERNAME=postgres
+      - DATABASE_PASSWORD=postgres
+      - DATABASE_ENABLED=true
+      - DATABASE_SSL_MODE=disable
+      
+      # Application Configuration
+      - APP_HOST_NAME=http://localhost:8091/
+      - APP_STATE_LEVEL_TENANT_ID=pb
+      - APP_IS_MULTI_INSTANCE=false
+      - APP_IS_CENTRAL_INSTANCE=true
+      - APP_STATE_LEVEL_TENANT_ID_LENGTH=2
+      
+      # HashIDs Configuration
+      - HASHIDS_SALT=docker-development-salt
+      - HASHIDS_MIN_LENGTH=5
+      
+      # Debug Configuration
+      - GIN_MODE=debug
+      - LOG_LEVEL=info
+    depends_on:
+      postgres:
+        condition: service_healthy
+      redis:
+        condition: service_healthy
+    networks:
+      - url-shortener-network
+    healthcheck:
+      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8091/health"]
+      interval: 30s
+      timeout: 10s
+      retries: 3
+      start_period: 40s
+
+  postgres:
+    image: postgres:15-alpine
+    environment:
+      - POSTGRES_DB=devdb
+      - POSTGRES_USER=postgres
+      - POSTGRES_PASSWORD=postgres
+    ports:
+      - "5432:5432"
+    volumes:
+      - postgres_data:/var/lib/postgresql/data
+      - ./migrations:/docker-entrypoint-initdb.d
+    networks:
+      - url-shortener-network
+    healthcheck:
+      test: ["CMD-SHELL", "pg_isready -U postgres"]
+      interval: 10s
+      timeout: 5s
+      retries: 5
+
+  redis:
+    image: redis:7-alpine
+    ports:
+      - "6379:6379"
+    volumes:
+      - redis_data:/data
+    networks:
+      - url-shortener-network
+    healthcheck:
+      test: ["CMD", "redis-cli", "ping"]
+      interval: 10s
+      timeout: 5s
+      retries: 5
+
+  # Optional: Redis Commander for Redis management
+  redis-commander:
+    image: rediscommander/redis-commander:latest
+    environment:
+      - REDIS_HOSTS=local:redis:6379
+    ports:
+      - "8081:8081"
+    depends_on:
+      redis:
+        condition: service_healthy
+    networks:
+      - url-shortener-network
+    profiles:
+      - debug
+
+  # Optional: pgAdmin for PostgreSQL management
+  pgadmin:
+    image: dpage/pgadmin4:latest
+    environment:
+      - PGADMIN_DEFAULT_EMAIL=admin@example.com
+      - PGADMIN_DEFAULT_PASSWORD=admin
+    ports:
+      - "8080:80"
+    depends_on:
+      postgres:
+        condition: service_healthy
+    networks:
+      - url-shortener-network
+    profiles:
+      - debug
+
+volumes:
+  postgres_data:
+    driver: local
+  redis_data:
+    driver: local
+
+networks:
+  url-shortener-network:
+    driver: bridge