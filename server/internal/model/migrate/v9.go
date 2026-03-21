package migrate

import "gorm.io/gorm"

func dropPortProfileEnabledColumns(db *gorm.DB, backend string) error {
	if backend == "sqlite" {
		// SQLite is already tolerant of extra columns for this schema, while
		// glebarez/sqlite may panic during DropColumn table recreation.
		return nil
	}
	if db.Migrator().HasColumn("port_profiles", "enabled") {
		if err := db.Migrator().DropColumn("port_profiles", "enabled"); err != nil {
			return err
		}
	}
	if db.Migrator().HasColumn("port_profile_templates", "enabled") {
		if err := db.Migrator().DropColumn("port_profile_templates", "enabled"); err != nil {
			return err
		}
	}
	return nil
}

func V9(hooks Hooks) Migration {
	return Migration{
		FromVersion: 8,
		ToVersion:   9,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
				NamedStep{Name: "drop_port_profile_enabled_columns", Run: dropPortProfileEnabledColumns},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
