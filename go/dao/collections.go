package dao

import (
	"context"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"golang.org/x/exp/maps"
)

type Collection struct {
	Id      int
	Name    string
	Domain  string
	Version string
	Fields  map[string]CollectionField
	Table   string
}

type CollectionField struct {
	Name   string
	Type   string
	Ref    int
	IsList bool
}

// FindCollection implements Dao.
func (o *daoImpl) FindCollections(
	ctx context.Context,
	names []string,
) (map[int]*Collection, error) {
	collectionQuery := sq.Select("id", "name", "domain").
		From("_collections")
	if len(names) > 0 {
		collectionQuery = collectionQuery.Where(sq.Eq{"name": names})
	}
	collectionRows, err := collectionQuery.
		RunWith(o.db).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer collectionRows.Close()

	collections := map[int]*Collection{}
	for collectionRows.Next() {
		collection := &Collection{}
		if err := collectionRows.Scan(&collection.Id, &collection.Name, &collection.Domain); err != nil {
			return nil, err
		}
		collection.Fields = map[string]CollectionField{}
		collections[collection.Id] = collection
	}

	fieldRows, err := sq.Select("collection_id", "name", "type", "ref", "is_list").
		From("_collection_fields").
		Where(sq.Eq{"collection_id": maps.Keys(collections)}).
		RunWith(o.db).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		field := CollectionField{}
		var collectionId int
		if err := fieldRows.Scan(&collectionId, &field.Name, &field.Type, &field.Ref, &field.IsList); err != nil {
			return nil, err
		}
		collections[collectionId].Fields[field.Name] = field
	}

	return collections, nil
}

// AddCollection implements Dao.
func (o *daoImpl) AddCollection(ctx context.Context, collection *Collection) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := sq.Insert("_collections").
		Columns(
			"name",
			"domain",
			"version",
		).
		Values(
			collection.Name,
			collection.Domain,
			collection.Version,
		).
		RunWith(tx).ExecContext(ctx)
	if err != nil {
		return err
	}
	collectionId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	sqlCols := []string{}

	for _, field := range collection.Fields {
		_, err := sq.Insert("_collection_fields").
			Columns(
				"collection_id",
				"name",
				"type",
				"ref",
				"is_list",
			).
			Values(
				collectionId,
				field.Name,
				field.Type,
				field.Ref,
				field.IsList,
			).
			RunWith(tx).ExecContext(ctx)
		if err != nil {
			return err
		}
		sqlCols = append(sqlCols, field.Name+" "+field.Type)
	}

	createTable, _, err := sq.ConcatExpr(`CREATE TABLE `, collection.Name, ` (
		id INTEGER PRIMARY KEY,`,
		strings.Join(sqlCols, ", "),
		`)`,
	).ToSql()

	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, createTable)
	if err != nil {
		return err
	}

	return tx.Commit()
}
