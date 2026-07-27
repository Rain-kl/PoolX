package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	maxServerBodyBytes = 256 << 20
	maxRequestTimeout  = 24 * time.Hour
	maxReadTimeout     = time.Hour

	minJWTSecretLength = 32
	credentialKeyBytes = 32

	// Database drivers accepted by database.driver / FOAM_DATABASE_DRIVER.
	DatabaseDriverSQLite   = "sqlite"
	DatabaseDriverPostgres = "postgres"

	defaultListen          = "127.0.0.1:8000"
	defaultMaxBodyBytes    = 32 << 20
	defaultReadTimeout     = 15 * time.Minute
	defaultRequestTimeout  = 2 * time.Hour
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
	defaultSQLitePath      = "./data/backend.db"
	defaultStaticPath      = "./frontend/dist"
	defaultPostgresMaxOpen = 50
	defaultPostgresMaxIdle = 10
	maxPostgresOpenConns   = 1000
)

// Config 表示后端运行配置。
type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Frontend       FrontendConfig       `yaml:"frontend"`
	Database       DatabaseConfig       `yaml:"database"`
	Auth           AuthConfig           `yaml:"auth"`
	Secrets        Secrets              `yaml:"secrets"`
	BootstrapAdmin BootstrapAdminConfig `yaml:"bootstrapAdmin"`
}

type ServerConfig struct {
	Listen         string   `yaml:"listen"`
	MaxBodyBytes   int64    `yaml:"maxBodyBytes"`
	ReadTimeout    Duration `yaml:"readTimeout"`
	RequestTimeout Duration `yaml:"requestTimeout"`
}

type FrontendConfig struct {
	PublicAPIBaseURL string `yaml:"publicApiBaseURL"`
	StaticPath       string `yaml:"staticPath"`
}

const DefaultPublicAPIBaseURL = "http://127.0.0.1:8000"

// EffectivePublicAPIBaseURL 按配置文件、内置默认值的顺序解析公开地址。
func (c FrontendConfig) EffectivePublicAPIBaseURL() string {
	if value := strings.TrimRight(strings.TrimSpace(c.PublicAPIBaseURL), "/"); value != "" {
		return value
	}
	return DefaultPublicAPIBaseURL
}

type DatabaseConfig struct {
	Driver   string                 `yaml:"driver"`
	SQLite   SQLiteDatabaseConfig   `yaml:"sqlite"`
	Postgres PostgresDatabaseConfig `yaml:"postgres"`
}

type SQLiteDatabaseConfig struct {
	Path string `yaml:"path"`
}

type PostgresDatabaseConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
}

type AuthConfig struct {
	AccessTokenTTL  Duration `yaml:"accessTokenTTL"`
	RefreshTokenTTL Duration `yaml:"refreshTokenTTL"`
	SecureCookies   bool     `yaml:"secureCookies"`
}

type Secrets struct {
	JWTSecret               string `yaml:"jwtSecret"`
	CredentialEncryptionKey string `yaml:"credentialEncryptionKey"`
}

type BootstrapAdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Duration 支持在 YAML 中使用 10m、1h 等可读时间格式。
type Duration time.Duration

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	parsed, err := time.ParseDuration(node.Value)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

func (d Duration) MarshalYAML() (any, error) { return d.String(), nil }

func (d Duration) Value() time.Duration { return time.Duration(d) }

func (d Duration) String() string {
	value := d.Value().String()
	if strings.HasSuffix(value, "m0s") {
		value = strings.TrimSuffix(value, "0s")
	}
	if strings.HasSuffix(value, "h0m") {
		value = strings.TrimSuffix(value, "0m")
	}
	return value
}

// Load 加载启动配置，优先级：
//
//	环境变量（含可选 .env 注入的未设置变量） > config.yaml > 代码默认值
//
// CLI 标志（如 --listen）在 Load 之后由调用方覆盖，优先级更高。
func Load(path string) (Config, error) {
	loadDotEnvIfPresent()

	cfg := defaultConfig()
	loadedFrom, err := mergeYAMLFile(&cfg, path)
	if err != nil {
		return Config{}, err
	}

	if err := applyEnvOverrides(&cfg); err != nil {
		return Config{}, fmt.Errorf("环境变量配置: %w", err)
	}

	if loadedFrom != "" {
		if err := resolveRelativePaths(&cfg, loadedFrom); err != nil {
			return Config{}, err
		}
	} else if err := resolveRelativePathsAgainstCWD(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// mergeYAMLFile overlays a single YAML document onto cfg when path exists.
// Returns the path that was loaded (empty if missing).
func mergeYAMLFile(cfg *Config, path string) (loadedFrom string, err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is CLI/FOAM_CONFIG controlled
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("读取配置文件: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("解析配置文件: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return "", fmt.Errorf("解析配置文件: %w", err)
		}
		return "", errors.New("配置文件只能包含一个 YAML 文档")
	}
	return path, nil
}

func resolveRelativePaths(cfg *Config, configPath string) error {
	absoluteConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("解析配置文件路径: %w", err)
	}
	return resolveRelativePathsInBase(cfg, filepath.Dir(absoluteConfigPath))
}

func resolveRelativePathsAgainstCWD(cfg *Config) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("解析工作目录: %w", err)
	}
	return resolveRelativePathsInBase(cfg, cwd)
}

