package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	User     string
	Password string
	DB       string
	Host     string
	Port     string
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password, c.Host, c.Port, c.DB)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadConfig() Config {
	config := Config{
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DB:       os.Getenv("POSTGRES_DB"),
		Host:     getenv("POSTGRES_HOST", "localhost"),
		Port:     getenv("POSTGRES_PORT", "5432"),
	}

	var missing []string
	if config.User == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if config.Password == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if config.DB == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if len(missing) > 0 {
		missingFields := strings.Join(missing, ", ")
		log.Fatalf("missing required environment variables: %s", missingFields)
	}

	return config
}
