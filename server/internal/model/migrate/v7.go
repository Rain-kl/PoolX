package migrate

import "gorm.io/gorm"

func V7(hooks Hooks) Migration {
	return Migration{
		FromVersion: 6,
		ToVersion:   7,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
