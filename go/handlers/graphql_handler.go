package handlers

import (
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

type GraphqlHandler struct {
	handler *handler.Handler
}

func NewGraphqlHandler(schema *graphql.Schema) *GraphqlHandler {
	return &GraphqlHandler{
		handler: handler.New(&handler.Config{
			Schema:     schema,
			GraphiQL:   false,
			Playground: true,
		}),
	}
}

func (h *GraphqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *GraphqlHandler) Route() string {
	return "/graphql"
}
