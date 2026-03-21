package migrate

import "gorm.io/gorm"

func V10(hooks Hooks) Migration {
	return Migration{
		FromVersion: 9,
		ToVersion:   10,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
