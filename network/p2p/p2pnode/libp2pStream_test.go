package p2pnode_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/onflow/flow-go/network/p2p"

	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/irrecoverable"
	"github.com/onflow/flow-go/network/internal/p2pfixtures"
	"github.com/onflow/flow-go/network/internal/p2putils"
	"github.com/onflow/flow-go/network/p2p/p2pnode"
	"github.com/onflow/flow-go/network/p2p/unicast"
	"github.com/onflow/flow-go/network/p2p/utils"
	"github.com/onflow/flow-go/utils/unittest"
)

// TestStreamClosing tests 1-1 communication with streams closed using libp2p2 handler.FullClose
func TestStreamClosing(t *testing.T) {
	count := 10
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	var msgRegex = regexp.MustCompile("^hello[0-9]")

	handler, streamCloseWG := mockStreamHandlerForMessages(t, ctx, count, msgRegex)

	// Creates nodes
	nodes, identities := p2pfixtures.NodesFixture(t,
		unittest.IdentifierFixture(),
		"test_stream_closing",
		2,
		p2pfixtures.WithDefaultStreamHandler(handler))

	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	nodeInfo1, err := utils.PeerAddressInfo(*identities[1])
	require.NoError(t, err)

	senderWG := sync.WaitGroup{}
	senderWG.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			// Create stream from node 1 to node 2 (reuse if one already exists)
			nodes[0].Host().Peerstore().AddAddrs(nodeInfo1.ID, nodeInfo1.Addrs, peerstore.AddressTTL)
			s, err := nodes[0].CreateStream(ctx, nodeInfo1.ID)
			assert.NoError(t, err)
			w := bufio.NewWriter(s)

			// Send message from node 1 to 2
			msg := fmt.Sprintf("hello%d\n", i)
			_, err = w.WriteString(msg)
			assert.NoError(t, err)

			// Flush the stream
			assert.NoError(t, w.Flush())

			// close the stream
			err = s.Close()
			require.NoError(t, err)

			senderWG.Done()
		}(i)
	}

	// wait for stream to be closed
	unittest.RequireReturnsBefore(t, senderWG.Wait, 1*time.Second, "could not send messages on time")
	unittest.RequireReturnsBefore(t, streamCloseWG.Wait, 1*time.Second, "could not close stream at receiver side")
}

// mockStreamHandlerForMessages creates a stream handler that expects receiving `msgCount` unique messages that match the input regexp.
// The returned wait group will be unlocked when all messages are completely received and associated streams are closed.
func mockStreamHandlerForMessages(t *testing.T, ctx context.Context, msgCount int, msgRegexp *regexp.Regexp) (network.StreamHandler, *sync.WaitGroup) {
	streamCloseWG := &sync.WaitGroup{}
	streamCloseWG.Add(msgCount)

	h := func(s network.Stream) {
		go func(s network.Stream) {
			rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
			for {
				str, err := rw.ReadString('\n')
				if err != nil {
					if errors.Is(err, io.EOF) {
						err := s.Close()
						require.NoError(t, err)

						streamCloseWG.Done()
						return
					}
					require.Fail(t, fmt.Sprintf("received error %v", err))
					err = s.Reset()
					require.NoError(t, err)
					return
				}
				select {
				case <-ctx.Done():
					return
				default:
					require.True(t, msgRegexp.MatchString(str), str)
				}
			}
		}(s)

	}
	return h, streamCloseWG
}

// TestCreateStream_WithDefaultUnicast evaluates correctness of creating default (tcp) unicast streams between two libp2p nodes.
func TestCreateStream_WithDefaultUnicast(t *testing.T) {
	sporkId := unittest.IdentifierFixture()
	testCreateStream(t,
		sporkId,
		nil, // sends nil as preferred unicast so that nodes run on default plain tcp streams.
		unicast.FlowProtocolID(sporkId))
}

// TestCreateStream_WithPreferredGzipUnicast evaluates correctness of creating gzip-compressed tcp unicast streams between two libp2p nodes.
func TestCreateStream_WithPreferredGzipUnicast(t *testing.T) {
	sporkId := unittest.IdentifierFixture()
	testCreateStream(t,
		sporkId,
		[]unicast.ProtocolName{unicast.GzipCompressionUnicast},
		unicast.FlowGzipProtocolId(sporkId))
}

