package resolvers

import (
	"context"
	"database/sql"

	"github.com/graphql-go/graphql"
	"github.com/sashankg/hold/core"
	"github.com/sashankg/hold/dao"
)

type CollectionsResolver struct {
	db          *sql.DB
	dao         dao.Dao
	queryFields graphql.Fields
	types       []graphql.Type
}

func NewCollectionsResolver(
	dao dao.Dao,
	db *sql.DB,
) (*CollectionsResolver, error) {
	r := &CollectionsResolver{
		dao:         dao,
		queryFields: graphql.Fields{},
	}

	collections, err := dao.ListCollections(context.Background())
	if err != nil {
		return nil, err
	}
	r.registerCollections(collections)

	return r, nil
}

// QueryFields implements core.QueryResolver.
func (r *CollectionsResolver) QueryFields() graphql.Fields {
	return r.queryFields
}

func (r *CollectionsResolver) registerCollections(collections []*dao.Collection) {
	typeMap := map[int]*graphql.Object{}
	for _, collection := range collections {
		typeMap[collection.Id] = graphql.NewObject(graphql.ObjectConfig{
			Name:   collection.Name,
			Fields: graphql.Fields{},
		})
	}
	for _, t := range typeMap {
		r.types = append(r.types, t)
	}
	for _, collection := range collections {
		for _, field := range collection.Fields {
			typeMap[collection.Id].AddFieldConfig(field.Name, &graphql.Field{
				Name: field.Name,
				Type: collectionFieldGqlType(field, typeMap),
			})
		}
		r.queryFields[collection.Name] = &graphql.Field{
			Type: typeMap[collection.Id],
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				r.dao
			},
		}
	}
}

func collectionFieldGqlType(
	field dao.CollectionField,
	typeMap map[int]*graphql.Object,
) graphql.Type {
	var t graphql.Type
	if field.Ref > 0 {
		t = typeMap[field.Ref]
	} else {
		switch field.Type {
		case "string":
			t = graphql.String
		case "int":
			t = graphql.Int
		case "float":
			t = graphql.Float
		case "bool":
			t = graphql.Boolean
		default:
			return graphql.String
		}
	}
	if field.IsList {
		t = graphql.NewList(t)
	}
	return t
}

type JsonCompletedObject struct {
	json string
}

// Serialize implements graphql.ResolvedObject.
func (o *JsonCompletedObject) CompletedObjectResult() interface{} {
	return o
}

func (o *JsonCompletedObject) MarshalJSON() ([]byte, error) {
	return []byte(o.json), nil
}

var _ graphql.CompletedObject = (*JsonCompletedObject)(nil)

// Types implements core.TypeSource.
func (r *CollectionsResolver) Types() []graphql.Type {
	return r.types
}

var _ core.QueryResolver = (*CollectionsResolver)(nil)
var _ core.TypeSource = (*CollectionsResolver)(nil)
