package migrator

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUpInitializesSQLiteSchema(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open("file:foam-migrator-test?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	report, err := Up(context.Background(), gormDB, "sqlite")
	if err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if report.Backend != "sqlite" {
		t.Fatalf("Backend = %q", report.Backend)
	}
	if report.Version < 202607240001 {
		t.Fatalf("Version = %d, want >= 202607240001", report.Version)
	}
	if !report.Applied {
		t.Fatal("Applied = false, want true on empty database")
	}

	for _, table := range []string{"admins", "admin_sessions", "examples", "runtime_settings", "goose_db_version"} {
		var name string
		err := gormDB.Raw(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
			table,
		).Scan(&name).Error
		if err != nil {
			t.Fatalf("lookup table %s: %v", table, err)
		}
		if name != table {
			t.Fatalf("table %s missing after migrate", table)
		}
	}

	// Idempotent second run.
	report2, err := Up(context.Background(), gormDB, "sqlite")
	if err != nil {
		t.Fatalf("second Up() error = %v", err)
	}
	if report2.Applied {
		t.Fatal("second Up Applied = true, want false")
	}
	if report2.Version != report.Version {
		t.Fatalf("second Version = %d, want %d", report2.Version, report.Version)
	}
}

func TestUpRejectsUnknownDriver(t *testing.T) {
	_, err := Up(context.Background(), nil, "mysql")
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}
