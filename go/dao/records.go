package dao

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type Selection struct {
	FieldName     string
	Subselections []Selection
}

type Record struct {
	collection *Collection
	fields     map[string]interface{}
}

func (o *daoImpl) GetRecord(
	ctx context.Context,
	id int,
	selection []Selection,
	collectionId int,
	collectionMap map[int]*Collection,
) (string, error) {
	var buildRecordQuery func(sq.Sqlizer, []Selection, int) sq.SelectBuilder
	buildRecordQuery = func(id sq.Sqlizer, selection []Selection, collectionId int) sq.SelectBuilder {
		collection := collectionMap[collectionId]
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
					buildRecordQuery(
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

	recordQuery := buildRecordQuery(sq.Expr(`?`, id), selection, collectionId)
	fmt.Println(recordQuery.ToSql())

	var json string
	err := recordQuery.RunWith(o.db).QueryRowContext(ctx).Scan(&json)
	return json, err
}

// func (o *daoImpl) AddRecord(
//     ctx context.Context,
//     collection *Collection,

// ) (int, error) {
//     o.db.ExecContext()

// }
