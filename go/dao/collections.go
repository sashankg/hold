package dao

import (
	"context"
	"strings"

	sq "github.com/Masterminds/squirrel"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"
)

type CollectionDao interface {
	FindCollections(ctx context.Context, names []string) (map[int]*Collection, error)
	FindCollectionFieldsBySpec(
		ctx context.Context,
		fieldMap map[CollectionSpec]mapset.Set[string],
	) (map[CollectionSpec]*Collection, error)
	FindCollectionFieldsByCollectionId(
		ctx context.Context,
		fieldMap map[int]mapset.Set[string],
	) (map[int]*Collection, error)
	AddCollection(ctx context.Context, collection *Collection) error
}

var _ CollectionDao = (*daoImpl)(nil)

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

type CollectionSpec struct {
	Name      string
	Namespace string
}

// FindCollection implements CollectionDao.
func (o *daoImpl) FindCollections(
	ctx context.Context,
	names []string,
) (map[int]*Collection, error) {
	collectionQuery := sq.Select("id", "name", "domain").
		From("collections")
	if len(names) > 0 {
		collectionQuery = collectionQuery.Where(sq.Eq{"name": names})
	}
	collectionRows, err := collectionQuery.
		RunWith(o.schemaDb).
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
		From("collection_fields").
		Where(sq.Eq{"collection_id": maps.Keys(collections)}).
		RunWith(o.schemaDb).
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

// FindCollectionFieldsBySpec implements CollectionDao.
func (o *daoImpl) FindCollectionFieldsBySpec(
	ctx context.Context,
	fieldMap map[CollectionSpec]mapset.Set[string],
) (map[CollectionSpec]*Collection, error) {
	collections := map[CollectionSpec]*Collection{}
	matchesCollectionSpecsClause := sq.Expr(``)
	for collectionSpec := range fieldMap {
		matchesCollectionSpecsClause = sq.Or{
			matchesCollectionSpecsClause,
			sq.Eq{
				"collections.name":       collectionSpec.Name,
				"collections.domain":     collectionSpec.Namespace,
				"collection_fields.name": fieldMap[collectionSpec].ToSlice(),
			},
		}
	}
	fieldRows, err := sq.Select("collections.id", "collections.name", "collections.domain", "collection_fields.name", "type", "ref", "is_list").
		From("collection_fields").
		Join("collections ON collections.id = collection_fields.collection_id").
		Where(matchesCollectionSpecsClause).
		RunWith(o.schemaDb).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer fieldRows.Close()
	for fieldRows.Next() {
		collection := &Collection{}
		var ref *int
		var isList *bool
		var field CollectionField
		if err := fieldRows.Scan(
			&collection.Id,
			&collection.Name,
			&collection.Domain,
			&field.Name,
			&field.Type,
			&ref,
			&isList,
		); err != nil {
			return nil, err
		}
		if ref != nil {
			field.Ref = *ref
		}
		if isList != nil {
			field.IsList = *isList
		}
		spec := CollectionSpec{Name: collection.Name, Namespace: collection.Domain}
		if _, ok := collections[spec]; !ok {
			collections[spec] = collection
		}
		collections[spec].Fields[field.Name] = field
	}
	return collections, nil
}

// FindCollectionFieldsByCollectionId implements CollectionDao.
func (o *daoImpl) FindCollectionFieldsByCollectionId(
	ctx context.Context,
	fieldMap map[int]mapset.Set[string],
) (map[int]*Collection, error) {
	collections := map[int]*Collection{}
	matchesCollectionIdsClause := sq.Expr(``)
	for id := range fieldMap {
		matchesCollectionIdsClause = sq.Or{
			matchesCollectionIdsClause,
			sq.Eq{
				"collection_id": id,
				"name":          fieldMap[id].ToSlice(),
			},
		}
	}
	fieldRows, err := sq.Select("collections.id", "collections.name", "collections.domain", "name", "type", "ref", "is_list").
		From("collection_fields").
		Join("collections ON collections.id = collection_fields.collection_id").
		Where(matchesCollectionIdsClause).
		RunWith(o.schemaDb).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer fieldRows.Close()
	for fieldRows.Next() {
		collection := &Collection{}
		var field CollectionField
		if err := fieldRows.Scan(&collection.Id, &collection.Name, &collection.Domain, &field.Name, &field.Type, &field.Ref, &field.IsList); err != nil {
			return nil, err
		}
		if _, ok := collections[collection.Id]; !ok {
			collections[collection.Id] = collection
		}
		collections[collection.Id].Fields[field.Name] = field
	}
	return collections, nil
}

// AddCollection implements CollectionDao.
func (o *daoImpl) AddCollection(ctx context.Context, collection *Collection) error {
	tx, err := o.schemaDb.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := sq.Insert("collections").
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
		_, err := sq.Insert("collection_fields").
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
