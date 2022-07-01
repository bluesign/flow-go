// (c) 2022 Dapper Labs - ALL RIGHTS RESERVED

package storage

import (
	"github.com/onflow/flow-go/engine/execution"
	"github.com/onflow/flow-go/model/flow"
)

// ComputationResults interface defines storage operations for ComputationResult.
type ComputationResults interface {
	// Store inserts ComputationResult into persistent storage with given ID.
	Store(computationResultID flow.Identifier, computationResult *execution.ComputationResult) error

	// GetAllIDs returns all IDs of stored ComputationResult instances.
	GetAllIDs() ([]flow.Identifier, error)

	// ByID returns an instance of ComputationResult with given ID.
	ByID(computationResultID flow.Identifier) (*execution.ComputationResult, error)

	// Remove removes an instance of ComputationResult with given ID.
	Remove(computationResultID flow.Identifier) error
}
