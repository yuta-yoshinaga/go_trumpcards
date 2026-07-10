//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestPrsi returns a freshly reset 4-player game.
func newTestPrsi(t *testing.T) *domain.Prsi {
	t.Helper()
	g := domain.NewDefaultPrsi()
	g.Reset()
	return g
}

// setHand replaces a player's hand with the given cards.
func setPrsiHand(p *domain.PrsiPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewTrumpCardsPrsi_Has32DistinctCards(t *testing.T) {
	deck := domain.NewTrumpCardsPrsi()
	assert.Equal(t, 32, deck.GetTotalCount())

	seen := make(map[int]bool)
	want := map[int]bool{1: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	count := 0
	for {
		c := deck.DrawCard()
		if c == nil {
			break
		}
		count++
		assert.True(t, want[c.GetValue()], "unexpected value %d in Prsi deck", c.GetValue())
		key := c.GetDesign()*100 + c.GetValue()
		assert.False(t, seen[key], "duplicate card design=%d value=%d", c.GetDesign(), c.GetValue())
		seen[key] = true
	}
	assert.Equal(t, 32, count)
	assert.Len(t, seen, 32)
}

func TestPrsi_ResetDealsHands(t *testing.T) {
	g := newTestPrsi(t)
	assert.Equal(t, domain.PrsiPhasePlay, g.GetPhase())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.NotNil(t, g.GetDiscardTop())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.PrsiHandSize, g.GetPlayer(i).GetCardsSize())
	}
	// 32 - 4*5 - 1 = 11 in stock
	assert.Equal(t, 11, g.GetDrawPileCount())
}

func TestPrsi_PlayMatchingSuitOrRank(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	// Discard top ♠9
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})

	// Match by suit: ♠13 (K)
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignSpade, 13, false), domain.NewCard(domain.CardDesignHeart, 12, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, 13, g.GetDiscardTop().GetValue())
	assert.Equal(t, domain.CardDesignSpade, g.GetDiscardTop().GetDesign())
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // advanced
}

func TestPrsi_InvalidPlayRejected(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	// ♥12 matches neither suit (♠) nor rank (9)
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignHeart, 12, false))
	err := g.PlayerPlay(0)
	assert.Error(t, err)
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize()) // unchanged
}

func TestPrsi_SevenForcesDrawTwoAndStacks(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
	})

	// P0 plays ♠7 -> penalty 2, P1 to act
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignSpade, 7, false), domain.NewCard(domain.CardDesignHeart, 9, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, 2, g.GetPenaltyDrawCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	// Move the turn back to the human seat (the stacking rule is seat-agnostic;
	// PlayerPlay only accepts the human seat). Stack ♥7 -> penalty 4.
	g.SetCurrentPlayerIdx(0)
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignHeart, 7, false), domain.NewCard(domain.CardDesignHeart, 13, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, 4, g.GetPenaltyDrawCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	// Under penalty, the human may only play a 7; ♥13 is rejected.
	g.SetCurrentPlayerIdx(0)
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignHeart, 13, false))
	assert.Error(t, g.PlayerPlay(0))

	// Cannot stack -> draws the 4 penalty cards, turn passes.
	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDraw())
	assert.Equal(t, before+4, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 0, g.GetPenaltyDrawCount())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestPrsi_AceSkipsNextPlayer(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	// P0 plays ♠A (value 1) -> skip P1, P2 to act
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignSpade, 1, false), domain.NewCard(domain.CardDesignHeart, 9, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, 0, g.GetPendingSkips()) // consumed in advance
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
}

func TestPrsi_UnderSkipsNextPlayer(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	// P0 plays ♠J (Under, value 11) -> skip P1, P2 to act
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignSpade, 11, false), domain.NewCard(domain.CardDesignHeart, 9, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
}

func TestPrsi_DrawWhenNoMatch(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 12, false)})
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignHeart, 13, false))
	require.NoError(t, g.PlayerDraw())
	assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize()) // drew 1
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())       // turn passes
}

func TestPrsi_DrawPassesWhenStockEmpty(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile(nil)
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignHeart, 13, false))
	require.NoError(t, g.PlayerDraw())
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize()) // nothing drawn
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestPrsi_WinConditionEmptyHand(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	// P0 has exactly one playable card; playing it empties the hand -> win.
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignSpade, 13, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
	assert.Equal(t, domain.PrsiPhaseGameEnd, g.GetPhase())
	assert.True(t, g.GetPlayer(0).GetIsFinished())
	// Further actions rejected
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerDraw(), domain.ErrGameEnded)
}

