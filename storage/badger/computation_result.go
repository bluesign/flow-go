// (c) 2022 Dapper Labs - ALL RIGHTS RESERVED

package badger

import (
	"github.com/dgraph-io/badger/v2"

	"github.com/onflow/flow-go/engine/execution"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/storage/badger/operation"
)

type ComputationResults struct {
	db *badger.DB
}

func NewComputationResults(db *badger.DB) *ComputationResults {
	return &ComputationResults{
		db: db,
	}
}

func (c *ComputationResults) Store(computationResultID flow.Identifier,
	computationResult *execution.ComputationResult) error {
	return operation.RetryOnConflict(c.db.Update,
		operation.InsertComputationResult(computationResultID, computationResult))
}

func (c *ComputationResults) GetAllIDs() ([]flow.Identifier, error) {
	ids := make([]flow.Identifier, 0)
	err := c.db.View(operation.GetAllComputationResultIDs(&ids))
	return ids, err
}

func (c *ComputationResults) ByID(computationResultID flow.Identifier) (*execution.ComputationResult, error) {
	var ret execution.ComputationResult
	err := c.db.View(func(btx *badger.Txn) error {
		return operation.GetComputationResult(computationResultID, &ret)(btx)
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *ComputationResults) Remove(computationResultID flow.Identifier) error {
	return operation.RetryOnConflict(c.db.Update, func(btx *badger.Txn) error {
		return operation.RemoveComputationResult(computationResultID)(btx)
	})
}
