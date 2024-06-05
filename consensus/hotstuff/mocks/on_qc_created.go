// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// OnQCCreated is an autogenerated mock type for the OnQCCreated type
type OnQCCreated struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *OnQCCreated) Execute(_a0 *flow.QuorumCertificate) {
	_m.Called(_a0)
}

// NewOnQCCreated creates a new instance of OnQCCreated. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOnQCCreated(t interface {
	mock.TestingT
	Cleanup(func())
}) *OnQCCreated {
	mock := &OnQCCreated{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
