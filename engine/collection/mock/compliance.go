// Code generated by mockery v2.21.4. DO NOT EDIT.

package mock

import (
	flow "github.com/onflow/flow-go/model/flow"
	irrecoverable "github.com/onflow/flow-go/module/irrecoverable"

	messages "github.com/onflow/flow-go/model/messages"

	mock "github.com/stretchr/testify/mock"
)

// Compliance is an autogenerated mock type for the Compliance type
type Compliance struct {
	mock.Mock
}

// Done provides a mock function with given fields:
func (_m *Compliance) Done() <-chan struct{} {
	ret := _m.Called()

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// OnClusterBlockProposal provides a mock function with given fields: proposal
func (_m *Compliance) OnClusterBlockProposal(proposal flow.Slashable[*messages.ClusterBlockProposal]) {
	_m.Called(proposal)
}

// OnSyncedClusterBlock provides a mock function with given fields: block
func (_m *Compliance) OnSyncedClusterBlock(block flow.Slashable[*messages.ClusterBlockProposal]) {
	_m.Called(block)
}

// Ready provides a mock function with given fields:
func (_m *Compliance) Ready() <-chan struct{} {
	ret := _m.Called()

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// Start provides a mock function with given fields: _a0
func (_m *Compliance) Start(_a0 irrecoverable.SignalerContext) {
	_m.Called(_a0)
}

type mockConstructorTestingTNewCompliance interface {
	mock.TestingT
	Cleanup(func())
}

// NewCompliance creates a new instance of Compliance. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCompliance(t mockConstructorTestingTNewCompliance) *Compliance {
	mock := &Compliance{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}