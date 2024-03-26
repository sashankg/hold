package resolvers

import (
	"database/sql"

	"github.com/graphql-go/graphql"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/parser"
)

type RegisterResolver struct {
	schema *graphql.Schema
	db     *sql.DB
}

func NewRegisterResolver(schema *graphql.Schema, db *sql.DB) *RegisterResolver {
	db.Exec(`CREATE TABLE IF NOT EXISTS graph (
		id INTEGER PRIMARY KEY,
		type TEXT,
		value JSONB
	)`)
	return &RegisterResolver{
		schema: schema,
		db:     db,
	}
}

// Key implements core.MutationResolver.
func (*RegisterResolver) Key() string {
	return "register"
}

// MutationArgs implements core.MutationResolver.
func (*RegisterResolver) MutationArgs() graphql.FieldConfigArgument {
	return graphql.FieldConfigArgument{
		"schema": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	}
}

// MutationType implements core.MutationResolver.
func (*RegisterResolver) MutationType() graphql.Type {
	return graphql.String
}

// ResolveMutation implements core.MutationResolver.
func (r *RegisterResolver) ResolveMutation(p graphql.ResolveParams) (interface{}, error) {
	schemaStr := p.Args["schema"]
	_, err := parser.Parse(parser.ParseParams{
		Source: schemaStr,
	})
	if err != nil {
		return nil, err
	}

	// for _, def := range doc.Definitions {
	//     switch def.GetKind() {
	//     case kinds.ObjectDefinition:
	//         def := def.(*ast.ObjectDefinition)
	//         if err != nil {
	//             return nil, err
	//         }
	//         for _, field := range def.Fields {
	//             if err != nil {
	//                 return nil, err
	//             }
	//         }
	//     }
	// }

	return "ok", nil
}

func gqlTypeToSqlType(t ast.Type) string {
	if t.GetKind() == kinds.Named {
		switch t.(*ast.Named).Name.Value {
		case "String":
			return "TEXT"
		case "Int":
			return "INTEGER"
		case "Float":
			return "REAL"
		case "Boolean":
			return "INTEGER"
		}
	}
	return "JSONB"
}

// var _ core.MutationResolver = (*RegisterResolver)(nil)
