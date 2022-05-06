package database

import (
	"io/fs"

	"zombiezen.com/go/sqlite/sqlitemigration"
)

type Embeds struct {
	AppID int32

	// Migrations
	MigrationsFS        fs.FS
	RepeatableMigration string

	// Queries
	QueriesFS            fs.FS
	PrepareConnQueriesFS fs.FS
}

func (e Embeds) NewSchema() (sqlitemigration.Schema, error) {
	return NewSchema(e.MigrationsFS, e.RepeatableMigration, e.AppID)
}
