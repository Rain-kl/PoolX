package app

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"poolx/internal/middleware"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"poolx/internal/pkg/utils/geoip"
	"poolx/internal/router"
	"poolx/internal/service"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func RunServer(assetFS fs.FS, buildDir string, indexPage []byte, zashboardDir string) {
	common.SetupGinLog()
	slog.Info("PoolX started", "version", common.Version)
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	err := model.InitDB()
	if err != nil {
		slog.Error("initialize database failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			slog.Error("close database failed", "error", err)
			os.Exit(1)
		}
	}()

	err = common.InitRedisClient()
	if err != nil {
		slog.Error("initialize redis failed", "error", err)
		os.Exit(1)
	}

	model.InitOptionMap()
	geoip.InitGeoIP()
	if err = service.AppLog.Push(model.AppLogClassificationSystem, model.AppLogLevelInfo, fmt.Sprintf("PoolX server started | version=%s | db_backend=%s | redis_enabled=%t", common.Version, dbBackend(common.SQLDSN), common.RedisEnabled)); err != nil {
		slog.Warn("write startup app log failed", "error", err)
	}

	server := gin.Default()
	server.Use(middleware.CORS())

	if common.RedisEnabled {
		opt := common.ParseRedisOption()
		store, _ := redis.NewStore(opt.MinIdleConns, opt.Network, opt.Addr, opt.Password, []byte(common.SessionSecret))
		server.Use(sessions.Sessions("session", store))
	} else {
		store := cookie.NewStore([]byte(common.SessionSecret))
		server.Use(sessions.Sessions("session", store))
	}

	router.SetRouter(server, assetFS, buildDir, indexPage, zashboardDir)
	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	slog.Info(
		"server config",
		"port", port,
		"gin_mode", gin.Mode(),
		"log_level", common.GetLogLevel(),
		"db_backend", dbBackend(common.SQLDSN),
		"sqlite_path", common.SQLitePath,
		"redis_enabled", common.RedisEnabled,
		"upload_path", common.UploadPath,
		"log_dir", valueOrDefault(*common.LogDir, "stdout"),
	)
	slog.Info("server listening", "address", fmt.Sprintf(":%s", port))
	err = server.Run(":" + port)
	if err != nil {
		slog.Error("server run failed", "error", err)
	}
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func dbBackend(dsn string) string {
	if dsn != "" {
		return "postgres"
	}
	return "sqlite"
}
