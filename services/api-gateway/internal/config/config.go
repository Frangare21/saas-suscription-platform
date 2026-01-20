package config

import "os"

type Config struct {
	HTTPAddr          string
	JWTSecret         string
	AuthServiceURL    string
	UserServiceURL    string
	BillingServiceURL string
}

func Load() Config {
	return Config{
		HTTPAddr:          getEnv("GATEWAY_HTTP_ADDR", ":8080"),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://user-service:8081"),
		BillingServiceURL: getEnv("BILLING_SERVICE_URL", "http://billing-service:8083"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
