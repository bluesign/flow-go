// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	context "context"

	flow "github.com/dapperlabs/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// IngestRPC is an autogenerated mock type for the IngestRPC type
type IngestRPC struct {
	mock.Mock
}

// ExecuteScriptAtBlockID provides a mock function with given fields: ctx, script, blockID
func (_m *IngestRPC) ExecuteScriptAtBlockID(ctx context.Context, script []byte, blockID flow.Identifier) ([]byte, error) {
	ret := _m.Called(ctx, script, blockID)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, []byte, flow.Identifier) []byte); ok {
		r0 = rf(ctx, script, blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []byte, flow.Identifier) error); ok {
		r1 = rf(ctx, script, blockID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAccount provides a mock function with given fields: ctx, address, blockID
func (_m *IngestRPC) GetAccount(ctx context.Context, address flow.Address, blockID flow.Identifier) (*flow.Account, error) {
	ret := _m.Called(ctx, address, blockID)

	var r0 *flow.Account
	if rf, ok := ret.Get(0).(func(context.Context, flow.Address, flow.Identifier) *flow.Account); ok {
		r0 = rf(ctx, address, blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, flow.Address, flow.Identifier) error); ok {
		r1 = rf(ctx, address, blockID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
