diff --git a/core-services/egov-url-shortening-go/setup-local.sh b/core-services/egov-url-shortening-go/setup-local.sh
--- a/core-services/egov-url-shortening-go/setup-local.sh
+++ b/core-services/egov-url-shortening-go/setup-local.sh
@@ -0,0 +1,60 @@
+#!/bin/bash
+
+# Setup script for egov-url-shortening-go
+# Run this script on your local machine to create the complete project
+
+set -e
+
+echo "🚀 Setting up eGov URL Shortening Service (Go)"
+
+# Create project directory
+PROJECT_DIR="egov-url-shortening-go"
+if [ -d "$PROJECT_DIR" ]; then
+    echo "Directory $PROJECT_DIR already exists. Remove it? (y/n)"
+    read -r response
+    if [[ "$response" == "y" ]]; then
+        rm -rf "$PROJECT_DIR"
+    else
+        echo "Aborting setup."
+        exit 1
+    fi
+fi
+
+mkdir -p "$PROJECT_DIR"
+cd "$PROJECT_DIR"
+
+# Create directory structure
+echo "📁 Creating directory structure..."
+mkdir -p config models utils repository service handlers migrations
+
+echo "📝 Creating files..."
+
+# Create go.mod
+cat > go.mod << 'EOF'
+module egov-url-shortening-go
+
+go 1.21
+
+require (
+	github.com/gin-gonic/gin v1.9.1
+	github.com/go-redis/redis/v8 v8.11.5
+	github.com/jackc/pgx/v5 v5.5.1
+	github.com/speps/go-hashids/v2 v2.0.1
+	github.com/kelseyhightower/envconfig v1.4.0
+	github.com/sirupsen/logrus v1.9.3
+)
+EOF
+
+echo "✅ Created go.mod"
+
+echo ""
+echo "🎉 Project structure created successfully!"
+echo ""
+echo "Next steps:"
+echo "1. cd $PROJECT_DIR"
+echo "2. Copy the remaining files from the workspace"
+echo "3. Run: go mod tidy"
+echo "4. Run: make docker-run"
+echo ""
+echo "Or download the complete project from:"
+echo "📦 /workspace/core-services/egov-url-shortening-go/egov-url-shortening-go.tar.gz"