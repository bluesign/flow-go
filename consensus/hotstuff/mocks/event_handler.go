// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	context "context"
	"time"

	hotstuff "github.com/onflow/flow-go/consensus/hotstuff"
	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"

	model "github.com/onflow/flow-go/consensus/hotstuff/model"
)

// EventHandler is an autogenerated mock type for the EventHandler type
type EventHandler struct {
	mock.Mock
}

// OnLocalTimeout provides a mock function with given fields: info
func (_m *EventHandler) OnLocalTimeout() error {
	ret := _m.Called(info)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.TimerInfo) error); ok {
		r0 = rf(info)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnPartialTcCreated provides a mock function with given fields: partialTC
func (_m *EventHandler) OnPartialTcCreated(partialTC *hotstuff.PartialTcCreated) error {
	ret := _m.Called(partialTC)

	var r0 error
	if rf, ok := ret.Get(0).(func(*hotstuff.PartialTcCreated) error); ok {
		r0 = rf(partialTC)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnReceiveProposal provides a mock function with given fields: proposal
func (_m *EventHandler) OnReceiveProposal(proposal *model.Proposal) error {
	ret := _m.Called(proposal)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Proposal) error); ok {
		r0 = rf(proposal)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnReceiveQc provides a mock function with given fields: qc
func (_m *EventHandler) OnReceiveQc(qc *flow.QuorumCertificate) error {
	ret := _m.Called(qc)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.QuorumCertificate) error); ok {
		r0 = rf(qc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnReceiveTc provides a mock function with given fields: tc
func (_m *EventHandler) OnReceiveTc(tc *flow.TimeoutCertificate) error {
	ret := _m.Called(tc)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.TimeoutCertificate) error); ok {
		r0 = rf(tc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields: ctx
func (_m *EventHandler) Start(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TimeoutChannel provides a mock function with given fields:
func (_m *EventHandler) TimeoutChannel() <-chan time.Time {
	ret := _m.Called()

	var r0 <-chan model.TimerInfo
	if rf, ok := ret.Get(0).(func() <-chan model.TimerInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan model.TimerInfo)
		}
	}

	return r0
}

type mockConstructorTestingTNewEventHandler interface {
	mock.TestingT
	Cleanup(func())
}

// NewEventHandler creates a new instance of EventHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEventHandler(t mockConstructorTestingTNewEventHandler) *EventHandler {
	mock := &EventHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
