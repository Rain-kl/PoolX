package migrate

import "gorm.io/gorm"

func V3(hooks Hooks) Migration {
	return Migration{
		FromVersion: 2,
		ToVersion:   3,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
