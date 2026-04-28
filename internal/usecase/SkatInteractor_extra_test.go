//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// newSkatMocksLocal mirrors newSkatMocks; duplicated to keep this file
// independent of the original test file.
func newSkatMocksLocal() (*presenter.MockSkatPresenter, *interfaces.MockSkatGame) {
	sp := new(presenter.MockSkatPresenter)
	sp.On("Output", mock.Anything, mock.Anything).Return(skatMockOutput)
	g := new(interfaces.MockSkatGame)
	return sp, g
}

// TestSkatInteractor_GameEndShortCircuits exercises the gameEndFlag=true
// guards on every action method.
func TestSkatInteractor_GameEndShortCircuits(t *testing.T) {
	cases := []struct {
		name string
		act  func(si *usecase.SkatInteractor) string
	}{
		{"PickSkat", func(si *usecase.SkatInteractor) string { return si.PickSkat(true) }},
		{"Discard", func(si *usecase.SkatInteractor) string { return si.Discard(0, 1) }},
		{"DeclareGame", func(si *usecase.SkatInteractor) string {
			return si.DeclareGame(domain.SkatGameSuit, domain.CardDesignSpade)
		}},
		{"NextTrick", func(si *usecase.SkatInteractor) string { return si.NextTrick() }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sp, g := newSkatMocksLocal()
			g.On("GetGameEndFlag").Return(true)
			si := usecase.NewSkatInteractor(g, sp)
			assert.Equal(t, skatMockOutput, c.act(si))
			// No PlayerXxx call must have been made — the gameEnd guard short
			// circuits before reaching the delegated method.
			g.AssertNotCalled(t, "PlayerPickSkat", mock.Anything)
			g.AssertNotCalled(t, "PlayerDiscard", mock.Anything, mock.Anything)
			g.AssertNotCalled(t, "PlayerDeclareGame", mock.Anything, mock.Anything)
			g.AssertNotCalled(t, "NextTrick")
		})
	}
}

// TestSkatInteractor_NextRoundShortCircuitsWhenGameEnded covers the
// post-ScoreRound gameEnd path where NextRound() bails out before advancing.
func TestSkatInteractor_NextRoundShortCircuitsWhenGameEnded(t *testing.T) {
	sp, g := newSkatMocksLocal()
	g.On("ScoreRound").Return()
	g.On("GetGameEndFlag").Return(true)
	si := usecase.NewSkatInteractor(g, sp)
	assert.Equal(t, skatMockOutput, si.NextRound())
	g.AssertCalled(t, "ScoreRound")
	g.AssertNotCalled(t, "NextRound")
}

// TestSkatInteractor_RunCpuDeclarerPhases drives runCpuAutoPhases through a
// CPU declarer's full Pickup → Discard → Declaration → Play chain. Each
// GetPhase call returns the next phase in sequence to simulate real state
// progression.
func TestSkatInteractor_RunCpuDeclarerPhases(t *testing.T) {
	sp, g := newSkatMocksLocal()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)

	// Phase progression for runCpuAutoPhases:
	//   runCpuBids:          GetPhase=Bid (loop exits because human bid turn)
	//   runCpuDeclarerPhases: SkatPickup → Discard → Declaration → Play
	//   runCpuAutoPhases tail: GetPhase=Play
	//   runCpuTurns:          GetPhase=Play; IsHumanTurn=true (loop exits)
	g.On("GetPhase").Return(domain.SkatPhaseBid).Once()
	g.On("IsHumanBidTurn").Return(true).Once()
	g.On("GetPhase").Return(domain.SkatPhaseSkatPickup).Once()
	g.On("IsHumanDeclarerTurn").Return(false).Once()
	g.On("CpuPickSkat").Return().Once()
	g.On("GetPhase").Return(domain.SkatPhaseDiscard).Once()
	g.On("IsHumanDeclarerTurn").Return(false).Once()
	g.On("CpuDiscard").Return().Once()
	g.On("GetPhase").Return(domain.SkatPhaseGameDeclaration).Once()
	g.On("IsHumanDeclarerTurn").Return(false).Once()
	g.On("CpuDeclareGame").Return().Once()
	g.On("GetPhase").Return(domain.SkatPhasePlay).Once() // exits declarer loop
	g.On("GetPhase").Return(domain.SkatPhasePlay).Once() // runCpuAutoPhases tail check
	g.On("GetPhase").Return(domain.SkatPhasePlay)        // runCpuTurns loop reads
	g.On("IsHumanTurn").Return(true)                     // exits CPU turn loop immediately

	si := usecase.NewSkatInteractor(g, sp)
	si.Reset()

	g.AssertCalled(t, "CpuPickSkat")
	g.AssertCalled(t, "CpuDiscard")
	g.AssertCalled(t, "CpuDeclareGame")
}

// TestSkatInteractor_RunCpuDeclarerPhases_HumanDeclarer covers the early-
// return branches where IsHumanDeclarerTurn=true at each phase.
func TestSkatInteractor_RunCpuDeclarerPhases_HumanDeclarer(t *testing.T) {
	sp, g := newSkatMocksLocal()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SkatPhaseBid).Once()
	g.On("IsHumanBidTurn").Return(true).Once()
	g.On("GetPhase").Return(domain.SkatPhaseSkatPickup)
	g.On("IsHumanDeclarerTurn").Return(true)

	si := usecase.NewSkatInteractor(g, sp)
	si.Reset()
	g.AssertNotCalled(t, "CpuPickSkat")
}

// TestSkatInteractor_RunCpuDeclarerPhases_PassedOut verifies the default
// (non-declarer-phase) branch in runCpuDeclarerPhases — the loop exits
// when phase is something other than Pickup/Discard/Declaration.
func TestSkatInteractor_RunCpuDeclarerPhases_PassedOut(t *testing.T) {
	sp, g := newSkatMocksLocal()
	g.On("Reset").Return()
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.SkatPhaseBid).Once()
	g.On("IsHumanBidTurn").Return(true).Once()
	// runCpuDeclarerPhases enters with phase RoundEnd and falls through default.
	g.On("GetPhase").Return(domain.SkatPhaseRoundEnd).Once()
	g.On("GetPhase").Return(domain.SkatPhaseRoundEnd) // tail check in runCpuAutoPhases

	si := usecase.NewSkatInteractor(g, sp)
	si.Reset()
}
