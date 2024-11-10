// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/int128/kauthproxy/internal/env (interfaces: Interface)
//
// Generated by this command:
//
//	mockgen -destination internal/mocks/mock_env/mock.go github.com/int128/kauthproxy/internal/env Interface
//

// Package mock_env is a generated GoMock package.
package mock_env

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
	isgomock struct{}
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// AllocateLocalPort mocks base method.
func (m *MockInterface) AllocateLocalPort() (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllocateLocalPort")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllocateLocalPort indicates an expected call of AllocateLocalPort.
func (mr *MockInterfaceMockRecorder) AllocateLocalPort() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllocateLocalPort", reflect.TypeOf((*MockInterface)(nil).AllocateLocalPort))
}