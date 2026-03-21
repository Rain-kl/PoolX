package migrate

import "gorm.io/gorm"

const v2DefaultAppLogLevel = "info"

type v2AppLog struct {
	ID             int    `gorm:"primaryKey;autoIncrement"`
	Classification string `gorm:"size:32;index;not null"`
	Level          string `gorm:"size:16;index;not null;default:info"`
	Message        string `gorm:"type:text;not null"`
}

func (v2AppLog) TableName() string {
	return "app_logs"
}

func migrateV2AppLogLevels(db *gorm.DB, backend string) error {
	_ = backend
	appLogTable := &v2AppLog{}
	if db == nil || !db.Migrator().HasTable(appLogTable) {
		return nil
	}
	if !db.Migrator().HasColumn(appLogTable, "Level") {
		if err := db.Migrator().AddColumn(appLogTable, "Level"); err != nil {
			return err
		}
	}
	return db.Model(appLogTable).
		Where("level IS NULL OR TRIM(level) = ''").
		Update("level", v2DefaultAppLogLevel).
		Error
}

func V2(hooks Hooks) Migration {
	return Migration{
		FromVersion: 1,
		ToVersion:   2,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
				NamedStep{Name: "backfill_app_log_levels", Run: migrateV2AppLogLevels},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
