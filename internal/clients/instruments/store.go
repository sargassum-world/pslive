// Package instruments provides a high-level client for management of imaging instruments
package instruments

import (
	_ "embed"

	"github.com/sargassum-world/godest/database"
)

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{
		db: db,
	}
}
