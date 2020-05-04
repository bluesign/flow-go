// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package consensus

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/dgraph-io/badger/v2"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module/mempool"
	"github.com/dapperlabs/flow-go/storage/badger/operation"
	"github.com/dapperlabs/flow-go/storage/badger/procedure"
)

// Builder is the builder for consensus block payloads. Upon providing a payload
// hash, it also memorizes which entities were included into the payload.
type Builder struct {
	db         *badger.DB
	guarantees mempool.Guarantees
	seals      mempool.Seals
	cfg        Config
	// cache of block ID -> height for checking transaction expiry
	cache map[flow.Identifier]uint64
}

// NewBuilder creates a new block builder.
func NewBuilder(db *badger.DB, guarantees mempool.Guarantees, seals mempool.Seals, options ...func(*Config)) *Builder {

	// initialize default config
	cfg := Config{
		minInterval:  500 * time.Millisecond,
		maxInterval:  10 * time.Second,
		expiryBlocks: 64,
	}

	// apply option parameters
	for _, option := range options {
		option(&cfg)
	}

	b := &Builder{
		db:         db,
		guarantees: guarantees,
		seals:      seals,
		cfg:        cfg,
		cache:      make(map[flow.Identifier]uint64),
	}
	return b
}

// BuildOn creates a new block header build on the provided parent, using the given view and applying the
// custom setter function to allow the caller to make changes to the header before storing it.
func (b *Builder) BuildOn(parentID flow.Identifier, setter func(*flow.Header)) (*flow.Header, error) {
	var header *flow.Header
	err := b.db.Update(func(tx *badger.Txn) error {

		// STEP ONE: get the payload entity IDs for all entities that are included
		// in ancestor blocks which are not finalized yet; this allows us to avoid
		// including them in a block on the same fork twice

		// first, we need to know what the latest finalized block number is
		var boundary uint64
		err := operation.RetrieveBoundary(&boundary)(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve boundary: %w", err)
		}

		// calculate how many blocks we look back
		limit := boundary - b.cfg.expiryBlocks
		if limit > boundary { // overflow check
			limit = 0
		}

		// get the last finalized block ID
		var finalizedID flow.Identifier
		err = operation.RetrieveNumber(boundary, &finalizedID)(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve finalized ID: %w", err)
		}

		// for each unfinalized ancestor of the payload we are building, we retrieve
		// a list of all pending IDs for guarantees and seals; we can use them to
		// exclude entities from being included in two blocks on the same fork.
		ancestorID := parentID
		guaranteeLookup := make(map[flow.Identifier]struct{})
		sealLookup := make(map[flow.Identifier]struct{})
		for {

			// retrieve the header for the ancestor
			var ancestor flow.Header
			err = operation.RetrieveHeader(ancestorID, &ancestor)(tx)
			if err != nil {
				return fmt.Errorf("could not retrieve ancestor (%x): %w", ancestorID, err)
			}

			// if we have reached the limit, stop indexing
			if ancestor.Height <= limit {
				break
			}

			// look up the ancestor's guarantees
			var guaranteeIDs []flow.Identifier
			err = operation.LookupGuaranteePayload(ancestor.Height, ancestorID, ancestor.ParentID, &guaranteeIDs)(tx)
			if err != nil {
				return fmt.Errorf("could not look up ancestor guarantees (%x): %w", ancestor.PayloadHash, err)
			}

			// look up the ancestor's seals
			var sealIDs []flow.Identifier
			err = operation.LookupSealPayload(ancestor.Height, ancestorID, ancestor.ParentID, &sealIDs)(tx)
			if err != nil {
				return fmt.Errorf("could not look up ancestor seals (%x): %w", ancestor.PayloadHash, err)
			}

			// insert guarantees and seals into the lookups
			for _, guaranteeID := range guaranteeIDs {
				guaranteeLookup[guaranteeID] = struct{}{}
			}
			for _, sealID := range sealIDs {
				sealLookup[sealID] = struct{}{}
			}

			// continue with the next ancestor in the chain
			ancestorID = ancestor.ParentID
		}

		// STEP TWO: build a payload that includes as much of the memory pool
		// contents as possible, while remaining valid on the respective fork;
		// for guarantees, this implies they were not included in an unfinalized
		// ancestor yet; for block seals, it means they were not included yet
		// *and* they are a valid extension of the last valid execution state

		// collect guarantees from memory pool, excluding those already pending
		// on this fork
		// also find the minHeight of the ref. block id among all the guarantees in the pool
		guaranteeIDs, minHeight, err := b.getCandidateGuarantees(tx, guaranteeLookup)
		if err != nil {
			return fmt.Errorf("could not get candidate guarantees: %w", err)
		}

		// find any guarantees between the boundary and going back till minimal block height that conflict with FINALIZED blocks
		var invalidIDs map[flow.Identifier]struct{}
		err = operation.CheckGuaranteePayload(minHeight, boundary, finalizedID, guaranteeIDs, &invalidIDs)(tx)
		if err != nil {
			return fmt.Errorf("could not check guarantee payload: %w", err)
		}

		var guarantees []*flow.CollectionGuarantee
		for _, guaranteeID := range guaranteeIDs {

			_, isInvalid := invalidIDs[guaranteeID]
			if isInvalid {
				// remove from mempool, it will never be valid
				b.guarantees.Rem(guaranteeID)
				continue
			}

			// add ONLY non-conflicting guarantees to the final payload
			guarantee, err := b.guarantees.ByID(guaranteeID)
			if err != nil {
				return fmt.Errorf("could not get guarantee from pool: %w", err)
			}
			guarantees = append(guarantees, guarantee)
		}

		// get the finalized state commitment at the parent
		lastSeal := &flow.Seal{}
		err = procedure.LookupSealByBlock(parentID, lastSeal)(tx)
		if err != nil {
			return fmt.Errorf("could not get parent seal: %w", err)
		}

		// collect all block headers from the last sealed block to the parent
		var ancestorIDs []flow.Identifier
		ancestorID = parentID
		sealedID := lastSeal.BlockID
		for ancestorID != sealedID {

			// get the ancestor
			var ancestor flow.Header
			err = operation.RetrieveHeader(ancestorID, &ancestor)(tx)
			if err != nil {
				return fmt.Errorf("could not get ancestor: %w", err)
			}

			// sanity check; should never be going that long without seal
			if ancestor.Height <= limit {
				break
			}

			// add to list
			ancestorIDs = append(ancestorIDs, ancestorID)
			ancestorID = ancestor.ParentID
		}

		// for each ancestor on the path, we can now include the pending seals
		// if available
		var seals []*flow.Seal
		for i := len(ancestorIDs); i > 0; i-- {

			// get the ancestor from the list
			ancestorID := ancestorIDs[i-1]

			// try to get the seal from the memory pool
			seal, err := b.seals.ByBlockID(ancestorID)
			if err == mempool.ErrNotFound {
				break
			}
			if err != nil {
				return fmt.Errorf("could not get seal from cache: %w", err)
			}

			// add to list of seals to include
			seals = append(seals, seal)
		}

		// sanity check: each seal should connect to previous final state
		finalState := lastSeal.FinalState
		for _, seal := range seals {

			// check that the seal connects to previous initial state
			if !bytes.Equal(seal.InitialState, finalState) {
				return fmt.Errorf("state transition failure!")
			}

			// forward the final state
			finalState = seal.FinalState
		}

		// STEP THREE: we have the guarantees and seals we can validly include
		// in the payload built on top of the given block. Now we need to build
		// and store the block header, as well as index the payload contents.

		// build the payload so we can get the hash
		payload := flow.Payload{
			Identities: nil,
			Guarantees: guarantees,
			Seals:      seals,
		}

		// retrieve the parent to set the height
		var parent flow.Header
		err = operation.RetrieveHeader(parentID, &parent)(tx)
		if err != nil {
			return fmt.Errorf("could not retrieve parent: %w", err)
		}

		// calculate the timestamp and cutoffs
		timestamp := time.Now().UTC()
		from := parent.Timestamp.Add(b.cfg.minInterval)
		to := parent.Timestamp.Add(b.cfg.maxInterval)

		// adjust timestamp if outside of cutoffs
		if timestamp.Before(from) {
			timestamp = from
		}
		if timestamp.After(to) {
			timestamp = to
		}

		// construct default block on top of the provided parent
		header = &flow.Header{
			ChainID:     parent.ChainID,
			ParentID:    parentID,
			Height:      parent.Height + 1,
			Timestamp:   timestamp,
			PayloadHash: payload.Hash(),

			// the following fields should be set by the custom function as needed
			// NOTE: we could abstract all of this away into an interface{} field,
			// but that would be over the top as we will probably always use hotstuff
			View:           0,
			ParentVoterIDs: nil,
			ParentVoterSig: nil,
			ProposerID:     flow.ZeroID,
			ProposerSig:    nil,
		}

		// apply the custom fields setter of the consensus algorithm
		setter(header)

		// insert the header into the DB
		err = operation.InsertHeader(header)(tx)
		if err != nil {
			return fmt.Errorf("could not insert header: %w", err)
		}

		// insert the payload into the DB
		err = procedure.InsertPayload(&payload)(tx)
		if err != nil {
			return fmt.Errorf("could not insert payload: %w", err)
		}

		// index the payload for the block
		err = procedure.IndexPayload(header, &payload)(tx)
		if err != nil {
			return fmt.Errorf("could not index payload: %w", err)
		}

		// update the last seal if we have new ones included
		if len(seals) > 0 {
			lastSeal = seals[len(seals)-1]
		}

		// index the last seal for this block
		err = operation.IndexSealIDByBlock(header.ID(), lastSeal.ID())(tx)
		if err != nil {
			return fmt.Errorf("could not index commit: %w", err)
		}

		return nil
	})

	return header, err
}

