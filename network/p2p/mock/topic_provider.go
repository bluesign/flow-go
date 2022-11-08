// Code generated by mockery v2.13.1. DO NOT EDIT.

package mockp2p

import (
	mock "github.com/stretchr/testify/mock"

	peer "github.com/libp2p/go-libp2p/core/peer"
)

// TopicProvider is an autogenerated mock type for the TopicProvider type
type TopicProvider struct {
	mock.Mock
}

// GetTopics provides a mock function with given fields:
func (_m *TopicProvider) GetTopics() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// ListPeers provides a mock function with given fields: topic
func (_m *TopicProvider) ListPeers(topic string) []peer.ID {
	ret := _m.Called(topic)

	var r0 []peer.ID
	if rf, ok := ret.Get(0).(func(string) []peer.ID); ok {
		r0 = rf(topic)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]peer.ID)
		}
	}

	return r0
}

type mockConstructorTestingTNewTopicProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewTopicProvider creates a new instance of TopicProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTopicProvider(t mockConstructorTestingTNewTopicProvider) *TopicProvider {
	mock := &TopicProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}