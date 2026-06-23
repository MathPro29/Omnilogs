package configs

import "testing"

func TestEnvValidate(t *testing.T) {
	valid := &Env{
		AppEnv:                    "development",
		AppPort:                   "2910",
		DBHost:                    "localhost",
		DBPort:                    "5432",
		DBName:                    "omnilogs",
		DBUsername:                "postgres",
		DBPassword:                "postgres",
		JWTSecret:                 "super-secret",
		AccessTokenExpireSeconds:  900,
		RefreshTokenExpireSeconds: 604800,
		ElasticURL:                "http://localhost:9200",
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected env to be valid, got error: %v", err)
	}

	invalid := &Env{
		AppPort:                   "",
		DBHost:                    "",
		DBPort:                    "",
		DBName:                    "",
		DBUsername:                "",
		JWTSecret:                 "change-me",
		AccessTokenExpireSeconds:  0,
		RefreshTokenExpireSeconds: 0,
	}

	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid env to fail validation")
	}
}
