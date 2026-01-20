package config

import "os"

type Config struct {
	HTTPAddr       string
	JWTSecret      string
	UserServiceURL string
}

func Load() Config {
	return Config{
		HTTPAddr:       getEnv("AUTH_HTTP_ADDR", ":8080"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-secret"),
		UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
