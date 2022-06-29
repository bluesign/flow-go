package committees

import (
	"fmt"
	"sync"

	"github.com/onflow/flow-go/consensus/hotstuff"
	"github.com/onflow/flow-go/consensus/hotstuff/committees/leader"
	"github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/flow/filter"
	"github.com/onflow/flow-go/module/component"
	"github.com/onflow/flow-go/module/irrecoverable"
	"github.com/onflow/flow-go/state/protocol"
	"github.com/onflow/flow-go/state/protocol/events"
	"github.com/onflow/flow-go/state/protocol/seed"
)

// staticEpochInfo contains leader selection and the initial committee for one epoch.
// This data structure must not be mutated after construction.
type staticEpochInfo struct {
	firstView    uint64                  // first view of the epoch (inclusive)
	finalView    uint64                  // final view of the epoch (inclusive)
	randomSource []byte                  // random source of epoch
	leaders      *leader.LeaderSelection // pre-computed leader selection for the epoch
	// TODO: should use identity skeleton https://github.com/dapperlabs/flow-go/issues/6232
	initialCommittee     flow.IdentityList
	weightThresholdForQC uint64 // computed based on initial committee weights
	weightThresholdForTO uint64 // computed based on initial committee weights
	dkg                  hotstuff.DKG
}

// newStaticEpochInfo returns the static epoch information from the epoch.
// This can be cached and used for all by-view queries for this epoch.
func newStaticEpochInfo(epoch protocol.Epoch) (*staticEpochInfo, error) {
	firstView, err := epoch.FirstView()
	if err != nil {
		return nil, fmt.Errorf("could not get first view: %w", err)
	}
	finalView, err := epoch.FinalView()
	if err != nil {
		return nil, fmt.Errorf("could not get final view: %w", err)
	}
	randomSource, err := epoch.RandomSource()
	if err != nil {
		return nil, fmt.Errorf("could not get random source: %w", err)
	}
	leaders, err := leader.SelectionForConsensus(epoch)
	if err != nil {
		return nil, fmt.Errorf("could not get leader selection: %w", err)
	}
	initialIdentities, err := epoch.InitialIdentities()
	if err != nil {
		return nil, fmt.Errorf("could not initial identities: %w", err)
	}
	initialCommittee := initialIdentities.Filter(filter.IsVotingConsensusCommitteeMember)
	dkg, err := epoch.DKG()
	if err != nil {
		return nil, fmt.Errorf("could not get dkg: %w", err)
	}

	totalWeight := initialCommittee.TotalWeight()
	epochInfo := &staticEpochInfo{
		firstView:            firstView,
		finalView:            finalView,
		randomSource:         randomSource,
		leaders:              leaders,
		initialCommittee:     initialCommittee,
		weightThresholdForQC: WeightThresholdToBuildQC(totalWeight),
		weightThresholdForTO: WeightThresholdToTimeout(totalWeight),
		dkg:                  dkg,
	}
	return epochInfo, nil
}

// newEmergencyFallbackEpoch creates an artificial fallback epoch generated from
// the last committed epoch at the time epoch emergency fallback is triggered.
// The fallback epoch:
// * begins after the last committed epoch
// * lasts until the next spork (estimated 6 months)
// * has the same static committee as the last committed epoch
func newEmergencyFallbackEpoch(lastCommittedEpoch *staticEpochInfo) (*staticEpochInfo, error) {

	rng, err := seed.PRGFromRandomSource(lastCommittedEpoch.randomSource, seed.ProtocolConsensusLeaderSelection)
	if err != nil {
		return nil, fmt.Errorf("could not create rng from seed: %w", err)
	}
	leaders, err := leader.ComputeLeaderSelection(lastCommittedEpoch.finalView+1, rng, leader.EstimatedSixMonthOfViews, lastCommittedEpoch.initialCommittee)
	if err != nil {
		return nil, fmt.Errorf("could not compute leader selection for fallback epoch: %w", err)
	}
	epochInfo := &staticEpochInfo{
		firstView:            lastCommittedEpoch.finalView + 1,
		finalView:            lastCommittedEpoch.finalView + leader.EstimatedSixMonthOfViews,
		randomSource:         lastCommittedEpoch.randomSource,
		leaders:              leaders,
		initialCommittee:     lastCommittedEpoch.initialCommittee,
		weightThresholdForQC: lastCommittedEpoch.weightThresholdForQC,
		weightThresholdForTO: lastCommittedEpoch.weightThresholdForTO,
		dkg:                  lastCommittedEpoch.dkg,
	}
	return epochInfo, nil
}

