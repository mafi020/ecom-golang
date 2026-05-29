package config

import (
	"log"
	"time"
)

func (c *Config) JWTAccessExpiry() time.Duration {
	d, err := time.ParseDuration(c.JWT.Expiration)
	if err != nil {
		log.Fatalf("invalid JWT_EXPIRATION value: %v", err)
	}
	return d
}

func (c *Config) JWTRefreshExpiry() time.Duration {
	d, err := time.ParseDuration(c.JWT.RefreshExpiration)
	if err != nil {
		log.Fatalf("invalid JWT_REFRESH_EXPIRATION value: %v", err)
	}
	return d
}
