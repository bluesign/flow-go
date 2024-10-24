package utils

import (
	"fmt"
	"sync"

	"github.com/onflow/flow-go/fvm/state"
	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/model/flow"
)

// SimpleView provides a simple view for testing and migration purposes.
type SimpleView struct {
	Parent *SimpleView
	Ledger *MapLedger
}

func NewSimpleView() *SimpleView {
	return &SimpleView{
		Ledger: NewMapLedger(),
	}
}

func NewSimpleViewFromPayloads(payloads []ledger.Payload) *SimpleView {
	return &SimpleView{
		Ledger: NewMapLedgerFromPayloads(payloads),
	}
}

func (v *SimpleView) NewChild() state.View {
	ch := NewSimpleView()
	ch.Parent = v
	return ch
}

func (v *SimpleView) MergeView(o state.View) error {
	var other *SimpleView
	var ok bool
	if other, ok = o.(*SimpleView); !ok {
		return fmt.Errorf("can not merge: view type mismatch (given: %T, expected:SimpleView)", o)
	}

	for key, value := range other.Ledger.Registers {
		err := v.Ledger.Set(key, value)
		if err != nil {
			return fmt.Errorf("can not merge: %w", err)
		}
	}

	for k := range other.Ledger.RegisterTouches {
		v.Ledger.RegisterTouches[k] = struct{}{}
	}
	return nil
}

func (v *SimpleView) DropDelta() {
	v.Ledger.Registers = make(map[flow.RegisterID]flow.RegisterValue)
}

func (v *SimpleView) Set(id flow.RegisterID, value flow.RegisterValue) error {
	return v.Ledger.Set(id, value)
}

func (v *SimpleView) Get(id flow.RegisterID) (flow.RegisterValue, error) {
	value, err := v.Ledger.Get(id)
	if err != nil {
		return nil, err
	}
	if len(value) > 0 {
		return value, nil
	}

	if v.Parent != nil {
		return v.Parent.Get(id)
	}

	return nil, nil
}

// returns all the register ids that has been touched
func (v *SimpleView) AllRegisterIDs() []flow.RegisterID {
	res := make([]flow.RegisterID, 0, len(v.Ledger.RegisterTouches))
	for k := range v.Ledger.RegisterTouches {
		res = append(res, k)
	}
	return res
}

// returns all the register ids that has been updated
func (v *SimpleView) UpdatedRegisterIDs() []flow.RegisterID {
	res := make([]flow.RegisterID, 0, len(v.Ledger.RegisterUpdated))
	for k := range v.Ledger.RegisterUpdated {
		res = append(res, k)
	}
	return res
}

func (v *SimpleView) UpdatedRegisters() flow.RegisterEntries {
	entries := make(flow.RegisterEntries, 0, len(v.Ledger.RegisterUpdated))
	for key := range v.Ledger.RegisterUpdated {
		entries = append(
			entries,
			flow.RegisterEntry{
				Key:   key,
				Value: v.Ledger.Registers[key],
			})
	}
	return entries
}

func (v *SimpleView) Payloads() []ledger.Payload {
	return v.Ledger.Payloads()
}

// A MapLedger is a naive ledger storage implementation backed by a simple map.
//
// This implementation is designed for testing and migration purposes.
type MapLedger struct {
	sync.RWMutex
	Registers       map[flow.RegisterID]flow.RegisterValue
	RegisterTouches map[flow.RegisterID]struct{}
	RegisterUpdated map[flow.RegisterID]struct{}
}

// NewMapLedger returns an instance of map ledger (should only be used for
// testing and migration)
func NewMapLedger() *MapLedger {
	return &MapLedger{
		Registers:       make(map[flow.RegisterID]flow.RegisterValue),
		RegisterTouches: make(map[flow.RegisterID]struct{}),
		RegisterUpdated: make(map[flow.RegisterID]struct{}),
	}
}

// NewMapLedger returns an instance of map ledger with entries loaded from
// payloads (should only be used for testing and migration)
func NewMapLedgerFromPayloads(payloads []ledger.Payload) *MapLedger {
	ledger := NewMapLedger()
	for _, entry := range payloads {
		key, err := entry.Key()
		if err != nil {
			panic(err)
		}

		id := flow.NewRegisterID(
			string(key.KeyParts[0].Value),
			string(key.KeyParts[1].Value))

		ledger.Registers[id] = entry.Value()
	}

	return ledger
}

func (m *MapLedger) Set(id flow.RegisterID, value flow.RegisterValue) error {
	m.Lock()
	defer m.Unlock()

	m.RegisterTouches[id] = struct{}{}
	m.RegisterUpdated[id] = struct{}{}
	m.Registers[id] = value
	return nil
}

func (m *MapLedger) Get(id flow.RegisterID) (flow.RegisterValue, error) {
	m.Lock()
	defer m.Unlock()

	m.RegisterTouches[id] = struct{}{}
	return m.Registers[id], nil
}

func registerIdToLedgerKey(id flow.RegisterID) ledger.Key {
	keyParts := []ledger.KeyPart{
		ledger.NewKeyPart(0, []byte(id.Owner)),
		ledger.NewKeyPart(2, []byte(id.Key)),
	}

	return ledger.NewKey(keyParts)
}

func (m *MapLedger) Payloads() []ledger.Payload {
	m.RLock()
	defer m.RUnlock()

	ret := make([]ledger.Payload, 0, len(m.Registers))
	for id, val := range m.Registers {
		key := registerIdToLedgerKey(id)
		ret = append(ret, *ledger.NewPayload(key, ledger.Value(val)))
	}

	return ret
}
