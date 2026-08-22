//go:build test && (!js || !wasm || extra)

package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubSpeculationPresenter is a minimal presenter for snapshot tests.
type stubSpeculationPresenter struct{}

func (s *stubSpeculationPresenter) Output(_ interfaces.SpeculationGame, _ error) string { return `{}` }
func (s *stubSpeculationPresenter) ActionLogOutput(_ interfaces.SpeculationGame) string { return `{}` }
func (s *stubSpeculationPresenter) HintOutput(_ interfaces.SpeculationGame) string      { return `{}` }

// speculationSnapshotTable builds a table whose every persisted field carries a
// value no freshly dealt table would have, so a restore that ignores its input
// cannot accidentally reproduce it. Nothing here depends on the shuffle.
func speculationSnapshotTable(t *testing.T) *domain.Speculation {
	t.Helper()

	cfg := domain.NewDefaultSpeculationConfig()
	cfg.Players = 4
	cfg.Rounds = 7
	cfg.Stake = 15
	cfg.InitialChips = 300
	g := domain.NewSpeculation(cfg)

	seats := g.GetPlayers()
	require.Len(t, seats, 4)

	// Distinct, non-default chip counts: a fresh deal would leave every seat on
	// InitialChips-Stake, so four different values can only come from the bytes.
	seats[0].SetChips(123)
	seats[1].SetChips(45)
	seats[2].SetChips(678)
	seats[3].SetChips(9)

	// Distinct face-down counts, including an exhausted seat.
	seats[0].SetHidden([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignClover, 8, true),
		domain.NewCard(domain.CardDesignHeart, 11, true),
	})
	seats[1].SetHidden([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignSpade, 13, true),
	})
	seats[2].SetHidden([]*domain.Card{domain.NewCard(domain.CardDesignClover, 6, true)})
	seats[3].SetHidden(nil)

	// Seat 0 holds the best trump; a heart Queen.
	seats[0].SetBest(domain.NewCard(domain.CardDesignHeart, 12, true))
	seats[1].SetBest(nil)
	seats[2].SetBest(nil)
	seats[3].SetBest(nil)

	g.SetPhase(domain.SpeculationPhaseAuction)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPot(77)
	g.SetTurnSeat(3)
	g.SetBestSeat(0)
	g.SetOffer(2, 0, 34)
	g.SetRoundNo(5)
	return g
}

