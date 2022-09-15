// Package db contains application-specific db schemas and scripts
package db

import (
	"embed"
	"io/fs"

	"github.com/sargassum-world/godest/database"

	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

// Randomly-generated 32-bit integer for the pslive app, to prevent migration of database files
// from other applications.
const appID = 370761302

// Migrations

var DomainEmbeds map[string]database.DomainEmbeds = map[string]database.DomainEmbeds{
	"chat":        chat.NewDomainEmbeds(),
	"instruments": instruments.NewDomainEmbeds(),
}

var MigrationFiles []database.MigrationFile = []database.MigrationFile{
	{Domain: "chat", File: chat.MigrationFiles[0]},
	{Domain: "instruments", File: instruments.MigrationFiles[0]},
	{Domain: "chat", File: chat.MigrationFiles[1]},
	{Domain: "instruments", File: instruments.MigrationFiles[1]},
	{Domain: "instruments", File: instruments.MigrationFiles[2]},
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
