package migrate

import "gorm.io/gorm"

// BuildTemplateScheduler assembles the PoolX schema migration plan.
// The release baseline starts from schema V1 and does not keep historical
// upgrade chains from pre-release databases.
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
		RegisterMigration(V2(hooks))
}