func resolveRelativePathsInBase(cfg *Config, baseDir string) error {
	if cfg.Database.Driver == DatabaseDriverSQLite {
		path := strings.TrimSpace(cfg.Database.SQLite.Path)
		if path != "" && !filepath.IsAbs(path) {
			cfg.Database.SQLite.Path = filepath.Clean(filepath.Join(baseDir, path))
		}
	}
	staticPath := strings.TrimSpace(cfg.Frontend.StaticPath)
	if staticPath != "" && !filepath.IsAbs(staticPath) {
		cfg.Frontend.StaticPath = filepath.Clean(filepath.Join(baseDir, staticPath))
	}
	return nil
}

// Validate 校验启动所需的安全配置和运行边界。
func (c Config) Validate() error {
	if err := c.validateServer(); err != nil {
		return err
	}
	if err := c.validateFrontend(); err != nil {
		return err
	}
	if err := c.validateDatabase(); err != nil {
		return err
	}
	if err := c.validateSecrets(); err != nil {
		return err
	}
	if c.Auth.AccessTokenTTL.Value() <= 0 || c.Auth.RefreshTokenTTL.Value() <= 0 {
		return errors.New("JWT 有效期必须大于零")
	}
	return nil
}

func (c Config) validateServer() error {
	if strings.TrimSpace(c.Server.Listen) == "" {
		return errors.New("server.listen 不能为空")
	}
	if c.Server.MaxBodyBytes <= 0 || c.Server.MaxBodyBytes > maxServerBodyBytes {
		return fmt.Errorf("server.maxBodyBytes 必须在 1 到 %d 字节之间", maxServerBodyBytes)
	}
	if c.Server.ReadTimeout.Value() <= 0 || c.Server.ReadTimeout.Value() > maxReadTimeout {
		return errors.New("server.readTimeout 必须大于零且不超过 1 小时")
	}
	if c.Server.RequestTimeout.Value() <= 0 || c.Server.RequestTimeout.Value() > maxRequestTimeout {
		return errors.New("server.requestTimeout 必须大于零且不超过 24 小时")
	}
	return nil
}

func (c Config) validateFrontend() error {
	publicBase := strings.TrimSpace(c.Frontend.PublicAPIBaseURL)
	if publicBase == "" {
		return nil
	}
	publicAPIURL, err := url.ParseRequestURI(publicBase)
	if err != nil || (publicAPIURL.Scheme != "http" && publicAPIURL.Scheme != "https") || publicAPIURL.Host == "" || publicAPIURL.User != nil || publicAPIURL.RawQuery != "" || publicAPIURL.Fragment != "" {
		return errors.New("frontend.publicApiBaseURL 必须是不含凭据、查询参数和片段的 HTTP(S) URL")
	}
	return nil
}

func (c Config) validateDatabase() error {
	switch c.Database.Driver {
	case DatabaseDriverSQLite:
		if strings.TrimSpace(c.Database.SQLite.Path) == "" {
			return errors.New("database.sqlite.path 不能为空")
		}
	case DatabaseDriverPostgres:
		if strings.TrimSpace(c.Database.Postgres.DSN) == "" {
			return errors.New("database.postgres.dsn 不能为空")
		}
		if c.Database.Postgres.MaxOpenConns < 1 || c.Database.Postgres.MaxOpenConns > maxPostgresOpenConns || c.Database.Postgres.MaxIdleConns < 0 || c.Database.Postgres.MaxIdleConns > c.Database.Postgres.MaxOpenConns {
			return errors.New("database.postgres 连接池配置无效")
		}
	default:
		return errors.New("database.driver 必须是 sqlite 或 postgres")
	}
	return nil
}

func (c Config) validateSecrets() error {
	if len(c.Secrets.JWTSecret) < minJWTSecretLength {
		return errors.New("secrets.jwtSecret 至少需要 32 个字符")
	}
	if isExampleSecret(c.Secrets.JWTSecret) {
		return errors.New("secrets.jwtSecret 不能使用示例占位值")
	}
	if !validCredentialEncryptionKey(c.Secrets.CredentialEncryptionKey) {
		return errors.New("secrets.credentialEncryptionKey 必须是 Base64 编码的 32 字节密钥")
	}
	if isExampleSecret(c.Secrets.CredentialEncryptionKey) {
		return errors.New("secrets.credentialEncryptionKey 不能使用示例占位值")
	}
	if isExampleSecret(c.BootstrapAdmin.Password) {
		return errors.New("bootstrapAdmin.password 不能使用示例占位值")
	}
	return nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Listen:         defaultListen,
			MaxBodyBytes:   defaultMaxBodyBytes,
			ReadTimeout:    Duration(defaultReadTimeout),
			RequestTimeout: Duration(defaultRequestTimeout),
		},
		Frontend: FrontendConfig{PublicAPIBaseURL: DefaultPublicAPIBaseURL, StaticPath: defaultStaticPath},
		Database: DatabaseConfig{
			Driver:   DatabaseDriverSQLite,
			SQLite:   SQLiteDatabaseConfig{Path: defaultSQLitePath},
			Postgres: PostgresDatabaseConfig{MaxOpenConns: defaultPostgresMaxOpen, MaxIdleConns: defaultPostgresMaxIdle},
		},
		Auth: AuthConfig{
			AccessTokenTTL:  Duration(defaultAccessTokenTTL),
			RefreshTokenTTL: Duration(defaultRefreshTokenTTL),
		},
	}
}

func validCredentialEncryptionKey(value string) bool {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	return err == nil && len(decoded) == credentialKeyBytes
}

func isExampleSecret(value string) bool {
	switch strings.TrimSpace(value) {
	case "replace-with-at-least-32-characters", "replace-with-base64-key", "replace-with-a-strong-password":
		return true
	default:
		return false
	}
}
