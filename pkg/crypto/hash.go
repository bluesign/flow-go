package crypto

import (
	"fmt"
	"hash"

	"golang.org/x/crypto/sha3"
)

// NewHasher initializes and chooses a hashing algorithm
func NewHasher(name AlgoName) (Hasher, error) {
	if name == SHA3_256 {
		algo := &sha3_256Algo{
			commonHasher: &commonHasher{
				name:       name,
				outputSize: HashLengthSha3_256,
				Hash:       sha3.New256()}}

		// Output length sanity check, size() is provided by Hash.hash
		if algo.outputSize != algo.Size() {
			return nil, cryptoError{fmt.Sprintf("%s requires an output length %d", SHA3_256, algo.Size())}
		}
		return algo, nil
	}
	if name == SHA3_384 {
		algo := &sha3_384Algo{
			commonHasher: &commonHasher{
				name:       name,
				outputSize: HashLengthSha3_384,
				Hash:       sha3.New384()}}
		// Output length sanity check, size() is provided by Hash.hash
		if algo.outputSize != algo.Size() {
			return nil, cryptoError{fmt.Sprintf("%s requires an output length %d", SHA3_384, algo.Size())}
		}
		return algo, nil
	}
	return nil, cryptoError{fmt.Sprintf("the hashing algorithm %s is not supported.", name)}
}

// Hash is the hash algorithms output types
type Hash []byte

func (h Hash) IsEqual(input Hash) bool {
	if len(h) != len(input) {
		return false
	}
	for i := 0; i < len(h); i++ {
		if h[i] != input[i] {
			return false
		}
	}
	return true
}

// Hasher interface

type Hasher interface {
	Name() AlgoName
	// Size returns the hash output length
	Size() int
	// ComputeHash returns the hash output
	ComputeHash([]byte) Hash
	// Adds more bytes to the current hash state (a Hash.hash method)
	Add([]byte)
	// SumHash returns the hash output and resets the hash state a
	SumHash() Hash
	// Reset resets the hash state
	Reset()
}

// commonHasher holds the common data for all hashers
type commonHasher struct {
	name       AlgoName
	outputSize int
	hash.Hash
}

// Name returns the name of the algorithm
func (a *commonHasher) Name() AlgoName {
	return a.name
}

// Name returns the size of the output
func (a *commonHasher) Size() int {
	return a.outputSize
}

// AddBytes adds bytes to the current hash state
func (a *commonHasher) Add(data []byte) {
	a.Write(data)
}

func BytesToHash(b []byte) Hash {
	h := make([]byte, len(b))
	copy(h, b)
	return h
}

// HashesToBytes converts a slice of hashes to a slice of byte slices.
func HashesToBytes(hashes []Hash) [][]byte {
	b := make([][]byte, len(hashes))

	for i, h := range hashes {
		b[i] = h
	}

	return b
}
