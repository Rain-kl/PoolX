package model

import (
	"path/filepath"
	migratepkg "poolx/internal/model/migrate"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openBareTestSQLiteDB(t *testing.T, name string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return db
}

func openTestSQLiteDB(t *testing.T, name string) *gorm.DB {
	t.Helper()

	db := openBareTestSQLiteDB(t, name)
	if err := migratepkg.ComposeAutoMigrator(registeredModels()...)(db); err != nil {
		t.Fatalf("auto migrate db: %v", err)
	}
	return db
}

func findDBModelByTableName(t *testing.T, tableName string) dbModel {
	t.Helper()

	models, err := buildDBModels()
	if err != nil {
		t.Fatalf("build db models: %v", err)
	}
	for _, item := range models {
		if item.tableName == tableName {
			return item
		}
	}
	t.Fatalf("db model not found for table %s", tableName)
	return dbModel{}
}

func TestIsDatabaseEmpty(t *testing.T) {
	db := openTestSQLiteDB(t, "empty.db")
	checker := migratepkg.ComposeEmptyChecker(registeredModels()...)

	empty, err := checker(db)
	if err != nil {
		t.Fatalf("empty checker returned error: %v", err)
	}
	if !empty {
		t.Fatal("expected database to be empty")
	}

	if err := db.Create(&User{
		Username:    "alice",
		Password:    "secret",
		DisplayName: "Alice",
		Role:        1,
		Status:      1,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	empty, err = checker(db)
	if err != nil {
		t.Fatalf("empty checker after seed returned error: %v", err)
	}
	if empty {
		t.Fatal("expected database to be non-empty")
	}
}

func TestMigrateTableDataCopiesRows(t *testing.T) {
	source := openTestSQLiteDB(t, "source.db")
	target := openTestSQLiteDB(t, "target.db")

	user := User{
		Id:          1,
		Username:    "root",
		Password:    "hashed",
		DisplayName: "Root User",
		Role:        100,
		Status:      1,
	}
	option := Option{
		Key:   "SystemName",
		Value: "PoolX",
	}

	if err := source.Create(&user).Error; err != nil {
		t.Fatalf("seed source user: %v", err)
	}
	if err := source.Create(&option).Error; err != nil {
		t.Fatalf("seed source option: %v", err)
	}

	if err := migrateTableData(source, target, findDBModelByTableName(t, "users")); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := migrateTableData(source, target, findDBModelByTableName(t, "options")); err != nil {
		t.Fatalf("migrate options: %v", err)
	}

	var gotUser User
	if err := target.First(&gotUser, 1).Error; err != nil {
		t.Fatalf("query migrated user: %v", err)
	}
	if gotUser.Username != user.Username || gotUser.DisplayName != user.DisplayName {
		t.Fatalf("unexpected migrated user: %+v", gotUser)
	}

	var gotOption Option
	if err := target.First(&gotOption, "key = ?", option.Key).Error; err != nil {
		t.Fatalf("query migrated option: %v", err)
	}
	if gotOption.Value != option.Value {
		t.Fatalf("unexpected migrated option value: %s", gotOption.Value)
	}
}

func TestEnsureDatabaseSchemaUpToDateInitializesFreshDatabase(t *testing.T) {
	db := openBareTestSQLiteDB(t, "fresh-schema.db")

	if err := ensureDatabaseSchemaUpToDate(db, "sqlite"); err != nil {
		t.Fatalf("ensureDatabaseSchemaUpToDate: %v", err)
	}

	loadVersion, _ := migratepkg.NewSchemaVersionStore((&DatabaseSchemaVersion{}).TableName(), databaseSchemaVersionRowID)
	version, exists, err := loadVersion(db)
	if err != nil {
		t.Fatalf("loadDatabaseSchemaVersion: %v", err)
	}
	if !exists {
		t.Fatal("expected database schema version to be recorded")
	}
	if version != currentDatabaseSchemaVersion {
		t.Fatalf("unexpected schema version: got %d want %d", version, currentDatabaseSchemaVersion)
	}
}

func TestEnsureDatabaseSchemaUpToDateUpgradesLegacyDatabase(t *testing.T) {
	db := openBareTestSQLiteDB(t, "legacy-schema.db")
	if err := migratepkg.ComposeAutoMigrator(registeredModels()...)(db); err != nil {
		t.Fatalf("auto migrate db: %v", err)
	}
	if err := db.Create(&User{
		Username:    "legacy",
		Password:    "secret",
		DisplayName: "Legacy User",
		Role:        1,
		Status:      1,
	}).Error; err != nil {
		t.Fatalf("seed legacy user: %v", err)
	}

	if err := ensureDatabaseSchemaUpToDate(db, "sqlite"); err != nil {
		t.Fatalf("ensureDatabaseSchemaUpToDate: %v", err)
	}

	loadVersion, _ := migratepkg.NewSchemaVersionStore((&DatabaseSchemaVersion{}).TableName(), databaseSchemaVersionRowID)
	version, exists, err := loadVersion(db)
	if err != nil {
		t.Fatalf("loadDatabaseSchemaVersion: %v", err)
	}
	if !exists {
		t.Fatal("expected legacy database to gain a schema version record")
	}
	if version != currentDatabaseSchemaVersion {
		t.Fatalf("unexpected schema version: got %d want %d", version, currentDatabaseSchemaVersion)
	}
}

func TestV2MigrationBackfillsEmptyAppLogLevels(t *testing.T) {
	db := openTestSQLiteDB(t, "app-log-level-backfill.db")

	if err := db.Create(&AppLog{
		Classification: AppLogClassificationSystem,
		Level:          "",
		Message:        "legacy log",
	}).Error; err != nil {
		t.Fatalf("seed app log: %v", err)
	}

	if err := migratepkg.V2(migratepkg.Hooks{
		ApplySchema:    migrationHooks().ApplySchema,
		ValidateSchema: migrationHooks().ValidateSchema,
	}).Migrate(db, "sqlite"); err != nil {
		t.Fatalf("V2 migration: %v", err)
	}

	var row AppLog
	if err := db.First(&row).Error; err != nil {
		t.Fatalf("query migrated app log: %v", err)
	}
	if row.Level != AppLogLevelInfo {
		t.Fatalf("unexpected app log level: got %q want %q", row.Level, AppLogLevelInfo)
	}
}

func TestRunDatabaseSchemaMigrationDoesNotAdvanceVersionWhenValidationFails(t *testing.T) {
	db := openBareTestSQLiteDB(t, "failed-validation.db")
	loadVersion, saveVersion := migratepkg.NewSchemaVersionStore((&DatabaseSchemaVersion{}).TableName(), databaseSchemaVersionRowID)

	err := migratepkg.NewScheduler(
		db,
		"sqlite",
		loadVersion,
		saveVersion,
		migratepkg.ComposeEmptyChecker(registeredModels()...),
		migratepkg.ComposeAutoMigrator(schemaMetadataModels()...),
	).
		RegisterInitializer(migratepkg.V0(migratepkg.Hooks{
			ApplySchema: func(db *gorm.DB, backend string) error {
				_ = backend
				return migratepkg.ComposeAutoMigrator(schemaMetadataModels()...)(db)
			},
			ValidateSchema: migratepkg.ComposeSchemaValidator(
				(&DatabaseSchemaVersion{}).TableName(),
				registeredModels(),
				nil,
			),
		})).
		RegisterMigration(migratepkg.Migration{
			FromVersion: 0,
			ToVersion:   currentDatabaseSchemaVersion,
			Migrate: func(tx *gorm.DB, backend string) error {
				return migratepkg.ComposeAutoMigrator(schemaMetadataModels()...)(tx)
			},
			Validate: func(tx *gorm.DB, backend string) error {
				return gorm.ErrInvalidDB
			},
		}).
		Apply()
	if err == nil {
		t.Fatal("expected migration validation to fail")
	}

	_, exists, loadErr := loadVersion(db)
	if loadErr != nil {
		t.Fatalf("loadDatabaseSchemaVersion: %v", loadErr)
	}
	if exists {
		t.Fatal("expected schema version to remain unset after failed validation")
	}
}
