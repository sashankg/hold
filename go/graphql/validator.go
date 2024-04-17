package graphql

import (
	"context"
	"errors"
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/sashankg/hold/dao"
)

type Validator interface {
	ValidateRootSelections(
		ctx context.Context,
		doc *ast.Document,
	) error
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
) error {
	return iterateRootFields(doc, func(field *ast.Field) error {
		collectionSpec, schemaErr := rootFieldToCollectionSpec(field)
		if schemaErr != nil {
			return schemaErr
		}
		collection, err := h.dao.FindCollectionBySpec(ctx, *collectionSpec)
		if err != nil {
			return errors.Join(
				NewInvalidSchemaError("invalid collection: "+collectionSpec.Name, field.Loc),
				err,
			)
		}
		h.validateNestedSelections(ctx, field.SelectionSet, collection)
		return nil
	})
}

func (h *validatorImpl) validateNestedSelections(
	ctx context.Context,
	selections *ast.SelectionSet,
	collectionMap *dao.Collection, /*inout*/
) error {
	for _, sel := range selections.Selections {
		switch sel := sel.(type) {
		case *ast.Field:
			field, ok := collectionMap.Fields[sel.Name.Value]
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
			nestedCollection, err := h.dao.FindCollectionById(ctx, field.Ref)
			if err != nil {
				return NewInvalidSchemaError("invalid collection reference: "+sel.Name.Value, sel.Loc)
			}
			if err := h.validateNestedSelections(ctx, sel.SelectionSet, nestedCollection); err != nil {
				return err
			}
		}
	}
	return nil
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