// Consensus represents the main committee for consensus nodes. The consensus
// committee might be active for multiple successive epochs.
type Consensus struct {
	state                  protocol.State              // the protocol state
	me                     flow.Identifier             // the node ID of this node
	mu                     sync.RWMutex                // protects access to epochs
	epochs                 map[uint64]*staticEpochInfo // cache of initial committee & leader selection per epoch
	committedEpochsCh      chan protocol.Epoch         // protocol events for newly committed epochs
	epochEmergencyFallback chan struct{}               // protocol event for epoch emergency fallback
	events.Noop                                        // implements protocol.Consumer
	component.Component
}

var _ hotstuff.Replicas = (*Consensus)(nil)
var _ hotstuff.DynamicCommittee = (*Consensus)(nil)

func NewConsensusCommittee(state protocol.State, me flow.Identifier) (*Consensus, error) {

	com := &Consensus{
		state:                  state,
		me:                     me,
		epochs:                 make(map[uint64]*staticEpochInfo),
		committedEpochsCh:      make(chan protocol.Epoch),
		epochEmergencyFallback: make(chan struct{}),
	}

	com.Component = component.NewComponentManagerBuilder().
		AddWorker(com.handleProtocolEvents).
		Build()

	final := state.Final()

	// pre-compute leader selection for all presently relevant committed epochs
	epochs := make([]protocol.Epoch, 0, 3)
	// we always prepare the current epoch
	epochs = append(epochs, final.Epochs().Current())

	// we prepare the previous epoch, if one exists
	exists, err := protocol.PreviousEpochExists(final)
	if err != nil {
		return nil, fmt.Errorf("could not check previous epoch exists: %w", err)
	}
	if exists {
		epochs = append(epochs, final.Epochs().Previous())
	}

	// we prepare the next epoch, if it is committed
	phase, err := final.Phase()
	if err != nil {
		return nil, fmt.Errorf("could not check epoch phase: %w", err)
	}
	if phase == flow.EpochPhaseCommitted {
		epochs = append(epochs, final.Epochs().Next())
	}

	// if epoch emergency fallback was triggered, inject the fallback epoch
	triggered, err := state.Params().EpochFallbackTriggered()
	if err != nil {
		return nil, fmt.Errorf("could not check epoch fallback: %w", err)
	}
	if triggered {
		err = com.onEpochEmergencyFallbackTriggered()
		if err != nil {
			return nil, fmt.Errorf("could not prepare emergency fallback epoch: %w", err)
		}
	}

	for _, epoch := range epochs {
		_, err = com.prepareEpoch(epoch)
		if err != nil {
			return nil, fmt.Errorf("could not prepare initial epochs: %w", err)
		}
	}

	return com, nil
}

func (c *Consensus) IdentitiesByBlock(blockID flow.Identifier) (flow.IdentityList, error) {
	il, err := c.state.AtBlockID(blockID).Identities(filter.IsVotingConsensusCommitteeMember)
	return il, err
}

func (c *Consensus) IdentityByBlock(blockID flow.Identifier, nodeID flow.Identifier) (*flow.Identity, error) {
	identity, err := c.state.AtBlockID(blockID).Identity(nodeID)
	if err != nil {
		if protocol.IsIdentityNotFound(err) {
			return nil, model.NewInvalidSignerErrorf("id %v is not a valid node id: %w", nodeID, err)
		}
		return nil, fmt.Errorf("could not get identity for node ID %x: %w", nodeID, err)
	}
	if !filter.IsVotingConsensusCommitteeMember(identity) {
		return nil, model.NewInvalidSignerErrorf("node %v is not an authorized hotstuff voting participant", nodeID)
	}
	return identity, nil
}

// IdentitiesByEpoch returns the committee identities in the epoch which contains
// the given view.
//
// Error returns:
//   * model.ErrViewForUnknownEpoch if no committed epoch containing the given view is known.
//     This is an expected error and must be handled.
//   * unspecific error in case of unexpected problems and bugs
func (c *Consensus) IdentitiesByEpoch(view uint64) (flow.IdentityList, error) {
	epochInfo, err := c.staticEpochInfoByView(view)
	if err != nil {
		return nil, err
	}
	return epochInfo.initialCommittee, nil
}

