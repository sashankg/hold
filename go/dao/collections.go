package dao

import (
	"context"

	"golang.org/x/exp/maps"
)

type Collection struct {
	Id      int
	Name    string
	Domain  string
	Version string
	Fields  []CollectionField
}

type CollectionField struct {
	Name   string
	Type   string
	Ref    int
	IsList bool
}

// FindCollection implements Dao.
func (o *daoImpl) FindCollection(
	ctx context.Context,
	name string,
	domain string,
) (*Collection, error) {
	row := o.db.QueryRowContext(ctx, `
		SELECT 
			id,
			name,
			domain
		FROM 
			_collections
		WHERE 
			name = ? AND domain = ?
		`,
		name,
		domain,
	)

	collection := &Collection{}

	if err := row.Scan(&collection.Id, &collection.Name, &collection.Domain); err != nil {
		return nil, err
	}

	rows, err := o.db.QueryContext(ctx,
		`SELECT id, type, ref, is_list FROM _collection_fields WHERE collection_id = ?`,
		collection.Id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		field := CollectionField{}
		if err := rows.Scan(&field.Name, &field.Type, &field.Ref, &field.IsList); err != nil {
			return nil, err
		}
		collection.Fields = append(collection.Fields, field)
	}

	return collection, nil
}

// ListCollections implements Dao.
func (o *daoImpl) ListCollections(ctx context.Context) ([]*Collection, error) {
	collections := map[int]*Collection{}
	collectionRows, err := o.db.QueryContext(ctx, `SELECT id, name, domain FROM _collections`)
	if err != nil {
		return nil, err
	}
	defer collectionRows.Close()
	for collectionRows.Next() {
		collection := &Collection{}
		if err := collectionRows.Scan(&collection.Id, &collection.Name, &collection.Domain); err != nil {
			return nil, err
		}
		collections[collection.Id] = collection
	}

	fieldRows, err := o.db.QueryContext(
		ctx,
		`SELECT collection_id, name, type, ref, is_list FROM _collection_fields`,
	)
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
		if _, ok := collections[collectionId]; ok {
			collections[collectionId].Fields = append(collections[collectionId].Fields, field)
		}
	}
	return maps.Values(collections), nil
}

// AddCollection implements Dao.
func (o *daoImpl) AddCollection(ctx context.Context, collection *Collection) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO _collections (name, domain, version) VALUES (?, ?, ?)`,
		collection.Name,
		collection.Domain,
		collection.Version,
	)
	if err != nil {
		return err
	}
	collectionId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for _, field := range collection.Fields {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO _collection_fields (collection_id, name, type, ref, is_list) VALUES (?, ?, ?, ?, ?)`,
			collectionId,
			field.Name,
			field.Type,
			field.Ref,
			field.IsList,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
