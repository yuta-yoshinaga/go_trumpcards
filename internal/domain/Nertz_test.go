//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// nertzGameForTest creates a Nertz with default 4-player config.
func nertzGameForTest(t *testing.T) *domain.Nertz {
	t.Helper()
	g := domain.NewDefaultNertz()
	g.Reset()
	return g
}

func TestNewDefaultNertz(t *testing.T) {
	g := domain.NewDefaultNertz()
	require.NotNil(t, g)
	assert.Equal(t, domain.NertzPhaseIdle, g.GetPhase())
}

func TestNertz_ResetInitialState(t *testing.T) {
	g := nertzGameForTest(t)

	assert.Equal(t, domain.NertzPhasePlaying, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNo())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetMatchWinner())

	cfg := g.GetConfig()
	assert.Equal(t, cfg.PlayerCount, len(g.GetPlayers()))
	for _, p := range g.GetPlayers() {
		assert.Equal(t, domain.NertzPileSize, p.NertzSize())
		assert.Equal(t, domain.NertzInitialStockSize, p.StockSize())
		assert.Equal(t, 0, p.WasteSize())
		for c := 0; c < domain.NertzTableauCnt; c++ {
			assert.Equal(t, 1, p.TableauSize(c))
		}
	}
	// foundations: fixed 4 * playerCount empty slots
	assert.Equal(t, 4*cfg.PlayerCount, len(g.GetFoundations()))
	for _, f := range g.GetFoundations() {
		assert.True(t, f.IsEmpty())
	}
}

func TestNertz_ResetWithConfig(t *testing.T) {
	g := domain.NewDefaultNertz()
	cfg := domain.DefaultNertzConfig()
	cfg.PlayerCount = 3
	cfg.DrawCount = 1
	g.ResetWithConfig(cfg)

	assert.Equal(t, 3, len(g.GetPlayers()))
	assert.Equal(t, 12, len(g.GetFoundations()))
	assert.Equal(t, 1, g.GetConfig().DrawCount)
}

func TestNertz_ResetWithInvalidConfigFallsBackToDefault(t *testing.T) {
	g := domain.NewDefaultNertz()
	bad := domain.NertzConfig{PlayerCount: 99, DrawCount: 7, TargetScore: 0}
	g.ResetWithConfig(bad)
	assert.Equal(t, domain.NertzPlayerCntDefault, g.GetConfig().PlayerCount)
}

// --- Helpers for board manipulation ---

func clearFoundations(g *domain.Nertz) {
	founds := g.GetFoundations()
	for i := range founds {
		founds[i] = domain.NewNertzFoundation()
	}
	g.SetFoundations(founds)
}

func clearPlayerPiles(p *domain.NertzPlayer) {
	p.ResetRoundPiles()
}

func TestNertz_DrawStock_DefaultDraw3(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	stockBefore := p.StockSize()
	require.NoError(t, g.DrawStock(0))
	assert.Equal(t, stockBefore-3, p.StockSize())
	assert.Equal(t, 3, p.WasteSize())
}

func TestNertz_DrawStock_RecyclesWhenStockEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	// Set up: stock=empty, waste=2 cards
	p.PushWaste(newNertzCard(domain.CardDesignSpade, 5))
	p.PushWaste(newNertzCard(domain.CardDesignHeart, 9))

	require.NoError(t, g.DrawStock(0))
	assert.Equal(t, 0, p.WasteSize(), "all waste recycled into stock")
	assert.Equal(t, 2, p.StockSize())
}

func TestNertz_DrawStock_FailsWhenBothEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	err := g.DrawStock(0)
	assert.Error(t, err)
}

func TestNertz_DrawStock_RejectsInvalidPlayer(t *testing.T) {
	g := nertzGameForTest(t)
	assert.Error(t, g.DrawStock(-1))
	assert.Error(t, g.DrawStock(99))
}

func TestNertz_MoveNertzToFoundation_OpensNewFoundationWithAce(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	ace := newNertzCard(domain.CardDesignSpade, 1)
	p.PushNertz(ace)

	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	assert.Equal(t, 0, p.NertzSize())
	founds := g.GetFoundations()
	assert.Equal(t, 1, founds[0].Size())
	assert.Equal(t, domain.CardDesignSpade, founds[0].Suit())
}

func TestNertz_MoveNertzToFoundation_RejectsInvalid(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)
	p.PushNertz(newNertzCard(domain.CardDesignSpade, 5))
	// Foundation is empty; only Ace allowed
	err := g.MoveNertzToFoundation(0, 0)
	assert.Error(t, err)
	assert.Equal(t, 1, p.NertzSize())
}

