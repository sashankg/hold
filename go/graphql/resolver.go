package graphql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/sashankg/hold/dao"
)

type Resolver interface {
	Resolve(context.Context, *ast.Document, *ValidationResult) ([]byte, error)
}

type resolverImpl struct {
	dao dao.RecordDao
}

var _ Resolver = (*resolverImpl)(nil)

func NewResolver(dao dao.RecordDao) *resolverImpl {
	return &resolverImpl{
		dao: dao,
	}
}

// Resolve implements Resolver.
func (r *resolverImpl) Resolve(
	ctx context.Context,
	doc *ast.Document,
	valRes *ValidationResult,
) ([]byte, error) {
	result := map[string]JsonValue{}
	iterateRootFields(doc, func(field *ast.Field) error {
		collectionSpec, schemaErr := rootFieldToCollectionSpec(field)
		if schemaErr != nil {
			return schemaErr
		}
		collectionId, ok := valRes.RootFieldCollectionIds[*collectionSpec]
		if !ok {
			return fmt.Errorf("collection not found for root field %s", field.Name.Value)
		}
		recordId, err := getRecordId(field)
		if err != nil {
			return err
		}
		json, err := r.dao.GetRecord(
			ctx,
			recordId,
			getDaoSelection(field.SelectionSet),
			collectionId,
			valRes.CollectionMap,
		)
		result[field.Name.Value] = JsonValue(json)
		return nil
	})
	return json.Marshal(result)
}

type JsonValue []byte

func (v JsonValue) MarshalJSON() ([]byte, error) {
	return []byte(v), nil
}

func getRecordId(field *ast.Field) (int, error) {
	for _, arg := range field.Arguments {
		if arg.Name.Value == "id" {
			value := arg.Value.GetValue()
			if value, ok := value.(int); ok {
				return value, nil
			}
			return 0, fmt.Errorf("id arg needs to be int")
		}
	}
	return 0, fmt.Errorf("no id arg")
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
