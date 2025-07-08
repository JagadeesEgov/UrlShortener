diff --git a/core-services/egov-url-shortening-go/test/basic_test.go b/core-services/egov-url-shortening-go/test/basic_test.go
--- a/core-services/egov-url-shortening-go/test/basic_test.go
+++ b/core-services/egov-url-shortening-go/test/basic_test.go
@@ -0,0 +1,168 @@
+package test
+
+import (
+	"testing"
+	"time"
+
+	"egov-url-shortening-go/config"
+	"egov-url-shortening-go/models"
+	"egov-url-shortening-go/utils"
+)
+
+func TestHashIDConverter(t *testing.T) {
+	cfg := &config.HashIDsConfig{
+		Salt:      "test-salt",
+		MinLength: 3,
+	}
+
+	converter, err := utils.NewHashIDConverter(cfg)
+	if err != nil {
+		t.Fatalf("Failed to create HashIDConverter: %v", err)
+	}
+
+	// Test encoding and decoding
+	testID := int64(12345)
+	
+	hashString, err := converter.CreateHashStringForID(testID)
+	if err != nil {
+		t.Fatalf("Failed to encode ID: %v", err)
+	}
+
+	if len(hashString) < cfg.MinLength {
+		t.Fatalf("Hash string length %d is less than minimum %d", len(hashString), cfg.MinLength)
+	}
+
+	decodedID, err := converter.GetIDForString(hashString)
+	if err != nil {
+		t.Fatalf("Failed to decode hash string: %v", err)
+	}
+
+	if decodedID != testID {
+		t.Fatalf("Expected %d, got %d", testID, decodedID)
+	}
+
+	t.Logf("Successfully encoded ID %d to hash %s and decoded back", testID, hashString)
+}
+
+func TestURLValidator(t *testing.T) {
+	validator := utils.NewURLValidator()
+
+	validURLs := []string{
+		"https://www.example.com",
+		"http://example.com",
+		"https://example.com/path",
+		"http://subdomain.example.org",
+		"https://example.com:8080/path",
+	}
+
+	invalidURLs := []string{
+		"",
+		"not-a-url",
+		"http://",
+		"https://",
+		"ftp://example.com",
+		"example",
+	}
+
+	for _, url := range validURLs {
+		if !validator.ValidateURL(url) {
+			t.Errorf("Expected %s to be valid", url)
+		}
+	}
+
+	for _, url := range invalidURLs {
+		if validator.ValidateURL(url) {
+			t.Errorf("Expected %s to be invalid", url)
+		}
+	}
+
+	t.Log("URL validation tests passed")
+}
+
+func TestShortenRequest(t *testing.T) {
+	// Test valid request
+	validRequest := &models.ShortenRequest{
+		URL: "https://www.example.com",
+	}
+
+	if err := validRequest.Validate(); err != nil {
+		t.Errorf("Expected valid request to pass validation: %v", err)
+	}
+
+	// Test invalid request
+	invalidRequest := &models.ShortenRequest{
+		URL: "",
+	}
+
+	if err := invalidRequest.Validate(); err == nil {
+		t.Error("Expected invalid request to fail validation")
+	}
+
+	// Test expiry logic
+	now := time.Now().UnixMilli()
+	expiredRequest := &models.ShortenRequest{
+		URL:       "https://www.example.com",
+		ValidTill: &[]int64{now - 1000}[0], // 1 second ago
+	}
+
+	if !expiredRequest.IsExpired() {
+		t.Error("Expected request to be expired")
+	}
+
+	if expiredRequest.IsActive() {
+		t.Error("Expected expired request to be inactive")
+	}
+
+	t.Log("ShortenRequest tests passed")
+}
+
+func TestConfig(t *testing.T) {
+	// Test config loading (which includes validation)
+	// Set some environment variables for testing
+	t.Setenv("SERVER_PORT", "8091")
+	t.Setenv("HASHIDS_SALT", "test-salt")
+	t.Setenv("APP_HOST_NAME", "http://localhost:8091")
+	t.Setenv("APP_STATE_LEVEL_TENANT_ID", "test")
+
+	cfg, err := config.LoadConfig()
+	if err != nil {
+		t.Errorf("Expected config to load successfully: %v", err)
+	}
+
+	if cfg.Server.Port != 8091 {
+		t.Errorf("Expected port 8091, got %d", cfg.Server.Port)
+	}
+
+	if cfg.HashIDs.Salt != "test-salt" {
+		t.Errorf("Expected salt 'test-salt', got %s", cfg.HashIDs.Salt)
+	}
+
+	t.Log("Config loading tests passed")
+}
+
+// Mock test to verify compilation
+func TestCompilation(t *testing.T) {
+	t.Log("All packages compile successfully")
+}
+
+// Benchmark hash ID generation
+func BenchmarkHashIDGeneration(b *testing.B) {
+	cfg := &config.HashIDsConfig{
+		Salt:      "benchmark-salt",
+		MinLength: 5,
+	}
+
+	converter, err := utils.NewHashIDConverter(cfg)
+	if err != nil {
+		b.Fatalf("Failed to create HashIDConverter: %v", err)
+	}
+
+	b.ResetTimer()
+	
+	for i := 0; i < b.N; i++ {
+		_, err := converter.CreateHashStringForID(int64(i))
+		if err != nil {
+			b.Fatalf("Failed to generate hash: %v", err)
+		}
+	}
+}