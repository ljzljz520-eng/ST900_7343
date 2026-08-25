package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address      string
	DatabasePath string
	ReadTimeout  int
	WriteTimeout int
	MaxBodyBytes int64
}

func Default() Config {
	return Config{Address: "127.0.0.1:8080", DatabasePath: "studio.db", ReadTimeout: 10, WriteTimeout: 10, MaxBodyBytes: 1 << 20}
}

func FromEnv() (Config, error) {
	cfg := Default()
	if value := strings.TrimSpace(os.Getenv("STUDIO_ADDRESS")); value != "" {
		cfg.Address = value
	}
	if value := strings.TrimSpace(os.Getenv("STUDIO_DB_PATH")); value != "" {
		cfg.DatabasePath = value
	}
	var err error
	if cfg.ReadTimeout, err = integerEnv("STUDIO_READ_TIMEOUT", cfg.ReadTimeout); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = integerEnv("STUDIO_WRITE_TIMEOUT", cfg.WriteTimeout); err != nil {
		return Config{}, err
	}
	if value := strings.TrimSpace(os.Getenv("STUDIO_MAX_BODY_BYTES")); value != "" {
		parsed, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil || parsed < 1024 {
			return Config{}, errors.New("STUDIO_MAX_BODY_BYTES must be at least 1024")
		}
		cfg.MaxBodyBytes = parsed
	}
	return cfg, cfg.Validate()
}

func integerEnv(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > 300 {
		return 0, fmt.Errorf("%s must be between 1 and 300", name)
	}
	return parsed, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return errors.New("server address is required")
	}
	if strings.TrimSpace(c.DatabasePath) == "" {
		return errors.New("database path is required")
	}
	if c.ReadTimeout < 1 || c.WriteTimeout < 1 {
		return errors.New("server timeouts must be positive")
	}
	if c.MaxBodyBytes < 1024 {
		return errors.New("request body limit is too small")
	}
	return nil
}
