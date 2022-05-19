package database

import (
	_ "embed"
	"strings"

	"github.com/pkg/errors"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed select-last-insert-rowid.sql
var rawSelectLastInsertRowIDQuery string
var selectLastInsertRowIDQuery string = strings.TrimSpace(rawSelectLastInsertRowIDQuery)

func ExecuteInsertion(
	conn *sqlite.Conn, query string, namedParams map[string]interface{},
) (rowID int64, err error) {
	defer sqlitex.Save(conn)(&err)

	if err = sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
		Named: namedParams,
	}); err != nil {
		return 0, errors.Wrapf(err, "couldn't execute insertion statement")
	}

	if err = sqlitex.Execute(conn, selectLastInsertRowIDQuery, &sqlitex.ExecOptions{
		ResultFunc: func(s *sqlite.Stmt) error {
			// TODO: instead
			rowID = s.GetInt64("row_id")
			return nil
		},
	}); err != nil {
		return 0, errors.Wrapf(err, "couldn't look up id of inserted row")
	}
	return rowID, err
}

func ExecuteUpdate(conn *sqlite.Conn, query string, namedParams map[string]interface{}) error {
	return errors.Wrap(
		sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
			Named: namedParams,
		}),
		"couldn't execute update statement",
	)
}

func ExecuteDelete(conn *sqlite.Conn, query string, namedParams map[string]interface{}) error {
	return errors.Wrap(
		sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
			Named: namedParams,
		}),
		"couldn't execute delete statement",
	)
}

func ExecuteSelection(
	conn *sqlite.Conn, query string, namedParams map[string]interface{},
	resultFunc func(s *sqlite.Stmt) error,
) error {
	return errors.Wrap(
		sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
			Named:      namedParams,
			ResultFunc: resultFunc,
		}),
		"couldn't execute selection statement",
	)
}
