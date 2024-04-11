package handlers

import (
	"fmt"
	"net/http"
	"regexp"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	gql_parser "github.com/graphql-go/graphql/language/parser"
	gql_source "github.com/graphql-go/graphql/language/source"
	gql_handler "github.com/graphql-go/handler"
	"github.com/sashankg/hold/dao"
)

type GraphqlHandler struct {
	dao dao.Dao
}

func NewGraphqlHandler(dao dao.Dao) *GraphqlHandler {
	return &GraphqlHandler{}
}

func (h *GraphqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := gql_handler.NewRequestOptions(r)
	source := gql_source.NewSource(&gql_source.Source{
		Body: []byte(opts.Query),
	})

	ast, err := gql_parser.Parse(gql_parser.ParseParams{Source: source})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	_, err = collectRequiredCollectionSpecs(ast)
}

func collectRequiredCollectionSpecs(
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

// func collectRequiredCollectionNames(node any) (mapset.Set[CollectionSpec], error) {
//     switch node := node.(type) {
//     case *ast.Document:
//         for _, def := range node.Definitions {
//             return collectRequiredCollectionNames(def)
//         }
//     case *ast.OperationDefinition:
//         operationSpec, err := operationNameToCollectionName(node)
//         if err != nil {
//             return nil, err
//         }
//         selectionSetSpecs, err := collectRequiredCollectionNames(node.SelectionSet)
//         if err != nil {
//             return nil, err
//         }
//         selectionSetSpecs.Add(*operationSpec)
//         return selectionSetSpecs, nil
//     case *ast.SelectionSet:
//         for _, sel := range node.Selections {
//             return collectRequiredCollectionNames(sel)
//         }
//     case *ast.Field:
//         node.GetSelectionSet()
//     }
//     return mapset.NewSet[CollectionSpec](), nil
// }

type InvalidSchemaError struct {
	reason   string
	location *ast.Location
}

func NewInvalidSchemaError(reason string, location *ast.Location) *InvalidSchemaError {
	return &InvalidSchemaError{reason: reason, location: location}
}

func (e *InvalidSchemaError) Error() string {
	return fmt.Sprintf("invalid schema: %s at %i", e.reason, e.location.Start)
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
	namespace, err := getNamespace(def.Directives)
	if err != nil {
		return nil, err
	}
	return &dao.CollectionSpec{Namespace: namespace, Name: matches[1]}, nil
}

type HasDirectives interface {
	GetDirectives() []*ast.Directive
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

func (h *GraphqlHandler) Route() string {
	return "/graphql"
}
