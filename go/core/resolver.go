package core

import "github.com/graphql-go/graphql"

type QueryResolver interface {
	QueryFields() graphql.Fields
}

type MutationResolver interface {
	MutationFields() graphql.Fields
}

type TypeSource interface {
	Types() []graphql.Type
}
