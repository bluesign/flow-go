package epochs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/onflow/flow-go/model/flow"
	mockstate "github.com/onflow/flow-go/state/protocol/mock"
	"github.com/onflow/flow-go/utils/unittest"
)

func TestEpochFallbackStateMachine(t *testing.T) {
	suite.Run(t, new(EpochFallbackStateMachineSuite))
}

// ProtocolStateMachineSuite is a dedicated test suite for testing happy path state machine.
type EpochFallbackStateMachineSuite struct {
	BaseStateMachineSuite
	params *mockstate.GlobalParams

	stateMachine *FallbackStateMachine
}

func (s *EpochFallbackStateMachineSuite) SetupTest() {
	s.BaseStateMachineSuite.SetupTest()

	s.params = mockstate.NewGlobalParams(s.T())
	s.params.On("EpochCommitSafetyThreshold").Return(uint64(200))
	s.parentProtocolState.InvalidEpochTransitionAttempted = true

	var err error
	s.stateMachine, err = NewFallbackStateMachine(s.params, s.candidate.View, s.parentProtocolState.Copy())
	require.NoError(s.T(), err)
}

// ProcessEpochSetupIsNoop ensures that processing epoch setup event is noop.
func (s *EpochFallbackStateMachineSuite) TestProcessEpochSetupIsNoop() {
	setup := unittest.EpochSetupFixture()
	applied, err := s.stateMachine.ProcessEpochSetup(setup)
	require.NoError(s.T(), err)
	require.False(s.T(), applied)
	updatedState, stateID, hasChanges := s.stateMachine.Build()
	require.False(s.T(), hasChanges)
	require.Equal(s.T(), s.parentProtocolState.ID(), updatedState.ID())
	require.Equal(s.T(), updatedState.ID(), stateID)
	require.Equal(s.T(), s.parentProtocolState.ID(), s.stateMachine.ParentState().ID())
}

// ProcessEpochCommitIsNoop ensures that processing epoch commit event is noop.
func (s *EpochFallbackStateMachineSuite) TestProcessEpochCommitIsNoop() {
	commit := unittest.EpochCommitFixture()
	applied, err := s.stateMachine.ProcessEpochCommit(commit)
	require.NoError(s.T(), err)
	require.False(s.T(), applied)
	updatedState, stateID, hasChanges := s.stateMachine.Build()
	require.False(s.T(), hasChanges)
	require.Equal(s.T(), s.parentProtocolState.ID(), updatedState.ID())
	require.Equal(s.T(), updatedState.ID(), stateID)
	require.Equal(s.T(), s.parentProtocolState.ID(), s.stateMachine.ParentState().ID())
}

// TestTransitionToNextEpoch ensures that transition to next epoch is not possible.
func (s *EpochFallbackStateMachineSuite) TestTransitionToNextEpoch() {
	err := s.stateMachine.TransitionToNextEpoch()
	require.NoError(s.T(), err)
	updatedState, updateStateID, hasChanges := s.stateMachine.Build()
	require.False(s.T(), hasChanges)
	require.Equal(s.T(), updatedState.ID(), updateStateID)
	require.Equal(s.T(), s.parentProtocolState.ID(), updateStateID)
}