func TestNertz_MoveNertzToFoundation_RoundEndOnEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)
	p.PushNertz(newNertzCard(domain.CardDesignDiamond, 1))

	require.NoError(t, g.MoveNertzToFoundation(0, 0))
	assert.Equal(t, domain.NertzPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestNertz_MoveNertzToTableau_RequiresValidPlacement(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	// Put a black K on tableau col 0 so a red Q can land
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	p.PushNertz(newNertzCard(domain.CardDesignHeart, 12))

	require.NoError(t, g.MoveNertzToTableau(0, 0))
	assert.Equal(t, 0, p.NertzSize())
	assert.Equal(t, 2, p.TableauSize(0))
}

func TestNertz_MoveNertzToTableau_RejectsBadColor(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	p.PushNertz(newNertzCard(domain.CardDesignClover, 12)) // black-on-black rejected

	err := g.MoveNertzToTableau(0, 0)
	assert.Error(t, err)
}

func TestNertz_MoveNertzToTableau_EmptyColumnOnlyAcceptsTopCard(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	// Empty tableau col 0 accepts ANY card from Nertz pile (Nertz has free placement on empty cols)
	p.PushNertz(newNertzCard(domain.CardDesignHeart, 7))
	require.NoError(t, g.MoveNertzToTableau(0, 0))
	assert.Equal(t, 1, p.TableauSize(0))
}

func TestNertz_MoveWasteToFoundation_NeedsAceWhenEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)
	p.PushWaste(newNertzCard(domain.CardDesignSpade, 1))

	require.NoError(t, g.MoveWasteToFoundation(0, 0))
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 1, g.GetFoundations()[0].Size())
}

func TestNertz_MoveWasteToFoundation_ContinuesSuit(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	founds := g.GetFoundations()
	require.NoError(t, founds[2].Push(newNertzCard(domain.CardDesignDiamond, 1), 1))
	g.SetFoundations(founds)
	p.PushWaste(newNertzCard(domain.CardDesignDiamond, 2))

	require.NoError(t, g.MoveWasteToFoundation(0, 2))
	assert.Equal(t, 2, g.GetFoundations()[2].Size())
	assert.Equal(t, 1, g.GetFoundations()[2].CountByContributor(0))
}

func TestNertz_MoveWasteToTableau(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)
	p.PushTableau(1, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignClover, 8), FaceUp: true})
	p.PushWaste(newNertzCard(domain.CardDesignHeart, 7))

	require.NoError(t, g.MoveWasteToTableau(0, 1))
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 2, p.TableauSize(1))
}

func TestNertz_MoveTableauToFoundation_AutoFillsFromNertz(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	// Foundation 3 has Heart-A; we'll play Heart-2 from a single-card tableau col
	founds := g.GetFoundations()
	require.NoError(t, founds[3].Push(newNertzCard(domain.CardDesignHeart, 1), 0))
	g.SetFoundations(founds)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 2), FaceUp: true})

	// Nertz pile has a card to auto-fill the empty column with
	autoCard := newNertzCard(domain.CardDesignSpade, 9)
	p.PushNertz(autoCard)

	require.NoError(t, g.MoveTableauToFoundation(0, 0, 3))
	assert.Equal(t, 1, p.TableauSize(0), "column auto-filled from Nertz pile")
	assert.Equal(t, autoCard, p.TableauTop(0))
	assert.Equal(t, 0, p.NertzSize())
}

func TestNertz_MoveTableauToFoundation_NoAutoFillWhenNertzEmpty(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearFoundations(g)
	clearPlayerPiles(p)

	founds := g.GetFoundations()
	require.NoError(t, founds[3].Push(newNertzCard(domain.CardDesignHeart, 1), 0))
	g.SetFoundations(founds)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 2), FaceUp: true})

	require.NoError(t, g.MoveTableauToFoundation(0, 0, 3))
	assert.Equal(t, 0, p.TableauSize(0))
}

func TestNertz_MoveTableauToTableau_Substack(t *testing.T) {
	g := nertzGameForTest(t)
	p := g.GetPlayers()[0]
	clearPlayerPiles(p)

	// Source col 0: Spade-K, Heart-Q, Clover-J (alternating colors descending)
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignSpade, 13), FaceUp: true})
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 12), FaceUp: true})
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignClover, 11), FaceUp: true})
	// Destination col 1: empty (any card accepted)
	require.NoError(t, g.MoveTableauToTableau(0, 0, 1, 1))
	// Take starting at idx 1 = Heart-Q + Clover-J (2 cards)
	assert.Equal(t, 1, p.TableauSize(0))
	assert.Equal(t, 2, p.TableauSize(1))
}

func TestNertz_MoveTableauToTableau_RejectsSameColumn(t *testing.T) {
	g := nertzGameForTest(t)
	err := g.MoveTableauToTableau(0, 0, 0, 0)
	assert.Error(t, err)
}

func TestNertz_RejectsActionsWhenNotPlaying(t *testing.T) {
	g := nertzGameForTest(t)
	g.SetPhase(domain.NertzPhaseRoundEnd)
	assert.Error(t, g.DrawStock(0))
	assert.Error(t, g.MoveNertzToFoundation(0, 0))
}

func TestNertz_JSON(t *testing.T) {
	g := nertzGameForTest(t)
	// take an action so the state is interesting
	_ = g.DrawStock(0)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	restored := domain.NewDefaultNertz()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, len(g.GetPlayers()), len(restored.GetPlayers()))
	assert.Equal(t, len(g.GetFoundations()), len(restored.GetFoundations()))
	assert.Equal(t, g.GetRoundNo(), restored.GetRoundNo())
}
