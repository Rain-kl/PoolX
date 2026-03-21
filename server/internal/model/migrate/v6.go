package migrate

import "gorm.io/gorm"

func V6(hooks Hooks) Migration {
	return Migration{
		FromVersion: 5,
		ToVersion:   6,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
