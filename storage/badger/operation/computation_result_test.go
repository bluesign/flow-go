// (c) 2022 Dapper Labs - ALL RIGHTS RESERVED

package operation

import (
	"reflect"
	"testing"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/common/pathfinder"
	"github.com/onflow/flow-go/ledger/complete"
	"github.com/onflow/flow-go/model/flow"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/engine/execution"
	"github.com/onflow/flow-go/utils/unittest"
)

func TestInsertAndRetrieveComputationResults(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		expected := generateComputationResult(t)

		t.Run("Insert and Retrieve ComputationResult", func(t *testing.T) {
			expectedId := expected.ExecutableBlock.ID()
			err := db.Update(InsertComputationResult(expectedId, expected))
			require.NoError(t, err)

			var actual execution.ComputationResult
			err = db.View(GetComputationResult(expectedId, &actual))
			require.NoError(t, err)

			actual.ExecutableBlock.ID()
			assert.Equal(t, expected, &actual)
		})
	})
}

func TestRemoveComputationResults(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		expected := generateComputationResult(t)

		t.Run("Remove ComputationResult", func(t *testing.T) {
			expectedId := expected.ExecutableBlock.ID()

			err := db.Update(InsertComputationResult(expectedId, expected))
			require.NoError(t, err)

			var actual execution.ComputationResult
			err = db.View(GetComputationResult(expectedId, &actual))
			require.NoError(t, err)

			actual.ExecutableBlock.ID()
			assert.Equal(t, expected, &actual)

			err = db.Update(RemoveComputationResult(expectedId))
			require.NoError(t, err)

			err = db.View(GetComputationResult(expectedId, &actual))
			assert.NotNil(t, err)
		})
	})
}

func TestListComputationResults(t *testing.T) {
	unittest.RunWithBadgerDB(t, func(db *badger.DB) {
		expected := [...]*execution.ComputationResult{
			generateComputationResult(t),
			generateComputationResult(t),
		}
		t.Run("List all ComputationResult", func(t *testing.T) {
			expectedIDs := make(map[string]bool, 0)
			// Store a list of ComputationResult instances first
			for _, cr := range expected {
				expectedId := cr.ExecutableBlock.ID()
				expectedIDs[expectedId.String()] = true
				err := db.Update(InsertComputationResult(expectedId, cr))
				require.NoError(t, err)
			}

			// Get the list of IDs of stored ComputationResult
			crIDs := make([]flow.Identifier, 0)
			err := db.View(GetAllComputationResultIDs(&crIDs))
			require.NoError(t, err)
			crIDsStrMap := make(map[string]bool, 0)
			for _, crID := range crIDs {
				crIDsStrMap[crID.String()] = true
			}

			assert.True(t, reflect.DeepEqual(crIDsStrMap, expectedIDs))
		})
	})
}

// Generate ComputationResult for testing purposes
func generateComputationResult(t *testing.T) *execution.ComputationResult {

	update1, err := ledger.NewUpdate(
		ledger.State(unittest.StateCommitmentFixture()),
		[]ledger.Key{
			ledger.NewKey([]ledger.KeyPart{ledger.NewKeyPart(3, []byte{33})}),
			ledger.NewKey([]ledger.KeyPart{ledger.NewKeyPart(1, []byte{11})}),
			ledger.NewKey([]ledger.KeyPart{ledger.NewKeyPart(2, []byte{1, 1}), ledger.NewKeyPart(3, []byte{2, 5})}),
		},
		[]ledger.Value{
			[]byte{21, 37},
			nil,
			[]byte{3, 3, 3, 3, 3},
		},
	)
	require.NoError(t, err)

	trieUpdate1, err := pathfinder.UpdateToTrieUpdate(update1, complete.DefaultPathFinderVersion)
	require.NoError(t, err)

	update2, err := ledger.NewUpdate(
		ledger.State(unittest.StateCommitmentFixture()),
		[]ledger.Key{},
		[]ledger.Value{},
	)
	require.NoError(t, err)

	trieUpdate2, err := pathfinder.UpdateToTrieUpdate(update2, complete.DefaultPathFinderVersion)
	require.NoError(t, err)

	update3, err := ledger.NewUpdate(
		ledger.State(unittest.StateCommitmentFixture()),
		[]ledger.Key{
			ledger.NewKey([]ledger.KeyPart{ledger.NewKeyPart(9, []byte{6})}),
		},
		[]ledger.Value{
			[]byte{21, 37},
		},
	)
	require.NoError(t, err)

	trieUpdate3, err := pathfinder.UpdateToTrieUpdate(update3, complete.DefaultPathFinderVersion)
	require.NoError(t, err)

	update4, err := ledger.NewUpdate(
		ledger.State(unittest.StateCommitmentFixture()),
		[]ledger.Key{
			ledger.NewKey([]ledger.KeyPart{ledger.NewKeyPart(9, []byte{6})}),
		},
		[]ledger.Value{
			[]byte{21, 37},
		},
	)
	require.NoError(t, err)

	trieUpdate4, err := pathfinder.UpdateToTrieUpdate(update4, complete.DefaultPathFinderVersion)
	require.NoError(t, err)

	return &execution.ComputationResult{
		ExecutableBlock: unittest.ExecutableBlockFixture([][]flow.Identifier{
			{unittest.IdentifierFixture()},
			{unittest.IdentifierFixture()},
			{unittest.IdentifierFixture()},
		}),
		StateSnapshots: nil,
		StateCommitments: []flow.StateCommitment{
			unittest.StateCommitmentFixture(),
			unittest.StateCommitmentFixture(),
			unittest.StateCommitmentFixture(),
			unittest.StateCommitmentFixture(),
		},
		Proofs: nil,
		Events: []flow.EventsList{
			{
				unittest.EventFixture("what", 0, 0, unittest.IdentifierFixture(), 2),
				unittest.EventFixture("ever", 0, 1, unittest.IdentifierFixture(), 22),
			},
			{},
			{
				unittest.EventFixture("what", 2, 0, unittest.IdentifierFixture(), 2),
				unittest.EventFixture("ever", 2, 1, unittest.IdentifierFixture(), 22),
				unittest.EventFixture("ever", 2, 2, unittest.IdentifierFixture(), 2),
				unittest.EventFixture("ever", 2, 3, unittest.IdentifierFixture(), 22),
			},
			{}, // system chunk events
		},
		EventsHashes:  nil,
		ServiceEvents: nil,
		TransactionResults: []flow.TransactionResult{
			{
				TransactionID:   unittest.IdentifierFixture(),
				ErrorMessage:    "",
				ComputationUsed: 23,
			},
			{
				TransactionID:   unittest.IdentifierFixture(),
				ErrorMessage:    "fail",
				ComputationUsed: 1,
			},
		},
		ComputationUsed: 0,
		StateReads:      0,
		TrieUpdates: []*ledger.TrieUpdate{
			trieUpdate1,
			trieUpdate2,
			trieUpdate3,
			trieUpdate4,
		},
	}
}
