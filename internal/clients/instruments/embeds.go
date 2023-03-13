package instruments

import (
	"embed"
	"io/fs"

	"github.com/sargassum-world/godest/database"
)

// Migrations

var (
	//go:embed migrations/*
	migrationsEFS   embed.FS
	migrationsFS, _ = fs.Sub(migrationsEFS, "migrations")
)

var MigrationFiles []string = []string{
	"1-initialize-schema-v0.1.7",
	"2-rename-user-to-identity-v0.1.11",
	"3-add-camera-controller-index-v0.1.15",
	"4-add-enabled-flag-v0.3.4",
	"5-enabled-not-null-v0.3.5",
	"6-add-automation-jobs-v0.3.5",
}

// Embeds

func NewDomainEmbeds() database.DomainEmbeds {
	return database.DomainEmbeds{
		MigrationsFS: migrationsFS,
	}
}
