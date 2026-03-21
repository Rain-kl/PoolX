package migrate

import "gorm.io/gorm"

func V5(hooks Hooks) Migration {
	return Migration{
		FromVersion: 4,
		ToVersion:   5,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
