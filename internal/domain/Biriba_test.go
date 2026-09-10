//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// Biriba is a configured Canasta (UsePozzetto). These tests exercise the
// pozzetto-specific branches that the Canasta tests (UsePozzetto=false) do not.

func newTestBiriba() *domain.Biriba {
	players := []*domain.BiribaPlayer{
		domain.NewBiribaPlayer(true),
		domain.NewBiribaPlayer(false),
	}
	return domain.NewBiriba(domain.NewTrumpCardsWithDecks(2, 4), players, domain.DefaultBiribaConfig())
}

func TestNewDefaultBiriba(t *testing.T) {
	g := domain.NewDefaultBiriba()
	g.Reset()
	assert.True(t, g.GetConfig().UsePozzetto)
	assert.True(t, g.GetConfig().UseBiriba)
	assert.Equal(t, 2, g.GetPlayerCnt())
	assert.Equal(t, domain.BiribaHandSize, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 2, g.GetPozzettoCount())
}

func TestBiriba_CpuMeld_TakesPozzetto(t *testing.T) {
	g := newTestBiriba()
	g.Reset()
	g.SetPhase(domain.BiribaPhaseMeld)
	g.SetCurrentPlayerIdx(1) // CPU

	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.SetHasInitMeld(true)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

	g.CpuPlay() // cpuMeld melds the sequence, emptying the hand → take pozzetto

	assert.True(t, cpu.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
}

func TestBiriba_CpuDiscard_EmptyHandTakesPozzetto(t *testing.T) {
	g := newTestBiriba()
	g.Reset()
	g.SetPhase(domain.BiribaPhaseDiscard)
	g.SetCurrentPlayerIdx(1) // CPU

	cpu := g.GetPlayer(1)
	cpu.Reset() // empty hand, pozzetto not yet taken

	g.CpuPlay() // cpuDiscard: empty hand + no pozzetto → take it, then discard

	assert.True(t, cpu.GetTookPozzetto())
}

func TestBiriba_CpuDiscard_EmptyHandAfterPozzettoAdvances(t *testing.T) {
	g := newTestBiriba()
	g.Reset()
	g.SetPhase(domain.BiribaPhaseDiscard)
	g.SetCurrentPlayerIdx(1) // CPU

	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.SetTookPozzetto(true) // already took pozzetto, no biriba, empty hand

	g.CpuPlay() // cpuDiscard: cannot discard/go out → advance turn

	assert.Equal(t, domain.BiribaPhaseDraw, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
}

func TestBiriba_DefaultConfig(t *testing.T) {
	cfg := domain.DefaultBiribaConfig()
	assert.True(t, cfg.UsePozzetto)
	assert.True(t, cfg.UseBiriba)
	assert.Equal(t, domain.BiribaDefaultPointLimit, cfg.PointLimit)
	assert.Equal(t, 2005, cfg.PointLimit)
}

func TestBiriba_PlayerMeld_UsesSameSuitSequence(t *testing.T) {
	g := newTestBiriba()
	g.SetPhase(domain.BiribaPhaseMeld)
	g.SetCurrentPlayerIdx(0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	for value := 5; value <= 7; value++ {
		player.AddCard(domain.NewCard(domain.CardDesignHeart, value, false))
	}

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Len(t, player.GetMelds(), 1)
}

func TestBiriba_PlayerMeld_RejectsSameRankSet(t *testing.T) {
	g := newTestBiriba()
	g.SetPhase(domain.BiribaPhaseMeld)
	g.SetCurrentPlayerIdx(0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	for _, design := range []int{domain.CardDesignHeart, domain.CardDesignSpade, domain.CardDesignDiamond} {
		player.AddCard(domain.NewCard(design, 7, false))
	}

	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Empty(t, player.GetMelds())
}

func TestBurraco_PlayerMeld_AcceptsSameRankSet(t *testing.T) {
	players := []*domain.CanastaPlayer{domain.NewCanastaPlayer(true), domain.NewCanastaPlayer(false)}
	cfg := domain.DefaultCanastaConfig()
	cfg.UsePozzetto = true
	g := domain.NewCanasta(domain.NewTrumpCardsWithDecks(2, 4), players, cfg)
	g.SetPhase(domain.CanastaPhaseMeld)
	g.SetCurrentPlayerIdx(0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	for _, design := range []int{domain.CardDesignHeart, domain.CardDesignSpade, domain.CardDesignDiamond} {
		player.AddCard(domain.NewCard(design, 7, false))
	}

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Len(t, player.GetMelds(), 1)
}

func TestBiriba_CpuPlay_FindsSequenceMeld(t *testing.T) {
	g := newTestBiriba()
	g.SetPhase(domain.BiribaPhaseMeld)
	g.SetCurrentPlayerIdx(1)
	player := g.GetPlayer(1)
	player.Reset()
	player.SetHasInitMeld(true)
	for value := 5; value <= 7; value++ {
		player.AddCard(domain.NewCard(domain.CardDesignClover, value, false))
	}

	g.CpuPlay()

	require.Len(t, player.GetMelds(), 1)
	assert.Equal(t, 3, len(player.GetMelds()[0].Cards))
	assert.Equal(t, domain.CanastaPhaseDiscard, g.GetPhase())
}

func TestBiriba_PlayerMeld_PureBiribaBonusIsScored(t *testing.T) {
	g := newTestBiriba()
	g.SetPhase(domain.BiribaPhaseMeld)
	g.SetCurrentPlayerIdx(0)
	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true)
	player.SetTookPozzetto(true)
	for value := 6; value <= 12; value++ {
		player.AddCard(domain.NewCard(domain.CardDesignSpade, value, false))
	}

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2, 3, 4, 5, 6}}))
	assert.Equal(t, 60+domain.CanastaPureBiribaBonus+domain.CanastaGoingOutBonus, player.GetRoundScore())
}

