package configs

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	AppEnv            string
	AppPort           string
	CORSAllowedOrigins string
	DataEncryptionKey string

	DBHost     string
	DBPort     string
	DBName     string
	DBUsername string
	DBPassword string
	DBSSLMode  string
	

	DBMaxOpenConns           int
	DBMaxIdleConns           int
	DBConnMaxIdleTimeSeconds int
	DBConnMaxLifetimeSeconds int
	DBSlowQueryThresholdMS   int

	APIRequestTimeoutSeconds  int
	APISlowRequestThresholdMS int

	JWTSecret                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int

	RateLimitEnabled       bool
	RateLimitRequests      int
	RateLimitWindowSeconds int

	AuthRateLimitRequests        int
	AuthRateLimitWindowSeconds   int
	UploadRateLimitRequests      int
	UploadRateLimitWindowSeconds int

	SwaggerEnabled                    bool
	ElasticURL                        string
	ElasticAPIKey                     string
	ElasticsearchQueryTimeoutSeconds  int
	ElasticsearchSlowQueryThresholdMS int
	NATSURL                           string
	NATSUserCreds                     string
	NATSCredsPath                     string
	NATSStream                        string
	NATSSubject                       string
	NATSConsumer                      string
	NATSFetchBatchSize                int
	NATSFetchMaxWaitMS                int
	NATSMaxAckPending                 int
	RetentionWorkerPollSeconds        int
	RetentionWorkerLeadSeconds        int

	MaxLogPayloadSizeBytes int
	MaxLogJSONDepth        int
	MaxLogFieldCount       int
	MaxLogStringLength     int
	MaxLogArrayLength      int

	// UploadProvider    string
	// UploadMockBaseURL string

	// R2Endpoint        string
	// R2Region          string
	// R2AccessKeyID     string
	// R2SecretAccessKey string
	// R2Bucket          string
	// R2PublicBaseURL   string
	// R2ObjectPrefix    string
}

func LoadEnv() *Env {
	loadDotEnv()

	return &Env{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "2910"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", ""),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "omnilogs-api"),
		DBUsername: getEnv("DB_USERNAME", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "admin"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		DBMaxOpenConns:           getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:           getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxIdleTimeSeconds: getEnvInt("DB_CONN_MAX_IDLE_TIME_SECONDS", 300),
		DBConnMaxLifetimeSeconds: getEnvInt("DB_CONN_MAX_LIFETIME_SECONDS", 1800),
		DBSlowQueryThresholdMS:   getEnvInt("DB_SLOW_QUERY_THRESHOLD_MS", 500),

		APIRequestTimeoutSeconds:  getEnvInt("API_REQUEST_TIMEOUT_SECONDS", 15),
		APISlowRequestThresholdMS: getEnvInt("API_SLOW_REQUEST_THRESHOLD_MS", 2000),

		JWTSecret:                 getEnv("JWT_SECRET", "change-me"),
		AccessTokenExpireSeconds:  getEnvInt("ACCESS_TOKEN_EXPIRE_SECONDS", 900),
		RefreshTokenExpireSeconds: getEnvInt("REFRESH_TOKEN_EXPIRE_SECONDS", 604800),

		RateLimitEnabled:       getEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequests:      getEnvInt("RATE_LIMIT_REQUESTS", 60),
		RateLimitWindowSeconds: getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),

		AuthRateLimitRequests:        getEnvInt("AUTH_RATE_LIMIT_REQUESTS", 10),
		AuthRateLimitWindowSeconds:   getEnvInt("AUTH_RATE_LIMIT_WINDOW_SECONDS", 60),
		UploadRateLimitRequests:      getEnvInt("UPLOAD_RATE_LIMIT_REQUESTS", 20),
		UploadRateLimitWindowSeconds: getEnvInt("UPLOAD_RATE_LIMIT_WINDOW_SECONDS", 60),

		SwaggerEnabled:                    getEnvBool("SWAGGER_ENABLED", true),
		ElasticURL:                        getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticAPIKey:                     getEnv("ELASTICSEARCH_API_KEY", ""),
		ElasticsearchQueryTimeoutSeconds:  getEnvInt("ELASTICSEARCH_QUERY_TIMEOUT_SECONDS", 10),
		ElasticsearchSlowQueryThresholdMS: getEnvInt("ELASTICSEARCH_SLOW_QUERY_THRESHOLD_MS", 1000),
		NATSURL:                           getEnv("NATS_URL", "nats://localhost:4222"),
		NATSUserCreds:                     getEnv("NATS_USER_CREDS", ""),
		NATSCredsPath:                     getEnv("NATS_CREDS_PATH", getEnv("NATS_USER_CREDS", "")),
		NATSStream:                        getEnv("NATS_STREAM", "OMNILOGS_LOGS"),
		NATSSubject:                       getEnv("NATS_SUBJECT", "omnilogs.logs.ingest"),
		NATSConsumer:                      getEnv("NATS_CONSUMER", "omnilogs-worker"),
		NATSFetchBatchSize:                getEnvInt("NATS_FETCH_BATCH_SIZE", 100),
		NATSFetchMaxWaitMS:                getEnvInt("NATS_FETCH_MAX_WAIT_MS", 250),
		NATSMaxAckPending:                 getEnvInt("NATS_MAX_ACK_PENDING", 200),
		RetentionWorkerPollSeconds:        getEnvInt("RETENTION_WORKER_POLL_SECONDS", 30),
		RetentionWorkerLeadSeconds:        getEnvInt("RETENTION_WORKER_LEAD_SECONDS", 30),

		MaxLogPayloadSizeBytes: getEnvInt("MAX_LOG_PAYLOAD_SIZE_BYTES", 262144),
		MaxLogJSONDepth:        getEnvInt("MAX_LOG_JSON_DEPTH", 20),
		MaxLogFieldCount:       getEnvInt("MAX_LOG_FIELD_COUNT", 500),
		MaxLogStringLength:     getEnvInt("MAX_LOG_STRING_LENGTH", 65536),
		MaxLogArrayLength:      getEnvInt("MAX_LOG_ARRAY_LENGTH", 1000),

		DataEncryptionKey: getEnv("DATA_ENCRYPTION_KEY", ""),
	}
}

