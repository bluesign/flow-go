// Code generated by mockery v2.43.2. DO NOT EDIT.

package mock

import (
	flow "github.com/onflow/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"
)

// StateMachineTelemetryConsumer is an autogenerated mock type for the StateMachineTelemetryConsumer type
type StateMachineTelemetryConsumer struct {
	mock.Mock
}

// OnInvalidServiceEvent provides a mock function with given fields: event, err
func (_m *StateMachineTelemetryConsumer) OnInvalidServiceEvent(event flow.ServiceEvent, err error) {
	_m.Called(event, err)
}

// OnServiceEventProcessed provides a mock function with given fields: event
func (_m *StateMachineTelemetryConsumer) OnServiceEventProcessed(event flow.ServiceEvent) {
	_m.Called(event)
}

// OnServiceEventReceived provides a mock function with given fields: event
func (_m *StateMachineTelemetryConsumer) OnServiceEventReceived(event flow.ServiceEvent) {
	_m.Called(event)
}

// NewStateMachineTelemetryConsumer creates a new instance of StateMachineTelemetryConsumer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStateMachineTelemetryConsumer(t interface {
	mock.TestingT
	Cleanup(func())
}) *StateMachineTelemetryConsumer {
	mock := &StateMachineTelemetryConsumer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}