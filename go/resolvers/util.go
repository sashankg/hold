package resolvers

import "github.com/graphql-go/graphql"

type JsonCompletedObject struct {
	json string
}

// Serialize implements graphql.ResolvedObject.
func (o *JsonCompletedObject) CompletedObjectResult() interface{} {
	return o
}

func (o *JsonCompletedObject) MarshalJSON() ([]byte, error) {
	return []byte(o.json), nil
}

var _ graphql.CompletedObject = (*JsonCompletedObject)(nil)
