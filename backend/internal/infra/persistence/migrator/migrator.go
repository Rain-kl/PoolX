// Package migrator runs versioned SQL migrations with pressly/goose (Wavelet-style).
//
// Layout:
//
//	goose/postgres/*.sql
//	goose/sqlite/*.sql
//
// Paired dialects must share the same version number and filename.
// Schema evolution is SQL-only; GORM AutoMigrate is not used in production.
package migrator

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sync"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// migrationFS contains SQL migrations under goose/<dialect>.
//
//go:embed goose/postgres/*.sql goose/sqlite/*.sql
var migrationFS embed.FS

const (
	dialectSqlite3  = "sqlite3"
	dialectPostgres = "postgres"
	dirSqlite       = "goose/sqlite"
	dirPostgres     = "goose/postgres"
)

// gooseMu serializes goose global setters (SetBaseFS / SetDialect) for parallel tests.
var gooseMu sync.Mutex

// Report is the migration outcome observed at startup.
type Report struct {
	Backend string // "sqlite" | "postgres"
	Version int64
	Applied bool
}

// Up applies all pending migrations for the given Foam database driver
// ("sqlite" or "postgres") against the provided GORM handle.
func Up(ctx context.Context, gormDB *gorm.DB, driver string) (Report, error) {
	_ = ctx
	if _, _, _, err := resolveDriver(driver); err != nil {
		return Report{}, err
	}
	if gormDB == nil {
		return Report{}, fmt.Errorf("migrator: gorm db is nil")
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return Report{}, fmt.Errorf("migrator: load sql.DB: %w", err)
	}
	return upSQL(sqlDB, driver)
}

// UpSQL applies migrations against a raw *sql.DB (tests / tooling).
func UpSQL(sqlDB *sql.DB, driver string) (Report, error) {
	return upSQL(sqlDB, driver)
}

func upSQL(sqlDB *sql.DB, driver string) (Report, error) {
	gooseDialect, dir, backend, err := resolveDriver(driver)
	if err != nil {
		return Report{}, err
	}

	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(migrationFS)
	if err := goose.SetDialect(gooseDialect); err != nil {
		return Report{}, fmt.Errorf("migrator: set dialect %s: %w", gooseDialect, err)
	}

	previousVersion, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return Report{}, fmt.Errorf("migrator: get version: %w", err)
	}
	if err := goose.Up(sqlDB, dir); err != nil {
		return Report{}, fmt.Errorf("migrator: goose up (%s): %w", backend, err)
	}
	currentVersion, err := goose.GetDBVersion(sqlDB)
	if err != nil {
		return Report{}, fmt.Errorf("migrator: get migrated version: %w", err)
	}

	return Report{
		Backend: backend,
		Version: currentVersion,
		Applied: currentVersion != previousVersion,
	}, nil
}

func resolveDriver(driver string) (gooseDialect, dir, backend string, err error) {
	switch driver {
	case "sqlite", "sqlite3":
		return dialectSqlite3, dirSqlite, "sqlite", nil
	case "postgres", "postgresql":
		return dialectPostgres, dirPostgres, "postgres", nil
	default:
		return "", "", "", fmt.Errorf("migrator: unsupported database driver %q", driver)
	}
}