// testCreateStreams checks if a new streams of "preferred" type is created each time when CreateStream is called and an existing stream is not
// reused. The "preferred" stream type is the one with the largest index in `unicasts` list.
// To check that the streams are of "preferred" type, it evaluates the protocol id of established stream against the input `protocolID`.
func testCreateStream(t *testing.T, sporkId flow.Identifier, unicasts []unicast.ProtocolName, protocolID core.ProtocolID) {
	count := 2
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	nodes, identities := p2pfixtures.NodesFixture(t,
		sporkId,
		"test_create_stream",
		count,
		p2pfixtures.WithPreferredUnicasts(unicasts))

	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	id2 := identities[1]

	// Assert that there is no outbound stream to the target yet
	require.Equal(t, 0, p2putils.CountStream(nodes[0].Host(), nodes[1].Host().ID(), protocolID, network.DirOutbound))

	// Now attempt to create another 100 outbound stream to the same destination by calling CreateStream
	streamCount := 100
	var streams []network.Stream
	for i := 0; i < streamCount; i++ {
		pInfo, err := utils.PeerAddressInfo(*id2)
		require.NoError(t, err)
		nodes[0].Host().Peerstore().AddAddrs(pInfo.ID, pInfo.Addrs, peerstore.AddressTTL)
		anotherStream, err := nodes[0].CreateStream(ctx, pInfo.ID)
		// Assert that a stream was returned without error
		require.NoError(t, err)
		require.NotNil(t, anotherStream)
		// assert that the stream count within libp2p incremented (a new stream was created)
		require.Equal(t, i+1, p2putils.CountStream(nodes[0].Host(), nodes[1].Host().ID(), protocolID, network.DirOutbound))
		// assert that the same connection is reused
		require.Len(t, nodes[0].Host().Network().Conns(), 1)
		streams = append(streams, anotherStream)
	}

	// reverse loop to close all the streams
	for i := streamCount - 1; i >= 0; i-- {
		s := streams[i]
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := s.Close()
			assert.NoError(t, err)
			wg.Done()
		}()
		unittest.RequireReturnsBefore(t, wg.Wait, 1*time.Second, "could not close streams on time")
		// assert that the stream count within libp2p decremented
		require.Equal(t, i, p2putils.CountStream(nodes[0].Host(), nodes[1].Host().ID(), protocolID, network.DirOutbound))
	}
}

// TestCreateStream_FallBack checks two libp2p nodes with conflicting supported unicast protocols fall back
// to default (tcp) unicast protocol during their negotiation.
// To do this, a node with preferred gzip-compressed tcp unicast tries creating stream to another node that only
// supports default plain tcp unicast. The test evaluates that the unicast stream established between two nodes
// are of type default plain tcp.
func TestCreateStream_FallBack(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	// Creates two nodes: one with preferred gzip, and other one with default protocol
	sporkId := unittest.IdentifierFixture()
	thisNode, _ := p2pfixtures.NodeFixture(t,
		sporkId,
		"test_create_stream_fallback",
		p2pfixtures.WithPreferredUnicasts([]unicast.ProtocolName{unicast.GzipCompressionUnicast}))
	otherNode, otherId := p2pfixtures.NodeFixture(t, sporkId, "test_create_stream_fallback")

	nodes := []p2p.LibP2PNode{thisNode, otherNode}
	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	// Assert that there is no outbound stream to the target yet (neither default nor preferred)
	defaultProtocolId := unicast.FlowProtocolID(sporkId)
	preferredProtocolId := unicast.FlowGzipProtocolId(sporkId)
	require.Equal(t, 0, p2putils.CountStream(thisNode.Host(), otherNode.Host().ID(), defaultProtocolId, network.DirOutbound))
	require.Equal(t, 0, p2putils.CountStream(thisNode.Host(), otherNode.Host().ID(), preferredProtocolId, network.DirOutbound))

	// Now attempt to create another 100 outbound stream to the same destination by calling CreateStream
	streamCount := 100
	var streams []network.Stream
	for i := 0; i < streamCount; i++ {
		pInfo, err := utils.PeerAddressInfo(otherId)
		require.NoError(t, err)
		thisNode.Host().Peerstore().AddAddrs(pInfo.ID, pInfo.Addrs, peerstore.AddressTTL)

		// a new stream must be created
		anotherStream, err := thisNode.CreateStream(ctx, pInfo.ID)
		require.NoError(t, err)
		require.NotNil(t, anotherStream)

		// number of default-protocol streams must be incremented, while preferred ones must be zero, since the other node
		// only supports default ones.
		require.Equal(t, i+1, p2putils.CountStream(thisNode.Host(), otherNode.Host().ID(), defaultProtocolId, network.DirOutbound))
		require.Equal(t, 0, p2putils.CountStream(thisNode.Host(), otherNode.Host().ID(), preferredProtocolId, network.DirOutbound))

		// assert that the same connection is reused
		require.Len(t, thisNode.Host().Network().Conns(), 1)
		streams = append(streams, anotherStream)
	}

	// reverse loop to close all the streams
	for i := streamCount - 1; i >= 0; i-- {
		s := streams[i]
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := s.Close()
			assert.NoError(t, err)
			wg.Done()
		}()
		unittest.RequireReturnsBefore(t, wg.Wait, 1*time.Second, "could not close streams on time")

		// number of default-protocol streams must be decremented, while preferred ones must be zero, since the other node
		// only supports default ones.
		require.Equal(t, i, p2putils.CountStream(thisNode.Host(), otherNode.Host().ID(), defaultProtocolId, network.DirOutbound))
		require.Equal(t, 0, p2putils.CountStream(thisNode.Host(), otherNode.Host().ID(), preferredProtocolId, network.DirOutbound))
	}
}

