// (c) 2022 Dapper Labs - ALL RIGHTS RESERVED

package operation

import (
	"github.com/dgraph-io/badger/v2"

	"github.com/onflow/flow-go/engine/execution"
	"github.com/onflow/flow-go/model/flow"
)

// InsertComputationResult addes given instance of ComputationResult into local BadgerDB.
func InsertComputationResult(computationResultID flow.Identifier,
	computationResult *execution.ComputationResult) func(*badger.Txn) error {
	return insert(makePrefix(codeComputationResults, computationResultID), computationResult)
}

// RemoveComputationResult removes an instance of ComputationResult with given ID.
func RemoveComputationResult(computationResultID flow.Identifier) func(*badger.Txn) error {
	return remove(makePrefix(codeComputationResults, computationResultID))
}

// GetComputationResult returns stored ComputationResult instance with given ID.
func GetComputationResult(computationResultID flow.Identifier,
	computationResult *execution.ComputationResult) func(*badger.Txn) error {
	return retrieve(makePrefix(codeComputationResults, computationResultID), computationResult)
}

// GetAllComputationResultIDs returns all IDs of stored ComputationResult instances.
func GetAllComputationResultIDs(computationResultIDs *[]flow.Identifier) func(*badger.Txn) error {
	return traverse(makePrefix(codeComputationResults), func() (checkFunc, createFunc, handleFunc) {
		check := func(key []byte) bool {
			return true
		}

		var val execution.ComputationResult
		create := func() interface{} {
			return &val
		}

		handle := func() error {
			if computationResultIDs != nil && val.ExecutableBlock != nil {
				// key format: code(1 byte) + all 32 bytes of stored flow.Identifier
				*computationResultIDs = append(*computationResultIDs, val.ExecutableBlock.ID())
			}
			return nil
		}
		return check, create, handle
	})
}
