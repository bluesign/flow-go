// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	hotstuff "github.com/onflow/flow-go/consensus/hotstuff"
	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// DynamicCommittee is an autogenerated mock type for the DynamicCommittee type
type DynamicCommittee struct {
	mock.Mock
}

// DKG provides a mock function with given fields: view
func (_m *DynamicCommittee) DKG(view uint64) (hotstuff.DKG, error) {
	ret := _m.Called(view)

	var r0 hotstuff.DKG
	if rf, ok := ret.Get(0).(func(uint64) hotstuff.DKG); ok {
		r0 = rf(view)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(hotstuff.DKG)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(view)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IdentitiesByBlock provides a mock function with given fields: blockID
func (_m *DynamicCommittee) IdentitiesByBlock(blockID flow.Identifier) (flow.IdentityList, error) {
	ret := _m.Called(blockID)

	var r0 flow.IdentityList
	if rf, ok := ret.Get(0).(func(flow.Identifier) flow.IdentityList); ok {
		r0 = rf(blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.IdentityList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier) error); ok {
		r1 = rf(blockID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IdentitiesByEpoch provides a mock function with given fields: view
func (_m *DynamicCommittee) IdentitiesByEpoch(view uint64) (flow.IdentityList, error) {
	ret := _m.Called(view)

	var r0 flow.IdentityList
	if rf, ok := ret.Get(0).(func(uint64) flow.IdentityList); ok {
		r0 = rf(view)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.IdentityList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(view)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IdentityByBlock provides a mock function with given fields: blockID, participantID
func (_m *DynamicCommittee) IdentityByBlock(blockID flow.Identifier, participantID flow.Identifier) (*flow.Identity, error) {
	ret := _m.Called(blockID, participantID)

	var r0 *flow.Identity
	if rf, ok := ret.Get(0).(func(flow.Identifier, flow.Identifier) *flow.Identity); ok {
		r0 = rf(blockID, participantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Identity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier, flow.Identifier) error); ok {
		r1 = rf(blockID, participantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IdentityByEpoch provides a mock function with given fields: view, participantID
func (_m *DynamicCommittee) IdentityByEpoch(view uint64, participantID flow.Identifier) (*flow.Identity, error) {
	ret := _m.Called(view, participantID)

	var r0 *flow.Identity
	if rf, ok := ret.Get(0).(func(uint64, flow.Identifier) *flow.Identity); ok {
		r0 = rf(view, participantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Identity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64, flow.Identifier) error); ok {
		r1 = rf(view, participantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LeaderForView provides a mock function with given fields: view
func (_m *DynamicCommittee) LeaderForView(view uint64) (flow.Identifier, error) {
	ret := _m.Called(view)

	var r0 flow.Identifier
	if rf, ok := ret.Get(0).(func(uint64) flow.Identifier); ok {
		r0 = rf(view)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.Identifier)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(view)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QuorumThresholdForView provides a mock function with given fields: view
func (_m *DynamicCommittee) QuorumThresholdForView(view uint64) (uint64, error) {
	ret := _m.Called(view)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(uint64) uint64); ok {
		r0 = rf(view)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(view)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Self provides a mock function with given fields:
func (_m *DynamicCommittee) Self() flow.Identifier {
	ret := _m.Called()

	var r0 flow.Identifier
	if rf, ok := ret.Get(0).(func() flow.Identifier); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.Identifier)
		}
	}

	return r0
}

// TimeoutThresholdForView provides a mock function with given fields: view
func (_m *DynamicCommittee) TimeoutThresholdForView(view uint64) (uint64, error) {
	ret := _m.Called(view)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(uint64) uint64); ok {
		r0 = rf(view)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(view)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewDynamicCommitteeT interface {
	mock.TestingT
	Cleanup(func())
}

// NewDynamicCommittee creates a new instance of DynamicCommittee. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDynamicCommittee(t NewDynamicCommitteeT) *DynamicCommittee {
	mock := &DynamicCommittee{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
