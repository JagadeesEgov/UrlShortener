diff --git a/core-services/egov-url-shortening-go/migrations/001_create_url_shortener_table.sql b/core-services/egov-url-shortening-go/migrations/001_create_url_shortener_table.sql
--- a/core-services/egov-url-shortening-go/migrations/001_create_url_shortener_table.sql
+++ b/core-services/egov-url-shortening-go/migrations/001_create_url_shortener_table.sql
@@ -0,0 +1,12 @@
+-- Create sequence for URL shortener IDs
+DROP SEQUENCE IF EXISTS eg_url_shorter_id;
+CREATE SEQUENCE eg_url_shorter_id;
+
+-- Create URL shortener table (matching the Java DDL)
+CREATE TABLE IF NOT EXISTS "eg_url_shortener" (
+    "id" VARCHAR(128) NOT NULL,
+    "validform" BIGINT,
+    "validto" BIGINT,
+    "url" VARCHAR(1024) NOT NULL,
+    PRIMARY KEY ("id")
+);