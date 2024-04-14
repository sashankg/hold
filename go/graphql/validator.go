package graphql

import (
	"context"
	"fmt"
	"regexp"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/sashankg/hold/dao"
)

type Validator interface {
	ValidateRootSelections(
		ctx context.Context,
		doc *ast.Document,
	) error
}

type validatorImpl struct {
	dao dao.Dao
}

func NewValidator(dao dao.Dao) *validatorImpl {
	return &validatorImpl{
		dao: dao,
	}
}

func (h *validatorImpl) ValidateRootSelections(
	ctx context.Context,
	doc *ast.Document,
) error {
	fieldMap, err := collectRequiredFieldsBySpec(doc)
	if err != nil {
		return err
	}
	collections, err := h.dao.FindCollectionFieldsBySpec(ctx, fieldMap)
	if err != nil {
		return err
	}
	nestedSelections := map[int][]*ast.SelectionSet{}
	for _, def := range doc.Definitions {
		switch def := def.(type) {
		case *ast.OperationDefinition:
			for _, opSel := range def.SelectionSet.Selections {
				switch opSel := opSel.(type) {
				case *ast.Field:
					collectionSpec, err := rootFieldToCollectionSpec(opSel)
					if err != nil {
						return err
					}
					collection, ok := collections[*collectionSpec]
					if !ok {
						// not a real collection
						return NewInvalidSchemaError("invalid collection: "+collectionSpec.Name, opSel.Loc)
					}
					collectionNestedSelections(collection, opSel.SelectionSet, nestedSelections)
				}
			}
		}
	}
	return h.validateNestedSelections(ctx, nestedSelections)
}

func (h *validatorImpl) validateNestedSelections(
	ctx context.Context,
	selections map[int][]*ast.SelectionSet,
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
	nestedSelections := map[int][]*ast.SelectionSet{}
	for ref, selectionSets := range selections {
		collection, ok := collections[ref]
		if !ok {
			return fmt.Errorf("invalid collection id: %d", ref)
		}
		for _, selectionSet := range selectionSets {
			collectionNestedSelections(collection, selectionSet, nestedSelections)
		}
	}
	return h.validateNestedSelections(ctx, nestedSelections)
}

func collectionNestedSelections(
	collection map[string]dao.CollectionField,
	selectionSet *ast.SelectionSet,
	nestedSelections map[int][]*ast.SelectionSet, /*inout*/
) error {
	for _, sel := range selectionSet.Selections {
		switch sel := sel.(type) {
		case *ast.Field:
			field, ok := collection[sel.Name.Value]
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

func collectRequiredFieldsBySpec(
	doc *ast.Document,
) (map[dao.CollectionSpec]mapset.Set[string], error) {
	fields := map[dao.CollectionSpec]mapset.Set[string]{}
	for _, def := range doc.Definitions {
		switch def := def.(type) {
		case *ast.OperationDefinition:
			for _, opSel := range def.SelectionSet.Selections {
				switch opSel := opSel.(type) {
				case *ast.Field:
					collectionSpec, err := rootFieldToCollectionSpec(opSel)
					if err != nil {
						return nil, err
					}
					if _, ok := fields[*collectionSpec]; !ok {
						fields[*collectionSpec] = mapset.NewSet[string]()
					}
					for _, sel := range opSel.SelectionSet.Selections {
						switch sel := sel.(type) {
						case *ast.Field:
							fields[*collectionSpec].Add(sel.Name.Value)
						}
					}
				}
			}
		}
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

func rootFieldToCollectionSpec(
	def *ast.Field,
) (*dao.CollectionSpec, *InvalidSchemaError) {
	matcher, err := regexp.Compile("^(?:find|list|patch|set)([A-Z][a-zA-Z]*)$")
	if err != nil {
		panic(err)
	}
	matches := matcher.FindStringSubmatch(def.Name.Value)
	if len(matches) != 2 {
		return nil, NewInvalidSchemaError(
			"invalid operation name: "+def.Name.Value+" should match "+matcher.String(),
			def.Loc,
		)
	}
	namespace, schemaErr := getNamespace(def.Directives)
	if schemaErr != nil {
		return nil, schemaErr
	}
	return &dao.CollectionSpec{Namespace: namespace, Name: matches[1]}, nil
}

type HasDirectives interface {
	GetDirectives() []*ast.Directive
}

func getNamespace(directives []*ast.Directive) (string, *InvalidSchemaError) {
	for _, directive := range directives {
		if directive.Name.Value != "namespace" {
			continue
		}
		if len(directive.Arguments) > 1 || directive.Arguments[0].Name.Value != "name" ||
			directive.Arguments[0].Value.GetKind() != kinds.StringValue {
			return "", NewInvalidSchemaError(
				"namespace directive should take one string argument named 'name'",
				directive.Loc,
			)
		}
		return directive.Arguments[0].Value.GetValue().(string), nil
	}
	return "", nil
}
