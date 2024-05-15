package graphql

import (
	"regexp"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/sashankg/hold/dao"
)

func iterateRootFields(doc *ast.Document, yield func(*ast.Field) error) error {
	for _, def := range doc.Definitions {
		switch def := def.(type) {
		case *ast.OperationDefinition:
			for _, opSel := range def.SelectionSet.Selections {
				switch opSel := opSel.(type) {
				case *ast.Field:
					if err := yield(opSel); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func rootFieldToCollectionSpec(
	def *ast.Field,
) (*dao.CollectionSpec, error) {
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

func getNamespace(directives []*ast.Directive) (string, error) {
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

func mergeCollectionFields(
	collectionA *dao.Collection,
	collectionB *dao.Collection,
) *dao.Collection {
	if collectionA == nil {
		return collectionB
	}
	if collectionB == nil {
		return collectionA
	}
	for fieldName, field := range collectionB.Fields {
		collectionA.Fields[fieldName] = field
	}
	return collectionA
}