func TestPrsi_CpuPlayCompletesTurn(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(1) // a CPU
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 12, false)})
	setPrsiHand(g.GetPlayer(1), domain.NewCard(domain.CardDesignSpade, 13, false))
	g.CpuPlay()
	// CPU played its ♠K and won (single card)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestPrsi_CpuPlayDrawsWhenStuck(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(1)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 12, false)})
	setPrsiHand(g.GetPlayer(1), domain.NewCard(domain.CardDesignHeart, 13, false))
	g.CpuPlay()
	assert.Equal(t, 2, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
}

func TestPrsi_NotHumanTurnRejected(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(1) // CPU
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
	assert.ErrorIs(t, g.PlayerDraw(), domain.ErrNotHumanTurn)
}

func TestPrsi_WrongPhaseRejected(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.PrsiPhaseGameEnd)
	// gameEndFlag is false here, so ErrWrongPhase is returned (not ErrGameEnded)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)
	assert.ErrorIs(t, g.PlayerDraw(), domain.ErrWrongPhase)
}

func TestPrsi_PlayOutOfRangeRejected(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(999))
}

func TestPrsi_RecycleDiscardIntoStock(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	// Empty stock; discard pile has several cards. Drawing should recycle all
	// but the top into a new stock and then draw one.
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignSpade, 9, false), // top
	})
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignClover, 13, false))
	require.NoError(t, g.PlayerDraw())
	assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize()) // drew 1
	// 2 recycled - 1 drawn = 1 remaining
	assert.Equal(t, 1, g.GetDrawPileCount())
}

func TestPrsi_Accessors(t *testing.T) {
	g := domain.NewDefaultPrsi()
	g.SetPhase(domain.PrsiPhasePlay)
	assert.Equal(t, domain.PrsiPhasePlay, g.GetPhase())
	g.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	g.SetPenaltyDrawCount(6)
	assert.Equal(t, 6, g.GetPenaltyDrawCount())
	g.SetPendingSkips(1)
	assert.Equal(t, 1, g.GetPendingSkips())
	pile := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	g.SetDiscardPile(pile)
	assert.Equal(t, pile, g.GetDiscardPile())
	assert.Equal(t, 7, g.GetDiscardTop().GetValue())
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, false)})
	assert.Equal(t, 1, g.GetDrawPileCount())
	cfg := domain.PrsiConfig{CpuDifficulty: domain.PrsiCpuDifficultyHard}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	assert.False(t, g.GetPlayer(1).GetIsHuman())

	// IsHumanTurn: true on the human seat, false for an out-of-range index.
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())
}

func TestPrsi_DiscardTopNilWhenEmpty(t *testing.T) {
	g := domain.NewPrsi(domain.NewTrumpCardsPrsi(), []*domain.PrsiPlayer{
		domain.NewPrsiPlayer(true), domain.NewPrsiPlayer(false),
		domain.NewPrsiPlayer(false), domain.NewPrsiPlayer(false),
	}, domain.DefaultPrsiConfig())
	assert.Nil(t, g.GetDiscardTop())
}

func TestPrsi_HasPlayableCardAndValidIndices(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	setPrsiHand(g.GetPlayer(0),
		domain.NewCard(domain.CardDesignSpade, 13, false), // playable (suit)
		domain.NewCard(domain.CardDesignHeart, 12, false), // not playable
	)
	assert.True(t, g.HasPlayableCard(0))
	assert.Equal(t, []int{0}, g.GetValidPlayIndices(0))
	assert.False(t, g.HasPlayableCard(99))
	assert.Nil(t, g.GetValidPlayIndices(-1))
}

