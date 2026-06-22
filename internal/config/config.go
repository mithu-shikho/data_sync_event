// Package config loads and validates environment-driven configuration.
package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration.
type Config struct {
	// Control plane / persistence.
	ControlMongoURI string
	ControlDBName   string

	// Data plane.
	NATSURL       string
	NumPartitions int

	// URI encryption: hex-encoded 32-byte key for AES-256-GCM. Empty disables
	// encryption (connection URIs stored in plaintext — dev only).
	URIEncryptionKey []byte

	// Auth.
	JWTSecret         string
	OAuthClientID     string
	OAuthClientSecret string
	OAuthRedirectURL  string
	SessionKey        string
	AllowlistDomain   string

	// HTTP.
	HTTPAddr string
}

// Load reads configuration from the environment, applies defaults, and validates.
func Load() (*Config, error) {
	cfg := &Config{
		ControlMongoURI:   env("CONTROL_MONGO_URI", "mongodb://localhost:27017"),
		ControlDBName:     env("CONTROL_DB_NAME", "data_sync_control"),
		NATSURL:           env("NATS_URL", "nats://localhost:4222"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		OAuthClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		OAuthClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		OAuthRedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		SessionKey:        os.Getenv("SESSION_KEY"),
		AllowlistDomain:   os.Getenv("ALLOWLIST_DOMAIN"),
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
	}

	parts, err := strconv.Atoi(env("NUM_PARTITIONS", "16"))
	if err != nil {
		return nil, fmt.Errorf("NUM_PARTITIONS: %w", err)
	}
	if parts < 1 {
		return nil, fmt.Errorf("NUM_PARTITIONS must be >= 1, got %d", parts)
	}
	cfg.NumPartitions = parts

	if raw := os.Getenv("URI_ENCRYPTION_KEY"); raw != "" {
		key, err := hex.DecodeString(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("URI_ENCRYPTION_KEY must be hex-encoded: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("URI_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
		}
		cfg.URIEncryptionKey = key
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.ControlMongoURI == "" {
		return fmt.Errorf("CONTROL_MONGO_URI is required")
	}
	if c.ControlDBName == "" {
		return fmt.Errorf("CONTROL_DB_NAME is required")
	}
	if c.NATSURL == "" {
		return fmt.Errorf("NATS_URL is required")
	}
	return nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
