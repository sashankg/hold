package graphql_test

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql/language/parser"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/testing/util"
	"github.com/stretchr/testify/require"
)

func TestRegisterSchema(t *testing.T) {
	testDao := util.NewMemoryDao(t)
	registrar := graphql.NewRegistrar(testDao)

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

	postCollection, err := testDao.FindCollectionBySpec(context.Background(), dao.CollectionSpec{
		Name: "Post",
	})
	require.NoError(t, err)
	require.Greater(t, postCollection.Id, 0)

	personCollection, err := testDao.FindCollectionBySpec(context.Background(), dao.CollectionSpec{
		Name: "Person",
	})
	require.NoError(t, err)
	require.Greater(t, personCollection.Id, 0)

	require.Equal(t, postCollection, &dao.Collection{
		Name:   "Post",
		Domain: "",
		Id:     postCollection.Id,
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
				Ref:  personCollection.Id,
			},
		},
	})

	require.Equal(t, personCollection, &dao.Collection{
		Name:   "Person",
		Domain: "",
		Id:     personCollection.Id,
		Fields: map[string]dao.CollectionField{
			"name": {
				Name: "name",
				Type: "String",
			},
			"friends": {
				Name:   "friends",
				Type:   "Person",
				Ref:    personCollection.Id,
				IsList: true,
			},
		},
	})

	testField := func(collectionName, fieldName, expectedFieldType string) {
		var fieldType string
		require.NoError(t, testDao.RecordDb.QueryRow(`
			SELECT type FROM pragma_table_info(?) WHERE name = ?
			`, collectionName, fieldName).
			Scan(&fieldType))
		require.Equal(t, fieldType, expectedFieldType)
	}

	testField("post", "title", "TEXT")
	testField("post", "body", "TEXT")
	testField("post", "author", "INTEGER")
	testField("person", "name", "TEXT")
	testField("person", "friends", "INTEGER")
}
