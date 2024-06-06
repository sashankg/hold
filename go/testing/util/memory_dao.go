package util

import (
	"database/sql"
	"os"
	"path"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/sashankg/hold/dao"
	"github.com/stretchr/testify/require"
)

type memoryDao struct {
	dao.Dao
	SchemaDb *sql.DB
	RecordDb *sql.DB
}

func NewMemoryDao(t *testing.T) *memoryDao {
	schemaDb, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(os.DirFS(path.Join(cwd, "..", "migrations")))
	require.NoError(t, goose.SetDialect("sqlite3"))
	require.NoError(t, goose.Up(schemaDb, "."))

	recordDb, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	return &memoryDao{
		Dao:      dao.NewDao(schemaDb, recordDb),
		SchemaDb: schemaDb,
		RecordDb: recordDb,
	}
}
