package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDurationAndSecretsFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`server:
  requestTimeout: 2m
secrets:
  jwtSecret: "12345678901234567890123456789012"
  credentialEncryptionKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
bootstrapAdmin:
  username: "admin"
  password: "password123"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.RequestTimeout.Value() != 2*time.Minute {
		t.Fatalf("requestTimeout = %s", cfg.Server.RequestTimeout.Value())
	}
	if cfg.Server.ReadTimeout.Value() != 15*time.Minute {
		t.Fatalf("readTimeout = %s", cfg.Server.ReadTimeout.Value())
	}
	if cfg.Secrets.JWTSecret != "12345678901234567890123456789012" {
		t.Fatalf("jwtSecret = %q", cfg.Secrets.JWTSecret)
	}
	if cfg.Secrets.CredentialEncryptionKey != "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" {
		t.Fatalf("credentialEncryptionKey = %q", cfg.Secrets.CredentialEncryptionKey)
	}
	if cfg.BootstrapAdmin.Username != "admin" || cfg.BootstrapAdmin.Password != "password123" {
		t.Fatalf("bootstrapAdmin = %#v", cfg.BootstrapAdmin)
	}
	expectedDatabasePath := filepath.Join(dir, "data", "backend.db")
	if cfg.Database.SQLite.Path != expectedDatabasePath {
		t.Fatalf("database path = %q, want %q", cfg.Database.SQLite.Path, expectedDatabasePath)
	}
}

func TestValidateRejectsPlaceholderSecrets(t *testing.T) {
	cfg := defaultConfig()
	cfg.Secrets.JWTSecret = "replace-with-at-least-32-characters"
	cfg.Secrets.CredentialEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	cfg.BootstrapAdmin.Password = "password123"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected placeholder jwt secret to fail validation")
	}

	cfg = defaultConfig()
	cfg.Secrets.JWTSecret = "12345678901234567890123456789012"
	cfg.Secrets.CredentialEncryptionKey = "replace-with-base64-key"
	cfg.BootstrapAdmin.Password = "password123"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected placeholder credential key to fail validation")
	}

	cfg = defaultConfig()
	cfg.Secrets.JWTSecret = "12345678901234567890123456789012"
	cfg.Secrets.CredentialEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	cfg.BootstrapAdmin.Password = "replace-with-a-strong-password"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected placeholder bootstrap password to fail validation")
	}
}

func TestValidateRejectsUnknownDatabaseDriver(t *testing.T) {
	cfg := defaultConfig()
	cfg.Secrets.JWTSecret = "12345678901234567890123456789012"
	cfg.Secrets.CredentialEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	cfg.BootstrapAdmin.Password = "password123"
	cfg.Database.Driver = "mysql"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid database driver to fail validation")
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`server:
  listen: "127.0.0.1:8000"
provider:
  build:
    baseURL: "https://example.com"
secrets:
  jwtSecret: "12345678901234567890123456789012"
  credentialEncryptionKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
bootstrapAdmin:
  username: "admin"
  password: "password123"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected unknown fields to fail")
	}
}

func TestLoadEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`server:
  listen: "127.0.0.1:8000"
secrets:
  jwtSecret: "12345678901234567890123456789012"
  credentialEncryptionKey: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
bootstrapAdmin:
  username: "admin"
  password: "password123"
database:
  driver: sqlite
  sqlite:
    path: "./data/backend.db"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvServerListen, "0.0.0.0:9000")
	t.Setenv(EnvSecretsJWTSecret, "env-jwt-secret-must-be-32-chars!!")
	t.Setenv(EnvAuthSecureCookies, "true")
	t.Setenv(EnvDatabaseDriver, DatabaseDriverPostgres)
	t.Setenv(EnvDatabasePostgresDSN, "postgres://u:p@127.0.0.1:5432/foam?sslmode=disable")
	t.Setenv(EnvDatabasePostgresMaxOpenConns, "20")
	t.Setenv(EnvDatabasePostgresMaxIdleConns, "5")
	t.Setenv(EnvServerRequestTimeout, "45m")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Listen != "0.0.0.0:9000" {
		t.Fatalf("listen = %q", cfg.Server.Listen)
	}
	if cfg.Secrets.JWTSecret != "env-jwt-secret-must-be-32-chars!!" {
		t.Fatalf("jwtSecret = %q", cfg.Secrets.JWTSecret)
	}
	if !cfg.Auth.SecureCookies {
		t.Fatal("secureCookies want true")
	}
	if cfg.Database.Driver != DatabaseDriverPostgres {
		t.Fatalf("driver = %q", cfg.Database.Driver)
	}
	if cfg.Database.Postgres.DSN != "postgres://u:p@127.0.0.1:5432/foam?sslmode=disable" {
		t.Fatalf("dsn = %q", cfg.Database.Postgres.DSN)
	}
	if cfg.Database.Postgres.MaxOpenConns != 20 || cfg.Database.Postgres.MaxIdleConns != 5 {
		t.Fatalf("pool = %d/%d", cfg.Database.Postgres.MaxOpenConns, cfg.Database.Postgres.MaxIdleConns)
	}
	if cfg.Server.RequestTimeout.Value() != 45*time.Minute {
		t.Fatalf("requestTimeout = %s", cfg.Server.RequestTimeout.Value())
	}
	// YAML bootstrap password still applied when env not set.
	if cfg.BootstrapAdmin.Password != "password123" {
		t.Fatalf("bootstrap password = %q", cfg.BootstrapAdmin.Password)
	}
}

func TestLoadDotEnvDoesNotOverrideProcessEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("FOAM_SERVER_LISTEN=127.0.0.1:1111\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FOAM_ENV_FILE", envPath)
	t.Setenv(EnvServerListen, "127.0.0.1:2222")
	t.Setenv(EnvSecretsJWTSecret, "12345678901234567890123456789012")
	t.Setenv(EnvSecretsCredentialEncryptionKey, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	t.Setenv(EnvBootstrapAdminPassword, "password123")

	// No yaml file; env-only load path.
	cfg, err := Load(filepath.Join(dir, "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Listen != "127.0.0.1:2222" {
		t.Fatalf("listen = %q, process env must win over .env", cfg.Server.Listen)
	}
}

func TestApplyEnvOverridesInvalidBool(t *testing.T) {
	cfg := defaultConfig()
	t.Setenv(EnvAuthSecureCookies, "not-a-bool")
	if err := applyEnvOverrides(&cfg); err == nil {
		t.Fatal("expected invalid bool error")
	}
}
