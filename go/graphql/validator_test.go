package graphql_test

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/testing/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateRootSelections(t *testing.T) {
	doc, err := parseGraphql(`
		query {
			findPost {
				title
				body
				author {
					name
					friends {
						name
					}
				}
			}
		}
	`)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	mockDao := mocks.NewMockDao(ctrl)

	mockDao.EXPECT().
		FindCollectionBySpec(gomock.Any(), gomock.Eq(dao.CollectionSpec{Name: "Post"})).
		Return(&dao.Collection{
			Name: "Post",
			Fields: map[string]dao.CollectionField{
				"title": {
					Name: "title",
					Type: "String",
				},
				"body": {
					Name: "body",
					Type: "String",
				},
				"author": {
					Name: "author",
					Type: "Person",
					Ref:  2,
				},
			},
		}, nil)
	mockDao.EXPECT().
		FindCollectionById(gomock.Any(), gomock.Eq(2)).
		Return(&dao.Collection{
			Name: "Person",
			Fields: map[string]dao.CollectionField{
				"name": {
					Name: "name",
					Type: "String",
				},
				"friends": {
					Name:   "friends",
					Type:   "Person",
					Ref:    2,
					IsList: true,
				},
			},
		}, nil).Times(2)

	validator := graphql.NewValidator(mockDao)

	err = validator.ValidateRootSelections(context.Background(), doc)
	assert.NoError(t, err)
}

func parseGraphql(query string) (*ast.Document, error) {
	return parser.Parse(parser.ParseParams{
		Source: query,
	})
}
