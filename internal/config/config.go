package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL    string
	ClerkSecretKey string
	ClerkJWTKey    string
	PlatePepper    string
	Port           string
	CORSOrigin     string
}

func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		ClerkSecretKey: os.Getenv("CLERK_SECRET_KEY"),
		ClerkJWTKey:    os.Getenv("CLERK_JWT_KEY"),
		PlatePepper:    os.Getenv("PLATE_PEPPER"),
		Port:           os.Getenv("PORT"),
		CORSOrigin:     os.Getenv("CORS_ORIGIN"),
	}

	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.ClerkSecretKey == "" {
		missing = append(missing, "CLERK_SECRET_KEY")
	}
	if c.ClerkJWTKey == "" {
		missing = append(missing, "CLERK_JWT_KEY")
	}
	if c.PlatePepper == "" {
		missing = append(missing, "PLATE_PEPPER")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	if c.Port == "" {
		c.Port = "8080"
	}
	if c.CORSOrigin == "" {
		c.CORSOrigin = "*"
	}

	return c, nil
}
