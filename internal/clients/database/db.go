// Package database provides a SQLite-backed store with migration support
package database

import (
	"context"
	"io/fs"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
)

// Options

type DBOption func(*DB)

func WithQueries(queries fs.FS) DBOption {
	return func(db *DB) {
		db.queries = queries
	}
}

func WithPrepareConnQueries(queries fs.FS) DBOption {
	return func(db *DB) {
		db.prepareConnQueries = queries
	}
}

// DB

type DB struct {
	Config             Config
	queries            fs.FS
	prepareConnQueries fs.FS

	// Connection management
	connPools  map[*sqlite.Conn]*sqlitex.Pool
	connPoolsL *sync.RWMutex
	writePool  *sqlitex.Pool
	readPool   *sqlitex.Pool
}

func NewDB(c Config, opts ...DBOption) (db *DB) {
	db = &DB{
		Config:     c,
		connPools:  make(map[*sqlite.Conn]*sqlitex.Pool),
		connPoolsL: &sync.RWMutex{},
	}

	for _, storeOption := range opts {
		storeOption(db)
	}
	return db
}

func (db *DB) Open() (err error) {
	if db.writePool, err = sqlitex.Open(
		db.Config.URI, db.Config.Flags|sqlite.OpenReadWrite|sqlite.OpenCreate, db.Config.WritePoolSize,
	); err != nil {
		return errors.Wrap(err, "couldn't open writer pool")
	}
	if db.readPool, err = sqlitex.Open(
		db.Config.URI, db.Config.Flags|sqlite.OpenReadOnly, db.Config.ReadPoolSize,
	); err != nil {
		return errors.Wrap(err, "couldn't open reader pool")
	}
	return nil
}

func (db *DB) Close() error {
	if err := db.writePool.Close(); err != nil {
		return errors.Wrap(err, "couldn't close writer pool")
	}
	if err := db.readPool.Close(); err != nil {
		return errors.Wrap(err, "couldn't close reader pool")
	}
	return nil
}

// Connection Acquisition

func (db *DB) prepare(conn *sqlite.Conn, pool *sqlitex.Pool) error {
	if db.prepareConnQueries == nil {
		return nil
	}

	db.connPoolsL.RLock()
	_, initialized := db.connPools[conn]
	db.connPoolsL.RUnlock()
	if initialized {
		return nil
	}

	// TODO: embed the pragma queries for foreign keys, synchronous, and auto-checkpoint so that we
	// can parameterize them from environment variables
	queries, err := readQueries(db.prepareConnQueries, filterQuery)
	if err != nil {
		return errors.Wrap(err, "couldn't read connection preparation queries")
	}
	for _, query := range queries {
		// We run these as transient queries because non-transient query caching is per-connection, so
		// query caching provides no benefit for queries which are only run once per connection.
		if err := sqlitex.ExecuteTransient(conn, strings.TrimSpace(query), nil); err != nil {
			return errors.Wrap(err, "couldn't run connection preparation query")
		}
	}

	db.connPoolsL.Lock()
	defer db.connPoolsL.Unlock()

	db.connPools[conn] = pool
	return nil
}

func (db *DB) acquire(ctx context.Context, writable bool) (*sqlite.Conn, error) {
	pool := db.readPool
	if writable {
		pool = db.writePool
	}

	conn := pool.Get(ctx)
	if conn == nil {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "couldn't get connection from pool")
		}
		return nil, errors.New("couldn't get connection from a closed pool")
	}
	if err := db.prepare(conn, pool); err != nil {
		pool.Put(conn)
		return nil, errors.Wrap(err, "couldn't prepare connection")
	}
	return conn, nil
}

func (db *DB) AcquireReader(ctx context.Context) (*sqlite.Conn, error) {
	conn, err := db.acquire(ctx, false)
	return conn, errors.Wrap(err, "couldn't acquire reader")
}

func (db *DB) AcquireWriter(ctx context.Context) (*sqlite.Conn, error) {
	conn, err := db.acquire(ctx, true)
	return conn, errors.Wrap(err, "couldn't acquire writer")
}

func (db *DB) ReleaseReader(conn *sqlite.Conn) {
	go db.readPool.Put(conn)
}

func (db *DB) ReleaseWriter(conn *sqlite.Conn) {
	// TODO: for writer connections, run the sqlite PRAGMA optimize command in the goroutine, with a
	// PRAGMA analysis_limit=1000 on the connection, if it hasn't been run on that connection for a
	// while. Log an error if it fails.
	go db.writePool.Put(conn)
}

func (db *DB) Migrate(ctx context.Context, schema sqlitemigration.Schema) error {
	// TODO: also implement down-migrations
	conn, err := db.AcquireWriter(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't acquire connection to migrate schemas")
	}
	defer db.ReleaseWriter(conn)

	err = sqlitemigration.Migrate(ctx, conn, schema)
	return errors.Wrap(err, "couldn't migrate schemas")
}

// Statement Execution

func (db *DB) Execute(conn *sqlite.Conn, queryFile string, opts *sqlitex.ExecOptions) error {
	return sqlitex.ExecuteFS(conn, db.queries, queryFile, opts)
}

func (db *DB) ExecuteScript(conn *sqlite.Conn, queryFile string, opts *sqlitex.ExecOptions) error {
	return sqlitex.ExecuteScriptFS(conn, db.queries, queryFile, opts)
}
