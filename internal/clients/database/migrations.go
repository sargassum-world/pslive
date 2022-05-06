package database

import (
	"io/fs"
	"strings"

	"github.com/pkg/errors"
	"zombiezen.com/go/sqlite/sqlitemigration"
)

// Migrations

const migrationUpFileExt = ".up.sql"

func filterMigrationUp(path string) bool {
	return strings.HasSuffix(path, migrationUpFileExt)
}

func NewSchema(
	migrationFS fs.FS, repeatableMigration string, appID int32,
) (sqlitemigration.Schema, error) {
	migrations, err := readQueries(migrationFS, filterMigrationUp)
	if err != nil {
		return sqlitemigration.Schema{}, errors.Wrap(err, "couldn't read migrations")
	}
	return sqlitemigration.Schema{
		Migrations:          migrations,
		RepeatableMigration: repeatableMigration,
		AppID:               appID,
	}, nil
}
