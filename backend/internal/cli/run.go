package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Rain-kl/Foam/backend/internal/app"
	"github.com/Rain-kl/Foam/backend/internal/infra/config"
	"github.com/Rain-kl/Foam/backend/internal/infra/observability"
)

// Run 解析启动参数并运行后端服务。
func Run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	cfg, err := config.Load(options.configPath)
	if err != nil {
		return err
	}
	if options.listen != "" {
		cfg.Server.Listen = options.listen
		if err := cfg.Validate(); err != nil {
			return err
		}
	}
	logger := observability.NewLogger()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			logger.Warn("application_close_failed", "error", closeErr)
		}
	}()
	return application.Run(ctx)
}

type runOptions struct {
	configPath string
	listen     string
}

func parseOptions(args []string) (runOptions, error) {
	options := runOptions{configPath: defaultConfigPath()}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--config":
			if index+1 >= len(args) {
				return runOptions{}, errors.New("--config 缺少路径")
			}
			options.configPath = args[index+1]
			index++
		case "--listen":
			if index+1 >= len(args) {
				return runOptions{}, errors.New("--listen 缺少地址")
			}
			options.listen = args[index+1]
			index++
		default:
			return runOptions{}, fmt.Errorf("不支持的启动参数: %s", args[index])
		}
	}
	return options, nil
}

func defaultConfigPath() string {
	// FOAM_CONFIG / CONFIG_PATH override the default search path (env > file discovery).
	for _, key := range []string{"FOAM_CONFIG", "CONFIG_PATH"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	for _, candidate := range []string{"config.yaml", filepath.Join("..", "config.yaml")} {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate
		}
	}
	return "config.yaml"
}
