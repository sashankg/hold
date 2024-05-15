package graphql_test

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/mocks"
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

	validator := graphql.NewValidator(mockDao)

	err = validator.ValidateRootSelections(context.Background(), doc)
	assert.NoError(t, err)
}

func parseGraphql(query string) (*ast.Document, error) {
	return parser.Parse(parser.ParseParams{
		Source: query,
	})
}