func TestBiriba_PlayerMeld_SequenceWildcardsStayWithinRankRange(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		wilds  int
		valid  bool
	}{
		{name: "ace and king with three wilds still need an impossible gap", values: []int{1, 13}, wilds: 3, valid: false},
		{name: "king with one wild can extend below jack", values: []int{11, 12, 13}, wilds: 1, valid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestBiriba()
			g.SetPhase(domain.BiribaPhaseMeld)
			g.SetCurrentPlayerIdx(0)
			player := g.GetPlayer(0)
			player.Reset()
			player.SetHasInitMeld(true)
			for _, value := range tt.values {
				player.AddCard(domain.NewCard(domain.CardDesignHeart, value, false))
			}
			for i := 0; i < tt.wilds; i++ {
				player.AddCard(domain.NewCard(domain.CardDesignJoker, 1, false))
			}

			err := g.PlayerMeld(intSliceToIndices(len(tt.values) + tt.wilds))
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func intSliceToIndices(n int) [][]int {
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	return [][]int{indices}
}

func TestBiriba_Reset_DealsElevenPlusTwoPozzetti(t *testing.T) {
	g := newTestBiriba()
	g.Reset()

	for i := 0; i < 2; i++ {
		assert.Equal(t, domain.BiribaHandSize, g.GetPlayer(i).GetCardsSize(),
			"player %d should be dealt 11 cards", i)
	}
	assert.Equal(t, 2, g.GetPozzettoCount(), "two pozzetti are set aside")
	assert.Equal(t, 2*domain.BiribaPozzettoSize, g.GetPozzettoCardCount())

	// 108 = hands + red3s + discard + draw + pozzetti
	total := g.GetDrawPileCount() + g.GetDiscardPileCount() + g.GetPozzettoCardCount()
	for i := 0; i < 2; i++ {
		total += g.GetPlayer(i).GetCardsSize() + len(g.GetPlayer(i).GetRed3s())
	}
	assert.Equal(t, 108, total)
}

func TestBiriba_TakePozzetto_OnMeldEmptyingHand(t *testing.T) {
	g := newTestBiriba()
	g.Reset()
	g.SetPhase(domain.BiribaPhaseMeld)
	g.SetCurrentPlayerIdx(0)
	require.Equal(t, 2, g.GetPozzettoCount())

	player := g.GetPlayer(0)
	player.Reset()
	player.SetHasInitMeld(true) // bypass initial-meld minimum
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))

	require.NoError(t, g.PlayerMeld([][]int{{0, 1, 2}}))

	// Hand emptied → pozzetto taken (one pile consumed), player keeps playing
	assert.True(t, player.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
	assert.Greater(t, player.GetCardsSize(), 0)
	assert.Equal(t, domain.BiribaPhaseDiscard, g.GetPhase())
}

func TestBiriba_TakePozzetto_OnDiscardEmptyingHand(t *testing.T) {
	g := newTestBiriba()
	g.Reset()
	g.SetPhase(domain.BiribaPhaseDiscard)
	g.SetCurrentPlayerIdx(0)

	player := g.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false)) // last card

	require.NoError(t, g.PlayerDiscard(0))

	assert.True(t, player.GetTookPozzetto())
	assert.Equal(t, 1, g.GetPozzettoCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // turn advanced to opponent
}

