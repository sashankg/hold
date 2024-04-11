package resolvers_test

import (
	"database/sql"
	"testing"

	"github.com/graphql-go/graphql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashankg/hold/resolvers"
	"github.com/stretchr/testify/assert"
)

func TestPut(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	r := resolvers.NewKvResolver(db)

	schema, err := resolvers.NewGraphqlSchema([]any{r})
	assert.NoError(t, err)

	params := graphql.Params{Schema: *schema, RequestString: `
		mutation {
			kv {
				put(key: "foo", value: "[1, 2, 3]")
			}
		}
	`}
	res := graphql.Do(params)

	assert.Equal(t, res.Data, map[string]interface{}{
		"kv": map[string]interface{}{
			"put": "[1, 2, 3]",
		},
	})
}
