package migrate

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type StepFunc func(db *gorm.DB, backend string) error
type VersionLoader func(db *gorm.DB) (int, bool, error)
type VersionSaver func(db *gorm.DB, version int) error
type EmptyChecker func(db *gorm.DB) (bool, error)
type MetadataMigrator func(db *gorm.DB) error

type Initializer struct {
	Version    int
	Initialize StepFunc
	Validate   StepFunc
}

type Migration struct {
	FromVersion int
	ToVersion   int
	Migrate     StepFunc
	Validate    StepFunc
}

type Hooks struct {
	ApplySchema     StepFunc
	ValidateSchema  StepFunc
	AfterInitialize StepFunc
}

type NamedStep struct {
	Name string
	Run  StepFunc
}

type RequiredColumn struct {
	Model  any
	Table  string
	Column string
}

type schemaVersionRow struct {
	ID      int `gorm:"primaryKey"`
	Version int `gorm:"not null"`
}

func runNamedSteps(db *gorm.DB, backend string, steps ...NamedStep) error {
	for _, step := range steps {
		if step.Run == nil {
			continue
		}
		if err := step.Run(db, backend); err != nil {
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}
	}
	return nil
}

func ComposeAutoMigrator(models ...any) MetadataMigrator {
	return func(db *gorm.DB) error {
		for _, item := range models {
			if err := db.AutoMigrate(item); err != nil {
				return err
			}
		}
		return nil
	}
}

func ComposeEmptyChecker(models ...any) EmptyChecker {
	return func(db *gorm.DB) (bool, error) {
		for _, item := range models {
			if !db.Migrator().HasTable(item) {
				continue
			}
			var count int64
			if err := db.Model(item).Limit(1).Count(&count).Error; err != nil {
				return false, err
			}
			if count > 0 {
				return false, nil
			}
		}
		return true, nil
	}
}

// ComposeSchemaApplier builds a reusable current-schema applier.
// It belongs to the migrate package because it is migration infrastructure,
// not application startup logic.
func ComposeSchemaApplier(ensureSchemaMeta MetadataMigrator, applyBusinessSchema MetadataMigrator) StepFunc {
	return func(db *gorm.DB, backend string) error {
		if err := ensureSchemaMeta(db); err != nil {
			return err
		}
		if err := applyBusinessSchema(db); err != nil {
			return err
		}
		_ = backend
		return nil
	}
}

func ComposeSchemaValidator(versionTableName string, requiredModels []any, requiredColumns []RequiredColumn) StepFunc {
	return func(db *gorm.DB, backend string) error {
		_ = backend
		if db == nil {
			return fmt.Errorf("database handle is nil")
		}
		if !db.Migrator().HasTable(versionTableName) {
			return fmt.Errorf("table %s is missing", versionTableName)
		}
		for _, item := range requiredModels {
			if !db.Migrator().HasTable(item) {
				return fmt.Errorf("required table is missing")
			}
		}
		for _, item := range requiredColumns {
			if !db.Migrator().HasColumn(item.Model, item.Column) {
				return fmt.Errorf("column %s.%s is missing", item.Table, item.Column)
			}
		}
		return nil
	}
}

func NewSchemaVersionStore(tableName string, rowID int) (VersionLoader, VersionSaver) {
	load := func(db *gorm.DB) (int, bool, error) {
		if db == nil {
			return 0, false, nil
		}
		if !db.Migrator().HasTable(tableName) {
			return 0, false, nil
		}
		var state schemaVersionRow
		err := db.Table(tableName).Where("id = ?", rowID).First(&state).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil
		}
		if err != nil {
			return 0, false, err
		}
		return state.Version, true, nil
	}

	save := func(db *gorm.DB, version int) error {
		return db.Table(tableName).
			Where("id = ?", rowID).
			Assign(map[string]any{
				"id":      rowID,
				"version": version,
			}).
			FirstOrCreate(&schemaVersionRow{}).
			Error
	}

	return load, save
}
