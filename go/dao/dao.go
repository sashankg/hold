package dao

import (
	"database/sql"
)

type Dao interface {
	CollectionDao
	RecordDao
}

type daoImpl struct {
	schemaDb *sql.DB
	recordDb *sql.DB
}

var _ Dao = (*daoImpl)(nil)

func NewDao(schemaDb *sql.DB, recordDb *sql.DB) Dao {
	return &daoImpl{
		schemaDb,
		recordDb,
	}
}