// getCandidateGuarantees creates a set of candidate guarantees from the
// mempool. It guarantees that all candidates are not expired and do not
// reference un-finalized blocks.
func (b *Builder) getCandidateGuarantees(tx *badger.Txn, guaranteeLookup map[flow.Identifier]struct{}) ([]flow.Identifier, uint64, error) {

	// retrieve the finalized head ON THE MAIN CHAIN in order to know which
	// guarantees have expired and should be discarded
	var final flow.Header
	err := procedure.RetrieveLatestFinalizedHeader(&final)(tx)
	if err != nil {
		return nil, 0, fmt.Errorf("could not retrieve finalized header: %w", err)
	}

	// build a payload that includes as many guarantees from the memory
	// that have not expired
	var guaranteeIDs []flow.Identifier
	var minHeight uint64 = math.MaxUint64
	for _, guarantee := range b.guarantees.All() {

		_, ok := guaranteeLookup[guarantee.ID()]
		if ok {
			continue
		}

		refBlkID := guarantee.ReferenceBlockID

		refHeight, cached := b.cache[refBlkID]
		// the block isn't in our cache, retrieve it from storage
		if !cached {
			var ref flow.Header
			err = operation.RetrieveHeader(refBlkID, &ref)(tx)
			if err != nil {
				return nil, 0, fmt.Errorf("could not retrieve reference block: %w", err)
			}

			// sanity check: ensure the reference block is from the main chain
			if ref.ChainID != flow.DefaultChainID {
				return nil, 0, fmt.Errorf("invalid reference block (chain_id=%s)", ref.ChainID)
			}

			refHeight = ref.Height
			b.cache[refBlkID] = ref.Height
		}

		// for now, disallow un-finalized reference blocks
		if final.Height < refHeight {
			continue
		}

		if minHeight < refHeight {
			minHeight = refHeight
		}

		// ensure the reference block is not too old
		// TODO: Define a buffer similar to Transaction and then use (flow.DefaultTransactionExpiry -b.cfg.expiryBuffer)
		if final.Height-refHeight > flow.DefaultTransactionExpiry {
			// the guarantee is expired, it will never be valid
			b.guarantees.Rem(guarantee.ID())
			continue
		}

		guaranteeIDs = append(guaranteeIDs, guarantee.ID())
	}

	// invalidate expired items in reference block ID cache
	// NOTE: the maximum number of items here is 100s, so this linear-time
	// invalidation should be OK
	for id, height := range b.cache {
		if final.Height-height > flow.DefaultTransactionExpiry {
			delete(b.cache, id)
		}
	}

	return guaranteeIDs, minHeight, nil
}