// TestCreateStreamIsConcurrencySafe tests that the CreateStream is concurrency safe
func TestCreateStreamIsConcurrencySafe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	// create two nodes
	nodes, identities := p2pfixtures.NodesFixture(t, unittest.IdentifierFixture(), "test_create_stream_is_concurrency_safe", 2)
	require.Len(t, identities, 2)

	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	nodeInfo1, err := utils.PeerAddressInfo(*identities[1])
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	// create a gate which gates the call to CreateStream for all concurrent go routines
	gate := make(chan struct{})

	createStream := func() {
		<-gate
		nodes[0].Host().Peerstore().AddAddrs(nodeInfo1.ID, nodeInfo1.Addrs, peerstore.AddressTTL)
		_, err := nodes[0].CreateStream(ctx, nodeInfo1.ID)
		assert.NoError(t, err) // assert that stream was successfully created
		wg.Done()
	}

	// kick off 10 concurrent calls to CreateStream
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go createStream()
	}
	// open the gate by closing the channel
	close(gate)

	// no call should block
	unittest.AssertReturnsBefore(t, wg.Wait, 10*time.Second)
}

// TestNoBackoffWhenCreatingStream checks that backoff is not enabled between attempts to connect to a remote peer
// for one-to-one direct communication.
func TestNoBackoffWhenCreatingStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup per node contexts so they can be stopped independently
	ctx1, cancel1 := context.WithCancel(ctx)
	signalerCtx1 := irrecoverable.NewMockSignalerContext(t, ctx1)

	ctx2, cancel2 := context.WithCancel(ctx)
	signalerCtx2 := irrecoverable.NewMockSignalerContext(t, ctx2)

	count := 2
	// Creates nodes
	nodes, identities := p2pfixtures.NodesFixture(t,
		unittest.IdentifierFixture(),
		"test_no_backoff_when_create_stream",
		count,
	)
	node1 := nodes[0]
	node2 := nodes[1]

	p2pfixtures.StartNode(t, signalerCtx1, node1, 100*time.Millisecond)
	p2pfixtures.StartNode(t, signalerCtx2, node2, 100*time.Millisecond)

	// stop node 2 immediately
	p2pfixtures.StopNode(t, node2, cancel2, 100*time.Millisecond)
	defer p2pfixtures.StopNode(t, node1, cancel1, 100*time.Millisecond)

	id2 := identities[1]
	pInfo, err := utils.PeerAddressInfo(*id2)
	require.NoError(t, err)
	nodes[0].Host().Peerstore().AddAddrs(pInfo.ID, pInfo.Addrs, peerstore.AddressTTL)
	maxTimeToWait := p2pnode.MaxConnectAttempt * unicast.MaxConnectAttemptSleepDuration * time.Millisecond

	// need to add some buffer time so that RequireReturnsBefore waits slightly longer than maxTimeToWait to avoid
	// a race condition
	someGraceTime := 100 * time.Millisecond
	totalWaitTime := maxTimeToWait + someGraceTime

	//each CreateStream() call may try to connect up to MaxConnectAttempt (3) times.

	//there are 2 scenarios that we need to account for:
	//
	//1. machines where a timeout occurs on the first connection attempt - this can be due to local firewall rules or other processes running on the machine.
	//   In this case, we need to create a scenario where a backoff would have normally occured. This is why we initiate a second connection attempt.
	//   Libp2p remembers the peer we are trying to connect to between CreateStream() calls and would have initiated a backoff if backoff wasn't turned off.
	//   The second CreateStream() call will make a second connection attempt MaxConnectAttempt times and that should never result in a backoff error.
	//
	//2. machines where a timeout does NOT occur on the first connection attempt - this is on CI machines and some local dev machines without a firewall / too many other processes.
	//   In this case, there will be MaxConnectAttempt (3) connection attempts on the first CreateStream() call and MaxConnectAttempt (3) attempts on the second CreateStream() call.

	// make two separate stream creation attempt and assert that no connection back off happened
	for i := 0; i < 2; i++ {

		// limit the maximum amount of time to wait for a connection to be established by using a context that times out
		ctx, cancel := context.WithTimeout(ctx, maxTimeToWait)

		unittest.RequireReturnsBefore(t, func() {
			_, err = node1.CreateStream(ctx, pInfo.ID)
		}, totalWaitTime, fmt.Sprintf("create stream did not error within %s", totalWaitTime.String()))
		require.Error(t, err)
		require.NotContainsf(t, err.Error(), swarm.ErrDialBackoff.Error(), "swarm dialer unexpectedly did a back off for a one-to-one connection")
		cancel()
	}
}

