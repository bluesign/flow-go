package pebble

import (
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/vmihailenco/msgpack"

	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/irrecoverable"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/storage"
)

type ChunkDataPacks struct {
	db             *pebble.DB
	collections    storage.Collections
	byChunkIDCache *Cache[flow.Identifier, *storage.StoredChunkDataPack]
}

var _ storage.ChunkDataPacks = (*ChunkDataPacks)(nil)

func NewChunkDataPacks(collector module.CacheMetrics, db *pebble.DB, collections storage.Collections, byChunkIDCacheSize uint) *ChunkDataPacks {

	retrieve := func(key flow.Identifier) func(pebble.Reader) (*storage.StoredChunkDataPack, error) {
		return func(r pebble.Reader) (*storage.StoredChunkDataPack, error) {
			var c storage.StoredChunkDataPack
			err := RetrieveChunkDataPack(key, &c)(r)
			return &c, err
		}
	}

	cache := newCache(collector, metrics.ResourceChunkDataPack,
		withLimit[flow.Identifier, *storage.StoredChunkDataPack](byChunkIDCacheSize),
		withRetrieve(retrieve),
	)

	return &ChunkDataPacks{
		db:             db,
		collections:    collections,
		byChunkIDCache: cache,
	}
}

func (ch *ChunkDataPacks) Store(cs []*flow.ChunkDataPack) error {
	batch := ch.db.NewBatch()
	defer batch.Close()

	scs := make([]*storage.StoredChunkDataPack, 0, len(cs))
	for _, c := range cs {
		sc, err := ch.batchStore(c, batch)
		if err != nil {
			return fmt.Errorf("cannot store chunk data pack: %w", err)
		}
		scs = append(scs, sc)
	}

	err := batch.Commit(pebble.Sync)
	if err != nil {
		return fmt.Errorf("cannot commit batch: %w", err)
	}

	// TODO: move to batchStore
	for _, sc := range scs {
		ch.byChunkIDCache.Insert(sc.ChunkID, sc)
	}

	return nil
}

func (ch *ChunkDataPacks) Remove(cs []flow.Identifier) error {
	batch := ch.db.NewBatch()

	for _, c := range cs {
		err := ch.batchRemove(c, batch)
		if err != nil {
			return fmt.Errorf("cannot remove chunk data pack: %w", err)
		}
	}

	err := batch.Commit(pebble.Sync)
	if err != nil {
		return fmt.Errorf("cannot commit batch: %w", err)
	}

	for _, c := range cs {
		ch.byChunkIDCache.Remove(c)
	}

	return nil
}

func (ch *ChunkDataPacks) ByChunkID(chunkID flow.Identifier) (*flow.ChunkDataPack, error) {
	var sc storage.StoredChunkDataPack
	err := RetrieveChunkDataPack(chunkID, &sc)(ch.db)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve stored chunk data pack: %w", err)
	}

	chdp := &flow.ChunkDataPack{
		ChunkID:           sc.ChunkID,
		StartState:        sc.StartState,
		Proof:             sc.Proof,
		Collection:        nil, // to be filled in later
		ExecutionDataRoot: sc.ExecutionDataRoot,
	}
	if !sc.SystemChunk {
		collection, err := ch.collections.ByID(sc.CollectionID)
		if err != nil {
			return nil, fmt.Errorf("could not retrive collection (id: %x) for stored chunk data pack: %w", sc.CollectionID, err)
		}

		chdp.Collection = collection
	}
	return chdp, nil
}

func (ch *ChunkDataPacks) BatchRemove(chunkID flow.Identifier, batch storage.BatchStorage) error {
	return fmt.Errorf("not implemented")
}

func (ch *ChunkDataPacks) batchRemove(chunkID flow.Identifier, batch pebble.Writer) error {
	return batch.Delete(makeKey(codeChunkDataPack, chunkID), nil)
}

func (ch *ChunkDataPacks) batchStore(c *flow.ChunkDataPack, batch *pebble.Batch) (*storage.StoredChunkDataPack, error) {
	sc := storage.ToStoredChunkDataPack(c)
	err := InsertChunkDataPack(sc)(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to store chunk data pack: %w", err)
	}
	return sc, nil
}

// TODO: move to operation package
func InsertChunkDataPack(sc *storage.StoredChunkDataPack) func(w pebble.Writer) error {
	key := makeKey(codeChunkDataPack, sc.ChunkID)
	return insert(key, sc)
}

func RetrieveChunkDataPack(chunkID flow.Identifier, sc *storage.StoredChunkDataPack) func(r pebble.Reader) error {
	key := makeKey(codeChunkDataPack, chunkID)
	return retrieve(key, sc)
}

func insert(key []byte, val interface{}) func(pebble.Writer) error {
	return func(w pebble.Writer) error {
		value, err := msgpack.Marshal(val)
		if err != nil {
			return irrecoverable.NewExceptionf("failed to encode value: %w", err)
		}

		err = w.Set(key, value, nil)
		if err != nil {
			return irrecoverable.NewExceptionf("failed to store data: %w", err)
		}

		return nil
	}
}

func retrieve(key []byte, sc interface{}) func(r pebble.Reader) error {
	return func(r pebble.Reader) error {
		val, closer, err := r.Get(key)
		if err != nil {
			return convertNotFoundError(err)
		}
		defer closer.Close()

		err = msgpack.Unmarshal(val, &sc)
		if err != nil {
			return irrecoverable.NewExceptionf("failed to decode value: %w", err)
		}
		return nil
	}
}

const (
	codeChunkDataPack = 100
)

func makeKey(prefix byte, chunkID flow.Identifier) []byte {
	return append([]byte{prefix}, chunkID[:]...)
}
