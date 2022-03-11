package packer

import (
	"fmt"

	"github.com/onflow/flow-go/model/flow"
)

// EncodeSignerIdentifiersToIndices encodes the given signerIDs into compacted bit vector
func EncodeSignerIdentifiersToIndices(fullIdentities []flow.Identifier, signerIDs flow.IdentifierList) ([]byte, error) {
	signersLookup := signerIDs.Lookup()

	indices := make([]int, 0, len(fullIdentities))
	for i, member := range fullIdentities {
		if _, ok := signersLookup[member]; ok {
			indices = append(indices, i)
			delete(signersLookup, member)
		}
	}

	if len(signersLookup) > 0 {
		return nil, fmt.Errorf("unknown signers %v", signersLookup)
	}

	signerIndices, err := EncodeSignerIndices(indices, len(fullIdentities))
	if err != nil {
		return nil, err
	}

	return signerIndices, nil
}

// DecodeSignerIdentifiersFromIndices decodes the given compacted bit vector into signerIDs
func DecodeSignerIdentifiersFromIndices(fullIdentities []flow.Identifier, signerIndices []byte) ([]flow.Identifier, error) {
	indices, err := DecodeSignerIndices(signerIndices, len(fullIdentities))
	if err != nil {
		return nil, err
	}

	signerIDs := make([]flow.Identifier, 0, len(fullIdentities))
	for _, index := range indices {
		signerIDs = append(signerIDs, fullIdentities[index])
	}
	return signerIDs, nil
}