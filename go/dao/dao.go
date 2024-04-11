package dao

import (
	"database/sql"
)

type Dao interface {
	CollectionDao
	RecordDao
}

type daoImpl struct{ db *sql.DB }

var _ Dao = (*daoImpl)(nil)

func NewDao(db *sql.DB) Dao {
	return &daoImpl{db: db}
}
