package dao

import (
	"context"
	"database/sql"
)

type Dao interface {
	FindCollection(ctx context.Context, name string, domain string) (*Collection, error)
	ListCollections(ctx context.Context) ([]*Collection, error)
	AddCollection(ctx context.Context, collection *Collection) error
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
