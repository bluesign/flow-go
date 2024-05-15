package migrations

import (
	"github.com/onflow/cadence/runtime/common"

	"github.com/onflow/flow-go/cmd/util/ledger/util/registers"
	"github.com/onflow/flow-go/fvm/meter"
	snapshot2 "github.com/onflow/flow-go/fvm/storage/snapshot"
	"github.com/onflow/flow-go/fvm/storage/state"
	"github.com/onflow/flow-go/model/flow"
)

type TransactionStateProxy struct {
	object   state.NestedTransactionPreparer
	snapshot registers.StorageSnapshot
	changes  map[flow.RegisterID]flow.RegisterValue
}

func NewTransactionStateProxy(object state.NestedTransactionPreparer, snapshot registers.StorageSnapshot) state.NestedTransactionPreparer {
	return &TransactionStateProxy{
		object:   object,
		snapshot: snapshot,
		changes:  make(map[flow.RegisterID]flow.RegisterValue),
	}
}

func (t TransactionStateProxy) MeterComputation(kind common.ComputationKind, intensity uint) error {
	return nil
}

func (t TransactionStateProxy) ComputationAvailable(kind common.ComputationKind, intensity uint) bool {
	return true
}

func (t TransactionStateProxy) ComputationIntensities() meter.MeteredComputationIntensities {
	return t.object.ComputationIntensities()
}

func (t TransactionStateProxy) TotalComputationLimit() uint {
	return (1 << 64) - 1
}

func (t TransactionStateProxy) TotalComputationUsed() uint64 {
	return 0
}

func (t TransactionStateProxy) MeterMemory(kind common.MemoryKind, intensity uint) error {
	return nil
}

func (t TransactionStateProxy) MemoryIntensities() meter.MeteredMemoryIntensities {
	return t.object.MemoryIntensities()
}

func (t TransactionStateProxy) TotalMemoryEstimate() uint64 {
	return 0
}

func (t TransactionStateProxy) InteractionUsed() uint64 {
	return 0
}

func (t TransactionStateProxy) MeterEmittedEvent(byteSize uint64) error {
	return nil
}

func (t TransactionStateProxy) TotalEmittedEventBytes() uint64 {
	return 0
}

func (t TransactionStateProxy) RunWithAllLimitsDisabled(f func()) {
	t.object.RunWithAllLimitsDisabled(f)
}

func (t TransactionStateProxy) NumNestedTransactions() int {
	return t.object.NumNestedTransactions()
}

func (t TransactionStateProxy) IsParseRestricted() bool {
	return t.object.IsParseRestricted()
}

func (t TransactionStateProxy) MainTransactionId() state.NestedTransactionId {
	return t.object.MainTransactionId()
}

func (t TransactionStateProxy) IsCurrent(id state.NestedTransactionId) bool {
	return t.object.IsCurrent(id)
}

func (t TransactionStateProxy) InterimReadSet() map[flow.RegisterID]struct{} {
	return t.object.InterimReadSet()
}

func (t TransactionStateProxy) FinalizeMainTransaction() (*snapshot2.ExecutionSnapshot, error) {
	return &snapshot2.ExecutionSnapshot{
		ReadSet:     nil,
		WriteSet:    t.changes,
		SpockSecret: nil,
		Meter:       nil,
	}, nil
}

func (t TransactionStateProxy) BeginNestedTransaction() (state.NestedTransactionId, error) {
	return t.object.BeginNestedTransaction()
}

func (t TransactionStateProxy) BeginNestedTransactionWithMeterParams(params meter.MeterParameters) (state.NestedTransactionId, error) {
	return t.object.BeginNestedTransactionWithMeterParams(params)
}

func (t TransactionStateProxy) BeginParseRestrictedNestedTransaction(location common.AddressLocation) (state.NestedTransactionId, error) {
	return t.object.BeginParseRestrictedNestedTransaction(location)
}

func (t TransactionStateProxy) CommitNestedTransaction(expectedId state.NestedTransactionId) (*snapshot2.ExecutionSnapshot, error) {
	return t.object.CommitNestedTransaction(expectedId)
}

func (t TransactionStateProxy) CommitParseRestrictedNestedTransaction(location common.AddressLocation) (*snapshot2.ExecutionSnapshot, error) {
	return t.object.CommitParseRestrictedNestedTransaction(location)
}

func (t TransactionStateProxy) AttachAndCommitNestedTransaction(cachedSnapshot *snapshot2.ExecutionSnapshot) error {
	return t.object.AttachAndCommitNestedTransaction(cachedSnapshot)
}

func (t TransactionStateProxy) RestartNestedTransaction(id state.NestedTransactionId) error {
	return t.object.RestartNestedTransaction(id)
}

func (t TransactionStateProxy) Get(id flow.RegisterID) (flow.RegisterValue, error) {
	if v, exist := t.changes[id]; exist {
		return v, nil
	}

	return t.snapshot.Get(id)
}

func (t TransactionStateProxy) Set(id flow.RegisterID, value flow.RegisterValue) error {
	t.changes[id] = value
	return nil
}

var _ state.NestedTransactionPreparer = &TransactionStateProxy{}
