package graphql

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sashankg/hold/dao"
)

type Validator interface {
	ValidateRootSelections(
		ctx context.Context,
		doc *ast.Document,
	) (*ValidationResult, error)
}

type validatorImpl struct {
	dao dao.CollectionDao
}

var _ Validator = (*validatorImpl)(nil)

func NewValidator(dao dao.CollectionDao) *validatorImpl {
	return &validatorImpl{
		dao: dao,
	}
}

type ValidationResult struct {
	CollectionMap          map[int]*dao.Collection
	RootFieldCollectionIds map[dao.CollectionSpec]int
}

func (h *validatorImpl) ValidateRootSelections(
	ctx context.Context,
	doc *ast.Document,
) (*ValidationResult, error) {
	fieldMap, err := aggregateRequiredFieldsBySpec(doc)
	if err != nil {
		return nil, err
	}
	collectionMap := map[int]*dao.Collection{}
	rootFieldCollectionIds := map[dao.CollectionSpec]int{}
	collections, err := h.dao.FindCollectionFieldsBySpec(ctx, fieldMap)
	if err != nil {
		return nil, err
	}
	nestedSelections := map[int][]*ast.SelectionSet{}
	err = iterateRootFields(doc, func(field *ast.Field) error {
		collectionSpec, err := rootFieldToCollectionSpec(field)
		if err != nil {
			return err
		}
		collection, ok := collections[*collectionSpec]
		if !ok {
			return NewInvalidSchemaError("invalid collection: "+collectionSpec.Name, field.Loc)
		}
		collectionMap[collection.Id] = mergeCollectionFields(
			collectionMap[collection.Id],
			collection,
		)
		rootFieldCollectionIds[*collectionSpec] = collection.Id
		if !ok {
			// not a real collection
			return NewInvalidSchemaError("invalid collection: "+collectionSpec.Name, field.Loc)
		}
		aggregateNestedSelections(collection, field.SelectionSet, nestedSelections)
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = h.validateNestedSelections(ctx, nestedSelections, collectionMap)
	if err != nil {
		return nil, err
	}
	return &ValidationResult{
		CollectionMap:          collectionMap,
		RootFieldCollectionIds: rootFieldCollectionIds,
	}, nil
}

func (h *validatorImpl) validateNestedSelections(
	ctx context.Context,
	selections map[int][]*ast.SelectionSet,
	collectionMap map[int]*dao.Collection, /*inout*/
) error {
	if len(selections) == 0 {
		return nil
	}
	fields := map[int]mapset.Set[string]{}
	for ref, selectionSets := range selections {
		if _, ok := fields[ref]; !ok {
			fields[ref] = mapset.NewSet[string]()
		}
		for _, selectionSet := range selectionSets {
			for _, sel := range selectionSet.Selections {
				switch sel := sel.(type) {
				case *ast.Field:
					fields[ref].Add(sel.Name.Value)
				}
			}
		}
	}
	collections, err := h.dao.FindCollectionFieldsByCollectionId(ctx, fields)
	if err != nil {
		return err
	}
	for _, collection := range collections {
		collectionMap[collection.Id] = mergeCollectionFields(
			collectionMap[collection.Id],
			collection,
		)
	}
	nestedSelections := map[int][]*ast.SelectionSet{}
	for ref, selectionSets := range selections {
		collection, ok := collections[ref]
		if !ok {
			return fmt.Errorf("invalid collection id: %d", ref)
		}
		for _, selectionSet := range selectionSets {
			aggregateNestedSelections(collection, selectionSet, nestedSelections)
		}
	}
	return h.validateNestedSelections(ctx, nestedSelections, collectionMap)
}

func aggregateNestedSelections(
	collection *dao.Collection,
	selectionSet *ast.SelectionSet,
	nestedSelections map[int][]*ast.SelectionSet, /*inout*/
) error {
	for _, sel := range selectionSet.Selections {
		switch sel := sel.(type) {
		case *ast.Field:
			field, ok := collection.Fields[sel.Name.Value]
			if !ok {
				// not a real field
				return NewInvalidSchemaError("invalid field: "+sel.Name.Value, sel.Loc)
			}
			if field.Ref == 0 && sel.SelectionSet != nil {
				// not a reference field
				return NewInvalidSchemaError("field not object type: "+sel.Name.Value, sel.Loc)
			}
			if field.Ref > 0 && sel.SelectionSet == nil {
				// not a reference field
				return NewInvalidSchemaError("need a selection set for object fields: "+sel.Name.Value, sel.Loc)
			}
			if sel.SelectionSet == nil {
				continue
			}
			nestedSelections[field.Ref] = append(nestedSelections[field.Ref], sel.SelectionSet)
		}
	}
	return nil
}

func aggregateRequiredFieldsBySpec(
	doc *ast.Document,
) (map[dao.CollectionSpec]mapset.Set[string], error) {
	fields := map[dao.CollectionSpec]mapset.Set[string]{}
	err := iterateRootFields(doc, func(field *ast.Field) error {
		collectionSpec, err := rootFieldToCollectionSpec(field)
		if err != nil {
			return err
		}
		if _, ok := fields[*collectionSpec]; !ok {
			fields[*collectionSpec] = mapset.NewSet[string]()
		}
		for _, sel := range field.SelectionSet.Selections {
			switch sel := sel.(type) {
			case *ast.Field:
				fields[*collectionSpec].Add(sel.Name.Value)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fields, nil
}

type InvalidSchemaError struct {
	reason   string
	location *ast.Location
}

func NewInvalidSchemaError(reason string, location *ast.Location) *InvalidSchemaError {
	return &InvalidSchemaError{reason: reason, location: location}
}

func (e *InvalidSchemaError) Error() string {
	return fmt.Sprintf("invalid schema: %s at %d", e.reason, e.location.Start)
}
