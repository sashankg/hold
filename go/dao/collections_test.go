package dao_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/test_util"
	"github.com/stretchr/testify/assert"
)

func TestFindCollectionFields(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(test_util.SQLQueryMatcher))
	assert.NoError(t, err)
	o := dao.NewDao(db)

	mock.ExpectQuery(`
		SELECT 
			_collections.name, 
			_collections.domain,
			name,
			type,
			ref,
			is_list
		FROM _collection_fields JOIN _collections 
		ON _collections.id = _collection_fields.collection_id
		WHERE 
			((_collection_fields.name IN (?,?) 
			AND _collections.domain = ? 
		    AND _collections.name = ?)
			OR _collection_fields.name IN (?,?) 
			AND _collections.domain = ? 
		    AND _collections.name = ?)
	`).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"name", "domain", "name", "type", "ref", "is_list"}))
	_, err = o.FindCollectionFieldsBySpec(context.Background(), map[dao.CollectionSpec]mapset.Set[string]{
		{Namespace: "namespace", Name: "name1"}: mapset.NewSet("field1", "field2"),
		{Namespace: "namespace", Name: "name2"}: mapset.NewSet("field3", "field4"),
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
