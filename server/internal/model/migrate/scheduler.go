package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

type Scheduler struct {
	db                 *gorm.DB
	backend            string
	latestVersion      int
	loadVersion        VersionLoader
	saveVersion        VersionSaver
	isDatabaseEmpty    EmptyChecker
	ensureSchemaMeta   MetadataMigrator
	initializer        Initializer
	migrationsByTarget map[int]Migration
}

func NewScheduler(
	db *gorm.DB,
	backend string,
	loadVersion VersionLoader,
	saveVersion VersionSaver,
	isDatabaseEmpty EmptyChecker,
	ensureSchemaMeta MetadataMigrator,
) *Scheduler {
	return &Scheduler{
		db:                 db,
		backend:            backend,
		loadVersion:        loadVersion,
		saveVersion:        saveVersion,
		isDatabaseEmpty:    isDatabaseEmpty,
		ensureSchemaMeta:   ensureSchemaMeta,
		migrationsByTarget: make(map[int]Migration),
	}
}

func (s *Scheduler) RegisterInitializer(initializer Initializer) *Scheduler {
	s.initializer = initializer
	if initializer.Version > s.latestVersion {
		s.latestVersion = initializer.Version
	}
	return s
}

func (s *Scheduler) RegisterMigration(migration Migration) *Scheduler {
	s.migrationsByTarget[migration.ToVersion] = migration
	if migration.ToVersion > s.latestVersion {
		s.latestVersion = migration.ToVersion
	}
	return s
}

func (s *Scheduler) Apply() error {
	version, exists, err := s.loadVersion(s.db)
	if err != nil {
		return err
	}
	if exists {
		return s.upgrade(version)
	}

	empty, err := s.isDatabaseEmpty(s.db)
	if err != nil {
		return err
	}
	if empty {
		if err := s.initializeFresh(); err != nil {
			return err
		}
		return s.upgrade(s.initializer.Version)
	}

	if err := s.ensureSchemaMeta(s.db); err != nil {
		return err
	}
	if s.latestVersion == s.initializer.Version {
		if err := s.initializer.Validate(s.db, s.backend); err != nil {
			return err
		}
		return s.saveVersion(s.db, s.initializer.Version)
	}
	return s.upgrade(s.initializer.Version)
}

func (s *Scheduler) initializeFresh() error {
	if err := s.initializer.Initialize(s.db, s.backend); err != nil {
		return err
	}
	if err := s.initializer.Validate(s.db, s.backend); err != nil {
		return err
	}
	return s.saveVersion(s.db, s.initializer.Version)
}

func (s *Scheduler) upgrade(version int) error {
	if version > s.latestVersion {
		return fmt.Errorf("database schema version %d is newer than application version %d", version, s.latestVersion)
	}
	if version == s.latestVersion {
		return s.initializer.Validate(s.db, s.backend)
	}

	for nextVersion := version + 1; nextVersion <= s.latestVersion; nextVersion++ {
		migration, ok := s.migrationsByTarget[nextVersion]
		if !ok {
			return fmt.Errorf("database schema migration to v%d is not defined", nextVersion)
		}
		if err := s.runMigration(migration); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) runMigration(migration Migration) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := migration.Migrate(tx, s.backend); err != nil {
			return fmt.Errorf("migrate database schema from v%d to v%d failed: %w", migration.FromVersion, migration.ToVersion, err)
		}
		if err := migration.Validate(tx, s.backend); err != nil {
			return fmt.Errorf("validate database schema v%d failed: %w", migration.ToVersion, err)
		}
		if err := s.saveVersion(tx, migration.ToVersion); err != nil {
			return fmt.Errorf("persist database schema version v%d failed: %w", migration.ToVersion, err)
		}
		return nil
	})
}
