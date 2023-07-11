package internal

import (
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/network/channels"
	"github.com/onflow/flow-go/utils/unittest"
)

// TestNewRPCSentTracker ensures *RPCSenTracker is created as expected.
func TestNewRPCSentTracker(t *testing.T) {
	tracker := mockTracker()
	require.NotNil(t, tracker)
}

// TestRPCSentTracker_IHave ensures *RPCSentTracker tracks sent iHave control messages as expected.
func TestRPCSentTracker_IHave(t *testing.T) {
	tracker := mockTracker()
	require.NotNil(t, tracker)

	t.Run("WasIHaveRPCSent should return false for iHave message Id that has not been tracked", func(t *testing.T) {
		require.False(t, tracker.WasIHaveRPCSent("topic_id", "message_id"))
	})

	t.Run("WasIHaveRPCSent should return true for iHave message after it is tracked with OnIHaveRPCSent", func(t *testing.T) {
		topicID := channels.PushBlocks.String()
		messageID1 := unittest.IdentifierFixture().String()
		iHaves := []*pb.ControlIHave{{
			TopicID:    &topicID,
			MessageIDs: []string{messageID1},
		}}
		rpc := rpcFixture(withIhaves(iHaves))
		tracker.OnIHaveRPCSent(rpc.GetControl().GetIhave())
		require.True(t, tracker.WasIHaveRPCSent(topicID, messageID1))

		// manipulate last byte of message ID ensure false positive not returned
		messageID2 := []byte(messageID1)
		messageID2[len(messageID2)-1] = 'X'
		require.False(t, tracker.WasIHaveRPCSent(topicID, string(messageID2)))
	})
}

func mockTracker() *RPCSentTracker {
	logger := zerolog.Nop()
	sizeLimit := uint32(100)
	collector := metrics.NewNoopCollector()
	tracker := NewRPCSentTracker(logger, sizeLimit, collector)
	return tracker
}

type rpcFixtureOpt func(*pubsub.RPC)

func withIhaves(iHave []*pb.ControlIHave) rpcFixtureOpt {
	return func(rpc *pubsub.RPC) {
		rpc.Control.Ihave = iHave
	}
}

func rpcFixture(opts ...rpcFixtureOpt) *pubsub.RPC {
	rpc := &pubsub.RPC{
		RPC: pb.RPC{
			Control: &pb.ControlMessage{},
		},
	}
	for _, opt := range opts {
		opt(rpc)
	}
	return rpc
}