// TestNewEpochFallbackStateMachine tests that creating epoch fallback state machine sets
// `InvalidEpochTransitionAttempted` to true to record that we have entered epoch fallback mode[EFM].
// It tests scenarios where the EFM is entered in different phases of the epoch,
// and verifies protocol-compliant addition of epoch extensions, depending on the candidate view and epoch phase.
func (s *EpochFallbackStateMachineSuite) TestNewEpochFallbackStateMachine() {
	parentProtocolState := s.parentProtocolState.Copy()
	parentProtocolState.InvalidEpochTransitionAttempted = false

	thresholdView := parentProtocolState.CurrentEpochSetup.FinalView - s.params.EpochCommitSafetyThreshold()

	// The view we enter EFM is in the staking phase. The resulting epoch state should be unchanged to the
	// parent state _except_ that `InvalidEpochTransitionAttempted` is set to true.
	// We expect no epoch extension to be added since we have not reached the threshold view.
	s.Run("threshold-not-reached", func() {
		candidateView := thresholdView - 1
		stateMachine, err := NewFallbackStateMachine(s.params, candidateView, parentProtocolState.Copy())
		require.NoError(s.T(), err)
		require.Equal(s.T(), parentProtocolState.ID(), stateMachine.ParentState().ID())
		require.Equal(s.T(), candidateView, stateMachine.View())

		updatedState, stateID, hasChanges := stateMachine.Build()
		require.True(s.T(), hasChanges, "InvalidEpochTransitionAttempted has to be updated")
		require.Equal(s.T(), updatedState.ID(), stateID)
		require.NotEqual(s.T(), parentProtocolState.ID(), stateID)

		expectedProtocolState := &flow.ProtocolStateEntry{
			PreviousEpoch: parentProtocolState.PreviousEpoch,
			CurrentEpoch: flow.EpochStateContainer{
				SetupID:          parentProtocolState.CurrentEpoch.SetupID,
				CommitID:         parentProtocolState.CurrentEpoch.CommitID,
				ActiveIdentities: parentProtocolState.CurrentEpoch.ActiveIdentities,
				EpochExtensions:  nil,
			},
			NextEpoch:                       nil,
			InvalidEpochTransitionAttempted: true,
		}
		require.Equal(s.T(), expectedProtocolState, updatedState, "state should be equal to expected one")
	})

	// The view we enter EFM is in the staking phase. The resulting epoch state should set `InvalidEpochTransitionAttempted` to true.
	// We expect an epoch extension to be added since we have reached the threshold view.
	s.Run("staking-phase", func() {
		stateMachine, err := NewFallbackStateMachine(s.params, thresholdView, parentProtocolState.Copy())
		require.NoError(s.T(), err)
		require.Equal(s.T(), parentProtocolState.ID(), stateMachine.ParentState().ID())
		require.Equal(s.T(), thresholdView, stateMachine.View())

		updatedState, stateID, hasChanges := stateMachine.Build()
		require.True(s.T(), hasChanges, "InvalidEpochTransitionAttempted has to be updated")
		require.Equal(s.T(), updatedState.ID(), stateID)
		require.NotEqual(s.T(), parentProtocolState.ID(), stateID)

		expectedProtocolState := &flow.ProtocolStateEntry{
			PreviousEpoch: parentProtocolState.PreviousEpoch,
			CurrentEpoch: flow.EpochStateContainer{
				SetupID:          parentProtocolState.CurrentEpoch.SetupID,
				CommitID:         parentProtocolState.CurrentEpoch.CommitID,
				ActiveIdentities: parentProtocolState.CurrentEpoch.ActiveIdentities,
				EpochExtensions: []flow.EpochExtension{
					{
						FirstView:     parentProtocolState.CurrentEpochFinalView() + 1,
						FinalView:     parentProtocolState.CurrentEpochFinalView() + DefaultEpochExtensionViewCount,
						TargetEndTime: 0,
					},
				},
			},
			NextEpoch:                       nil,
			InvalidEpochTransitionAttempted: true,
		}
		require.Equal(s.T(), expectedProtocolState, updatedState, "state should be equal to expected one")
	})

	// The view we enter EFM is in the epoch setup phase. This means that a SetupEvent for the next epoch is in the parent block's
	// protocol state. We expect an epoch extension to be added and the outdated information for the next epoch to be removed.
	s.Run("setup-phase", func() {
		parentProtocolState := parentProtocolState.Copy()
		// setup next epoch but without commit event
		unittest.WithNextEpochProtocolState()(parentProtocolState)
		parentProtocolState.NextEpoch.CommitID = flow.ZeroID
		parentProtocolState.NextEpochCommit = nil

		stateMachine, err := NewFallbackStateMachine(s.params, thresholdView, parentProtocolState)
		require.NoError(s.T(), err)
		require.Equal(s.T(), parentProtocolState.ID(), stateMachine.ParentState().ID())
		require.Equal(s.T(), thresholdView, stateMachine.View())

		updatedState, stateID, hasChanges := stateMachine.Build()
		require.True(s.T(), hasChanges, "InvalidEpochTransitionAttempted has to be updated")
		require.Nil(s.T(), updatedState.NextEpoch, "outdated information for the next epoch should have been removed")
		require.Equal(s.T(), updatedState.ID(), stateID)
		require.NotEqual(s.T(), parentProtocolState.ID(), stateID)

		expectedProtocolState := &flow.ProtocolStateEntry{
			PreviousEpoch: parentProtocolState.PreviousEpoch,
			CurrentEpoch: flow.EpochStateContainer{
				SetupID:          parentProtocolState.CurrentEpoch.SetupID,
				CommitID:         parentProtocolState.CurrentEpoch.CommitID,
				ActiveIdentities: parentProtocolState.CurrentEpoch.ActiveIdentities,
				EpochExtensions: []flow.EpochExtension{
					{
						FirstView:     parentProtocolState.CurrentEpochFinalView() + 1,
						FinalView:     parentProtocolState.CurrentEpochFinalView() + DefaultEpochExtensionViewCount,
						TargetEndTime: 0,
					},
				},
			},
			NextEpoch:                       nil,
			InvalidEpochTransitionAttempted: true,
		}
		require.Equal(s.T(), expectedProtocolState, updatedState, "state should be equal to expected one")
	})

	// If the next epoch has been committed, the extension shouldn't be added to the current epoch (verified below). Instead, the
	// extension should be added to the next epoch when **next** epoch reaches its safety threshold, which is covered in separate test.
	s.Run("commit-phase", func() {
		parentProtocolState := parentProtocolState.Copy()
		// setup next committed epoch
		unittest.WithNextEpochProtocolState()(parentProtocolState)

		// if the next epoch has been committed, the extension shouldn't be added to the current epoch
		// instead it will be added to the next epoch when **next** epoch reaches its safety threshold.

		stateMachine, err := NewFallbackStateMachine(s.params, thresholdView, parentProtocolState)
		require.NoError(s.T(), err)
		require.Equal(s.T(), parentProtocolState.ID(), stateMachine.ParentState().ID())
		require.Equal(s.T(), thresholdView, stateMachine.View())

		updatedState, stateID, hasChanges := stateMachine.Build()
		require.True(s.T(), hasChanges, "InvalidEpochTransitionAttempted has to be updated")
		require.Equal(s.T(), updatedState.ID(), stateID)
		require.NotEqual(s.T(), parentProtocolState.ID(), stateID)

		expectedProtocolState := &flow.ProtocolStateEntry{
			PreviousEpoch: parentProtocolState.PreviousEpoch,
			CurrentEpoch: flow.EpochStateContainer{
				SetupID:          parentProtocolState.CurrentEpoch.SetupID,
				CommitID:         parentProtocolState.CurrentEpoch.CommitID,
				ActiveIdentities: parentProtocolState.CurrentEpoch.ActiveIdentities,
				EpochExtensions:  nil,
			},
			NextEpoch:                       parentProtocolState.NextEpoch,
			InvalidEpochTransitionAttempted: true,
		}
		require.Equal(s.T(), expectedProtocolState, updatedState, "state should be equal to expected one")
	})
}

