package dao

import (
	"context"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

type CollectionDao interface {
	FindCollectionBySpec(ctx context.Context, spec CollectionSpec) (*Collection, error)
	FindCollectionById(ctx context.Context, id int) (*Collection, error)
	GetCollectionId(ctx context.Context, spec CollectionSpec) (int, error)
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

func (o *daoImpl) FindCollectionBySpec(
	ctx context.Context,
	spec CollectionSpec,
) (*Collection, error) {
	collectionQuery := sq.Select("id", "name", "domain").
		From("collections").
		Where(sq.Eq{"name": spec.Name, "domain": spec.Namespace}).
		RunWith(o.schemaDb).
		QueryRowContext(ctx)
	collection := &Collection{}
	if err := collectionQuery.Scan(&collection.Id, &collection.Name, &collection.Domain); err != nil {
		return nil, err
	}
	if err := o.populateFields(ctx, collection); err != nil {
		return nil, err
	}
	return collection, nil
}

func (o *daoImpl) FindCollectionById(
	ctx context.Context,
	id int,
) (*Collection, error) {
	collectionQuery := sq.Select("id", "name", "domain").
		From("collections").
		Where(sq.Eq{"id": id}).
		RunWith(o.schemaDb).
		QueryRowContext(ctx)
	collection := &Collection{}
	if err := collectionQuery.Scan(&collection.Id, &collection.Name, &collection.Domain); err != nil {
		return nil, err
	}
	if err := o.populateFields(ctx, collection); err != nil {
		return nil, err
	}
	return collection, nil
}

func (o *daoImpl) populateFields(
	ctx context.Context,
	collection *Collection,
) error {
	var ref *int
	var isList *bool
	fieldRows, err := sq.Select("name", "type", "ref", "is_list").
		From("collection_fields").
		Where(sq.Eq{"collection_id": collection.Id}).
		RunWith(o.schemaDb).
		QueryContext(ctx)
	if err != nil {
		return err
	}
	defer fieldRows.Close()
	fields := map[string]CollectionField{}
	for fieldRows.Next() {
		field := CollectionField{}
		if err := fieldRows.Scan(&field.Name, &field.Type, &ref, &isList); err != nil {
			return err
		}
		if ref != nil {
			field.Ref = *ref
		}
		if isList != nil {
			field.IsList = *isList
		}
		fields[field.Name] = field
	}
	collection.Fields = fields
	return nil
}

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

// GetCollectionId implements CollectionDao.
func (o *daoImpl) GetCollectionId(ctx context.Context, spec CollectionSpec) (int, error) {
	collectionQuery := sq.Select("id").
		From("collections").
		Where(sq.Eq{"name": spec.Name, "domain": spec.Namespace}).
		RunWith(o.schemaDb).
		QueryRowContext(ctx)
	var id int
	if err := collectionQuery.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
