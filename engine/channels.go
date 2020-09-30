// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package engine

import (
	"fmt"

	"github.com/onflow/flow-go/model/flow"
)

// channel IDs
const (

	// Channels used for testing
	TestNetwork = "test-network"
	TestMetrics = "test-metrics"

	// Channels for consensus protocols
	ConsensusCommittee     = "consensus-committee"
	consensusClusterPrefix = "consensus-cluster" // use ChannelConsensusCluster

	// Channels for protocols actively synchronizing state across nodes
	SyncCommittee     = "sync-committee"
	syncClusterPrefix = "sync-cluster" // use ChannelSyncCluster
	SyncExecution     = "sync-execution"

	// Channels for actively pushing entities to subscribers
	PushTransactions = "push-transactions"
	PushGuarantees   = "push-guarantees"
	PushBlocks       = "push-blocks"
	PushReceipts     = "push-receipts"
	PushApprovals    = "push-approvals"

	// Channels for actively requesting missing entities
	RequestCollections       = "request-collections"
	RequestChunks            = "request-chunks"
	RequestReceiptsByBlockID = "request-receipts-by-block-id"

	// Channel aliases to make the code more readable / more robust to errors
	ReceiveTransactions = PushTransactions
	ReceiveGuarantees   = PushGuarantees
	ReceiveBlocks       = PushBlocks
	ReceiveReceipts     = PushReceipts
	ReceiveApprovals    = PushApprovals

	ProvideCollections       = RequestCollections
	ProvideChunks            = RequestChunks
	ProvideReceiptsByBlockID = RequestReceiptsByBlockID
)

// FullyQualifiedChannelName returns the unique channel name made up of channel name string suffixed with root block id
// The root block id is used to prevent cross talks between nodes on different sporks
func FullyQualifiedChannelName(channelID string, rootBlockID string) string {
	return fmt.Sprintf("%s/%s", channelID, rootBlockID)
}

// ChannelConsensusCluster returns a dynamic cluster consensus channel based on
// the chain ID of the cluster in question.
func ChannelConsensusCluster(clusterID flow.ChainID) string {
	return fmt.Sprint("%s-%s", consensusClusterPrefix, clusterID)
}

// ChannelSyncCluster returns a dynamic cluster sync channel based on the chain
// ID of the cluster in question.
func ChannelSyncCluster(clusterID flow.ChainID) string {
	return fmt.Sprint("%s-%s", syncClusterPrefix, clusterID)
}