// TestEpochFallbackStateMachineInjectsMultipleExtensions tests that the state machine injects multiple extensions
// as it reaches the safety threshold of the current epoch and the extensions themselves.
// In this test, we are simulating the scenario where the current epoch enters EFM at a view when the next epoch has not been committed yet.
// When the next epoch has been committed the extension should be added to the next epoch, this is covered in separate test.
func (s *EpochFallbackStateMachineSuite) TestEpochFallbackStateMachineInjectsMultipleExtensions() {
	parentStateInStakingPhase := s.parentProtocolState.Copy()
	parentStateInStakingPhase.InvalidEpochTransitionAttempted = false

	parentStateInSetupPhase := parentStateInStakingPhase.Copy()
	unittest.WithNextEpochProtocolState()(parentStateInSetupPhase)
	parentStateInSetupPhase.NextEpoch.CommitID = flow.ZeroID
	parentStateInSetupPhase.NextEpochCommit = nil

	for _, originalParentState := range []*flow.RichProtocolStateEntry{parentStateInStakingPhase, parentStateInSetupPhase} {
		// In the previous test `TestNewEpochFallbackStateMachine`, we verified that the first extension is added correctly. Below we
		// test proper addition of the subsequent extension. A new extension should be added when we reach `firstExtensionViewThreshold`.
		// When reaching (equality) this threshold, the next extension should be added
		firstExtensionViewThreshold := originalParentState.CurrentEpochSetup.FinalView + DefaultEpochExtensionViewCount - s.params.EpochCommitSafetyThreshold()
		secondExtensionViewThreshold := originalParentState.CurrentEpochSetup.FinalView + 2*DefaultEpochExtensionViewCount - s.params.EpochCommitSafetyThreshold()
		// We progress through views that are strictly smaller than threshold. Up to this point, only the initial extension should exist

		// we will be asserting the validity of extensions after producing multiple extensions,
		// we expect 2 extensions to be added to the current epoch
		// 1 after we reach the commit threshold of the epoch and another one after reaching the threshold of the extension themselves
		firstExtension := flow.EpochExtension{
			FirstView:     originalParentState.CurrentEpochSetup.FinalView + 1,
			FinalView:     originalParentState.CurrentEpochSetup.FinalView + DefaultEpochExtensionViewCount,
			TargetEndTime: 0,
		}
		secondExtension := flow.EpochExtension{
			FirstView:     firstExtension.FinalView + 1,
			FinalView:     firstExtension.FinalView + DefaultEpochExtensionViewCount,
			TargetEndTime: 0,
		}

		parentProtocolState := originalParentState.Copy()
		candidateView := originalParentState.CurrentEpochSetup.FirstView + 1
		// an utility function to progress state to target view
		// updates variables that are defined in outer context
		evolveStateToView := func(targetView uint64) {
			for ; candidateView < targetView; candidateView++ {
				stateMachine, err := NewFallbackStateMachine(s.params, candidateView, parentProtocolState.Copy())
				require.NoError(s.T(), err)
				updatedState, _, _ := stateMachine.Build()

				parentProtocolState, err = flow.NewRichProtocolStateEntry(updatedState,
					parentProtocolState.PreviousEpochSetup,
					parentProtocolState.PreviousEpochCommit,
					parentProtocolState.CurrentEpochSetup,
					parentProtocolState.CurrentEpochCommit,
					nil, // as the next epoch was not yet committed, when we entered EFM, the state machine removed the outdated setup event
					nil)

				require.NoError(s.T(), err)
			}
		}

		type TestData struct {
			TargetView         uint64
			ExpectedExtensions []flow.EpochExtension
		}

		for _, data := range []TestData{
			{
				TargetView:         firstExtensionViewThreshold,
				ExpectedExtensions: []flow.EpochExtension{firstExtension},
			},
			{
				TargetView:         secondExtensionViewThreshold,
				ExpectedExtensions: []flow.EpochExtension{firstExtension, secondExtension},
			},
		} {
			evolveStateToView(data.TargetView)

			expectedState := &flow.ProtocolStateEntry{
				PreviousEpoch: originalParentState.PreviousEpoch,
				CurrentEpoch: flow.EpochStateContainer{
					SetupID:          originalParentState.CurrentEpoch.SetupID,
					CommitID:         originalParentState.CurrentEpoch.CommitID,
					ActiveIdentities: originalParentState.CurrentEpoch.ActiveIdentities,
					EpochExtensions:  data.ExpectedExtensions,
				},
				NextEpoch:                       nil,
				InvalidEpochTransitionAttempted: true,
			}
			require.Equal(s.T(), expectedState, parentProtocolState.ProtocolStateEntry)
			require.Greater(s.T(), parentProtocolState.CurrentEpochFinalView(), candidateView,
				"final view should be greater than final view of test")
		}
	}
}

