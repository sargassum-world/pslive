// Package instruments provides a high-level client for management of imaging instruments
package instruments

import (
	"context"

	"github.com/pkg/errors"
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

func executeComponentInsert[ID ~int64, Model SubcomponentInsertable[ID]](
	ctx context.Context, query string, model Model, db *database.DB,
) (insertedID ID, err error) {
	rowID, err := db.ExecuteInsertionForID(ctx, query, model.NewInsertion())
	if err != nil {
		return 0, errors.Wrapf(
			err, "couldn't add %T %d for instrument %d", model, model.GetID(), model.GetInstrumentID(),
		)
	}
	// TODO: instead of returning the raw ID, return the frontend-facing ID as a salted SHA-256 hash
	// of the ID to mitigate the insecure direct object reference vulnerability and avoid leaking
	// info about instrument creation
	return ID(rowID), nil
}

func executeUpdate[ID ~int64, Model Updateable[ID]](
	ctx context.Context, query string, model Model, db *database.DB,
) error {
	return errors.Wrapf(
		db.ExecuteUpdate(ctx, query, model.NewUpdate()), "couldn't update %T %d", model, model.GetID(),
	)
}

func executeDelete[ID ~int64, Model Deletable[ID]](
	ctx context.Context, query string, model Model, db *database.DB,
) error {
	return errors.Wrapf(
		db.ExecuteDelete(ctx, query, model.NewDelete()), "couldn't delete %T %d", model, model.GetID(),
	)
}