func TestPrsi_HardCpuPlaysActionCardWhenNextPlayerLow(t *testing.T) {
	g := domain.NewDefaultPrsi()
	g.Reset()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.PrsiCpuDifficultyHard
	g.SetConfig(cfg)
	g.SetCurrentPlayerIdx(1) // a CPU seat
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 10, false), domain.NewCard(domain.CardDesignDiamond, 12, false), domain.NewCard(domain.CardDesignDiamond, 13, false)})
	// Make P2 (next) low on cards so the hard CPU should attack with a 7.
	setPrsiHand(g.GetPlayer(2), domain.NewCard(domain.CardDesignClover, 12, false))
	// P1 (CPU) holds a ♠7 (action) and a ♠13 (plain). Both valid by suit.
	setPrsiHand(g.GetPlayer(1),
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
	)
	g.CpuPlay()
	// It should have played the 7 (penalty stack > 0).
	assert.Equal(t, 7, g.GetDiscardTop().GetValue())
	assert.Equal(t, 2, g.GetPenaltyDrawCount())
}

func TestPrsi_RoundTripJSON(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(2)
	g.SetPenaltyDrawCount(4)
	g.SetPendingSkips(1)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Prsi
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	assert.Equal(t, g.GetPenaltyDrawCount(), restored.GetPenaltyDrawCount())
	assert.Equal(t, g.GetPendingSkips(), restored.GetPendingSkips())
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetDrawPileCount(), restored.GetDrawPileCount())
}

func TestPrsi_UnmarshalRejectsInvalidState(t *testing.T) {
	base := newTestPrsi(t)
	good, err := json.Marshal(base)
	require.NoError(t, err)

	// Decode into a generic map so we can tamper individual fields.
	tamper := func(mut func(m map[string]any)) []byte {
		var m map[string]any
		require.NoError(t, json.Unmarshal(good, &m))
		mut(m)
		out, err := json.Marshal(m)
		require.NoError(t, err)
		return out
	}

	cases := []struct {
		name string
		data []byte
	}{
		{"bad phase", tamper(func(m map[string]any) { m["ps"] = 99 })},
		{"bad currentPlayerIdx", tamper(func(m map[string]any) { m["ci"] = 9 })},
		{"negative currentPlayerIdx", tamper(func(m map[string]any) { m["ci"] = -1 })},
		{"wrong player count", tamper(func(m map[string]any) { m["pl"] = []any{} })},
		{"bad config difficulty", tamper(func(m map[string]any) { m["cf"] = map[string]any{"cd": 99} })},
		{"nil discard element", tamper(func(m map[string]any) { m["dp"] = []any{nil} })},
		{"winner set while in progress", tamper(func(m map[string]any) { m["wi"] = 0 })},
		{"winner out of range at game end", tamper(func(m map[string]any) { m["ge"] = true; m["wi"] = 99 })},
		{"no winner at game end", tamper(func(m map[string]any) { m["ge"] = true; m["wi"] = -1 })},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var g domain.Prsi
			assert.Error(t, json.Unmarshal(tc.data, &g))
		})
	}
}

func TestPrsi_UnmarshalRejectsOversizedSlice(t *testing.T) {
	big := make([]map[string]any, 1001)
	for i := range big {
		big[i] = map[string]any{"v": 7, "d": 1, "dr": true}
	}
	payload := map[string]any{
		"pl": []any{
			map[string]any{}, map[string]any{}, map[string]any{}, map[string]any{},
		},
		"dp": big,
		"ci": 0,
		"ps": 0,
		"cf": map[string]any{"cd": 1},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	var g domain.Prsi
	err = json.Unmarshal(data, &g)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum allowed size")
}

func TestPrsi_UnmarshalClampsCounters(t *testing.T) {
	base := newTestPrsi(t)
	good, err := json.Marshal(base)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(good, &m))
	m["pd"] = -5
	m["sk"] = -3
	data, err := json.Marshal(m)
	require.NoError(t, err)

	var g domain.Prsi
	require.NoError(t, json.Unmarshal(data, &g))
	assert.Equal(t, 0, g.GetPenaltyDrawCount())
	assert.Equal(t, 0, g.GetPendingSkips())
}

func TestPrsi_GetActionLogPopulated(t *testing.T) {
	g := newTestPrsi(t)
	g.SetCurrentPlayerIdx(0)
	g.SetDiscardPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	g.SetDrawPile([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 12, false)})
	setPrsiHand(g.GetPlayer(0), domain.NewCard(domain.CardDesignSpade, 13, false), domain.NewCard(domain.CardDesignHeart, 9, false))
	require.NoError(t, g.PlayerPlay(0))
	assert.NotEmpty(t, g.GetActionLog())
}
