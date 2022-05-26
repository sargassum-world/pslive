package instruments

import (
	"embed"
	"io/fs"

	"github.com/sargassum-world/pslive/internal/clients/database"
)

// Migrations

var (
	//go:embed migrations/*
	migrationsEFS   embed.FS
	migrationsFS, _ = fs.Sub(migrationsEFS, "migrations")
)

var MigrationFiles []string = []string{
	"1-initialize-schema-v0.1.7",
	"2-rename-user-to-identity-v0.1.10",
}

// Embeds

func NewDomainEmbeds() database.DomainEmbeds {
	return database.DomainEmbeds{
		MigrationsFS: migrationsFS,
	}
}
