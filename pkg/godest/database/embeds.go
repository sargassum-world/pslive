package database

import (
	"io/fs"
	"strings"

	"github.com/pkg/errors"
	"zombiezen.com/go/sqlite/sqlitemigration"
)

type DomainEmbeds struct {
	MigrationsFS fs.FS
}

type MigrationFile struct {
	Domain string
	File   string
}

type Embeds struct {
	AppID int32

	DomainEmbeds   map[string]DomainEmbeds
	MigrationFiles []MigrationFile

	PrepareConnQueriesFS fs.FS
}

const (
	migrationUpFileExt   = ".up.sql"
	migrationDownFileExt = ".down.sql"
)

func (e Embeds) readMigrations(up bool) (migrations []string, err error) {
	fileExt := migrationUpFileExt
	if !up {
		fileExt = migrationDownFileExt
	}
	migrations = make([]string, len(e.MigrationFiles))
	for i, migrationFile := range e.MigrationFiles {
		domainEmbed, ok := e.DomainEmbeds[migrationFile.Domain]
		if !ok {
			return nil, errors.Errorf("couldn't find migration domain %s", migrationFile.Domain)
		}
		rawMigration, err := readFile(migrationFile.File+fileExt, domainEmbed.MigrationsFS)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't read migration file %s%s from domain %s",
				migrationFile.File, fileExt, migrationFile.Domain,
			)
		}
		migrations[i] = strings.TrimSpace(string(rawMigration))
	}
	return migrations, nil
}

func (e Embeds) ReadUpMigrations() (migrations []string, err error) {
	migrations, err = e.readMigrations(true)
	return migrations, errors.Wrap(err, "couldn't read up-migrations")
}

func (e Embeds) ReadDownMigrations() (migrations []string, err error) {
	migrations, err = e.readMigrations(true)
	return migrations, errors.Wrap(err, "couldn't read down-migrations")
}

func (e Embeds) NewSchema() (sqlitemigration.Schema, error) {
	migrations, err := e.ReadUpMigrations()
	if err != nil {
		return sqlitemigration.Schema{}, errors.Wrap(err, "couldn't read migrations")
	}
	return sqlitemigration.Schema{
		AppID:      e.AppID,
		Migrations: migrations,
		// TODO: implement RepeatableMigration support
	}, nil
}
