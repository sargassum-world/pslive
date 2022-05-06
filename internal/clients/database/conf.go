package database

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
	"zombiezen.com/go/sqlite"
)

// User-configured settings

const envPrefix = "DATABASE_"

type Config struct {
	URI           string
	Flags         sqlite.OpenFlags
	WritePoolSize int
	ReadPoolSize  int
}

func GetConfig() (c Config, err error) {
	c.URI = env.GetString(envPrefix+"URI", "file:db.sqlite3")

	c.Flags = sqlite.OpenURI | sqlite.OpenNoMutex | sqlite.OpenSharedCache | sqlite.OpenWAL
	memory, err := env.GetBool(envPrefix + "MEMORY")
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make SQLite in-memory config")
	}
	if memory {
		c.Flags = c.Flags | sqlite.OpenMemory
	}

	const defaultWritePoolSize = 1
	rawWritePoolSize, err := env.GetInt64(envPrefix+"WRITEPOOL", defaultWritePoolSize)
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make SQLite write pool size config")
	}
	c.WritePoolSize = int(rawWritePoolSize)
	const defaultReadPoolSize = 16
	rawReadPoolSize, err := env.GetInt64(envPrefix+"READPOOL", defaultReadPoolSize)
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make SQLite read pool size config")
	}
	c.ReadPoolSize = int(rawReadPoolSize)
	return c, nil
}
