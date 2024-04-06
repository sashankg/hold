package dao

import (
	"context"
	"database/sql"
)

type Dao interface {
	FindCollections(ctx context.Context, names []string) (map[int]*Collection, error)
	AddCollection(ctx context.Context, collection *Collection) error

	GetRecord(
		ctx context.Context,
		id int,
		selection []Selection,
		collectionId int,
		collectionMap map[int]*Collection,
	) (string, error)
}

type daoImpl struct{ db *sql.DB }

func NewDao(db *sql.DB) (Dao, error) {
	tx, err := db.Begin()
	defer tx.Rollback()

	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &daoImpl{db: db}, nil
}
