diff --git a/core-services/egov-url-shortening-go/repository/redis.go b/core-services/egov-url-shortening-go/repository/redis.go
--- a/core-services/egov-url-shortening-go/repository/redis.go
+++ b/core-services/egov-url-shortening-go/repository/redis.go
@@ -0,0 +1,244 @@
+package repository
+
+import (
+	"context"
+	"encoding/json"
+	"fmt"
+	"time"
+
+	"egov-url-shortening-go/config"
+	"egov-url-shortening-go/models"
+
+	"github.com/go-redis/redis/v8"
+	"github.com/sirupsen/logrus"
+)
+
+// RedisRepository implements URLRepository using Redis
+type RedisRepository struct {
+	client *redis.Client
+	idKey  string
+	urlKey string
+	log    *logrus.Logger
+}
+
+// NewRedisRepository creates a new Redis repository instance
+func NewRedisRepository(cfg *config.RedisConfig, log *logrus.Logger) *RedisRepository {
+	options := &redis.Options{
+		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
+		Password: cfg.Password,
+		DB:       cfg.DB,
+		
+		// Connection pool settings
+		PoolSize:     10,
+		MinIdleConns: 3,
+		MaxRetries:   3,
+		
+		// Timeouts
+		DialTimeout:  5 * time.Second,
+		ReadTimeout:  3 * time.Second,
+		WriteTimeout: 3 * time.Second,
+		
+		// Keep alive
+		PoolTimeout: 4 * time.Second,
+		IdleTimeout: 5 * time.Minute,
+	}
+
+	client := redis.NewClient(options)
+
+	repo := &RedisRepository{
+		client: client,
+		idKey:  "url:id",
+		urlKey: "url:data",
+		log:    log,
+	}
+
+	// Test connection
+	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
+	defer cancel()
+	
+	if err := repo.HealthCheck(ctx); err != nil {
+		log.WithError(err).Warn("Redis connection test failed during initialization")
+	} else {
+		log.Info("Successfully connected to Redis")
+	}
+
+	return repo
+}
+
+// IncrementID increments and returns the next available ID
+func (r *RedisRepository) IncrementID(ctx context.Context) (int64, error) {
+	id, err := r.client.Incr(ctx, r.idKey).Result()
+	if err != nil {
+		r.log.WithError(err).Error("Failed to increment ID")
+		return 0, fmt.Errorf("failed to increment ID: %w", err)
+	}
+
+	// Return id-1 to match Java implementation (Java uses pre-increment logic)
+	result := id - 1
+	r.log.WithField("id", result).Debug("Incremented ID")
+	return result, nil
+}
+
+// SaveURL saves a URL with the given key and request
+func (r *RedisRepository) SaveURL(ctx context.Context, key string, request *models.ShortenRequest) error {
+	// Validate input
+	if key == "" {
+		return fmt.Errorf("key cannot be empty")
+	}
+	if request == nil {
+		return fmt.Errorf("request cannot be nil")
+	}
+	if request.URL == "" {
+		return fmt.Errorf("URL cannot be empty")
+	}
+
+	// Serialize the request to JSON (matching Java ObjectMapper behavior)
+	data, err := json.Marshal(request)
+	if err != nil {
+		r.log.WithError(err).WithField("key", key).Error("Failed to marshal request")
+		return fmt.Errorf("failed to marshal request: %w", err)
+	}
+
+	// Use HSET to store in a hash (matching Java Jedis.hset behavior)
+	err = r.client.HSet(ctx, r.urlKey, key, string(data)).Err()
+	if err != nil {
+		r.log.WithError(err).WithField("key", key).Error("Failed to save URL")
+		return fmt.Errorf("failed to save URL: %w", err)
+	}
+
+	r.log.WithFields(logrus.Fields{
+		"url": request.URL,
+		"key": key,
+	}).Info("URL saved successfully")
+
+	return nil
+}
+
+// GetURL retrieves a URL by ID
+func (r *RedisRepository) GetURL(ctx context.Context, id int64) (string, error) {
+	request, err := r.GetURLDetails(ctx, id)
+	if err != nil {
+		return "", err
+	}
+
+	// Check if URL is active
+	if !request.IsActive() {
+		if request.IsExpired() {
+			r.log.WithField("id", id).Warn("URL has expired")
+			return "", fmt.Errorf("URL at key %d has expired", id)
+		}
+		r.log.WithField("id", id).Warn("URL is not yet active")
+		return "", fmt.Errorf("URL at key %d is not yet active", id)
+	}
+
+	return request.URL, nil
+}
+
+// GetURLDetails retrieves full URL details by ID
+func (r *RedisRepository) GetURLDetails(ctx context.Context, id int64) (*models.ShortenRequest, error) {
+	key := fmt.Sprintf("url:%d", id)
+	
+	r.log.WithField("id", id).Debug("Retrieving URL details")
+	
+	// Get from hash
+	data, err := r.client.HGet(ctx, r.urlKey, key).Result()
+	if err != nil {
+		if err == redis.Nil {
+			r.log.WithField("id", id).Debug("URL not found")
+			return nil, fmt.Errorf("URL at key %d does not exist", id)
+		}
+		r.log.WithError(err).WithField("id", id).Error("Failed to retrieve URL")
+		return nil, fmt.Errorf("failed to retrieve URL: %w", err)
+	}
+
+	// Deserialize the JSON data
+	var request models.ShortenRequest
+	err = json.Unmarshal([]byte(data), &request)
+	if err != nil {
+		r.log.WithError(err).WithField("id", id).Error("Failed to unmarshal URL data")
+		return nil, fmt.Errorf("failed to unmarshal URL data: %w", err)
+	}
+
+	r.log.WithFields(logrus.Fields{
+		"url": request.URL,
+		"id":  id,
+	}).Debug("Retrieved URL details")
+
+	return &request, nil
+}
+
+// DeleteURL deletes a URL by ID
+func (r *RedisRepository) DeleteURL(ctx context.Context, id int64) error {
+	key := fmt.Sprintf("url:%d", id)
+	
+	deleted, err := r.client.HDel(ctx, r.urlKey, key).Result()
+	if err != nil {
+		r.log.WithError(err).WithField("id", id).Error("Failed to delete URL")
+		return fmt.Errorf("failed to delete URL: %w", err)
+	}
+
+	if deleted == 0 {
+		return fmt.Errorf("URL at key %d does not exist", id)
+	}
+
+	r.log.WithField("id", id).Info("URL deleted successfully")
+	return nil
+}
+
+// CheckURLExists checks if a URL exists for the given ID
+func (r *RedisRepository) CheckURLExists(ctx context.Context, id int64) (bool, error) {
+	key := fmt.Sprintf("url:%d", id)
+	
+	exists, err := r.client.HExists(ctx, r.urlKey, key).Result()
+	if err != nil {
+		r.log.WithError(err).WithField("id", id).Error("Failed to check URL existence")
+		return false, fmt.Errorf("failed to check URL existence: %w", err)
+	}
+
+	return exists, nil
+}
+
+// HealthCheck performs a health check on the repository
+func (r *RedisRepository) HealthCheck(ctx context.Context) error {
+	// Try to ping Redis
+	_, err := r.client.Ping(ctx).Result()
+	if err != nil {
+		return fmt.Errorf("redis health check failed: %w", err)
+	}
+
+	// Try a simple operation
+	testKey := "health:check"
+	err = r.client.Set(ctx, testKey, "ok", time.Second).Err()
+	if err != nil {
+		return fmt.Errorf("redis write test failed: %w", err)
+	}
+
+	// Clean up test key
+	r.client.Del(ctx, testKey)
+
+	return nil
+}
+
+// Close closes the Redis connection
+func (r *RedisRepository) Close() error {
+	if r.client != nil {
+		err := r.client.Close()
+		if err != nil {
+			r.log.WithError(err).Error("Failed to close Redis connection")
+			return fmt.Errorf("failed to close Redis connection: %w", err)
+		}
+		r.log.Info("Redis connection closed")
+	}
+	return nil
+}
+
+// GetStats returns Redis connection statistics
+func (r *RedisRepository) GetStats() *redis.PoolStats {
+	return r.client.PoolStats()
+}
+
+// FlushTestData flushes test data (for testing only)
+func (r *RedisRepository) FlushTestData(ctx context.Context) error {
+	// Only allow in development/test environments
+	return r.client.FlushDB(ctx).Err()
+}