func loadDotEnv() {
	candidates := []string{
		".env",
		"api/.env",
		"backend/api/.env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "api", ".env"),
	}

	seen := make(map[string]struct{}, len(candidates))
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		if _, err := os.Stat(cleaned); err == nil {
			seen[cleaned] = struct{}{}
			paths = append(paths, cleaned)
		}
	}

	if len(paths) == 0 {
		return
	}

	_ = godotenv.Load(paths...)
}

func (e *Env) Validate() error {
	var problems []string

	if strings.TrimSpace(e.AppPort) == "" {
		problems = append(problems, "APP_PORT is required")
	}
	if strings.TrimSpace(e.DBHost) == "" {
		problems = append(problems, "DB_HOST is required")
	}
	if strings.TrimSpace(e.DBPort) == "" {
		problems = append(problems, "DB_PORT is required")
	}
	if strings.TrimSpace(e.DBName) == "" {
		problems = append(problems, "DB_NAME is required")
	}
	if strings.TrimSpace(e.DBUsername) == "" {
		problems = append(problems, "DB_USERNAME is required")
	}
	if strings.TrimSpace(e.JWTSecret) == "" || e.JWTSecret == "change-me" {
		problems = append(problems, "JWT_SECRET must be configured")
	}
	if e.AccessTokenExpireSeconds <= 0 {
		problems = append(problems, "ACCESS_TOKEN_EXPIRE_SECONDS must be greater than 0")
	}
	if e.RefreshTokenExpireSeconds <= 0 {
		problems = append(problems, "REFRESH_TOKEN_EXPIRE_SECONDS must be greater than 0")
	}
	if e.APIRequestTimeoutSeconds <= 0 {
		problems = append(problems, "API_REQUEST_TIMEOUT_SECONDS must be greater than 0")
	}
	if e.APISlowRequestThresholdMS <= 0 {
		problems = append(problems, "API_SLOW_REQUEST_THRESHOLD_MS must be greater than 0")
	}
	if e.DBSlowQueryThresholdMS <= 0 {
		problems = append(problems, "DB_SLOW_QUERY_THRESHOLD_MS must be greater than 0")
	}
	if strings.TrimSpace(e.ElasticURL) == "" {
		problems = append(problems, "ELASTICSEARCH_URL is required")
	}
	if e.ElasticsearchQueryTimeoutSeconds <= 0 {
		problems = append(problems, "ELASTICSEARCH_QUERY_TIMEOUT_SECONDS must be greater than 0")
	}
	if e.ElasticsearchSlowQueryThresholdMS <= 0 {
		problems = append(problems, "ELASTICSEARCH_SLOW_QUERY_THRESHOLD_MS must be greater than 0")
	}
	if strings.TrimSpace(e.NATSURL) == "" {
		problems = append(problems, "NATS_URL is required")
	}
	if strings.TrimSpace(e.NATSStream) == "" {
		problems = append(problems, "NATS_STREAM is required")
	}
	if strings.TrimSpace(e.NATSSubject) == "" {
		problems = append(problems, "NATS_SUBJECT is required")
	}
	if strings.TrimSpace(e.NATSConsumer) == "" {
		problems = append(problems, "NATS_CONSUMER is required")
	}
	if e.NATSFetchBatchSize <= 0 {
		problems = append(problems, "NATS_FETCH_BATCH_SIZE must be greater than 0")
	}
	if e.NATSFetchMaxWaitMS <= 0 {
		problems = append(problems, "NATS_FETCH_MAX_WAIT_MS must be greater than 0")
	}
	if e.NATSMaxAckPending <= 0 {
		problems = append(problems, "NATS_MAX_ACK_PENDING must be greater than 0")
	}

	if len(problems) == 0 {
		if strings.TrimSpace(e.DataEncryptionKey) == "" {
			problems = append(problems, "DATA_ENCRYPTION_KEY is required")
		}
		if len([]byte(e.DataEncryptionKey)) != 32 {
			problems = append(problems, "DATA_ENCRYPTION_KEY must be exactly 32 bytes")
		}
	}

	if len(problems) == 0 {
		return nil
	}

	return errors.New(strings.Join(problems, "; "))
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
