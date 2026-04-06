package model

import (
	"path/filepath"
	migratepkg "poolx/internal/model/migrate"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type legacySourceConfigV1 struct {
	ID             int       `gorm:"primaryKey;autoIncrement"`
	Filename       string    `gorm:"size:255;index;not null"`
	ContentHash    string    `gorm:"size:64;index;not null"`
	RawContent     string    `gorm:"type:text;not null"`
	Status         string    `gorm:"size:32;index;not null;default:parsed"`
	TotalNodes     int       `gorm:"not null;default:0"`
	ValidNodes     int       `gorm:"not null;default:0"`
	InvalidNodes   int       `gorm:"not null;default:0"`
	DuplicateNodes int       `gorm:"not null;default:0"`
	ImportedNodes  int       `gorm:"not null;default:0"`
	UploadedBy     string    `gorm:"size:64;index"`
	UploadedByID   int       `gorm:"index"`
	CreatedAt      time.Time `gorm:"index"`
	UpdatedAt      time.Time `gorm:"index"`
}

func (legacySourceConfigV1) TableName() string {
	return "source_configs"
}

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

func TestEnsureDatabaseSchemaUpToDateNormalizesHigherRecordedVersionToCurrentBaseline(t *testing.T) {
	db := openBareTestSQLiteDB(t, "normalized-schema-version.db")
	if err := migratepkg.ComposeAutoMigrator(schemaMetadataModels()...)(db); err != nil {
		t.Fatalf("auto migrate schema metadata: %v", err)
	}
	if err := migratepkg.ComposeAutoMigrator(registeredModels()...)(db); err != nil {
		t.Fatalf("auto migrate business schema: %v", err)
	}

	if err := db.Create(&DatabaseSchemaVersion{
		ID:      databaseSchemaVersionRowID,
		Version: 10,
	}).Error; err != nil {
		t.Fatalf("seed schema version: %v", err)
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
		t.Fatal("expected schema version record to exist")
	}
	if version != currentDatabaseSchemaVersion {
		t.Fatalf("unexpected normalized schema version: got %d want %d", version, currentDatabaseSchemaVersion)
	}
}

func TestEnsureDatabaseSchemaUpToDateAddsSourceConfigMetadataColumns(t *testing.T) {
	db := openBareTestSQLiteDB(t, "source-config-metadata-upgrade.db")

	if err := migratepkg.ComposeAutoMigrator(schemaMetadataModels()...)(db); err != nil {
		t.Fatalf("auto migrate schema metadata: %v", err)
	}
	if err := migratepkg.ComposeAutoMigrator(
		&AppLog{},
		&File{},
		&User{},
		&Option{},
		&legacySourceConfigV1{},
		&ProxyNode{},
		&PortProfile{},
		&PortProfileTemplate{},
		&PortProfileNode{},
		&RuntimeConfig{},
		&KernelInstance{},
	)(db); err != nil {
		t.Fatalf("auto migrate non-source-config tables: %v", err)
	}

	if err := db.Create(&legacySourceConfigV1{
		ID:             1,
		Filename:       "legacy.yaml",
		ContentHash:    "legacy_hash",
		RawContent:     "proxies: []",
		Status:         SourceConfigStatusParsed,
		TotalNodes:     0,
		ValidNodes:     0,
		InvalidNodes:   0,
		DuplicateNodes: 0,
		ImportedNodes:  0,
		UploadedBy:     "legacy-user",
		UploadedByID:   100,
	}).Error; err != nil {
		t.Fatalf("seed legacy source config: %v", err)
	}

	if err := db.Create(&DatabaseSchemaVersion{
		ID:      databaseSchemaVersionRowID,
		Version: 1,
	}).Error; err != nil {
		t.Fatalf("seed schema version: %v", err)
	}

	if err := ensureDatabaseSchemaUpToDate(db, "sqlite"); err != nil {
		t.Fatalf("ensureDatabaseSchemaUpToDate: %v", err)
	}

	type tableColumn struct {
		Name string `gorm:"column:name"`
	}
	var columns []tableColumn
	if err := db.Raw("PRAGMA table_info(source_configs)").Scan(&columns).Error; err != nil {
		t.Fatalf("query source_configs columns: %v", err)
	}
	columnSet := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		columnSet[column.Name] = struct{}{}
	}

	required := []string{"source_type", "source_url", "content_type", "fetched_at"}
	for _, column := range required {
		if _, ok := columnSet[column]; !ok {
			t.Fatalf("expected source_configs to contain column %s", column)
		}
	}

	type sourceConfigRow struct {
		ID         int    `gorm:"column:id"`
		Filename   string `gorm:"column:filename"`
		SourceType string `gorm:"column:source_type"`
	}
	var row sourceConfigRow
	if err := db.Raw("SELECT id, filename, source_type FROM source_configs WHERE id = ?", 1).Scan(&row).Error; err != nil {
		t.Fatalf("query migrated legacy source config: %v", err)
	}
	if row.ID != 1 {
		t.Fatalf("expected legacy source config row to exist after migration")
	}
	if row.Filename != "legacy.yaml" {
		t.Fatalf("unexpected legacy source config filename: %s", row.Filename)
	}
	if row.SourceType != SourceConfigSourceTypeUpload {
		t.Fatalf("unexpected migrated source_type: got %s want %s", row.SourceType, SourceConfigSourceTypeUpload)
	}

	loadVersion, _ := migratepkg.NewSchemaVersionStore((&DatabaseSchemaVersion{}).TableName(), databaseSchemaVersionRowID)
	version, exists, err := loadVersion(db)
	if err != nil {
		t.Fatalf("loadDatabaseSchemaVersion: %v", err)
	}
	if !exists {
		t.Fatal("expected schema version record to exist")
	}
	if version != currentDatabaseSchemaVersion {
		t.Fatalf("unexpected schema version: got %d want %d", version, currentDatabaseSchemaVersion)
	}
}

func TestEnsureDatabaseSchemaUpToDateRejectsSchemaVersionTwoWithoutSourceConfigMetadataColumns(t *testing.T) {
	db := openBareTestSQLiteDB(t, "source-config-metadata-missing-schema-v2.db")

	if err := migratepkg.ComposeAutoMigrator(schemaMetadataModels()...)(db); err != nil {
		t.Fatalf("auto migrate schema metadata: %v", err)
	}
	if err := migratepkg.ComposeAutoMigrator(
		&AppLog{},
		&File{},
		&User{},
		&Option{},
		&legacySourceConfigV1{},
		&ProxyNode{},
		&PortProfile{},
		&PortProfileTemplate{},
		&PortProfileNode{},
		&RuntimeConfig{},
		&KernelInstance{},
	)(db); err != nil {
		t.Fatalf("auto migrate business schema: %v", err)
	}

	if err := db.Create(&DatabaseSchemaVersion{
		ID:      databaseSchemaVersionRowID,
		Version: currentDatabaseSchemaVersion,
	}).Error; err != nil {
		t.Fatalf("seed schema version: %v", err)
	}

	err := ensureDatabaseSchemaUpToDate(db, "sqlite")
	if err == nil {
		t.Fatal("expected schema validation to fail when source_configs metadata columns are missing")
	}
	if !strings.Contains(err.Error(), "column source_configs.source_type is missing") {
		t.Fatalf("expected missing source_configs metadata column error, got: %v", err)
	}
}
