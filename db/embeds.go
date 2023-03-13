// Package db contains application-specific db schemas and scripts
package db

import (
	"embed"
	"io/fs"

	"github.com/sargassum-world/godest/database"
	sessions "github.com/sargassum-world/godest/session/sqlitestore"

	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

// Randomly-generated 32-bit integer for the pslive app, to prevent accidental migration of database
// files from other applications.
const appID = 370761302

// Migrations

var DomainEmbeds map[string]database.DomainEmbeds = map[string]database.DomainEmbeds{
	"chat":        chat.NewDomainEmbeds(),
	"instruments": instruments.NewDomainEmbeds(),
	"sessions":    sessions.NewDomainEmbeds(),
}

var MigrationFiles []database.MigrationFile = []database.MigrationFile{
	{Domain: "chat", File: chat.MigrationFiles[0]},
	{Domain: "instruments", File: instruments.MigrationFiles[0]},
	{Domain: "chat", File: chat.MigrationFiles[1]},
	{Domain: "instruments", File: instruments.MigrationFiles[1]},
	{Domain: "instruments", File: instruments.MigrationFiles[2]},
	{Domain: "sessions", File: sessions.MigrationFiles[0]},
	{Domain: "instruments", File: instruments.MigrationFiles[3]},
	{Domain: "instruments", File: instruments.MigrationFiles[4]},
	{Domain: "instruments", File: instruments.MigrationFiles[5]},
	{Domain: "instruments", File: instruments.MigrationFiles[6]},
}

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
		DomainEmbeds:         DomainEmbeds,
		MigrationFiles:       MigrationFiles,
		PrepareConnQueriesFS: prepareConnQueriesFS,
	}
}
