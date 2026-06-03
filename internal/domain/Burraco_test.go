//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// Burraco is a configured Canasta (UsePozzetto). These tests exercise the
// pozzetto-specific branches that the Canasta tests (UsePozzetto=false) do not.

func newTestBurraco() *domain.Burraco {
	players := []*domain.BurracoPlayer{
		domain.NewBurracoPlayer(true),
		domain.NewBurracoPlayer(false),
	}
	return domain.NewBurraco(domain.NewTrumpCardsWithDecks(2, 4), players, domain.DefaultBurracoConfig())
}

func TestNewDefaultBurraco(t *testing.T) {
	g := domain.NewDefaultBurraco()
	g.Reset()
	assert.True(t, g.GetConfig().UsePozzetto)
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.Equal(t, domain.BurracoHandSize, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 2, g.GetPozzettoCount())
}

func TestBurraco_CpuMeld_TakesPozzetto(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetPhase(domain.BurracoPhaseMeld)
	g.SetCurrentPlayerIdx(1) // CPU

	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.SetHasInitMeld(true)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))

	g.CpuPlay() // cpuMeld melds the three aces, emptying the hand → take pozzetto

	assert.True(t, cpu.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
}

func TestBurraco_CpuDiscard_EmptyHandTakesPozzetto(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetPhase(domain.BurracoPhaseDiscard)
	g.SetCurrentPlayerIdx(1) // CPU

	cpu := g.GetPlayer(1)
	cpu.Reset() // empty hand, pozzetto not yet taken

	g.CpuPlay() // cpuDiscard: empty hand + no pozzetto → take it, then discard

	assert.True(t, cpu.GetTookPozzetto())
}

func TestBurraco_CpuDiscard_EmptyHandAfterPozzettoAdvances(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetPhase(domain.BurracoPhaseDiscard)
	g.SetCurrentPlayerIdx(1) // CPU

	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.SetTookPozzetto(true) // already took pozzetto, no burraco, empty hand

	g.CpuPlay() // cpuDiscard: cannot discard/go out → advance turn

	assert.Equal(t, domain.BurracoPhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

func TestBurraco_DefaultConfig(t *testing.T) {
	cfg := domain.DefaultBurracoConfig()
	assert.True(t, cfg.UsePozzetto)
	assert.Equal(t, domain.BurracoDefaultPointLimit, cfg.PointLimit)
	assert.Equal(t, 2005, cfg.PointLimit)
}

func TestBurraco_Reset_DealsElevenPlusTwoPozzetti(t *testing.T) {
	g := newTestBurraco()
	g.Reset()

	for i := 0; i < 2; i++ {
		assert.Equal(t, domain.BurracoHandSize, g.GetPlayer(i).GetCardsSize(),
			"player %d should be dealt 11 cards", i)
	}
	assert.Equal(t, 2, g.GetPozzettoCount(), "two pozzetti are set aside")
	assert.Equal(t, 2*domain.BurracoPozzettoSize, g.GetPozzettoCardCount())

	// 108 = hands + red3s + discard + draw + pozzetti
	total := g.GetDrawPileCount() + g.GetDiscardPileCount() + g.GetPozzettoCardCount()
	for i := 0; i < 2; i++ {
		total += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetRed3s())
	}
	assert.Equal(t, 108, total)
}

func TestBurraco_TakePozzetto_OnMeldEmptyingHand(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetPhase(domain.BurracoPhaseMeld)
	g.SetCurrentPlayerIdx(0)
	require.Equal(t, 2, g.GetPozzettoCount())

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true) // bypass initial-meld minimum
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 1, false))

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))

	// Hand emptied → pozzetto taken (one pile consumed), player keeps playing
	assert.True(t, player.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
	assert.Greater(t, player.GetCardsSize(), 0)
	assert.Equal(t, domain.BurracoPhaseDiscard, g.GetPhase())
}

func TestBurraco_TakePozzetto_OnDiscardEmptyingHand(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetPhase(domain.BurracoPhaseDiscard)
	g.SetCurrentPlayerIdx(0)

	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false)) // last card

	require.NoError(t, g.PlayerDiscard(0))

	assert.True(t, player.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // turn advanced to opponent
}

func TestBurraco_GoOut_RequiresPozzetto(t *testing.T) {
	g := newTestBurraco()
	g.Reset()
	g.SetPhase(domain.BurracoPhaseDiscard)
	g.SetCurrentPlayerIdx(0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetMelds([]*domain.BurracoMeld{{
		Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
			domain.NewCard(domain.CardDesignDiamond, 7, false),
		},
		IsNatural: true,
	}})
	player.SetHasInitMeld(true)
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	// Has a burraco but has NOT taken the pozzetto → cannot go out
	assert.Error(t, g.PlayerGoOut())

	// After taking the pozzetto, going out succeeds
	player.SetTookPozzetto(true)
	require.NoError(t, g.PlayerGoOut())
	assert.True(t, g.GetPhase() == domain.BurracoPhaseRoundEnd || g.GetPhase() == domain.BurracoPhaseGameEnd)
}

func TestBurraco_JSON_RoundTrip_PreservesPozzetti(t *testing.T) {
	g := newTestBurraco()
	g.Reset()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Burraco
	require.NoError(t, json.Unmarshal(data, &g2))

	assert.Equal(t, g.GetPozzettoCount(), g2.GetPozzettoCount())
	assert.Equal(t, g.GetPozzettoCardCount(), g2.GetPozzettoCardCount())
	assert.True(t, g2.GetConfig().UsePozzetto)
}

func TestBurracoPlayer_TookPozzetto_JSON(t *testing.T) {
	p := domain.NewBurracoPlayer(true)
	p.SetTookPozzetto(true)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 domain.BurracoPlayer
	require.NoError(t, json.Unmarshal(data, &p2))
	assert.True(t, p2.GetTookPozzetto())
}

func TestBurracoMeld_IsBurraco(t *testing.T) {
	cards := make([]*domain.Card, 7)
	for i := range cards {
		cards[i] = domain.NewCard(domain.CardDesignSpade, 7, false)
	}
	m := &domain.BurracoMeld{Cards: cards, IsNatural: true}
	assert.True(t, m.IsBurraco())

	short := &domain.BurracoMeld{Cards: cards[:3], IsNatural: true}
	assert.False(t, short.IsBurraco())
}
