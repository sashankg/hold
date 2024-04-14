package handlers

import (
	"net/http"

	gql_parser "github.com/graphql-go/graphql/language/parser"
	gql_handler "github.com/graphql-go/handler"
	"github.com/sashankg/hold/graphql"
)

type GraphqlHandler struct {
	validator graphql.Validator
}

func NewGraphqlHandler(validator graphql.Validator) *GraphqlHandler {
	return &GraphqlHandler{
		validator,
	}
}

func (h *GraphqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := gql_handler.NewRequestOptions(r)

	doc, err := gql_parser.Parse(gql_parser.ParseParams{Source: opts.Query})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	h.validator.ValidateRootSelections(r.Context(), doc)
}

func (h *GraphqlHandler) Route() string {
	return "/graphql"
}
