package storage

import (
	storageErrors "github.com/onflow/flow-go/storage/errors"
)

var (
	// Note: there is another not found error: badger.ErrKeyNotFound. The difference between
	// badger.ErrKeyNotFound and storage.ErrNotFound is that:
	// badger.ErrKeyNotFound is the error returned by the badger API.
	// Modules in storage/badger and storage/badger/operation package both
	// return storage.ErrNotFound for not found error
	ErrNotFound = storageErrors.ErrNotFound

	ErrAlreadyExists = storageErrors.ErrAlreadyExists
	ErrDataMismatch  = storageErrors.ErrDataMismatch
)
