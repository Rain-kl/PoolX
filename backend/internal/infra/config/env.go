package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Environment variable names (FOAM_*). Process env wins over config.yaml and defaults.
// Optional .env files only fill unset keys (they never override an already-set process env).
const (
	EnvConfigPath = "FOAM_CONFIG" // config file path (also used by CLI default)

	EnvServerListen         = "FOAM_SERVER_LISTEN"
	EnvServerMaxBodyBytes   = "FOAM_SERVER_MAX_BODY_BYTES"
	EnvServerReadTimeout    = "FOAM_SERVER_READ_TIMEOUT"
	EnvServerRequestTimeout = "FOAM_SERVER_REQUEST_TIMEOUT"

	EnvAuthAccessTokenTTL  = "FOAM_AUTH_ACCESS_TOKEN_TTL"
	EnvAuthRefreshTokenTTL = "FOAM_AUTH_REFRESH_TOKEN_TTL"
	EnvAuthSecureCookies   = "FOAM_AUTH_SECURE_COOKIES"

	EnvSecretsJWTSecret               = "FOAM_SECRETS_JWT_SECRET"
	EnvSecretsCredentialEncryptionKey = "FOAM_SECRETS_CREDENTIAL_ENCRYPTION_KEY"

	EnvBootstrapAdminUsername = "FOAM_BOOTSTRAP_ADMIN_USERNAME"
	EnvBootstrapAdminPassword = "FOAM_BOOTSTRAP_ADMIN_PASSWORD"

	EnvFrontendPublicAPIBaseURL = "FOAM_FRONTEND_PUBLIC_API_BASE_URL"
	EnvFrontendStaticPath       = "FOAM_FRONTEND_STATIC_PATH"

	EnvDatabaseDriver               = "FOAM_DATABASE_DRIVER"
	EnvDatabaseSQLitePath           = "FOAM_DATABASE_SQLITE_PATH"
	EnvDatabasePostgresDSN          = "FOAM_DATABASE_POSTGRES_DSN"
	EnvDatabasePostgresMaxOpenConns = "FOAM_DATABASE_POSTGRES_MAX_OPEN_CONNS"
	EnvDatabasePostgresMaxIdleConns = "FOAM_DATABASE_POSTGRES_MAX_IDLE_CONNS"

	// EnvClashMihomoBinaryPath overrides the default mihomo binary path
	// (./data/core/mihomo). Docker images set this to /opt/mihomo.
	EnvClashMihomoBinaryPath = "FOAM_CLASH_MIHOMO_BINARY_PATH"
)

// loadDotEnvIfPresent loads KEY=VALUE pairs from the first existing candidate
// into the process environment, without overriding variables already set.
// Candidates: FOAM_ENV_FILE, .env, ../.env (relative to cwd).
func loadDotEnvIfPresent() {
	candidates := make([]string, 0, 3)
	if custom := strings.TrimSpace(os.Getenv("FOAM_ENV_FILE")); custom != "" {
		candidates = append(candidates, custom)
	}
	candidates = append(candidates, ".env", filepath.Join("..", ".env"))
	for _, path := range candidates {
		if err := loadDotEnvFile(path); err == nil {
			return
		}
	}
}

func loadDotEnvFile(path string) error {
	file, err := os.Open(path) //nolint:gosec // path is FOAM_ENV_FILE or a fixed .env candidate
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	const (
		scannerInitialCap = 64 * 1024
		scannerMaxCap     = 1024 * 1024
		minQuotedLen      = 2
	)
	scanner := bufio.NewScanner(file)
	// Allow long base64 keys.
	scanner.Buffer(make([]byte, 0, scannerInitialCap), scannerMaxCap)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= minQuotedLen {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}

// applyEnvOverrides overlays FOAM_* environment variables onto cfg.
// Priority for each field: non-empty env > previous value (yaml or default).
// Empty env values are ignored so docker-compose "${VAR:-}" placeholders do not wipe YAML.
func applyEnvOverrides(cfg *Config) error {
	var err error
	setString := func(key string, dest *string) {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			*dest = v
		}
	}
	setBool := func(key string, dest *bool) {
		if err != nil {
			return
		}
		v := strings.TrimSpace(os.Getenv(key))
		if v == "" {
			return
		}
		parsed, parseErr := parseBool(v)
		if parseErr != nil {
			err = fmt.Errorf("%s: %w", key, parseErr)
			return
		}
		*dest = parsed
	}
	setInt := func(key string, dest *int) {
		if err != nil {
			return
		}
		v := strings.TrimSpace(os.Getenv(key))
		if v == "" {
			return
		}
		parsed, parseErr := strconv.Atoi(v)
		if parseErr != nil {
			err = fmt.Errorf("%s: 必须是整数", key)
			return
		}
		*dest = parsed
	}
	setInt64 := func(key string, dest *int64) {
		if err != nil {
			return
		}
		v := strings.TrimSpace(os.Getenv(key))
		if v == "" {
			return
		}
		parsed, parseErr := strconv.ParseInt(v, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("%s: 必须是整数", key)
			return
		}
		*dest = parsed
	}
	setDuration := func(key string, dest *Duration) {
		if err != nil {
			return
		}
		v := strings.TrimSpace(os.Getenv(key))
		if v == "" {
			return
		}
		parsed, parseErr := time.ParseDuration(v)
		if parseErr != nil {
			err = fmt.Errorf("%s: 必须是合法 duration（如 15m、2h）", key)
			return
		}
		*dest = Duration(parsed)
	}

	setString(EnvServerListen, &cfg.Server.Listen)
	setInt64(EnvServerMaxBodyBytes, &cfg.Server.MaxBodyBytes)
	setDuration(EnvServerReadTimeout, &cfg.Server.ReadTimeout)
	setDuration(EnvServerRequestTimeout, &cfg.Server.RequestTimeout)

	setDuration(EnvAuthAccessTokenTTL, &cfg.Auth.AccessTokenTTL)
	setDuration(EnvAuthRefreshTokenTTL, &cfg.Auth.RefreshTokenTTL)
	setBool(EnvAuthSecureCookies, &cfg.Auth.SecureCookies)

	setString(EnvSecretsJWTSecret, &cfg.Secrets.JWTSecret)
	setString(EnvSecretsCredentialEncryptionKey, &cfg.Secrets.CredentialEncryptionKey)

	setString(EnvBootstrapAdminUsername, &cfg.BootstrapAdmin.Username)
	setString(EnvBootstrapAdminPassword, &cfg.BootstrapAdmin.Password)

	setString(EnvFrontendPublicAPIBaseURL, &cfg.Frontend.PublicAPIBaseURL)
	setString(EnvFrontendStaticPath, &cfg.Frontend.StaticPath)

	setString(EnvDatabaseDriver, &cfg.Database.Driver)
	setString(EnvDatabaseSQLitePath, &cfg.Database.SQLite.Path)
	setString(EnvDatabasePostgresDSN, &cfg.Database.Postgres.DSN)
	setInt(EnvDatabasePostgresMaxOpenConns, &cfg.Database.Postgres.MaxOpenConns)
	setInt(EnvDatabasePostgresMaxIdleConns, &cfg.Database.Postgres.MaxIdleConns)

	return err
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "yes", "y", "on":
		return true, nil
	case "0", "f", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("必须是 true/false")
	}
}
