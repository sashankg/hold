package handlers_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/handlers"
	"github.com/sashankg/hold/testing/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGraphqlHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockValidator := mocks.NewMockValidator(ctrl)
	mockResolver := mocks.NewMockResolver(ctrl)

	mockValidator.EXPECT().ValidateRootSelections(gomock.Any(), gomock.Any()).Return(nil)

	mockResolver.EXPECT().
		Resolve(gomock.Any(), gomock.Any()).
		Return(graphql.JsonValue(`"hello world"`), nil)

	req := httptest.NewRequest("GET", "/", strings.NewReader(`
		query {
			findPost {
				title
				body
				author {
					name
					friends {
						name
					}
				}
			}
		}
	`))
	req.Header.Set("Content-Type", "application/graphql")
	resp := httptest.NewRecorder()
	handlers.NewGraphqlHandler(mockValidator, mockResolver).ServeHTTP(resp, req)

	body, err := io.ReadAll(resp.Result().Body)
	require.NoError(t, err)
	require.Equal(t, `{"data":"hello world"}`, string(body))
	require.Equal(t, 200, resp.Code)
}
