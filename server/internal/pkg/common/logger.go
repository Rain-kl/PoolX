package common

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

type logLevel int

const (
	logLevelDebug logLevel = iota
	logLevelInfo
	logLevelWarn
	logLevelError
)

var currentLogLevel = logLevelInfo
var currentLogLevelName = "info"
var commonLogWriter io.Writer = os.Stdout
var errorLogWriter io.Writer = os.Stderr
var defaultLogger *slog.Logger

type customTextHandler struct {
	writer io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

type levelRouterHandler struct {
	commonHandler slog.Handler
	errorHandler  slog.Handler
}

func (h *customTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *customTextHandler) Handle(_ context.Context, record slog.Record) error {
	var builder strings.Builder
	builder.WriteString(record.Time.Format("2006-01-02 15:04:05.000"))
	builder.WriteString(" | ")
	builder.WriteString(fmt.Sprintf("%-8s", levelLabel(record.Level)))
	builder.WriteString(" | ")
	builder.WriteString(sourceLocation(record.PC))
	builder.WriteString(" - ")
	builder.WriteString(record.Message)

	attrs := make([]slog.Attr, 0, len(h.attrs)+record.NumAttrs())
	attrs = append(attrs, h.attrs...)
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	if len(attrs) > 0 {
		builder.WriteString(" | ")
		builder.WriteString(formatAttrs(h.groups, attrs))
	}
	builder.WriteByte('\n')
	_, err := io.WriteString(h.writer, builder.String())
	return err
}

func (h *customTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cloned := *h
	cloned.attrs = append(slices.Clone(h.attrs), attrs...)
	return &cloned
}

func (h *customTextHandler) WithGroup(name string) slog.Handler {
	if strings.TrimSpace(name) == "" {
		return h
	}
	cloned := *h
	cloned.groups = append(slices.Clone(h.groups), name)
	return &cloned
}

func (h *levelRouterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.commonHandler.Enabled(ctx, level) || h.errorHandler.Enabled(ctx, level)
}

func (h *levelRouterHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level >= slog.LevelError {
		return h.errorHandler.Handle(ctx, record)
	}
	return h.commonHandler.Handle(ctx, record)
}

func (h *levelRouterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelRouterHandler{
		commonHandler: h.commonHandler.WithAttrs(attrs),
		errorHandler:  h.errorHandler.WithAttrs(attrs),
	}
}

func (h *levelRouterHandler) WithGroup(name string) slog.Handler {
	return &levelRouterHandler{
		commonHandler: h.commonHandler.WithGroup(name),
		errorHandler:  h.errorHandler.WithGroup(name),
	}
}

func configureGinWriters() {
	if shouldLog(logLevelDebug) {
		gin.DefaultWriter = commonLogWriter
	} else {
		gin.DefaultWriter = io.Discard
	}
	gin.DefaultErrorWriter = errorLogWriter
}

func slogLevel() slog.Level {
	switch currentLogLevel {
	case logLevelDebug:
		return slog.LevelDebug
	case logLevelWarn:
		return slog.LevelWarn
	case logLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func ensureLogger() *slog.Logger {
	if defaultLogger != nil {
		return defaultLogger
	}
	defaultLogger = slog.New(&levelRouterHandler{
		commonHandler: &customTextHandler{writer: commonLogWriter, level: slogLevel()},
		errorHandler:  &customTextHandler{writer: errorLogWriter, level: slogLevel()},
	})
	slog.SetDefault(defaultLogger)
	return defaultLogger
}

func SetLogLevel(level string) {
	normalized := strings.TrimSpace(strings.ToLower(level))
	switch normalized {
	case "debug":
		currentLogLevel = logLevelDebug
		currentLogLevelName = "debug"
	case "warn", "warning":
		currentLogLevel = logLevelWarn
		currentLogLevelName = "warn"
	case "error":
		currentLogLevel = logLevelError
		currentLogLevelName = "error"
	default:
		currentLogLevel = logLevelInfo
		currentLogLevelName = "info"
	}
	configureGinWriters()
}

func GetLogLevel() string {
	return currentLogLevelName
}

func shouldLog(level logLevel) bool {
	return level >= currentLogLevel
}

func SetupGinLog() {
	if *LogDir != "" {
		commonLogPath := filepath.Join(*LogDir, "common.log")
		errorLogPath := filepath.Join(*LogDir, "error.log")
		commonFd, err := os.OpenFile(commonLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			_, _ = io.WriteString(os.Stderr, "failed to open common log file\n")
			os.Exit(1)
		}
		errorFd, err := os.OpenFile(errorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			_, _ = io.WriteString(os.Stderr, "failed to open error log file\n")
			os.Exit(1)
		}
		commonLogWriter = io.MultiWriter(os.Stdout, commonFd)
		errorLogWriter = io.MultiWriter(os.Stderr, errorFd)
	}
	configureGinWriters()
	defaultLogger = nil
	ensureLogger()
}

func levelLabel(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARNING"
	default:
		return "ERROR"
	}
}

func sourceLocation(pc uintptr) string {
	if pc == 0 {
		return "unknown:unknown:0"
	}
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	fileName := strings.TrimSuffix(filepath.Base(frame.File), filepath.Ext(frame.File))
	if fileName == "" {
		fileName = "unknown"
	}
	functionName := "unknown"
	if frame.Function != "" {
		parts := strings.Split(frame.Function, "/")
		functionName = parts[len(parts)-1]
		if dot := strings.LastIndex(functionName, "."); dot >= 0 && dot < len(functionName)-1 {
			functionName = functionName[dot+1:]
		}
	}
	return fmt.Sprintf("%s:%s:%d", fileName, functionName, frame.Line)
}

func formatAttrs(groups []string, attrs []slog.Attr) string {
	parts := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		key := attr.Key
		if key == "" {
			continue
		}
		if len(groups) > 0 {
			key = strings.Join(append(slices.Clone(groups), key), ".")
		}
		parts = append(parts, fmt.Sprintf("%s=%v", key, attr.Value.Any()))
	}
	return strings.Join(parts, " ")
}
