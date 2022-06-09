// Code generated by mockery v2.12.1. DO NOT EDIT.

package mock

import (
	flow "github.com/onflow/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// Backend is an autogenerated mock type for the Backend type
type Backend struct {
	mock.Mock
}

// ProcessTransaction provides a mock function with given fields: _a0
func (_m *Backend) ProcessTransaction(_a0 *flow.TransactionBody) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.TransactionBody) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewBackend creates a new instance of Backend. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewBackend(t testing.TB) *Backend {
	mock := &Backend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
