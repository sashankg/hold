package handlers

import (
	"encoding/json"
	"net/http"

	gql_parser "github.com/graphql-go/graphql/language/parser"
	gql_handler "github.com/graphql-go/handler"
	"github.com/sashankg/hold/graphql"
)

type GraphqlHandler struct {
	validator graphql.Validator
	resolver  graphql.Resolver
}

func NewGraphqlHandler(validator graphql.Validator, resolver graphql.Resolver) *GraphqlHandler {
	return &GraphqlHandler{
		validator,
		resolver,
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
	println("successfully parsed")

	valRes, err := h.validator.ValidateRootSelections(r.Context(), doc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	println("successfully validated")

	responseData, err := h.resolver.Resolve(r.Context(), doc, valRes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	println("successfully resolved")
	response, err := json.Marshal(map[string]graphql.JsonValue{
		"data": responseData,
	})
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *GraphqlHandler) Route() string {
	return "/graphql"
}
