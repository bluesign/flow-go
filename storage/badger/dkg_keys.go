package badger

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/onflow/flow-go/model/dkg"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/storage/badger/operation"
)

type DKGKeys struct {
	db    *badger.DB
	cache *Cache
}

func NewDKGKeys(collector module.CacheMetrics, db *badger.DB) *DKGKeys {

	store := func(key interface{}, val interface{}) func(*badger.Txn) error {
		epochCounter := key.(uint64)
		info := val.(*dkg.DKGParticipantPriv)
		return operation.InsertMyDKGPrivateInfo(epochCounter, info)
	}

	retrieve := func(key interface{}) func(*badger.Txn) (interface{}, error) {
		epochCounter := key.(uint64)
		var info dkg.DKGParticipantPriv
		return func(tx *badger.Txn) (interface{}, error) {
			err := operation.RetrieveMyDKGPrivateInfo(epochCounter, &info)(tx)
			return &info, err
		}
	}

	k := &DKGKeys{
		db: db,
		cache: newCache(collector,
			withLimit(4*flow.DefaultTransactionExpiry),
			withStore(store),
			withRetrieve(retrieve),
			withResource(metrics.ResourceEpochSetup)),
	}

	return k
}

func (k *DKGKeys) storeTx(epochCounter uint64, info *dkg.DKGParticipantPriv) func(tx *badger.Txn) error {
	return k.cache.Put(epochCounter, info)
}

func (k *DKGKeys) retrieveTx(epochCounter uint64) func(tx *badger.Txn) (*dkg.DKGParticipantPriv, error) {
	return func(tx *badger.Txn) (*dkg.DKGParticipantPriv, error) {
		val, err := k.cache.Get(epochCounter)(tx)
		if err != nil {
			return nil, err
		}
		return val.(*dkg.DKGParticipantPriv), nil
	}
}

func (k *DKGKeys) InsertMyDKGPrivateInfo(epochCounter uint64, info *dkg.DKGParticipantPriv) error {
	return operation.RetryOnConflict(k.db.Update, k.storeTx(epochCounter, info))
}

func (k *DKGKeys) RetrieveMyDKGPrivateInfo(epochCounter uint64) (*dkg.DKGParticipantPriv, error) {
	tx := k.db.NewTransaction(false)
	defer tx.Discard()
	return k.retrieveTx(epochCounter)(tx)
}
