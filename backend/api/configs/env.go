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
	DataEncryptionKey string

	DBHost     string
	DBPort     string
	DBName     string
	DBUsername string
	DBPassword string

	DBMaxOpenConns           int
	DBMaxIdleConns           int
	DBConnMaxIdleTimeSeconds int
	DBConnMaxLifetimeSeconds int

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

	SwaggerEnabled     bool
	ElasticURL         string
	NATSURL            string
	NATSStream         string
	NATSSubject        string
	NATSConsumer       string
	NATSFetchBatchSize int
	NATSFetchMaxWaitMS int

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
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "2910"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "omnilogs-api"),
		DBUsername: getEnv("DB_USERNAME", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "admin"),

		DBMaxOpenConns:           getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:           getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxIdleTimeSeconds: getEnvInt("DB_CONN_MAX_IDLE_TIME_SECONDS", 300),
		DBConnMaxLifetimeSeconds: getEnvInt("DB_CONN_MAX_LIFETIME_SECONDS", 1800),

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

		SwaggerEnabled:     getEnvBool("SWAGGER_ENABLED", true),
		ElasticURL:         getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		NATSURL:            getEnv("NATS_URL", "nats://localhost:4222"),
		NATSStream:         getEnv("NATS_STREAM", "OMNILOGS_LOGS"),
		NATSSubject:        getEnv("NATS_SUBJECT", "omnilogs.logs.ingest"),
		NATSConsumer:       getEnv("NATS_CONSUMER", "omnilogs-worker"),
		NATSFetchBatchSize: getEnvInt("NATS_FETCH_BATCH_SIZE", 500),
		NATSFetchMaxWaitMS: getEnvInt("NATS_FETCH_MAX_WAIT_MS", 2000),

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
	if strings.TrimSpace(e.ElasticURL) == "" {
		problems = append(problems, "ELASTICSEARCH_URL is required")
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
