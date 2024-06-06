package graphql

import (
	"context"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/sashankg/hold/dao"
)

const (
	cPlaceholderObjectType = "Object"
)

type Registrar interface {
	RegisterSchema(context.Context, *ast.Document) error
}

type registrarImpl struct {
	dao dao.CollectionDao
}

func NewRegistrar(dao dao.CollectionDao) Registrar {
	return &registrarImpl{
		dao,
	}
}

type objectFieldSpec struct {
	dao.CollectionField
	namespace string
}

// RegisterSchema implements Registrar.
func (r *registrarImpl) RegisterSchema(ctx context.Context, doc *ast.Document) error {
	collections := []*dao.Collection{}
	objectFields := [][]objectFieldSpec{}
	for _, def := range doc.Definitions {
		switch def := def.(type) {
		case *ast.ObjectDefinition:
			namespace, err := getNamespace(def.Directives)
			if err != nil {
				return err
			}
			collection := &dao.Collection{
				Name:   def.Name.Value,
				Domain: namespace,
				Fields: map[string]dao.CollectionField{},
			}
			i := len(collections)
			objectFields = append(objectFields, []objectFieldSpec{})
			for _, fieldDef := range def.Fields {
				field, isScalar := getScalarCollectionField(fieldDef.Name.Value, fieldDef.Type, false, true)
				if isScalar {
					collection.Fields[fieldDef.Name.Value] = field
				} else {
					fieldNamespace, err := getNamespace(fieldDef.Directives)
					if err != nil {
						return err
					}
					objectFields[i] = append(objectFields[i], objectFieldSpec{
						CollectionField: field,
						namespace:       fieldNamespace,
					})
				}
			}
			collections = append(collections, collection)
		}
	}
	err := r.dao.AddCollections(ctx, collections)
	if err != nil {
		return err
	}

	for i, fields := range objectFields {
		for _, spec := range fields {
			refCollection, err := r.dao.FindCollectionBySpec(
				ctx,
				dao.CollectionSpec{Name: spec.CollectionField.Type, Namespace: spec.namespace},
			)
			if err != nil {
				return err
			}
			field := spec.CollectionField
			field.Ref = refCollection.Id
			r.dao.AddCollectionField(ctx, collections[i], field)
		}
	}

	return nil
}

func getScalarCollectionField(
	fieldName string,
	fieldType ast.Type,
	isList bool,
	isNullable bool,
) (dao.CollectionField, bool) {
	switch fieldType := fieldType.(type) {
	case *ast.Named:
		switch fieldType.Name.Value {
		case "Int":
			fallthrough
		case "Float":
			fallthrough
		case "String":
			fallthrough
		case "Boolean":
			fallthrough
		case "ID":
			return dao.CollectionField{
				Name:   fieldName,
				Type:   fieldType.Name.Value,
				IsList: isList,
			}, true
		}
		return dao.CollectionField{
			Name:   fieldName,
			Type:   fieldType.Name.Value,
			IsList: isList,
		}, false
	case *ast.List:
		return getScalarCollectionField(fieldName, fieldType.Type, true, isNullable)
	case *ast.NonNull:
		return getScalarCollectionField(fieldName, fieldType.Type, isList, false)
	}
	panic("invalid field definition type")
}
