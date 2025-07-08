diff --git a/core-services/egov-url-shortening-go/utils/validator.go b/core-services/egov-url-shortening-go/utils/validator.go
--- a/core-services/egov-url-shortening-go/utils/validator.go
+++ b/core-services/egov-url-shortening-go/utils/validator.go
@@ -0,0 +1,165 @@
+package utils
+
+import (
+	"net/url"
+	"regexp"
+	"strings"
+)
+
+// URLValidator provides URL validation functionality
+type URLValidator struct {
+	urlRegex *regexp.Regexp
+}
+
+// NewURLValidator creates a new URLValidator instance
+func NewURLValidator() *URLValidator {
+	// Same regex pattern as the Java implementation
+	urlRegex := regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)
+	
+	return &URLValidator{
+		urlRegex: urlRegex,
+	}
+}
+
+// ValidateURL validates a URL using the same regex pattern as the Java implementation
+func (v *URLValidator) ValidateURL(inputURL string) bool {
+	if inputURL == "" {
+		return false
+	}
+	
+	// Trim whitespace
+	inputURL = strings.TrimSpace(inputURL)
+	if inputURL == "" {
+		return false
+	}
+	
+	// Convert to lowercase for regex matching (matching Java behavior)
+	lowerURL := strings.ToLower(inputURL)
+	
+	// Check against regex pattern (primary validation - matches Java)
+	if !v.urlRegex.MatchString(lowerURL) {
+		return false
+	}
+	
+	// Additional validation using Go's url.Parse for extra safety
+	return v.validateURLStructure(inputURL)
+}
+
+// validateURLStructure performs additional URL structure validation
+func (v *URLValidator) validateURLStructure(inputURL string) bool {
+	// Add protocol if missing (for parsing)
+	testURL := inputURL
+	if !strings.HasPrefix(strings.ToLower(testURL), "http://") && !strings.HasPrefix(strings.ToLower(testURL), "https://") {
+		testURL = "http://" + testURL
+	}
+	
+	// Parse URL
+	parsedURL, err := url.Parse(testURL)
+	if err != nil {
+		return false
+	}
+	
+	// Validate host
+	if parsedURL.Host == "" {
+		return false
+	}
+	
+	// Check for valid host format
+	host := parsedURL.Hostname()
+	if host == "" {
+		return false
+	}
+	
+	// Reject localhost and private IPs for security (unless in development)
+	if v.isPrivateOrLocalhost(host) {
+		return false
+	}
+	
+	return true
+}
+
+// isPrivateOrLocalhost checks if the host is localhost or private IP
+func (v *URLValidator) isPrivateOrLocalhost(host string) bool {
+	// Allow localhost in development (you can make this configurable)
+	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
+		// For now, allow localhost for testing
+		// In production, you might want to reject these
+		return false
+	}
+	
+	// Check for private IP ranges (10.x.x.x, 172.16-31.x.x, 192.168.x.x)
+	privateIPPatterns := []string{
+		`^10\.`,
+		`^172\.(1[6-9]|2[0-9]|3[0-1])\.`,
+		`^192\.168\.`,
+	}
+	
+	for _, pattern := range privateIPPatterns {
+		matched, _ := regexp.MatchString(pattern, host)
+		if matched {
+			return true
+		}
+	}
+	
+	return false
+}
+
+// SanitizeURL cleans and normalizes a URL
+func (v *URLValidator) SanitizeURL(inputURL string) string {
+	if inputURL == "" {
+		return ""
+	}
+	
+	// Trim whitespace
+	cleanURL := strings.TrimSpace(inputURL)
+	
+	// Add protocol if missing
+	if !strings.HasPrefix(strings.ToLower(cleanURL), "http://") && !strings.HasPrefix(strings.ToLower(cleanURL), "https://") {
+		cleanURL = "https://" + cleanURL
+	}
+	
+	// Parse and reconstruct URL to normalize it
+	parsedURL, err := url.Parse(cleanURL)
+	if err != nil {
+		return inputURL // Return original if parsing fails
+	}
+	
+	return parsedURL.String()
+}
+
+// IsValidDomain checks if a domain name is valid
+func (v *URLValidator) IsValidDomain(domain string) bool {
+	if domain == "" {
+		return false
+	}
+	
+	// Basic domain validation regex
+	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
+	
+	return domainRegex.MatchString(domain) && len(domain) <= 253
+}
+
+// ExtractDomain extracts domain from URL
+func (v *URLValidator) ExtractDomain(inputURL string) string {
+	// Add protocol if missing
+	if !strings.HasPrefix(strings.ToLower(inputURL), "http://") && !strings.HasPrefix(strings.ToLower(inputURL), "https://") {
+		inputURL = "https://" + inputURL
+	}
+	
+	parsedURL, err := url.Parse(inputURL)
+	if err != nil {
+		return ""
+	}
+	
+	return parsedURL.Hostname()
+}
+
+// ValidateAndSanitizeURL validates a URL and returns the sanitized version
+func (v *URLValidator) ValidateAndSanitizeURL(inputURL string) (string, bool) {
+	if !v.ValidateURL(inputURL) {
+		return "", false
+	}
+	
+	sanitized := v.SanitizeURL(inputURL)
+	return sanitized, true
+}