// TestEpochFallbackStateMachineInjectsMultipleExtensions_NextEpochCommitted tests that the state machine injects multiple extensions
// as it reaches the safety threshold of the current epoch and the extensions themselves.
// In this test we are simulating the scenario where the current epoch enters fallback mode when the next epoch has been committed.
// It is expected that it will transition into the next epoch (since it was committed)
// then reach the safety threshold and add the extension to the next epoch, which at that point will be considered 'current'.
func (s *EpochFallbackStateMachineSuite) TestEpochFallbackStateMachineInjectsMultipleExtensions_NextEpochCommitted() {
	unittest.SkipUnless(s.T(), unittest.TEST_TODO,
		"This test doesn't work since it misses logic to transition to the next epoch when we are in EFM.")
	originalParentState := s.parentProtocolState.Copy()
	originalParentState.InvalidEpochTransitionAttempted = false
	unittest.WithNextEpochProtocolState()(originalParentState)

	// assert the validity of extensions after producing multiple extensions
	// we expect 3 extensions to be added to the current epoch
	// 1 after we reach the commit threshold of the epoch and two more after reaching the threshold of the extensions themselves
	firstExtension := flow.EpochExtension{
		FirstView:     originalParentState.NextEpochSetup.FinalView + 1,
		FinalView:     originalParentState.NextEpochSetup.FinalView + DefaultEpochExtensionViewCount,
		TargetEndTime: 0,
	}
	secondExtension := flow.EpochExtension{
		FirstView:     firstExtension.FinalView + 1,
		FinalView:     firstExtension.FinalView + DefaultEpochExtensionViewCount,
		TargetEndTime: 0,
	}
	thirdExtension := flow.EpochExtension{
		FirstView:     secondExtension.FinalView + 1,
		FinalView:     secondExtension.FinalView + DefaultEpochExtensionViewCount,
		TargetEndTime: 0,
	}

	// In the previous test `TestNewEpochFallbackStateMachine`, we verified that the first extension is added correctly. Below we
	// test proper addition of the subsequent extension. A new extension should be added when we reach `firstExtensionViewThreshold`.
	// When reaching (equality) this threshold, the next extension should be added
	firstExtensionViewThreshold := originalParentState.NextEpochSetup.FinalView + DefaultEpochExtensionViewCount - s.params.EpochCommitSafetyThreshold()
	secondExtensionViewThreshold := originalParentState.NextEpochSetup.FinalView + 2*DefaultEpochExtensionViewCount - s.params.EpochCommitSafetyThreshold()
	thirdExtensionViewThreshold := originalParentState.NextEpochSetup.FinalView + 3*DefaultEpochExtensionViewCount - s.params.EpochCommitSafetyThreshold()
	currentEpochLastView := originalParentState.CurrentEpochSetup.FinalView

	parentProtocolState := originalParentState.Copy()
	candidateView := originalParentState.CurrentEpochSetup.FirstView + 1
	// an utility function to progress state to target view
	// updates variables that are defined in outer context
	evolveStateToView := func(targetView uint64) {
		for ; candidateView < targetView; candidateView++ {
			stateMachine, err := NewFallbackStateMachine(s.params, candidateView, parentProtocolState.Copy())
			require.NoError(s.T(), err)

			var (
				nextEpochSetup  *flow.EpochSetup
				nextEpochCommit *flow.EpochCommit
			)
			if candidateView > parentProtocolState.CurrentEpochFinalView() {
				require.NoError(s.T(), stateMachine.TransitionToNextEpoch())
			}

			updatedState, _, _ := stateMachine.Build()

			if candidateView < currentEpochLastView {
				// as the next epoch has been committed, we need to respect that in current epoch
				// when we transition to the next epoch we will pass nil instead.
				nextEpochSetup = parentProtocolState.NextEpochSetup
				nextEpochCommit = parentProtocolState.NextEpochCommit
			}

			parentProtocolState, err = flow.NewRichProtocolStateEntry(updatedState,
				parentProtocolState.PreviousEpochSetup,
				parentProtocolState.PreviousEpochCommit,
				parentProtocolState.CurrentEpochSetup,
				parentProtocolState.CurrentEpochCommit,
				nextEpochSetup,
				nextEpochCommit)

			require.NoError(s.T(), err)
		}
	}

	type TestData struct {
		TargetView         uint64
		ExpectedExtensions []flow.EpochExtension
	}

	for _, data := range []TestData{
		{
			TargetView:         firstExtensionViewThreshold,
			ExpectedExtensions: []flow.EpochExtension{firstExtension},
		},
		{
			TargetView:         secondExtensionViewThreshold,
			ExpectedExtensions: []flow.EpochExtension{firstExtension, secondExtension},
		},
		{
			TargetView:         thirdExtensionViewThreshold,
			ExpectedExtensions: []flow.EpochExtension{firstExtension, secondExtension, thirdExtension},
		},
	} {
		evolveStateToView(data.TargetView)

		expectedState := &flow.ProtocolStateEntry{
			PreviousEpoch: originalParentState.PreviousEpoch,
			CurrentEpoch: flow.EpochStateContainer{
				SetupID:          originalParentState.CurrentEpoch.SetupID,
				CommitID:         originalParentState.CurrentEpoch.CommitID,
				ActiveIdentities: originalParentState.CurrentEpoch.ActiveIdentities,
				EpochExtensions:  data.ExpectedExtensions,
			},
			NextEpoch:                       nil,
			InvalidEpochTransitionAttempted: true,
		}
		require.Equal(s.T(), expectedState, parentProtocolState.ProtocolStateEntry)
		require.Greater(s.T(), parentProtocolState.CurrentEpochFinalView(), candidateView,
			"final view should be greater than final view of test")
	}

}
