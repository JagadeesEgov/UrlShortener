diff --git a/core-services/egov-url-shortening-go/service/url_service.go b/core-services/egov-url-shortening-go/service/url_service.go
--- a/core-services/egov-url-shortening-go/service/url_service.go
+++ b/core-services/egov-url-shortening-go/service/url_service.go
@@ -0,0 +1,394 @@
+package service
+
+import (
+	"bytes"
+	"context"
+	"encoding/json"
+	"fmt"
+	"net/http"
+	"strings"
+	"time"
+
+	"egov-url-shortening-go/config"
+	"egov-url-shortening-go/models"
+	"egov-url-shortening-go/repository"
+	"egov-url-shortening-go/utils"
+
+	"github.com/sirupsen/logrus"
+)
+
+// URLService provides URL shortening and conversion functionality
+type URLService struct {
+	repository    repository.URLRepository
+	hashConverter *utils.HashIDConverter
+	validator     *utils.URLValidator
+	config        *config.Config
+	log           *logrus.Logger
+	httpClient    *http.Client
+}
+
+// NewURLService creates a new URL service instance
+func NewURLService(
+	repo repository.URLRepository,
+	hashConverter *utils.HashIDConverter,
+	validator *utils.URLValidator,
+	cfg *config.Config,
+	log *logrus.Logger,
+) *URLService {
+	// Configure HTTP client with reasonable timeouts
+	httpClient := &http.Client{
+		Timeout: 30 * time.Second,
+		Transport: &http.Transport{
+			MaxIdleConns:        10,
+			IdleConnTimeout:     30 * time.Second,
+			DisableCompression:  false,
+			MaxIdleConnsPerHost: 5,
+		},
+	}
+
+	return &URLService{
+		repository:    repo,
+		hashConverter: hashConverter,
+		validator:     validator,
+		config:        cfg,
+		log:           log,
+		httpClient:    httpClient,
+	}
+}
+
+// ShortenURL creates a shortened URL from the given request
+func (s *URLService) ShortenURL(ctx context.Context, request *models.ShortenRequest, tenantID string, multiInstance bool) (string, error) {
+	if request == nil {
+		return "", fmt.Errorf("request cannot be nil")
+	}
+
+	// Validate the request
+	if err := request.Validate(); err != nil {
+		return "", fmt.Errorf("request validation failed: %w", err)
+	}
+
+	s.log.WithFields(logrus.Fields{
+		"url":       request.URL,
+		"tenantId":  tenantID,
+		"multiInstance": multiInstance,
+	}).Info("Processing URL shortening request")
+
+	// Sanitize and validate URL
+	sanitizedURL, isValid := s.validator.ValidateAndSanitizeURL(request.URL)
+	if !isValid {
+		return "", fmt.Errorf("invalid URL format: %s", request.URL)
+	}
+
+	// Update request with sanitized URL
+	request.URL = sanitizedURL
+
+	// Get next ID
+	id, err := s.repository.IncrementID(ctx)
+	if err != nil {
+		return "", fmt.Errorf("failed to get next ID: %w", err)
+	}
+
+	// Create unique ID using hash
+	uniqueID, err := s.hashConverter.CreateHashStringForID(id)
+	if err != nil {
+		return "", fmt.Errorf("failed to create hash for ID %d: %w", id, err)
+	}
+
+	// Save URL
+	key := fmt.Sprintf("url:%d", id)
+	err = s.repository.SaveURL(ctx, key, request)
+	if err != nil {
+		return "", fmt.Errorf("failed to save URL: %w", err)
+	}
+
+	// Build shortened URL
+	shortenedURL, err := s.buildShortenedURL(uniqueID, tenantID, multiInstance)
+	if err != nil {
+		return "", fmt.Errorf("failed to build shortened URL: %w", err)
+	}
+
+	s.log.WithFields(logrus.Fields{
+		"originalUrl":  request.URL,
+		"shortenedUrl": shortenedURL,
+		"uniqueId":     uniqueID,
+		"id":           id,
+	}).Info("URL shortened successfully")
+
+	return shortenedURL, nil
+}
+
+// GetLongURLFromID retrieves the original URL from a shortened ID
+func (s *URLService) GetLongURLFromID(ctx context.Context, uniqueID string) (string, error) {
+	if uniqueID == "" {
+		return "", fmt.Errorf("unique ID cannot be empty")
+	}
+
+	s.log.WithField("uniqueID", uniqueID).Debug("Converting shortened URL back")
+
+	// Convert hash back to ID
+	id, err := s.hashConverter.GetIDForString(uniqueID)
+	if err != nil {
+		// If modern hash conversion fails, this could be a legacy ID
+		// For now, just return error - can implement legacy support later if needed
+		s.log.WithError(err).WithField("uniqueID", uniqueID).Warn("Failed to decode hash ID")
+		return "", fmt.Errorf("invalid shortened URL ID: %s", uniqueID)
+	}
+
+	// Get URL from repository
+	longURL, err := s.repository.GetURL(ctx, id)
+	if err != nil {
+		s.log.WithError(err).WithField("id", id).Error("Failed to retrieve URL from repository")
+		return "", err
+	}
+
+	if longURL == "" {
+		return "", fmt.Errorf("invalid request: URL not found")
+	}
+
+	s.log.WithFields(logrus.Fields{
+		"uniqueID": uniqueID,
+		"id":       id,
+		"longURL":  longURL,
+	}).Info("Successfully retrieved original URL")
+
+	return longURL, nil
+}
+
+// GetURLDetails retrieves full URL details from a shortened ID
+func (s *URLService) GetURLDetails(ctx context.Context, uniqueID string) (*models.ShortenRequest, error) {
+	if uniqueID == "" {
+		return nil, fmt.Errorf("unique ID cannot be empty")
+	}
+
+	// Convert hash back to ID
+	id, err := s.hashConverter.GetIDForString(uniqueID)
+	if err != nil {
+		return nil, fmt.Errorf("invalid shortened URL ID: %s", uniqueID)
+	}
+
+	// Get URL details from repository
+	return s.repository.GetURLDetails(ctx, id)
+}
+
+// ValidateURL validates a URL using the configured validator
+func (s *URLService) ValidateURL(url string) bool {
+	return s.validator.ValidateURL(url)
+}
+
+// DeleteURL deletes a shortened URL
+func (s *URLService) DeleteURL(ctx context.Context, uniqueID string) error {
+	if uniqueID == "" {
+		return fmt.Errorf("unique ID cannot be empty")
+	}
+
+	// Convert hash back to ID
+	id, err := s.hashConverter.GetIDForString(uniqueID)
+	if err != nil {
+		return fmt.Errorf("invalid shortened URL ID: %s", uniqueID)
+	}
+
+	// Delete from repository
+	err = s.repository.DeleteURL(ctx, id)
+	if err != nil {
+		return fmt.Errorf("failed to delete URL: %w", err)
+	}
+
+	s.log.WithFields(logrus.Fields{
+		"uniqueID": uniqueID,
+		"id":       id,
+	}).Info("URL deleted successfully")
+
+	return nil
+}
+
+// CheckURLExists checks if a URL exists for the given shortened ID
+func (s *URLService) CheckURLExists(ctx context.Context, uniqueID string) (bool, error) {
+	if uniqueID == "" {
+		return false, fmt.Errorf("unique ID cannot be empty")
+	}
+
+	// Convert hash back to ID
+	id, err := s.hashConverter.GetIDForString(uniqueID)
+	if err != nil {
+		return false, fmt.Errorf("invalid shortened URL ID: %s", uniqueID)
+	}
+
+	// Check existence in repository
+	return s.repository.CheckURLExists(ctx, id)
+}
+
+// HealthCheck performs a health check on the service
+func (s *URLService) HealthCheck(ctx context.Context) error {
+	// Check repository health
+	if err := s.repository.HealthCheck(ctx); err != nil {
+		return fmt.Errorf("repository health check failed: %w", err)
+	}
+
+	// Check hash converter configuration
+	if err := s.hashConverter.ValidateConfiguration(); err != nil {
+		return fmt.Errorf("hash converter validation failed: %w", err)
+	}
+
+	// Test URL validation
+	if !s.validator.ValidateURL("https://www.example.com") {
+		return fmt.Errorf("URL validator is not working correctly")
+	}
+
+	return nil
+}
+
+// buildShortenedURL constructs the complete shortened URL
+func (s *URLService) buildShortenedURL(uniqueID, tenantID string, multiInstance bool) (string, error) {
+	if uniqueID == "" {
+		return "", fmt.Errorf("unique ID cannot be empty")
+	}
+	if tenantID == "" {
+		return "", fmt.Errorf("tenant ID cannot be empty")
+	}
+
+	var hostName string
+
+	if multiInstance {
+		// Multi-instance: get hostname from tenant mapping
+		if s.config.App.UIAppHostMapParsed == nil {
+			return "", fmt.Errorf("UI app host map not configured for multi-instance")
+		}
+
+		var exists bool
+		hostName, exists = s.config.App.UIAppHostMapParsed[tenantID]
+		if !exists {
+			return "", fmt.Errorf("hostname for provided state level tenant has not been configured for tenantId: %s", tenantID)
+		}
+	} else {
+		// Single instance: use configured hostname
+		hostName = s.config.App.HostName
+	}
+
+	// Clean up hostname (remove trailing slash)
+	if strings.HasSuffix(hostName, "/") {
+		hostName = hostName[:len(hostName)-1]
+	}
+
+	// Clean up context path (remove leading slash)
+	contextPath := s.config.Server.ContextPath
+	if strings.HasPrefix(contextPath, "/") {
+		contextPath = contextPath[1:]
+	}
+
+	// Build the complete URL
+	var shortenedURL strings.Builder
+	shortenedURL.WriteString(hostName)
+	shortenedURL.WriteString("/")
+	shortenedURL.WriteString(contextPath)
+	if !strings.HasSuffix(contextPath, "/") {
+		shortenedURL.WriteString("/")
+	}
+	shortenedURL.WriteString(uniqueID)
+
+	return shortenedURL.String(), nil
+}
+
+// GetUserUUID retrieves user UUID by mobile number (optional functionality)
+func (s *URLService) GetUserUUID(ctx context.Context, mobileNumber string) (string, error) {
+	if mobileNumber == "" {
+		return "", fmt.Errorf("mobile number cannot be empty")
+	}
+
+	request := models.UserSearchRequest{
+		Type:     "CITIZEN",
+		TenantID: s.config.App.StateLevelTenantID,
+		UserName: mobileNumber,
+	}
+
+	jsonData, err := json.Marshal(request)
+	if err != nil {
+		return "", fmt.Errorf("failed to marshal user search request: %w", err)
+	}
+
+	url := s.config.App.UserHost + s.config.App.UserSearchPath
+	
+	// Create request with context
+	httpRequest, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
+	if err != nil {
+		return "", fmt.Errorf("failed to create HTTP request: %w", err)
+	}
+	httpRequest.Header.Set("Content-Type", "application/json")
+
+	resp, err := s.httpClient.Do(httpRequest)
+	if err != nil {
+		s.log.WithError(err).Error("Exception while fetching user")
+		return "", fmt.Errorf("failed to fetch user: %w", err)
+	}
+	defer resp.Body.Close()
+
+	if resp.StatusCode != http.StatusOK {
+		return "", fmt.Errorf("user service returned status %d", resp.StatusCode)
+	}
+
+	var response models.UserSearchResponse
+	err = json.NewDecoder(resp.Body).Decode(&response)
+	if err != nil {
+		s.log.WithError(err).Error("Failed to decode user search response")
+		return "", fmt.Errorf("failed to decode user search response: %w", err)
+	}
+
+	if len(response.User) > 0 {
+		return response.User[0].UUID, nil
+	}
+
+	return "", nil // User not found
+}
+
+// ExtractStateLevelTenant extracts state level tenant from ULB level tenant
+func (s *URLService) ExtractStateLevelTenant(ulbTenantID string) string {
+	if ulbTenantID == "" {
+		return s.config.App.StateLevelTenantID
+	}
+
+	// Example: if ULB tenant is "pb.amritsar", state tenant is "pb"
+	parts := strings.Split(ulbTenantID, ".")
+	if len(parts) > 0 && len(parts[0]) == s.config.App.StateLevelTenantIDLength {
+		return parts[0]
+	}
+	return s.config.App.StateLevelTenantID // fallback
+}
+
+// CleanupExpiredURLs removes expired URLs (if repository supports it)
+func (s *URLService) CleanupExpiredURLs(ctx context.Context) (int64, error) {
+	// Check if repository supports cleanup
+	if cleaner, ok := s.repository.(interface {
+		CleanupExpiredURLs(ctx context.Context) (int64, error)
+	}); ok {
+		return cleaner.CleanupExpiredURLs(ctx)
+	}
+
+	// If repository doesn't support cleanup, return 0
+	return 0, nil
+}
+
+// GetServiceStats returns service statistics
+func (s *URLService) GetServiceStats(ctx context.Context) map[string]interface{} {
+	stats := make(map[string]interface{})
+
+	// Add repository stats if available
+	if statsProvider, ok := s.repository.(interface {
+		GetStats() interface{}
+	}); ok {
+		stats["repository"] = statsProvider.GetStats()
+	}
+
+	// Add hash converter stats
+	stats["hashConverter"] = map[string]interface{}{
+		"salt":      s.hashConverter.GetSalt(),
+		"minLength": s.hashConverter.GetMinLength(),
+	}
+
+	// Add configuration info
+	stats["config"] = map[string]interface{}{
+		"multiInstance":     s.config.App.IsMultiInstance,
+		"databaseEnabled":   s.config.Database.Enabled,
+		"contextPath":       s.config.Server.ContextPath,
+	}
+
+	return stats
+}