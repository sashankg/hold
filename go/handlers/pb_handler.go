package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	pb_apis "github.com/pocketbase/pocketbase/apis"
	pb_core "github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/migrations/logs"
	"github.com/pocketbase/pocketbase/tools/migrate"
	"github.com/sashankg/hold/core"
)

type PbHandler struct {
	pbApi *echo.Echo
}

type migrationsConnection struct {
	DB             *dbx.DB
	MigrationsList migrate.MigrationsList
}

func NewPbHandler(pb *pb_core.BaseApp) (*PbHandler, error) {
	api, err := pb_apis.InitApi(pb)
	if err != nil {
		return nil, err
	}

	connections := []migrationsConnection{
		{
			DB:             pb.DB(),
			MigrationsList: migrations.AppMigrations,
		},
		{
			DB:             pb.LogsDB(),
			MigrationsList: logs.LogsMigrations,
		},
	}

	for _, c := range connections {
		runner, err := migrate.NewRunner(c.DB, c.MigrationsList)
		if err != nil {
			return nil, err
		}

		if _, err := runner.Up(); err != nil {
			return nil, err
		}
	}

	return &PbHandler{
		pbApi: api,
	}, nil

}

// ServeHTTP implements core.Route.
func (h *PbHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = r.URL.Path[3:]
	h.pbApi.ServeHTTP(w, r)
}

// Route implements core.Route.
func (*PbHandler) Route() string {
	return "/pb"
}

var _ core.Route = (*PbHandler)(nil)
