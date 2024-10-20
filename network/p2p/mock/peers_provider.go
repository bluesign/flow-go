// Code generated by mockery v2.13.1. DO NOT EDIT.

package mockp2p

import (
	mock "github.com/stretchr/testify/mock"

	peer "github.com/libp2p/go-libp2p/core/peer"
)

// PeersProvider is an autogenerated mock type for the PeersProvider type
type PeersProvider struct {
	mock.Mock
}

// Execute provides a mock function with given fields:
func (_m *PeersProvider) Execute() peer.IDSlice {
	ret := _m.Called()

	var r0 peer.IDSlice
	if rf, ok := ret.Get(0).(func() peer.IDSlice); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(peer.IDSlice)
		}
	}

	return r0
}

type mockConstructorTestingTNewPeersProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewPeersProvider creates a new instance of PeersProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPeersProvider(t mockConstructorTestingTNewPeersProvider) *PeersProvider {
	mock := &PeersProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
