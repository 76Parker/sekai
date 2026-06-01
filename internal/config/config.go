package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/76Parker/golib/loglib"
)

const (
	defaultEnvFile                  = ".env"
	defaultHTTPAddr                 = ":8080"
	defaultMigrationsDir            = "migrations"
	defaultMaxUncompressedSizeBytes = 100 * 1024 * 1024
)

type Config struct {
	HTTP      HTTPConfig
	Postgres  PostgresConfig
	Auth      AuthConfig
	Inspector InspectorConfig
	Log       loglib.SlogConfig
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	DatabaseURL   string
	MigrationsDir string
}

type AuthConfig struct {
	Enabled bool
	Issuer  string
	JWKSURL string
}

type InspectorConfig struct {
	MaxUncompressedSizeBytes int64
}

func Load() (Config, error) {
	if err := loadEnvFile(defaultEnvFile); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTP: HTTPConfig{
			Addr:            envOrDefault("HTTP_ADDR", defaultHTTPAddr),
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Postgres: PostgresConfig{
			DatabaseURL:   os.Getenv("DATABASE_URL"),
			MigrationsDir: envOrDefault("MIGRATIONS_DIR", defaultMigrationsDir),
		},
		Auth: AuthConfig{
			Enabled: true,
			Issuer:  os.Getenv("KEYCLOAK_ISSUER"),
			JWKSURL: os.Getenv("KEYCLOAK_JWKS_URL"),
		},
		Inspector: InspectorConfig{
			MaxUncompressedSizeBytes: defaultMaxUncompressedSizeBytes,
		},
		Log: loglib.SlogConfig{
			Level:      envOrDefault("LOG_LEVEL", "info"),
			OnlyStdout: true,
		},
	}

	maxUncompressedSize, err := envInt64("MAX_ARTIFACT_UNCOMPRESSED_SIZE_BYTES")
	if err != nil {
		return Config{}, err
	}
	if maxUncompressedSize > 0 {
		cfg.Inspector.MaxUncompressedSizeBytes = maxUncompressedSize
	}

	authEnabled, err := envBoolOrDefault("AUTH", true)
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.Enabled = authEnabled

	return cfg, cfg.validate()
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d must be KEY=VALUE", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("%s:%d key must not be empty", path, lineNumber)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.Trim(value, `"'`)
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (c Config) validate() error {
	var err error
	if c.Postgres.DatabaseURL == "" {
		err = errors.Join(err, errors.New("DATABASE_URL is required"))
	}
	if c.Postgres.MigrationsDir == "" {
		err = errors.Join(err, errors.New("MIGRATIONS_DIR is required"))
	}
	if c.Auth.Enabled && c.Auth.Issuer == "" {
		err = errors.Join(err, errors.New("KEYCLOAK_ISSUER is required"))
	}
	if c.Auth.Enabled && c.Auth.JWKSURL == "" {
		err = errors.Join(err, errors.New("KEYCLOAK_JWKS_URL is required"))
	}
	if c.Inspector.MaxUncompressedSizeBytes <= 0 {
		err = errors.Join(err, errors.New("MAX_ARTIFACT_UNCOMPRESSED_SIZE_BYTES must be greater than 0"))
	}
	return err
}

func envBoolOrDefault(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %w", key, err)
	}
	return parsed, nil
}

func envOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func envInt64(key string) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}
