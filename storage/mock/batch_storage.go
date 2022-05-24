// Code generated by mockery v2.12.1. DO NOT EDIT.

package mock

import (
	badger "github.com/dgraph-io/badger/v2"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// BatchStorage is an autogenerated mock type for the BatchStorage type
type BatchStorage struct {
	mock.Mock
}

// Flush provides a mock function with given fields:
func (_m *BatchStorage) Flush() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetWriter provides a mock function with given fields:
func (_m *BatchStorage) GetWriter() *badger.WriteBatch {
	ret := _m.Called()

	var r0 *badger.WriteBatch
	if rf, ok := ret.Get(0).(func() *badger.WriteBatch); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*badger.WriteBatch)
		}
	}

	return r0
}

// OnSucceed provides a mock function with given fields: callback
func (_m *BatchStorage) OnSucceed(callback func()) {
	_m.Called(callback)
}

// NewBatchStorage creates a new instance of BatchStorage. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewBatchStorage(t testing.TB) *BatchStorage {
	mock := &BatchStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
