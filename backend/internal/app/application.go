package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/application/adminauth"
	clashapp "github.com/Rain-kl/Foam/backend/internal/application/clash"
	exampleapp "github.com/Rain-kl/Foam/backend/internal/application/example"
	settingsapp "github.com/Rain-kl/Foam/backend/internal/application/settings"
	"github.com/Rain-kl/Foam/backend/internal/infra/config"
	"github.com/Rain-kl/Foam/backend/internal/infra/persistence/relational"
	"github.com/Rain-kl/Foam/backend/internal/infra/security"
	httpserver "github.com/Rain-kl/Foam/backend/internal/transport/http"
)

const (
	httpReadHeaderTimeout = 10 * time.Second
	httpShutdownTimeout   = 10 * time.Second
	readinessStateReady   = "ready"
)

// Application 管理后端进程生命周期。
type Application struct {
	logger   *slog.Logger
	database *relational.Database
	server   *http.Server
}

// New 完成数据库、管理员认证、示例 CRUD 与 HTTP 路由装配。
func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Application, error) {
	if logger == nil {
		logger = slog.Default()
	}

	var database *relational.Database
	var err error
	switch cfg.Database.Driver {
	case config.DatabaseDriverSQLite:
		database, err = relational.OpenSQLite(ctx, cfg.Database.SQLite.Path)
	case config.DatabaseDriverPostgres:
		database, err = relational.OpenPostgres(ctx, cfg.Database.Postgres.DSN, cfg.Database.Postgres.MaxOpenConns, cfg.Database.Postgres.MaxIdleConns)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Database.Driver)
	}
	if err != nil {
		return nil, err
	}
	// Schema via goose SQL (postgres/sqlite dual dialect); not GORM AutoMigrate.
	if err := database.InitializeSchema(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("database migrate: %w", err)
	}

	// Keep cipher validation so config still requires a real encryption key.
	if _, err := security.NewCipher(cfg.Secrets.CredentialEncryptionKey); err != nil {
		_ = database.Close()
		return nil, err
	}

	adminRepo := relational.NewAdminRepository(database)
	sessionRepo := relational.NewAdminSessionRepository(database)
	exampleRepo := relational.NewExampleRepository(database)
	settingsRepo := relational.NewRuntimeSettingsRepository(database)
	tokenService := security.NewTokenService(cfg.Secrets.JWTSecret)
	adminService := adminauth.NewService(
		adminRepo,
		sessionRepo,
		tokenService,
		cfg.Auth.AccessTokenTTL.Value(),
		cfg.Auth.RefreshTokenTTL.Value(),
	)
	if err := adminService.Bootstrap(ctx, cfg.BootstrapAdmin.Username, cfg.BootstrapAdmin.Password); err != nil {
		_ = database.Close()
		return nil, err
	}
	exampleService := exampleapp.NewService(exampleRepo)
	persistedSettings, settingsUpdatedAt, settingsRevision, err := settingsapp.LoadPersisted(ctx, cfg, settingsRepo)
	if err != nil {
		_ = database.Close()
		return nil, err
	}
	settingsService := settingsapp.NewService(cfg, persistedSettings, settingsUpdatedAt, settingsRevision, settingsRepo)

	clashSourceRepo := relational.NewSourceConfigRepository(database)
	clashNodeRepo := relational.NewProxyNodeRepository(database)
	clashTestResRepo := relational.NewNodeTestResultRepository(database)
	clashProfileRepo := relational.NewPortProfileRepository(database)
	clashTplRepo := relational.NewPortProfileTemplateRepository(database)
	clashRuntimeRepo := relational.NewRuntimeConfigRepository(database)
	clashKernelRepo := relational.NewKernelInstanceRepository(database)
	clashService := clashapp.NewService(clashSourceRepo, clashNodeRepo, clashTestResRepo, clashProfileRepo, clashTplRepo, clashRuntimeRepo, clashKernelRepo, settingsRepo)

	// HTTP stack uses per-request contexts from net/http; do not thread process ctx into handlers.
	router := httpserver.New(httpserver.Dependencies{ //nolint:contextcheck // request-scoped contexts are derived inside Gin middleware
		Logger:             logger,
		RequestTimeout:     cfg.Server.RequestTimeout.Value(),
		MaxBodyBytes:       cfg.Server.MaxBodyBytes,
		SecureCookies:      cfg.Auth.SecureCookies,
		PublicAPIBaseURL:   settingsService.PublicAPIBaseURL,
		FrontendStaticPath: cfg.Frontend.StaticPath,
		Readiness: func(context.Context) httpserver.ReadinessSnapshot {
			return httpserver.ReadinessSnapshot{
				Ready:     true,
				State:     readinessStateReady,
				UpdatedAt: time.Now().UTC(),
				Components: map[string]httpserver.ReadinessComponent{
					"database": {State: readinessStateReady},
				},
			}
		},
		AdminAuth: adminService,
		Examples:  exampleService,
		Settings:  settingsService,
		Clash:     clashService,
	})

	server := &http.Server{
		Addr:              cfg.Server.Listen,
		Handler:           router,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout.Value(),
	}

	app := &Application{logger: logger, database: database, server: server}

	// 若配置与内核二进制均可用，在后台异步自动启动内核，不阻塞 HTTP 服务启动。
	go clashService.AutoStartKernel(context.WithoutCancel(ctx), logger)

	return app, nil
}

// Run 启动 HTTP 服务，直到上下文取消。
func (a *Application) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("server_started", "listen", a.server.Addr)
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		// Parent ctx is already canceled; use a detached timeout for graceful drain.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), httpShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Warn("server_shutdown_failed", "error", err)
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

// Close 关闭数据库连接。
func (a *Application) Close() error {
	if a == nil || a.database == nil {
		return nil
	}
	return a.database.Close()
}
