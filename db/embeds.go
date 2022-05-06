// Package db contains application-specific db schemas and scripts
package db

import (
	"embed"
	"io/fs"

	"github.com/sargassum-world/pslive/internal/clients/database"
)

// Randomly-generated 32-bit integer for the pslive app, to prevent migration of database files
// from other applications.
const appID = 370761302

// Schemas

var (
	//go:embed schemas/migrations/*
	migrationsEFS   embed.FS
	migrationsFS, _ = fs.Sub(migrationsEFS, "schemas/migrations")
)

//go:embed schemas/repeatable-migration.sql
var repeatableMigration string

// Queries

var (
	//go:embed queries/*
	queriesEFS              embed.FS
	queriesFS, _            = fs.Sub(queriesEFS, "queries")
	prepareConnQueriesFS, _ = fs.Sub(queriesFS, "prepare-conn")
)

// Embeds

func NewEmbeds() database.Embeds {
	return database.Embeds{
		AppID:                appID,
		MigrationsFS:         migrationsFS,
		RepeatableMigration:  repeatableMigration,
		QueriesFS:            queriesFS,
		PrepareConnQueriesFS: prepareConnQueriesFS,
	}
}
