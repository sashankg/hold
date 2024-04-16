package resolvers

import (
	"context"
	"database/sql"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sashankg/hold/core"
	"github.com/sashankg/hold/dao"
	"golang.org/x/exp/maps"
)

type CollectionsResolver struct {
	db          *sql.DB
	dao         dao.Dao
	queryFields graphql.Fields
	types       []graphql.Type
}

func NewCollectionsResolver(
	dao dao.Dao,
) (*CollectionsResolver, error) {
	r := &CollectionsResolver{
		dao:         dao,
		queryFields: graphql.Fields{},
	}
	collections, err := dao.FindCollections(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	r.registerCollections(maps.Values(collections))
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
		collectionId := collection.Id
		for _, field := range collection.Fields {
			typeMap[collectionId].AddFieldConfig(field.Name, &graphql.Field{
				Name: field.Name,
				Type: collectionFieldGqlType(field, typeMap),
			})
		}
		r.queryFields[collection.Name] = &graphql.Field{
			Type: typeMap[collectionId],
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				collectionMap, err := r.dao.FindCollections(
					p.Context,
					collectRequiredCollections(
						p.Info.FieldASTs[0].SelectionSet,
						p.Info.ReturnType,
					),
				)
				json, err := r.dao.GetRecord(
					p.Context,
					p.Args["id"].(int),
					getDaoSelection(p.Info.FieldASTs[0].SelectionSet),
					collectionId,
					collectionMap,
				)
				if err != nil {
					return nil, err
				}
				return &JsonCompletedObject{json}, nil
			},
		}
	}
}

func collectRequiredCollections(
	selectionSet *ast.SelectionSet,
	outputType graphql.Output,
) []string {
	collections := []string{}
			objectType := outputType.(*graphql.Object)
	collections = append(collections, objectType.Name())
	for _, s := range selectionSet.Selections {
		if s, ok := s.(*ast.Field); ok && s.SelectionSet != nil &&
			len(s.SelectionSet.Selections) > 0 {
			collections = append(
				collections,
				collectRequiredCollections(
					s.SelectionSet,
					objectType.Fields()[s.Name.Value].Type,
				)...)
		}
	}
	return collections
}

func getDaoSelection(selectionSet *ast.SelectionSet) []dao.Selection {
	if selectionSet == nil {
		return nil
	}
	selection := []dao.Selection{}
	for _, s := range selectionSet.Selections {
		if s, ok := s.(*ast.Field); ok {
			selection = append(selection, dao.Selection{
				FieldName:     s.Name.Value,
				Subselections: getDaoSelection(s.GetSelectionSet()),
			})
		}
	}
	return selection
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

// Types implements core.TypeSource.
func (r *CollectionsResolver) Types() []graphql.Type {
	return r.types
}

var _ core.QueryResolver = (*CollectionsResolver)(nil)
var _ core.TypeSource = (*CollectionsResolver)(nil)
