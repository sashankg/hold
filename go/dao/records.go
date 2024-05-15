package dao

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

type RecordDao interface {
	GetRecord(
		ctx context.Context,
		id int,
		selection []Selection,
		collectionId int,
	) ([]byte, error)
}

type Selection struct {
	FieldName     string
	Subselections []Selection
}

type Record struct {
	collection *Collection
	fields     map[string]interface{}
}

var _ RecordDao = (*daoImpl)(nil)

func (o *daoImpl) GetRecord(
	ctx context.Context,
	id int,
	selection []Selection,
	collectionId int,
) ([]byte, error) {
	recordQuery := o.buildRecordQuery(ctx, sq.Expr(`?`, id), selection, collectionId)
	var json []byte
	err := recordQuery.RunWith(o.recordDb).QueryRowContext(ctx).Scan(&json)
	return json, err
}

func (o *daoImpl) buildRecordQuery(
	ctx context.Context,
	id sq.Sqlizer,
	selection []Selection,
	collectionId int,
) sq.SelectBuilder {
	collection, err := o.FindCollectionById(ctx, collectionId)
	if err != nil {
		// should have already validated
		panic(err)
	}
	builder := sq.Select().From(collection.Name).Where(sq.Expr(`id = ?`, id))
	objectArgs := sq.Expr(``)
	for i, s := range selection {
		if i > 0 {
			objectArgs = sq.ConcatExpr(objectArgs, `, `)
		}
		objectArgs = sq.ConcatExpr(objectArgs, sq.Expr(`?, `, s.FieldName))
		if len(s.Subselections) > 0 {

			objectArgs = sq.ConcatExpr(
				objectArgs,
				`(`,
				o.buildRecordQuery(
					ctx,
					sq.Expr(s.FieldName),
					s.Subselections,
					collection.Fields[s.FieldName].Ref,
				),
				`)`,
			)
		} else {
			objectArgs = sq.ConcatExpr(objectArgs, s.FieldName)
		}
	}
	jsonObject := sq.ConcatExpr(`json_object(`, objectArgs, `)`)
	return builder.Column(jsonObject)
}
