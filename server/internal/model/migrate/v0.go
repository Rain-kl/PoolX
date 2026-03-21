package migrate

import "gorm.io/gorm"

func V0(hooks Hooks) Initializer {
	return Initializer{
		Version: 0,
		Initialize: func(db *gorm.DB, backend string) error {
			if err := hooks.ApplySchema(db, backend); err != nil {
				return err
			}
			if hooks.AfterInitialize == nil {
				return nil
			}
			return hooks.AfterInitialize(db, backend)
		},
		Validate: hooks.ValidateSchema,
	}
}
