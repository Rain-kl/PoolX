package migrate

import "gorm.io/gorm"

func V8(hooks Hooks) Migration {
	return Migration{
		FromVersion: 7,
		ToVersion:   8,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
