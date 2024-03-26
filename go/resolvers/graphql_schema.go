package resolvers

import (
	"github.com/graphql-go/graphql"
	"github.com/sashankg/hold/core"
)

func NewGraphqlSchema(
	resolvers []any,
) (*graphql.Schema, error) {
	queryFields := graphql.Fields{}
	mutationFields := graphql.Fields{}
	types := []graphql.Type{}
	extensions := []graphql.Extension{}
	for _, r := range resolvers {
		if r, ok := r.(core.QueryResolver); ok {
			for k, v := range r.QueryFields() {
				queryFields[k] = v
			}
		}
		if r, ok := r.(core.MutationResolver); ok {
			for k, v := range r.MutationFields() {
				mutationFields[k] = v
			}
		}
		if r, ok := r.(core.TypeSource); ok {
			for _, t := range r.Types() {
				types = append(types, t)
			}
		}
	}
	query := graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: queryFields,
	})
	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootMutation",
		Fields: mutationFields,
	})
	println("mutation", mutation)
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: query,
		// Mutation:   mutation,
		Types:      types,
		Extensions: extensions,
	})
	return &schema, err
}