// TestUnicastOverStream_WithPlainStream checks two nodes can send and receive unicast messages on libp2p plain streams.
func TestUnicastOverStream_WithPlainStream(t *testing.T) {
	testUnicastOverStream(t)
}

// TestUnicastOverStream_WithGzipStreamCompression checks two nodes can send and receive unicast messages on gzip compressed streams
// when both nodes have gzip stream compression enabled.
func TestUnicastOverStream_WithGzipStreamCompression(t *testing.T) {
	testUnicastOverStream(t, p2pfixtures.WithPreferredUnicasts([]unicast.ProtocolName{unicast.GzipCompressionUnicast}))
}

// testUnicastOverStream sends a message from node 1 to node 2 and then from node 2 to node 1 over a unicast stream.
func testUnicastOverStream(t *testing.T, opts ...p2pfixtures.NodeFixtureParameterOption) {
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	// Creates nodes
	sporkId := unittest.IdentifierFixture()

	streamHandler1, inbound1 := p2pfixtures.StreamHandlerFixture(t)
	node1, id1 := p2pfixtures.NodeFixture(
		t,
		sporkId,
		t.Name(),
		append(opts, p2pfixtures.WithDefaultStreamHandler(streamHandler1))...)

	streamHandler2, inbound2 := p2pfixtures.StreamHandlerFixture(t)
	node2, id2 := p2pfixtures.NodeFixture(
		t,
		sporkId,
		t.Name(),
		append(opts, p2pfixtures.WithDefaultStreamHandler(streamHandler2))...)

	nodes := []p2p.LibP2PNode{node1, node2}
	ids := flow.IdentityList{&id1, &id2}
	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	p2pfixtures.LetNodesDiscoverEachOther(t, ctx, nodes, ids)

	p2pfixtures.EnsureMessageExchangeOverUnicast(
		t,
		ctx,
		nodes,
		[]chan string{inbound1, inbound2},
		p2pfixtures.LongStringMessageFactoryFixture(t))
}

// TestUnicastOverStream_Fallback checks two nodes with asymmetric sets of preferred unicast protocols can create streams and
// send and receive unicasts. Despite the asymmetry, the nodes must fall back to the libp2p plain stream during negotiation.
func TestUnicastOverStream_Fallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	// Creates nodes
	// node1: supports only plain unicast protocol
	// node2: supports plain and gzip
	sporkId := unittest.IdentifierFixture()

	streamHandler1, inbound1 := p2pfixtures.StreamHandlerFixture(t)
	node1, id1 := p2pfixtures.NodeFixture(
		t,
		sporkId,
		t.Name(),
		p2pfixtures.WithDefaultStreamHandler(streamHandler1),
	)

	streamHandler2, inbound2 := p2pfixtures.StreamHandlerFixture(t)
	node2, id2 := p2pfixtures.NodeFixture(
		t,
		sporkId,
		t.Name(),
		p2pfixtures.WithDefaultStreamHandler(streamHandler2),
		p2pfixtures.WithPreferredUnicasts([]unicast.ProtocolName{unicast.GzipCompressionUnicast}),
	)

	nodes := []p2p.LibP2PNode{node1, node2}
	ids := flow.IdentityList{&id1, &id2}
	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	p2pfixtures.LetNodesDiscoverEachOther(t, ctx, nodes, ids)
	p2pfixtures.EnsureMessageExchangeOverUnicast(t, ctx, nodes, []chan string{inbound1, inbound2}, p2pfixtures.LongStringMessageFactoryFixture(t))
}

