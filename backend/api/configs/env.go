package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUsername string
	DBPassword string

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
	_ = godotenv.Load()

	return &Env{
		AppPort: getEnv("APP_PORT", "2910"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "omnilogs-api"),
		DBUsername: getEnv("DB_USERNAME", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),

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

		// UploadProvider:    getEnv("UPLOAD_PROVIDER", "mock"),
		// UploadMockBaseURL: getEnv("UPLOAD_MOCK_BASE_URL", "https://mock-upload.local/files"),

		// R2Endpoint:        os.Getenv("R2_ENDPOINT"),
		// R2Region:          getEnv("R2_REGION", "auto"),
		// R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		// R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		// R2Bucket:          os.Getenv("R2_BUCKET"),
		// R2PublicBaseURL:   os.Getenv("R2_PUBLIC_BASE_URL"),
		// R2ObjectPrefix:    getEnv("R2_OBJECT_PREFIX", "local/uploads"),
	}
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
