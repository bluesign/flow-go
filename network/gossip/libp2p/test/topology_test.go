package test

import (
	"math"
	"os"
	"sort"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/network/codec/json"
	"github.com/onflow/flow-go/network/gossip/libp2p"
	"github.com/onflow/flow-go/network/gossip/libp2p/topology"
	"github.com/onflow/flow-go/utils/unittest"
)

// TopologyTestSuite tests the bare minimum requirements of a randomized
// topology that is needed for our network. It should not replace the information
// theory assumptions behind the schemes, e.g., random oracle model of hashes
type TopologyTestSuite struct {
	suite.Suite
	ids          flow.IdentityList // represents the identity list of all nodes in the system
	nets         *libp2p.Network   // represents the single network instance that creates topology
	count        int               // indicates size of system
	expectedSize int               // the expected topology size that will be generated by the RandPermTopology
}

// TestNetworkTestSuit starts all the tests in this test suite
func TestNetworkTestSuit(t *testing.T) {
	suite.Run(t, new(TopologyTestSuite))
}

// SetupTest initiates the test setups prior to each test
func (n *TopologyTestSuite) SetupTest() {
	n.count = 100
	n.ids = CreateIDs(n.count)
	rndSubsetSize := int(math.Ceil(float64(n.count+1) / 2))
	oneOfEachNodetype := 0 // there is only one node type in this test
	remaining := n.count - rndSubsetSize - oneOfEachNodetype
	halfOfRemainingNodes := int(math.Ceil(float64(remaining+1) / 2))
	n.expectedSize = rndSubsetSize + oneOfEachNodetype + halfOfRemainingNodes

	// takes firs id as the current nodes id
	me := n.ids[0]

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	key, err := GenerateNetworkingKey(me.NodeID)
	require.NoError(n.T(), err)

	metrics := metrics.NewNoopCollector()

	// creates a middleware instance
	mw, err := libp2p.NewMiddleware(logger,
		json.NewCodec(),
		"0.0.0.0:0",
		me.NodeID,
		key,
		metrics,
		libp2p.DefaultMaxUnicastMsgSize,
		libp2p.DefaultMaxPubSubMsgSize,
		unittest.IdentifierFixture().String())
	require.NoError(n.T(), err)

	// creates and mocks a network instance
	nets, err := createNetworks(logger, []*libp2p.Middleware{mw}, n.ids, 1, true, nil, nil)
	require.NoError(n.T(), err)
	require.Len(n.T(), nets, 1)
	n.nets = nets[0]
}

func (n *TopologyTestSuite) TestTopologySize() {
	// topology of size the entire network
	top, err := n.nets.Topology()
	require.NoError(n.T(), err)
	require.Len(n.T(), top, n.expectedSize)
}

// TestMembership evaluates every id in topology to be a protocol id
func (n *TopologyTestSuite) TestMembership() {
	top, err := n.nets.Topology()
	require.NoError(n.T(), err)
	require.Len(n.T(), top, n.expectedSize)

	// every id in topology should be an id of the protocol
	for id := range top {
		require.Contains(n.T(), n.ids.NodeIDs(), id)
	}
}

// TestDeteministicity verifies that the same seed generates the same topology
func (n *TopologyTestSuite) TestDeteministicity() {
	top, err := topology.NewRandPermTopology(flow.RoleCollection, unittest.IdentifierFixture())
	require.NoError(n.T(), err)
	// topology of size count/2
	topSize := uint(n.count / 2)
	var previous, current []string

	for i := 0; i < n.count; i++ {
		previous = current
		current = nil
		// generate a new topology with a the same ids, size and seed
		idMap, err := top.Subset(n.ids, topSize)
		require.NoError(n.T(), err)

		for _, v := range idMap {
			current = append(current, v.NodeID.String())
		}
		// no guarantees about order is made by Topology.Subset(), hence sort the return values before comparision
		sort.Strings(current)

		if previous == nil {
			continue
		}

		// assert that a different seed generates a different topology
		require.Equal(n.T(), previous, current)
	}
}

// TestUniqueness verifies that different seeds generates different topologies
func (n *TopologyTestSuite) TestUniqueness() {

	// topology of size count/2
	topSize := uint(n.count / 2)
	var previous, current []string

	for i := 0; i < n.count; i++ {
		previous = current
		current = nil
		// generate a new topology with a the same ids, size but a different seed for each iteration
		identity, _ := n.ids.ByIndex(uint(i))
		top, err := topology.NewRandPermTopology(flow.RoleCollection, identity.NodeID)
		require.NoError(n.T(), err)
		idMap, err := top.Subset(n.ids, topSize)
		require.NoError(n.T(), err)

		for _, v := range idMap {
			current = append(current, v.NodeID.String())
		}
		sort.Strings(current)

		if previous == nil {
			continue
		}

		// assert that a different seed generates a different topology
		require.NotEqual(n.T(), previous, current)
	}
}
