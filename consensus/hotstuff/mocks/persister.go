// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	hotstuff "github.com/onflow/flow-go/consensus/hotstuff"
	mock "github.com/stretchr/testify/mock"
)

// Persister is an autogenerated mock type for the Persister type
type Persister struct {
	mock.Mock
}

// GetLivenessData provides a mock function with given fields:
func (_m *Persister) GetLivenessData() (*hotstuff.LivenessData, error) {
	ret := _m.Called()

	var r0 *hotstuff.LivenessData
	if rf, ok := ret.Get(0).(func() *hotstuff.LivenessData); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*hotstuff.LivenessData)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSafetyData provides a mock function with given fields:
func (_m *Persister) GetSafetyData() (*hotstuff.SafetyData, error) {
	ret := _m.Called()

	var r0 *hotstuff.SafetyData
	if rf, ok := ret.Get(0).(func() *hotstuff.SafetyData); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*hotstuff.SafetyData)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutLivenessData provides a mock function with given fields: livenessData
func (_m *Persister) PutLivenessData(livenessData *hotstuff.LivenessData) error {
	ret := _m.Called(livenessData)

	var r0 error
	if rf, ok := ret.Get(0).(func(*hotstuff.LivenessData) error); ok {
		r0 = rf(livenessData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PutSafetyData provides a mock function with given fields: safetyData
func (_m *Persister) PutSafetyData(safetyData *hotstuff.SafetyData) error {
	ret := _m.Called(safetyData)

	var r0 error
	if rf, ok := ret.Get(0).(func(*hotstuff.SafetyData) error); ok {
		r0 = rf(safetyData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewPersister interface {
	mock.TestingT
	Cleanup(func())
}

// NewPersister creates a new instance of Persister. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPersister(t mockConstructorTestingTNewPersister) *Persister {
	mock := &Persister{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
