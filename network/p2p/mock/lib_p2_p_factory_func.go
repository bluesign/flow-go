// Code generated by mockery v2.21.4. DO NOT EDIT.

package mockp2p

import (
	p2p "github.com/onflow/flow-go/network/p2p"
	mock "github.com/stretchr/testify/mock"
)

// LibP2PFactoryFunc is an autogenerated mock type for the LibP2PFactoryFunc type
type LibP2PFactoryFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields:
func (_m *LibP2PFactoryFunc) Execute() (p2p.LibP2PNode, error) {
	ret := _m.Called()

	var r0 p2p.LibP2PNode
	var r1 error
	if rf, ok := ret.Get(0).(func() (p2p.LibP2PNode, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() p2p.LibP2PNode); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(p2p.LibP2PNode)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewLibP2PFactoryFunc interface {
	mock.TestingT
	Cleanup(func())
}

// NewLibP2PFactoryFunc creates a new instance of LibP2PFactoryFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLibP2PFactoryFunc(t mockConstructorTestingTNewLibP2PFactoryFunc) *LibP2PFactoryFunc {
	mock := &LibP2PFactoryFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}