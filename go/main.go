package main

import (
	"context"
	"database/sql"
	"embed"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	goose "github.com/pressly/goose/v3"

	"github.com/sashankg/hold/core"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/handlers"
	"github.com/sashankg/hold/resolvers"
	"github.com/sashankg/hold/server"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	ln, err := server.NewTcpListener()
	if err != nil {
		panic(err)
	}

	db, err := NewDb()

	if err != nil {
		panic(err)
	}

	daoObj, err := dao.NewDao(db)
	if err != nil {
		panic(err)
	}

	if err := daoObj.AddCollection(context.Background(), &dao.Collection{
		Name:    "posts",
		Domain:  "public",
		Version: "v1",
		Fields: []dao.CollectionField{
			{
				Name: "title",
				Type: "string",
			},
			{
				Name: "body",
				Type: "string",
			},
		},
	}); err != nil {
		panic(err)
	}

	collectionsResolver, err := resolvers.NewCollectionsResolver(daoObj)
	if err != nil {
		panic(err)
	}

	schema, err := resolvers.NewGraphqlSchema(
		[]any{
			// resolvers.NewKvResolver(db),
			collectionsResolver,
			// &LoggerExtension{},
		},
	)
	if err != nil {
		panic(err)
	}

	server := NewServer([]core.Route{
		handlers.NewUploadHandler(&handlers.FnvHasher{}),
		handlers.NewGraphqlHandler(
			schema,
		),
	})
	panic(server.Serve(ln))
}

func NewServer(routes []core.Route) *http.Server {
	serveMux := http.NewServeMux()
	for _, route := range routes {
		serveMux.Handle(route.Route(), route)
	}
	return &http.Server{
		Handler: serveMux,
	}
}

func NewDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
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
