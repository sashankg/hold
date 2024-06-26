// Code generated by MockGen. DO NOT EDIT.
// Source: graphql/validator.go
//
// Generated by this command:
//
//	mockgen -source=graphql/validator.go -destination=testing/mocks/validator_mock.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	ast "github.com/graphql-go/graphql/language/ast"
	gomock "go.uber.org/mock/gomock"
)

// MockValidator is a mock of Validator interface.
type MockValidator struct {
	ctrl     *gomock.Controller
	recorder *MockValidatorMockRecorder
}

// MockValidatorMockRecorder is the mock recorder for MockValidator.
type MockValidatorMockRecorder struct {
	mock *MockValidator
}

// NewMockValidator creates a new mock instance.
func NewMockValidator(ctrl *gomock.Controller) *MockValidator {
	mock := &MockValidator{ctrl: ctrl}
	mock.recorder = &MockValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockValidator) EXPECT() *MockValidatorMockRecorder {
	return m.recorder
}

// ValidateRootSelections mocks base method.
func (m *MockValidator) ValidateRootSelections(ctx context.Context, doc *ast.Document) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateRootSelections", ctx, doc)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateRootSelections indicates an expected call of ValidateRootSelections.
func (mr *MockValidatorMockRecorder) ValidateRootSelections(ctx, doc any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateRootSelections", reflect.TypeOf((*MockValidator)(nil).ValidateRootSelections), ctx, doc)
}
