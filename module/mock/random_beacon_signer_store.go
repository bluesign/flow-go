// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	module "github.com/onflow/flow-go/module"
	mock "github.com/stretchr/testify/mock"
)

// RandomBeaconKeyStore is an autogenerated mock type for the RandomBeaconKeyStore type
type RandomBeaconKeyStore struct {
	mock.Mock
}

// GetSigner provides a mock function with given fields: view
func (_m *RandomBeaconKeyStore) GetSigner(view uint64) (module.MsgSigner, error) {
	ret := _m.Called(view)

	var r0 module.MsgSigner
	if rf, ok := ret.Get(0).(func(uint64) module.MsgSigner); ok {
		r0 = rf(view)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(module.MsgSigner)
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
