package relational

import (
	"context"
	"fmt"

	"github.com/Rain-kl/Foam/backend/internal/infra/persistence/migrator"
	"gorm.io/gorm"
)

const postgresSchemaMigrationLockID int64 = 0x464F414D53434146 // "FOAMSC AF"

// InitializeSchema applies versioned goose SQL migrations (Wavelet-style).
// GORM AutoMigrate is not used; schema lives in migrator/goose/{postgres,sqlite}.
func (d *Database) InitializeSchema(ctx context.Context) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("database is nil")
	}
	if d.dialect == "postgres" {
		return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", postgresSchemaMigrationLockID).Error; err != nil {
				return fmt.Errorf("acquire PostgreSQL migration lock: %w", err)
			}
			if _, err := migrator.Up(ctx, tx, d.dialect); err != nil {
				return err
			}
			return nil
		})
	}
	if _, err := migrator.Up(ctx, d.db, d.dialect); err != nil {
		return err
	}
	return nil
}

// Gorm exposes the underlying *gorm.DB for packages that need raw access (e.g. migrator tests).
func (d *Database) Gorm() *gorm.DB {
	if d == nil {
		return nil
	}
	return d.db
}
