package resolvers

import (
	"database/sql"

	"github.com/graphql-go/graphql"
)

const (
	cKvKey   = "key"
	cKvValue = "value"
)

var (
	cKeyField = graphql.ArgumentConfig{
		Type: graphql.NewNonNull(graphql.String),
	}
	cValueField = graphql.ArgumentConfig{
		Type: graphql.String,
	}
)

type KvResolver struct {
	db *sql.DB
}

func NewKvResolver(db *sql.DB) *KvResolver {
	db.Exec(`CREATE TABLE IF NOT EXISTS kv (
		key TEXT PRIMARY KEY,
		value JSONB
	)`)
	return &KvResolver{
		db,
	}
}

func (*KvResolver) QueryArgs() graphql.FieldConfigArgument {
	return graphql.FieldConfigArgument{
		cKvKey: &cKeyField,
	}
}

func (r *KvResolver) ResolveQuery(p graphql.ResolveParams) (interface{}, error) {
	key := p.Args[cKvKey]
	row := r.db.QueryRow("SELECT json(value) FROM kv WHERE key = ?", key)
	var value string
	if err := row.Scan(&value); err != nil {
		return nil, err
	}
	return value, nil
}
func (*KvResolver) QueryType() graphql.Type {
	return graphql.String
}

func (*KvResolver) MutationArgs() graphql.FieldConfigArgument {
	return nil
}

func (*KvResolver) ResolveMutation(p graphql.ResolveParams) (interface{}, error) {
	return 0, nil
}

func (r *KvResolver) resolvePut(p graphql.ResolveParams) (interface{}, error) {
	row := r.db.QueryRow(
		"REPLACE INTO kv (key, value) VALUES (?, jsonb(?)) RETURNING json(value)",
		p.Args[cKvKey],
		p.Args[cKvValue],
	)
	var value string
	if err := row.Scan(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func (r *KvResolver) resolvePatch(p graphql.ResolveParams) (interface{}, error) {
	row := r.db.QueryRow(
		"UPDATE kv SET value = jsonb_patch(value, jsonb(?)) WHERE key = ? RETURNING json(value)",
		p.Args[cKvValue],
		p.Args[cKvKey],
	)
	var value string
	if err := row.Scan(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func (r *KvResolver) MutationType() graphql.Type {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "KvMutation",
		Fields: graphql.Fields{
			"put": &graphql.Field{
				Type:    graphql.String,
				Resolve: r.resolvePut,
				Args: graphql.FieldConfigArgument{
					cKvKey:   &cKeyField,
					cKvValue: &cValueField,
				},
			},
			"patch": &graphql.Field{
				Type:    graphql.String,
				Resolve: r.resolvePatch,
				Args: graphql.FieldConfigArgument{
					cKvKey:   &cKeyField,
					cKvValue: &cValueField,
				},
			},
		},
	})
}

func (*KvResolver) Key() string {
	return "kv"
}

// var _ core.QueryResolver = (*KvResolver)(nil)
// var _ core.MutationResolver = (*KvResolver)(nil)
