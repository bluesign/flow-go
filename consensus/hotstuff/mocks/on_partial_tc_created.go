// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// OnPartialTCCreated is an autogenerated mock type for the OnPartialTCCreated type
type OnPartialTCCreated struct {
	mock.Mock
}

// Execute provides a mock function with given fields: view
func (_m *OnPartialTCCreated) Execute(view uint64) {
	_m.Called(view)
}

type NewOnPartialTCCreatedT interface {
	mock.TestingT
	Cleanup(func())
}

// NewOnPartialTCCreated creates a new instance of OnPartialTCCreated. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOnPartialTCCreated(t NewOnPartialTCCreatedT) *OnPartialTCCreated {
	mock := &OnPartialTCCreated{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}