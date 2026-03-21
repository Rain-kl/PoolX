package migrate

import "gorm.io/gorm"

type v4NodeTestResult struct{}

func (v4NodeTestResult) TableName() string {
	return "node_test_results"
}

func dropNodeTestResultsTable(db *gorm.DB, backend string) error {
	_ = backend
	return db.Migrator().DropTable(&v4NodeTestResult{})
}

func V4(hooks Hooks) Migration {
	return Migration{
		FromVersion: 3,
		ToVersion:   4,
		Migrate: func(db *gorm.DB, backend string) error {
			return runNamedSteps(db, backend,
				NamedStep{Name: "apply_current_schema", Run: hooks.ApplySchema},
				NamedStep{Name: "drop_node_test_results_table", Run: dropNodeTestResultsTable},
			)
		},
		Validate: hooks.ValidateSchema,
	}
}
