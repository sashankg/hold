package handlers

import (
	"net/http"

	pb_core "github.com/pocketbase/pocketbase/core"
	"github.com/sashankg/hold/core"
)

type collectionsHandler struct {
	pbApp *pb_core.BaseApp
}

// Route implements core.Route.
func (*collectionsHandler) Route() string {
	return "/collections"
}

// ServeHTTP implements core.Route.
func (*collectionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
	}
}

func NewCollectionsHandler(pbApp *pb_core.BaseApp) *collectionsHandler {
	return &collectionsHandler{
		pbApp,
	}
}

var _ core.Route = (*collectionsHandler)(nil)
