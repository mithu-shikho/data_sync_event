// Package config loads and validates environment-driven configuration.
package config

// Config holds all runtime configuration. Populated in phase-01.
type Config struct {
	ControlMongoURI string
	NATSURL         string
	NumPartitions   int
	JWTSecret       string

	OAuthClientID     string
	OAuthClientSecret string
	OAuthRedirectURL  string
	SessionKey        string
	AllowlistDomain   string

	HTTPAddr string
}

// TODO(phase-01): Load() reads from env, applies defaults, and validates.
