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

	daoObj := dao.NewDao(db)

	// if err := daoObj.AddCollection(context.Background(), &dao.Collection{
	//     Name:    "posts",
	//     Domain:  "public",
	//     Version: "v1",
	//     Fields: map[string]dao.CollectionField{
	//         "title": {
	//             Name: "title",
	//             Type: "string",
	//         },
	//         "body": {
	//             Name: "body",
	//             Type: "string",
	//         },
	//         "author": {
	//             Name: "author",
	//             Type: "people",
	//         },
	//     },
	// }); err != nil {
	//     panic(err)
	// }

	// if err := daoObj.AddCollection(context.Background(), &dao.Collection{
	//     Name:    "people",
	//     Domain:  "public",
	//     Version: "v1",
	//     Fields: map[string]dao.CollectionField{
	//         "name": {
	//             Name: "name",
	//             Type: "string",
	//         },
	//     },
	// }); err != nil {
	//     panic(err)
	// }

	collectionsResolver, err := resolvers.NewCollectionsResolver(daoObj)
	if err != nil {
		panic(err)
	}

	_, err = resolvers.NewGraphqlSchema(
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
			daoObj,
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

func NewDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "data.db")
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
