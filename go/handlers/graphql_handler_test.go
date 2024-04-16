package handlers_test

import (
	"context"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/handlers"
	"github.com/sashankg/hold/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGraphqlHandler_ValidateRootSelections(t *testing.T) {
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
		FindCollectionFieldsBySpec(gomock.Any(), gomock.Eq(map[dao.CollectionSpec]mapset.Set[string]{
			{Namespace: "", Name: "Post"}: mapset.NewSet("title", "body", "author"),
		})).
		Return(map[dao.CollectionSpec]map[string]dao.CollectionField{
			{Namespace: "", Name: "Post"}: {
				"title": {
					Name: "title",
					Type: "string",
				},
				"body": {
					Name: "body",
					Type: "string",
				},
				"author": {
					Name: "author",
					Type: "Person",
					Ref:  1,
				},
			},
		}, nil)

	mockDao.EXPECT().
		FindCollectionFieldsByCollectionId(gomock.Any(), gomock.Eq(map[int]mapset.Set[string]{
			1: mapset.NewSet("name", "friends"),
		})).
		Return(map[int]map[string]dao.CollectionField{
			1: {
				"name": {
					Name: "name",
					Type: "string",
				},
				"friends": {
					Name:   "friends",
					Type:   "Person",
					Ref:    1,
					IsList: true,
				},
			},
		}, nil)

	mockDao.EXPECT().
		FindCollectionFieldsByCollectionId(gomock.Any(), gomock.Eq(map[int]mapset.Set[string]{
			1: mapset.NewSet("name"),
		})).
		Return(map[int]map[string]dao.CollectionField{
			1: {
				"name": {
					Name: "name",
					Type: "string",
				},
			},
		}, nil)

	h := handlers.NewGraphqlHandler(mockDao)

	err = h.Validate(context.Background(), doc)
	assert.NoError(t, err)
}

func parseGraphql(query string) (*ast.Document, error) {
	return parser.Parse(parser.ParseParams{
		Source: query,
	})
}