// IdentityByEpoch returns the identity for the given node ID, in the epoch which
// contains the given view.
//
// Error returns:
//   * model.ErrViewForUnknownEpoch if no committed epoch containing the given view is known.
//     This is an expected error and must be handled.
//   * model.InvalidSignerError if nodeID was not listed by the Epoch Setup event as an
//     authorized consensus participants.
//   * unspecific error in case of unexpected problems and bugs
//
func (c *Consensus) IdentityByEpoch(view uint64, nodeID flow.Identifier) (*flow.Identity, error) {
	epochInfo, err := c.staticEpochInfoByView(view)
	if err != nil {
		return nil, err
	}
	identity, ok := epochInfo.initialCommittee.ByNodeID(nodeID)
	if !ok {
		return nil, model.NewInvalidSignerErrorf("id %v is not a valid node id: %w", nodeID, err)
	}
	return identity, nil
}

// LeaderForView returns the node ID of the leader for the given view.
//
// Error returns:
//   * model.ErrViewForUnknownEpoch if no committed epoch containing the given view is known.
//     This is an expected error and must be handled.
//   * unspecific error in case of unexpected problems and bugs
func (c *Consensus) LeaderForView(view uint64) (flow.Identifier, error) {

	epochInfo, err := c.staticEpochInfoByView(view)
	if err != nil {
		return flow.ZeroID, err
	}
	return epochInfo.leaders.LeaderForView(view)
}

// QuorumThresholdForView returns the minimum weight required to build a valid
// QC in the given view. The weight threshold only changes at epoch boundaries
// and is computed based on the initial committee weights.
//
// Error returns:
//   * model.ErrViewForUnknownEpoch if no committed epoch containing the given view is known.
//     This is an expected error and must be handled.
//   * unspecific error in case of unexpected problems and bugs
func (c *Consensus) QuorumThresholdForView(view uint64) (uint64, error) {
	epochInfo, err := c.staticEpochInfoByView(view)
	if err != nil {
		return 0, err
	}
	return epochInfo.weightThresholdForQC, nil
}

func (c *Consensus) Self() flow.Identifier {
	return c.me
}

// TimeoutThresholdForView returns the minimum weight of observed timeout objects
// to safely immediately timeout for the current view. The weight threshold only
// changes at epoch boundaries and is computed based on the initial committee weights.
func (c *Consensus) TimeoutThresholdForView(view uint64) (uint64, error) {
	epochInfo, err := c.staticEpochInfoByView(view)
	if err != nil {
		return 0, err
	}
	return epochInfo.weightThresholdForTO, nil
}

// DKG returns the DKG for epoch which includes the given view.
//
// Error returns:
//   * model.ErrViewForUnknownEpoch if no committed epoch containing the given view is known.
//     This is an expected error and must be handled.
//   * unspecific error in case of unexpected problems and bugs
func (c *Consensus) DKG(view uint64) (hotstuff.DKG, error) {
	epochInfo, err := c.staticEpochInfoByView(view)
	if err != nil {
		return nil, err
	}
	return epochInfo.dkg, nil
}

// handleProtocolEvents is a worker routine which processes protocol events from
// the protocol state. When we observe a new epoch being committed, we compute
// the leader selection and cache static info for the epoch. When we observe
// epoch emergency fallback being triggered, we inject a fallback epoch.
func (c *Consensus) handleProtocolEvents(ctx irrecoverable.SignalerContext, ready component.ReadyFunc) {
	ready()

	for {
		select {
		case <-ctx.Done():
			return
		case epoch := <-c.committedEpochsCh:
			_, err := c.prepareEpoch(epoch)
			if err != nil {
				ctx.Throw(err)
			}
		case <-c.epochEmergencyFallback:
			err := c.onEpochEmergencyFallbackTriggered()
			if err != nil {
				ctx.Throw(err)
			}
		}
	}
}