// TestCreateStreamTimeoutWithUnresponsiveNode tests that the CreateStream call does not block longer than the
// timeout interval
func TestCreateStreamTimeoutWithUnresponsiveNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	// creates a regular node
	nodes, identities := p2pfixtures.NodesFixture(t,
		unittest.IdentifierFixture(),
		"test_create_stream_timeout_with_unresponsive_node",
		1,
	)
	require.Len(t, identities, 1)

	p2pfixtures.StartNodes(t, signalerCtx, nodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, nodes, cancel, 100*time.Millisecond)

	// create a silent node which never replies
	listener, silentNodeId := p2pfixtures.SilentNodeFixture(t)
	defer func() {
		require.NoError(t, listener.Close())
	}()

	silentNodeInfo, err := utils.PeerAddressInfo(silentNodeId)
	require.NoError(t, err)

	timeout := 1 * time.Second
	tctx, tcancel := context.WithTimeout(ctx, timeout)
	defer tcancel()

	// attempt to create a stream from node 1 to node 2 and assert that it fails after timeout
	grace := 100 * time.Millisecond
	unittest.AssertReturnsBefore(t,
		func() {
			nodes[0].Host().Peerstore().AddAddrs(silentNodeInfo.ID, silentNodeInfo.Addrs, peerstore.AddressTTL)
			_, err = nodes[0].CreateStream(tctx, silentNodeInfo.ID)
		},
		timeout+grace)
	assert.Error(t, err)
}

// TestCreateStreamIsConcurrent tests that CreateStream calls can be made concurrently such that one blocked call
// does not block another concurrent call.
func TestCreateStreamIsConcurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	signalerCtx := irrecoverable.NewMockSignalerContext(t, ctx)

	// create two regular node
	goodNodes, goodNodeIds := p2pfixtures.NodesFixture(t,
		unittest.IdentifierFixture(),
		"test_create_stream_is_concurrent",
		2,
	)
	require.Len(t, goodNodeIds, 2)

	p2pfixtures.StartNodes(t, signalerCtx, goodNodes, 100*time.Millisecond)
	defer p2pfixtures.StopNodes(t, goodNodes, cancel, 100*time.Millisecond)

	goodNodeInfo1, err := utils.PeerAddressInfo(*goodNodeIds[1])
	require.NoError(t, err)

	// create a silent node which never replies
	listener, silentNodeId := p2pfixtures.SilentNodeFixture(t)
	defer func() {
		require.NoError(t, listener.Close())
	}()
	silentNodeInfo, err := utils.PeerAddressInfo(silentNodeId)
	require.NoError(t, err)

	// creates a stream to unresponsive node and makes sure that the stream creation is blocked
	blockedCallCh := unittest.RequireNeverReturnBefore(t,
		func() {
			goodNodes[0].Host().Peerstore().AddAddrs(silentNodeInfo.ID, silentNodeInfo.Addrs, peerstore.AddressTTL)
			_, _ = goodNodes[0].CreateStream(ctx, silentNodeInfo.ID) // this call will block
		},
		1*time.Second,
		"CreateStream attempt to the unresponsive peer did not block")

	// requires same peer can still connect to the other regular peer without being blocked
	unittest.RequireReturnsBefore(t,
		func() {
			goodNodes[0].Host().Peerstore().AddAddrs(goodNodeInfo1.ID, goodNodeInfo1.Addrs, peerstore.AddressTTL)
			_, err := goodNodes[0].CreateStream(ctx, goodNodeInfo1.ID)
			require.NoError(t, err)
		},
		1*time.Second, "creating stream to a responsive node failed while concurrently blocked on unresponsive node")

	// requires the CreateStream call to the unresponsive node was blocked while we attempted the CreateStream to the
	// good address
	unittest.RequireNeverClosedWithin(t, blockedCallCh, 1*time.Millisecond,
		"CreateStream attempt to the unresponsive peer did not block after connecting to good node")

}