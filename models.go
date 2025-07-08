diff --git a/core-services/egov-url-shortening-go/models/models.go b/core-services/egov-url-shortening-go/models/models.go
--- a/core-services/egov-url-shortening-go/models/models.go
+++ b/core-services/egov-url-shortening-go/models/models.go
@@ -0,0 +1,195 @@
+package models
+
+import (
+	"encoding/json"
+	"time"
+)
+
+// ShortenRequest represents the request payload for URL shortening
+type ShortenRequest struct {
+	ID        string `json:"id,omitempty"`
+	URL       string `json:"url" binding:"required,url" validate:"required,url"`
+	ValidFrom *int64 `json:"validFrom,omitempty"`
+	ValidTill *int64 `json:"validTill,omitempty"`
+}
+
+// Validate checks if the ShortenRequest is valid
+func (s *ShortenRequest) Validate() error {
+	if s.URL == "" {
+		return &ValidationError{Field: "url", Message: "URL is required"}
+	}
+	
+	// Check if ValidFrom is before ValidTill
+	if s.ValidFrom != nil && s.ValidTill != nil {
+		if *s.ValidFrom >= *s.ValidTill {
+			return &ValidationError{Field: "validTill", Message: "ValidTill must be after ValidFrom"}
+		}
+	}
+	
+	return nil
+}
+
+// IsExpired checks if the URL has expired
+func (s *ShortenRequest) IsExpired() bool {
+	if s.ValidTill == nil {
+		return false
+	}
+	
+	now := time.Now().UnixMilli()
+	return now > *s.ValidTill
+}
+
+// IsActive checks if the URL is currently active
+func (s *ShortenRequest) IsActive() bool {
+	now := time.Now().UnixMilli()
+	
+	// Check if it's after ValidFrom (if set)
+	if s.ValidFrom != nil && now < *s.ValidFrom {
+		return false
+	}
+	
+	// Check if it's before ValidTill (if set)
+	if s.ValidTill != nil && now > *s.ValidTill {
+		return false
+	}
+	
+	return true
+}
+
+// ShortenResponse represents the response for URL shortening
+type ShortenResponse struct {
+	ShortenedURL string `json:"shortenedUrl"`
+	ID           string `json:"id,omitempty"`
+	OriginalURL  string `json:"originalUrl,omitempty"`
+	ExpiresAt    *int64 `json:"expiresAt,omitempty"`
+}
+
+// URLEntry represents a URL entry in the database
+type URLEntry struct {
+	ID        string     `json:"id" db:"id"`
+	URL       string     `json:"url" db:"url"`
+	ValidFrom *time.Time `json:"validFrom" db:"validform"`
+	ValidTo   *time.Time `json:"validTo" db:"validto"`
+	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
+	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
+}
+
+// ToShortenRequest converts URLEntry to ShortenRequest
+func (u *URLEntry) ToShortenRequest() *ShortenRequest {
+	req := &ShortenRequest{
+		ID:  u.ID,
+		URL: u.URL,
+	}
+	
+	if u.ValidFrom != nil {
+		validFrom := u.ValidFrom.UnixMilli()
+		req.ValidFrom = &validFrom
+	}
+	
+	if u.ValidTo != nil {
+		validTo := u.ValidTo.UnixMilli()
+		req.ValidTill = &validTo
+	}
+	
+	return req
+}
+
+// ErrorResponse represents an error response
+type ErrorResponse struct {
+	Code      string `json:"code"`
+	Message   string `json:"message"`
+	RequestID string `json:"requestId,omitempty"`
+	Timestamp int64  `json:"timestamp"`
+}
+
+// NewErrorResponse creates a new error response
+func NewErrorResponse(code, message string) *ErrorResponse {
+	return &ErrorResponse{
+		Code:      code,
+		Message:   message,
+		Timestamp: time.Now().UnixMilli(),
+	}
+}
+
+// ValidationError represents a validation error
+type ValidationError struct {
+	Field   string `json:"field"`
+	Message string `json:"message"`
+}
+
+// Error implements the error interface
+func (v *ValidationError) Error() string {
+	return v.Message
+}
+
+// UserSearchRequest represents a user search request
+type UserSearchRequest struct {
+	Type     string `json:"type" binding:"required"`
+	TenantID string `json:"tenantId" binding:"required"`
+	UserName string `json:"userName" binding:"required"`
+}
+
+// UserSearchResponse represents a user search response
+type UserSearchResponse struct {
+	User []UserInfo `json:"user"`
+}
+
+// UserInfo represents user information
+type UserInfo struct {
+	UUID string `json:"uuid"`
+	Name string `json:"name,omitempty"`
+	Type string `json:"type,omitempty"`
+}
+
+// HealthCheckResponse represents health check response
+type HealthCheckResponse struct {
+	Status    string            `json:"status"`
+	Service   string            `json:"service"`
+	Version   string            `json:"version,omitempty"`
+	Timestamp int64             `json:"timestamp"`
+	Checks    map[string]string `json:"checks,omitempty"`
+}
+
+// NewHealthCheckResponse creates a new health check response
+func NewHealthCheckResponse(service, version string) *HealthCheckResponse {
+	return &HealthCheckResponse{
+		Status:    "UP",
+		Service:   service,
+		Version:   version,
+		Timestamp: time.Now().UnixMilli(),
+		Checks:    make(map[string]string),
+	}
+}
+
+// AddCheck adds a health check result
+func (h *HealthCheckResponse) AddCheck(name, status string) {
+	if h.Checks == nil {
+		h.Checks = make(map[string]string)
+	}
+	h.Checks[name] = status
+}
+
+// Analytics represents URL analytics data
+type Analytics struct {
+	ShortURL    string `json:"shortUrl"`
+	OriginalURL string `json:"originalUrl"`
+	Clicks      int64  `json:"clicks"`
+	CreatedAt   int64  `json:"createdAt"`
+	LastClicked *int64 `json:"lastClicked,omitempty"`
+}
+
+// String returns JSON representation of any model
+func (s *ShortenRequest) String() string {
+	data, _ := json.Marshal(s)
+	return string(data)
+}
+
+func (s *ShortenResponse) String() string {
+	data, _ := json.Marshal(s)
+	return string(data)
+}
+
+func (e *ErrorResponse) String() string {
+	data, _ := json.Marshal(e)
+	return string(data)
+}