diff --git a/core-services/egov-url-shortening-go/config/config.go b/core-services/egov-url-shortening-go/config/config.go
--- a/core-services/egov-url-shortening-go/config/config.go
+++ b/core-services/egov-url-shortening-go/config/config.go
@@ -0,0 +1,163 @@
+package config
+
+import (
+	"encoding/json"
+	"fmt"
+	"log"
+
+	"github.com/kelseyhightower/envconfig"
+)
+
+// Config holds all configuration for the URL shortening service
+type Config struct {
+	Server   ServerConfig   `envconfig:"SERVER"`
+	Redis    RedisConfig    `envconfig:"REDIS"`
+	Database DatabaseConfig `envconfig:"DATABASE"`
+	Kafka    KafkaConfig    `envconfig:"KAFKA"`
+	App      AppConfig      `envconfig:"APP"`
+	HashIDs  HashIDsConfig  `envconfig:"HASHIDS"`
+}
+
+// ServerConfig holds server configuration
+type ServerConfig struct {
+	Port        int    `envconfig:"PORT" default:"8091"`
+	ContextPath string `envconfig:"CONTEXT_PATH" default:"/egov-url-shortening"`
+}
+
+// RedisConfig holds Redis configuration
+type RedisConfig struct {
+	Host     string `envconfig:"HOST" default:"localhost"`
+	Port     int    `envconfig:"PORT" default:"6379"`
+	Password string `envconfig:"PASSWORD" default:""`
+	DB       int    `envconfig:"DB" default:"0"`
+}
+
+// DatabaseConfig holds database configuration
+type DatabaseConfig struct {
+	Host     string `envconfig:"HOST" default:"localhost"`
+	Port     int    `envconfig:"PORT" default:"5432"`
+	Name     string `envconfig:"NAME" default:"devdb"`
+	Username string `envconfig:"USERNAME" default:"postgres"`
+	Password string `envconfig:"PASSWORD" default:"postgres"`
+	Enabled  bool   `envconfig:"ENABLED" default:"true"`
+	SSLMode  string `envconfig:"SSL_MODE" default:"disable"`
+}
+
+// KafkaConfig holds Kafka configuration
+type KafkaConfig struct {
+	BootstrapServers string `envconfig:"BOOTSTRAP_SERVERS" default:"localhost:9092"`
+	Topic            string `envconfig:"TOPIC" default:"save-url-shortening-details"`
+	Enabled          bool   `envconfig:"ENABLED" default:"false"`
+}
+
+// AppConfig holds application-specific configuration
+type AppConfig struct {
+	HostName                 string            `envconfig:"HOST_NAME" default:"https://qa.digit.org/"`
+	StateLevelTenantID       string            `envconfig:"STATE_LEVEL_TENANT_ID" default:"pb"`
+	UserHost                 string            `envconfig:"USER_HOST" default:"http://egov-user.egov:8080/"`
+	UserSearchPath           string            `envconfig:"USER_SEARCH_PATH" default:"user/_search"`
+	UIAppHostMap             string            `envconfig:"UI_APP_HOST_MAP" default:"{\"in\":\"https://central-instance.digit.org\",\"in.statea\":\"https://statea.digit.org\"}"`
+	UIAppHostMapParsed       map[string]string
+	IsMultiInstance          bool `envconfig:"IS_MULTI_INSTANCE" default:"false"`
+	IsCentralInstance        bool `envconfig:"IS_CENTRAL_INSTANCE" default:"true"`
+	StateLevelTenantIDLength int  `envconfig:"STATE_LEVEL_TENANT_ID_LENGTH" default:"2"`
+}
+
+// HashIDsConfig holds HashIDs configuration
+type HashIDsConfig struct {
+	Salt      string `envconfig:"SALT" default:"randomsalt"`
+	MinLength int    `envconfig:"MIN_LENGTH" default:"3"`
+}
+
+// LoadConfig loads configuration from environment variables
+func LoadConfig() (*Config, error) {
+	var cfg Config
+	err := envconfig.Process("", &cfg)
+	if err != nil {
+		return nil, fmt.Errorf("failed to process environment config: %w", err)
+	}
+
+	// Validate configuration
+	if err := cfg.validate(); err != nil {
+		return nil, fmt.Errorf("configuration validation failed: %w", err)
+	}
+
+	// Parse the UI app host map JSON
+	if cfg.App.UIAppHostMap != "" {
+		err = json.Unmarshal([]byte(cfg.App.UIAppHostMap), &cfg.App.UIAppHostMapParsed)
+		if err != nil {
+			log.Printf("Warning: Failed to parse UI app host map: %v", err)
+			cfg.App.UIAppHostMapParsed = make(map[string]string)
+		}
+	} else {
+		cfg.App.UIAppHostMapParsed = make(map[string]string)
+	}
+
+	return &cfg, nil
+}
+
+// validate checks if the configuration is valid
+func (c *Config) validate() error {
+	// Validate server port
+	if c.Server.Port <= 0 || c.Server.Port > 65535 {
+		return fmt.Errorf("invalid server port: %d", c.Server.Port)
+	}
+
+	// Validate context path
+	if c.Server.ContextPath == "" {
+		return fmt.Errorf("context path cannot be empty")
+	}
+
+	// Validate Redis config if database is disabled
+	if !c.Database.Enabled {
+		if c.Redis.Host == "" {
+			return fmt.Errorf("redis host cannot be empty when database is disabled")
+		}
+		if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
+			return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
+		}
+	}
+
+	// Validate database config if enabled
+	if c.Database.Enabled {
+		if c.Database.Host == "" {
+			return fmt.Errorf("database host cannot be empty when database is enabled")
+		}
+		if c.Database.Port <= 0 || c.Database.Port > 65535 {
+			return fmt.Errorf("invalid database port: %d", c.Database.Port)
+		}
+		if c.Database.Name == "" {
+			return fmt.Errorf("database name cannot be empty")
+		}
+		if c.Database.Username == "" {
+			return fmt.Errorf("database username cannot be empty")
+		}
+	}
+
+	// Validate HashIDs config
+	if c.HashIDs.Salt == "" {
+		return fmt.Errorf("hashids salt cannot be empty")
+	}
+	if c.HashIDs.MinLength < 1 {
+		return fmt.Errorf("hashids min length must be at least 1")
+	}
+
+	// Validate app config
+	if c.App.HostName == "" {
+		return fmt.Errorf("app host name cannot be empty")
+	}
+	if c.App.StateLevelTenantID == "" {
+		return fmt.Errorf("state level tenant ID cannot be empty")
+	}
+	if c.App.StateLevelTenantIDLength < 1 {
+		return fmt.Errorf("state level tenant ID length must be at least 1")
+	}
+
+	return nil
+}
+
+// GetConnectionString returns the database connection string
+func (d *DatabaseConfig) GetConnectionString() string {
+	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
+		d.Username, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
+}