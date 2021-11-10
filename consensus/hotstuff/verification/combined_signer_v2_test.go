package verification

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/consensus/hotstuff/mocks"
	"github.com/onflow/flow-go/consensus/hotstuff/model"
	"github.com/onflow/flow-go/model/encodable"
	"github.com/onflow/flow-go/model/encoding"
	"github.com/onflow/flow-go/module/local"
	modulemock "github.com/onflow/flow-go/module/mock"
	modulesig "github.com/onflow/flow-go/module/signature"
	storagemock "github.com/onflow/flow-go/storage/mock"
	"github.com/onflow/flow-go/utils/unittest"
)

// Test that when DKG key is available for a view, a signed block can pass the validation
// the sig include both staking sig and random beacon sig.
func TestCombinedSignWithDKGKey(t *testing.T) {
	// prepare data
	dkgKey := unittest.DKGParticipantPriv()
	pk := dkgKey.RandomBeaconPrivKey.PublicKey()
	signerID := dkgKey.NodeID
	view := uint64(20)

	fblock := unittest.BlockFixture()
	fblock.Header.ProposerID = signerID
	fblock.Header.View = view
	block := model.BlockFromFlow(fblock.Header, 10)

	epochCounter := uint64(3)
	epochLookup := &modulemock.EpochLookup{}
	epochLookup.On("EpochForViewWithFallback", view).Return(epochCounter, nil)

	keys := &storagemock.DKGKeys{}
	// there is DKG key for this epoch
	keys.On("RetrieveMyDKGPrivateInfo", epochCounter).Return(dkgKey, true, nil)

	beaconKeyStore := modulesig.NewEpochAwareRandomBeaconKeyStore(epochLookup, keys)

	stakingPriv := unittest.StakingPrivKeyFixture()
	nodeID := unittest.IdentityFixture()
	nodeID.NodeID = signerID
	nodeID.StakingPubKey = stakingPriv.PublicKey()

	me, err := local.New(nil, stakingPriv)
	require.NoError(t, err)
	staking := modulesig.NewSingleSigner(encoding.ConsensusVoteTag, me)
	signer := NewCombinedSignerV2(staking, beaconKeyStore, signerID)

	dkg := &mocks.DKG{}
	dkg.On("KeyShare", signerID).Return(pk, nil)

	committee := &mocks.Committee{}
	committee.On("DKG", mock.Anything).Return(dkg, nil)

	merger := modulesig.NewCombiner(encodable.ConsensusVoteSigLen, encodable.RandomBeaconSigLen)
	// TODO: to be replaced with factory methods that creates signer and verifier
	verifier := NewCombinedVerifierV2(committee, encoding.ConsensusVoteTag, encoding.RandomBeaconTag, merger)

	proposal, err := signer.CreateProposal(block)
	require.NoError(t, err)

	vote := proposal.ProposerVote()
	valid, err := verifier.VerifyVote(nodeID, vote.SigData, proposal.Block)
	require.NoError(t, err)
	require.Equal(t, true, valid)
}

// Test that when DKG key is not available for a view, a signed block can pass the validation
// the sig only include staking sig
func TestCombinedSignWithNoDKGKey(t *testing.T) {
	// prepare data
	dkgKey := unittest.DKGParticipantPriv()
	pk := dkgKey.RandomBeaconPrivKey.PublicKey()
	signerID := dkgKey.NodeID
	view := uint64(20)

	fblock := unittest.BlockFixture()
	fblock.Header.ProposerID = signerID
	fblock.Header.View = view
	block := model.BlockFromFlow(fblock.Header, 10)

	epochCounter := uint64(3)
	epochLookup := &modulemock.EpochLookup{}
	epochLookup.On("EpochForViewWithFallback", view).Return(epochCounter, nil)

	keys := &storagemock.DKGKeys{}
	// there is no DKG key for this epoch
	keys.On("RetrieveMyDKGPrivateInfo", epochCounter).Return(nil, false, nil)

	beaconKeyStore := modulesig.NewEpochAwareRandomBeaconKeyStore(epochLookup, keys)

	stakingPriv := unittest.StakingPrivKeyFixture()
	nodeID := unittest.IdentityFixture()
	nodeID.NodeID = signerID
	nodeID.StakingPubKey = stakingPriv.PublicKey()

	me, err := local.New(nil, stakingPriv)
	require.NoError(t, err)
	staking := modulesig.NewSingleSigner(encoding.ConsensusVoteTag, me)
	signer := NewCombinedSignerV2(staking, beaconKeyStore, signerID)

	dkg := &mocks.DKG{}
	dkg.On("KeyShare", signerID).Return(pk, nil)

	committee := &mocks.Committee{}
	committee.On("DKG", mock.Anything).Return(dkg, nil)

	merger := modulesig.NewCombiner(encodable.ConsensusVoteSigLen, encodable.RandomBeaconSigLen)
	// TODO: to be replaced with factory methods that creates signer and verifier
	verifier := NewCombinedVerifierV2(committee, encoding.ConsensusVoteTag, encoding.RandomBeaconTag, merger)

	proposal, err := signer.CreateProposal(block)
	require.NoError(t, err)

	vote := proposal.ProposerVote()
	valid, err := verifier.VerifyVote(nodeID, vote.SigData, proposal.Block)
	require.NoError(t, err)
	require.Equal(t, true, valid)
}
