package config

import "os"

type Config struct {
	HTTPAddr string
	DBDSN    string
}

func Load() Config {
	return Config{
		HTTPAddr: getEnv("BILLING_HTTP_ADDR", ":8083"),
		DBDSN:    getEnv("BILLING_DB_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