// onEpochEmergencyFallbackTriggered handles the protocol event for emergency epoch
// fallback mode being triggered. When this occurs, we inject a fallback epoch
// to the committee which extends the current epoch.
// This method must also be called on initialization, if emergency fallback mode
// was triggered in the past.
// No errors are expected during normal operation.
func (c *Consensus) onEpochEmergencyFallbackTriggered() error {
	currentEpochCounter, err := c.state.Final().Epochs().Current().Counter()
	if err != nil {
		return fmt.Errorf("could not get current epoch counter: %w", err)
	}

	c.mu.RLock()
	// sanity check: current epoch must be cached already
	currentEpoch, ok := c.epochs[currentEpochCounter]
	c.mu.RUnlock()
	if !ok {
		return fmt.Errorf("epoch fallback: could not find current epoch (counter=%d) info", currentEpochCounter)
	}
	// sanity check: next epoch must never be committed, therefore must not be cached
	c.mu.RLock()
	_, ok = c.epochs[currentEpochCounter+1]
	c.mu.RUnlock()
	if ok {
		return fmt.Errorf("epoch fallback: next epoch (counter=%d) is cached contrary to expectation", currentEpochCounter+1)
	}

	fallbackEpoch, err := newEmergencyFallbackEpoch(currentEpoch)
	if err != nil {
		return fmt.Errorf("could not construct fallback epoch: %w", err)
	}

	// cache the epoch info
	c.mu.Lock()
	c.epochs[currentEpochCounter+1] = fallbackEpoch
	c.mu.Unlock()

	return nil
}

// staticEpochInfoByView retrieves the previously cached static epoch info for
// the epoch which includes the given view. If no epoch is known for the given
// view, we will attempt to cache the next epoch.
//
// Error returns:
//   * model.ErrViewForUnknownEpoch if no committed epoch containing the given view is known
//   * unspecific error in case of unexpected problems and bugs
func (c *Consensus) staticEpochInfoByView(view uint64) (*staticEpochInfo, error) {

	// look for an epoch matching this view for which we have already pre-computed
	// leader selection. Epochs last ~500k views, so we find the epoch here 99.99%
	// of the time. Since epochs are long-lived and we only cache the most recent 3,
	// this linear map iteration is inexpensive.
	c.mu.RLock()
	for _, epoch := range c.epochs {
		if epoch.firstView <= view && view <= epoch.finalView {
			c.mu.RUnlock()
			return epoch, nil
		}
	}
	c.mu.RUnlock()

	return nil, model.ErrViewForUnknownEpoch
}

// prepareEpoch pre-computes and stores the static epoch information for the
// given epoch, including leader selection. Calling prepareEpoch multiple times
// for the same epoch returns cached static epoch information.
// Input must be a committed epoch.
// No errors are expected during normal operation.
func (c *Consensus) prepareEpoch(epoch protocol.Epoch) (*staticEpochInfo, error) {

	counter, err := epoch.Counter()
	if err != nil {
		return nil, fmt.Errorf("could not get counter for current epoch: %w", err)
	}

	// this is a no-op if we have already computed static info for this epoch
	c.mu.RLock()
	epochInfo, exists := c.epochs[counter]
	c.mu.RUnlock()
	if exists {
		return epochInfo, nil
	}

	epochInfo, err = newStaticEpochInfo(epoch)
	if err != nil {
		return nil, fmt.Errorf("could not create static epoch info for epch %d: %w", counter, err)
	}

	// sanity check: ensure new epoch has contiguous views with the prior epoch
	c.mu.RLock()
	prevEpochInfo, exists := c.epochs[counter-1]
	c.mu.RUnlock()
	if exists {
		if epochInfo.firstView != prevEpochInfo.finalView+1 {
			return nil, fmt.Errorf("non-contiguous view ranges between consecutive epochs (epoch_%d=[%d,%d], epoch_%d=[%d,%d])",
				counter-1, prevEpochInfo.firstView, prevEpochInfo.finalView,
				counter, epochInfo.firstView, epochInfo.finalView)
		}
	}

	// cache the epoch info
	c.mu.Lock()
	defer c.mu.Unlock()
	c.epochs[counter] = epochInfo
	// now prune any old epochs, if we have exceeded our maximum of 3
	// if we have fewer than 3 epochs, this is a no-op
	c.pruneEpochInfo()
	return epochInfo, nil
}

// pruneEpochInfo removes any epochs older than the most recent 3.
// NOTE: Not safe for concurrent use - the caller must first acquire the lock.
func (c *Consensus) pruneEpochInfo() {
	// find the maximum counter, including the epoch we just computed
	max := uint64(0)
	for counter := range c.epochs {
		if counter > max {
			max = counter
		}
	}

	// remove any epochs which aren't within the most recent 3
	for counter := range c.epochs {
		if counter+3 <= max {
			delete(c.epochs, counter)
		}
	}
}
