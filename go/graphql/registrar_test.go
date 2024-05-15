package graphql_test

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql/language/parser"
	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/testing/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegisterSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	dao := mocks.NewMockDao(ctrl)

	registrar := graphql.NewRegistrar(dao)

	// expectedCollections := []*dao.Collection{
	//     {},
	// }

	dao.EXPECT().AddCollections(gomock.Any() /*ctx*/)

	ast, err := parser.Parse(parser.ParseParams{
		Source: `
			type Post {
				title: String
				body: String
				author: Person

			}
			type Person {
				name: String
				friends: [Person]
			}
		`,
	})
	require.NoError(t, err)
	require.NoError(
		t,
		registrar.RegisterSchema(context.Background(), ast),
	)
}
