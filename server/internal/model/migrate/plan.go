package migrate

import "gorm.io/gorm"

// BuildTemplateScheduler assembles the PoolX schema migration plan.
// Concrete upgrade logic lives in v0.go, v1.go, v2.go... while Scheduler only
// performs version traversal, following an Android-style step-by-step upgrade flow.
func BuildTemplateScheduler(
	db *gorm.DB,
	backend string,
	loadVersion VersionLoader,
	saveVersion VersionSaver,
	isDatabaseEmpty EmptyChecker,
	ensureSchemaMeta MetadataMigrator,
	hooks Hooks,
) *Scheduler {
	return NewScheduler(
		db,
		backend,
		loadVersion,
		saveVersion,
		isDatabaseEmpty,
		ensureSchemaMeta,
	).
		RegisterInitializer(V0(hooks)).
		RegisterMigration(V1(hooks)).
		RegisterMigration(V2(hooks)).
		RegisterMigration(V3(hooks)).
		RegisterMigration(V4(hooks)).
		RegisterMigration(V5(hooks)).
		RegisterMigration(V6(hooks)).
		RegisterMigration(V7(hooks)).
		RegisterMigration(V8(hooks)).
		RegisterMigration(V9(hooks))
}