func TestBiriba_GoOut_RequiresPozzetto(t *testing.T) {
	g := newTestBiriba()
	g.Reset()
	g.SetPhase(domain.BiribaPhaseDiscard)
	g.SetCurrentPlayerIdx(0)

	player := g.GetPlayer(0)
	player.Reset()
	player.SetMelds([]*domain.BiribaMeld{{
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

	// Has a biriba but has NOT taken the pozzetto → cannot go out
	assert.Error(t, g.PlayerGoOut())

	// After taking the pozzetto, going out succeeds
	player.SetTookPozzetto(true)
	require.NoError(t, g.PlayerGoOut())
	assert.True(t, g.GetPhase() == domain.BiribaPhaseRoundEnd || g.GetPhase() == domain.BiribaPhaseGameEnd)
}

func TestBiriba_JSON_RoundTrip_PreservesPozzetti(t *testing.T) {
	g := newTestBiriba()
	g.Reset()

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Biriba
	require.NoError(t, json.Unmarshal(data, &g2))

	assert.Equal(t, g.GetPozzettoCount(), g2.GetPozzettoCount())
	assert.Equal(t, g.GetPozzettoCardCount(), g2.GetPozzettoCardCount())
	assert.True(t, g2.GetConfig().UsePozzetto)
}

func TestBiribaPlayer_TookPozzetto_JSON(t *testing.T) {
	p := domain.NewBiribaPlayer(true)
	p.SetTookPozzetto(true)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var p2 domain.BiribaPlayer
	require.NoError(t, json.Unmarshal(data, &p2))
	assert.True(t, p2.GetTookPozzetto())
}

func TestBiribaMeld_IsBiriba(t *testing.T) {
	cards := make([]*domain.Card, 7)
	for i := range cards {
		cards[i] = domain.NewCard(domain.CardDesignSpade, 7, false)
	}
	m := &domain.BiribaMeld{Cards: cards, IsNatural: true}
	assert.True(t, m.IsBiriba())

	short := &domain.BiribaMeld{Cards: cards[:3], IsNatural: true}
	assert.False(t, short.IsBiriba())
}

func TestBiribaPlayer_HasBiriba(t *testing.T) {
	player := domain.NewBiribaPlayer(true)
	assert.False(t, player.HasBiriba())

	cards := make([]*domain.Card, 7)
	for i := range cards {
		cards[i] = domain.NewCard(domain.CardDesignSpade, 7, false)
	}
	player.AddMeld(&domain.BiribaMeld{Cards: cards[:3], IsNatural: true})
	assert.False(t, player.HasBiriba())

	player.AddMeld(&domain.BiribaMeld{Cards: cards, IsNatural: true})
	assert.True(t, player.HasBiriba())
}