func TestSpeculationInteractorSnapshotRestore(t *testing.T) {
	g := speculationSnapshotTable(t)
	trumpCard := g.GetTrumpCard()
	logLen := len(g.GetActionLog())

	si := NewSpeculationInteractor(g, new(stubSpeculationPresenter))
	data, err := si.Snapshot()
	require.NoError(t, err)

	// **A struct of only unexported fields marshals to `{}` with no error.**
	// Without MarshalJSON the snapshot is two bytes and the Worker silently
	// rebuilds the table on every request, so pin the shape before the values.
	require.NotEqual(t, "{}", string(data), "snapshot lost every field; MarshalJSON is missing")
	require.Greater(t, len(data), 2)

	restored, err := RestoreSpeculationInteractor(data, new(stubSpeculationPresenter))
	require.NoError(t, err)
	require.NotNil(t, restored)

	got := restored.Game

	// Scalars that only survive if the bytes were actually decoded.
	assert.Equal(t, domain.SpeculationPhaseAuction, got.GetPhase())
	assert.Equal(t, 77, got.GetPot())
	assert.Equal(t, domain.CardDesignHeart, got.GetTrumpSuit())
	assert.Equal(t, 3, got.GetTurnSeat())
	assert.Equal(t, 0, got.GetBestSeat())
	assert.Equal(t, 5, got.GetRoundNo())

	// The offer is three separate fields; losing any one of them turns a live
	// auction into an unanswerable one.
	assert.Equal(t, 2, got.GetOfferFrom())
	assert.Equal(t, 0, got.GetOfferTo())
	assert.Equal(t, 34, got.GetOfferAmount())

	// Config.
	cfg := got.GetConfig()
	assert.Equal(t, 4, cfg.Players)
	assert.Equal(t, 7, cfg.Rounds)
	assert.Equal(t, 15, cfg.Stake)
	assert.Equal(t, 300, cfg.InitialChips)

	// The trump-fixing card itself, compared against what was actually dealt.
	require.NotNil(t, got.GetTrumpCard())
	require.NotNil(t, trumpCard)
	assert.Equal(t, trumpCard.GetDesign(), got.GetTrumpCard().GetDesign())
	assert.Equal(t, trumpCard.GetValue(), got.GetTrumpCard().GetValue())

	// Per-seat chips and face-down counts.
	seats := got.GetPlayers()
	require.Len(t, seats, 4)
	assert.Equal(t, []int{123, 45, 678, 9},
		[]int{seats[0].GetChips(), seats[1].GetChips(), seats[2].GetChips(), seats[3].GetChips()})
	assert.Equal(t, []int{3, 2, 1, 0},
		[]int{seats[0].GetHiddenCount(), seats[1].GetHiddenCount(), seats[2].GetHiddenCount(), seats[3].GetHiddenCount()})
	assert.Equal(t, []string{"You", "CPU1", "CPU2", "CPU3"},
		[]string{seats[0].GetName(), seats[1].GetName(), seats[2].GetName(), seats[3].GetName()})

	// The face-down cards themselves, in order -- the next Flip must turn up
	// the same card the pre-snapshot table would have.
	require.Len(t, seats[0].GetHidden(), 3)
	assert.Equal(t, domain.CardDesignSpade, seats[0].GetHidden()[0].GetDesign())
	assert.Equal(t, 4, seats[0].GetHidden()[0].GetValue())
	assert.Equal(t, domain.CardDesignHeart, seats[0].GetHidden()[2].GetDesign())
	assert.Equal(t, 11, seats[0].GetHidden()[2].GetValue())

	// The best trump, and only on the seat that held it.
	require.NotNil(t, seats[0].GetBest())
	assert.Equal(t, domain.CardDesignHeart, seats[0].GetBest().GetDesign())
	assert.Equal(t, 12, seats[0].GetBest().GetValue())
	assert.Nil(t, seats[1].GetBest())
	assert.Nil(t, seats[2].GetBest())
	assert.Nil(t, seats[3].GetBest())

	// The action log survives too.
	assert.Len(t, got.GetActionLog(), logLen)
	require.NotEmpty(t, got.GetActionLog())
	assert.Equal(t, "deal", got.GetActionLog()[0].ActionType)
}

// TestSpeculationInteractorSnapshotRestorePlayable proves the restored table is
// not merely populated but still runnable: declining the restored auction moves
// the game on exactly as it would have before the round trip.
func TestSpeculationInteractorSnapshotRestorePlayable(t *testing.T) {
	g := speculationSnapshotTable(t)
	si := NewSpeculationInteractor(g, new(stubSpeculationPresenter))

	data, err := si.Snapshot()
	require.NoError(t, err)

	restored, err := RestoreSpeculationInteractor(data, new(stubSpeculationPresenter))
	require.NoError(t, err)

	restored.Decline()

	assert.Equal(t, domain.SpeculationPhaseFlip, restored.Game.GetPhase())
	assert.Equal(t, -1, restored.Game.GetOfferFrom())
	assert.Equal(t, -1, restored.Game.GetOfferTo())
	assert.Equal(t, 0, restored.Game.GetOfferAmount())
	// Declining moves no chips.
	assert.Equal(t, 123, restored.Game.GetPlayers()[0].GetChips())
	assert.Equal(t, 678, restored.Game.GetPlayers()[2].GetChips())
}

func TestSpeculationInteractorRestoreRejectsGarbage(t *testing.T) {
	_, err := RestoreSpeculationInteractor([]byte("{not json"), new(stubSpeculationPresenter))
	assert.Error(t, err)
}
