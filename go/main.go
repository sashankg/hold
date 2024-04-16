package main

import (
	"database/sql"
	"embed"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	goose "github.com/pressly/goose/v3"

	"net/http/pprof"
	_ "net/http/pprof"

	"github.com/sashankg/hold/core"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/handlers"
	"github.com/sashankg/hold/server"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	ln, err := server.NewTcpListener()
	if err != nil {
		panic(err)
	}

	schemaDb, err := NewSchemaDb()
	if err != nil {
		panic(err)
	}

	recordDb, err := NewRecordDb()
	if err != nil {
		panic(err)
	}

	daoObj := dao.NewDao(schemaDb, recordDb)

	server := NewServer([]core.Route{
		handlers.NewUploadHandler(&handlers.FnvHasher{}),
		handlers.NewGraphqlHandler(
			graphql.NewValidator(daoObj),
			graphql.NewResolver(daoObj),
		),
	})
	panic(server.Serve(ln))
}

func NewServer(routes []core.Route) *http.Server {
	serveMux := http.NewServeMux()
	for _, route := range routes {
		serveMux.Handle(route.Route(), route)
	}
	serveMux.Handle("/debug/pprof/", pprof.Handler("heap"))
	return &http.Server{
		Handler: serveMux,
	}
}

func NewSchemaDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "schema.db")
	if err != nil {
		return nil, err
	}
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}
	return db, err
}

func NewRecordDb() (*sql.DB, error) {
	return sql.Open("sqlite3", "record.db")
}
