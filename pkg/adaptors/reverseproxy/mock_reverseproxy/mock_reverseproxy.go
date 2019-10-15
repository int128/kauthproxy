// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/int128/kauthproxy/pkg/adaptors/reverseproxy (interfaces: Interface)

// Package mock_reverseproxy is a generated GoMock package.
package mock_reverseproxy

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	reverseproxy "github.com/int128/kauthproxy/pkg/adaptors/reverseproxy"
	reflect "reflect"
)

// MockInterface is a mock of Interface interface
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// Run mocks base method
func (m *MockInterface) Run(arg0 context.Context, arg1 reverseproxy.Option) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockInterfaceMockRecorder) Run(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockInterface)(nil).Run), arg0, arg1)
}
