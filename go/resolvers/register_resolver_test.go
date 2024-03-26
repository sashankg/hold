package resolvers_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/graphql-go/graphql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashankg/hold/resolvers"
	"github.com/stretchr/testify/assert"
)

func TestRegisterResolveMutation(t *testing.T) {

	graphql.NewSchema(graphql.SchemaConfig{})

	r := resolvers.NewRegisterResolver(nil, nil)

	_, err := r.ResolveMutation(graphql.ResolveParams{
		Context: context.Background(),
		Args: map[string]any{
			"schema": `type Location { 
				lat: Float, 
				lon: Float 
			}`,
		},
	})
	assert.NoError(t, err)

	row := db.QueryRow("PRAGMA table_info(Location)")
	var result string
	println(row.Scan(&result).Error())
	println(result)

	t.Fail()
}
