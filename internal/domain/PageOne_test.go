//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestPageOne() *domain.PageOne {
	players := []*domain.PageOnePlayer{
		domain.NewPageOnePlayer(true),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
	}
	return domain.NewPageOne(domain.NewTrumpCards(0), players, domain.DefaultPageOneConfig())
}

func newTestPageOneWithDifficulty(d domain.PageOneCpuDifficulty) *domain.PageOne {
	players := []*domain.PageOnePlayer{
		domain.NewPageOnePlayer(true),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
	}
	cfg := domain.DefaultPageOneConfig()
	cfg.CpuDifficulty = d
	return domain.NewPageOne(domain.NewTrumpCards(0), players, cfg)
}

// setupPageOnePlayPhase sets up a deterministic play state.
func setupPageOnePlayPhase(g *domain.PageOne, currentIdx int, topCard *domain.Card) {
	g.SetPhase(domain.PageOnePhasePlay)
	g.SetCurrentPlayerIdx(currentIdx)
	g.SetDiscardPile([]*domain.Card{topCard})
}

func TestNewPageOne(t *testing.T) {
	g := newTestPageOne()
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
}

func TestPageOne_Reset(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	assert.Equal(t, domain.PageOnePhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, g.GetPlayer(i).GetRoundScore())
		assert.Equal(t, 0, g.GetPlayer(i).GetCumulativeScore())
		assert.False(t, g.GetPlayer(i).GetHasDeclared())
	}

	assert.Len(t, g.GetDiscardPile(), 1)
	assert.Equal(t, 31, g.GetDrawPileCount())
	assert.Nil(t, g.GetActionLog())
}

func TestPageOne_Reset_ClearsAllState(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	g.GetPlayer(0).SetCumulativeScore(300)
	g.SetPhase(domain.PageOnePhaseGameEnd)

	g.Reset()
	assert.Equal(t, domain.PageOnePhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestPageOne_Getters(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.NotNil(t, g.GetPlayer(0))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(4))

	cfg := g.GetConfig()
	assert.Equal(t, domain.PageOneCpuDifficultyNormal, cfg.CpuDifficulty)

	g.SetConfig(domain.PageOneConfig{CpuDifficulty: domain.PageOneCpuDifficultyHard, PointLimit: 100})
	assert.Equal(t, domain.PageOneCpuDifficultyHard, g.GetConfig().CpuDifficulty)
}

func TestPageOne_IsHumanTurn(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(-1)
	assert.False(t, g.IsHumanTurn())
}

func TestPageOne_PlayerPlay_MatchSuit(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false)) // matches suit
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 4, false))

	err := g.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestPageOne_PlayerPlay_MatchRank(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false)) // matches rank

	err := g.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestPageOne_PlayerPlay_Invalid(t *testing.T) {
	g := newTestPageOne()
	g.Reset()

	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // neither matches

	err := g.PlayerPlay(0)
	assert.Error(t, err)
}

func TestPageOne_PlayerPlay_OutOfRange(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)

	err := g.PlayerPlay(-1)
	assert.Error(t, err)
	err = g.PlayerPlay(999)
	assert.Error(t, err)
}

func TestPageOne_PlayerPlay_WrongPhase(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseRoundEnd)
	err := g.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPageOne_PlayerPlay_NotHumanTurn(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetCurrentPlayerIdx(1)
	err := g.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestPageOne_PlayerPlay_GameEnded(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseGameEnd)
	// Simulate gameEndFlag by reaching limit
	g.GetPlayer(0).SetCumulativeScore(999)
	// There's no direct setter for gameEndFlag, so use ScoreRound path; here we simulate via Marshal/Unmarshal
	data, _ := json.Marshal(g)
	raw := map[string]interface{}{}
	_ = json.Unmarshal(data, &raw)
	raw["ge"] = true
	b, _ := json.Marshal(raw)
	g2 := domain.NewPageOne(nil, nil, domain.DefaultPageOneConfig())
	require.NoError(t, json.Unmarshal(b, g2))
	assert.True(t, g2.GetGameEndFlag())
	err := g2.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestPageOne_PlayerPlay_TriggersDeclarationPhase(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	err := g.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, domain.PageOnePhaseMustDeclare, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx(), "turn stays on player who must declare")
}

func TestPageOne_PlayerPlay_LastCardEndsRound(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))

	err := g.PlayerPlay(0)
	require.NoError(t, err)
	assert.Equal(t, domain.PageOnePhaseRoundEnd, g.GetPhase())
	assert.True(t, p.GetIsFinished())
}

func TestPageOne_PlayerDeclare(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	g.SetPhase(domain.PageOnePhaseMustDeclare)
	g.SetCurrentPlayerIdx(0)

	err := g.PlayerDeclare()
	require.NoError(t, err)
	assert.True(t, g.GetPlayer(0).GetHasDeclared())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.PageOnePhasePlay, g.GetPhase())
}

func TestPageOne_PlayerDeclare_WrongPhase(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	err := g.PlayerDeclare()
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPageOne_PlayerDeclare_NotHumanTurn(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseMustDeclare)
	g.SetCurrentPlayerIdx(1)
	err := g.PlayerDeclare()
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestPageOne_PlayerSkipDeclare_AppliesPenalty(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseMustDeclare)
	g.SetCurrentPlayerIdx(0)
	before := g.GetPlayer(0).GetCardsSize()

	err := g.PlayerSkipDeclare()
	require.NoError(t, err)

	assert.Equal(t, before+domain.PageOnePenaltyDraw, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.PageOnePhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestPageOne_PlayerSkipDeclare_WrongPhase(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	err := g.PlayerSkipDeclare()
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPageOne_PlayerDraw(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)
	// Give player only unplayable cards → draw should pass turn
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

	before := p.GetCardsSize()
	err := g.PlayerDraw()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, p.GetCardsSize(), before)
}

func TestPageOne_PlayerDraw_WrongPhase(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseRoundEnd)
	err := g.PlayerDraw()
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestPageOne_CpuPlay(t *testing.T) {
	g := newTestPageOneWithDifficulty(domain.PageOneCpuDifficultyNormal)
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 1, top)

	p := g.GetPlayer(1)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	g.CpuPlay()
	assert.LessOrEqual(t, p.GetCardsSize(), 2)
}

func TestPageOne_CpuPlay_NoOpOnHumanTurn(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	before := g.GetPlayer(0).GetCardsSize()
	g.CpuPlay()
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
}

func TestPageOne_CpuDeclare_NormalAlwaysDeclares(t *testing.T) {
	g := newTestPageOneWithDifficulty(domain.PageOneCpuDifficultyNormal)
	g.Reset()
	g.GetPlayer(1).Reset()
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	g.SetPhase(domain.PageOnePhaseMustDeclare)
	g.SetCurrentPlayerIdx(1)

	g.CpuDeclare()
	assert.True(t, g.GetPlayer(1).GetHasDeclared())
}

func TestPageOne_CpuDeclare_NoOpWrongPhase(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetCurrentPlayerIdx(1)
	g.CpuDeclare()
	assert.False(t, g.GetPlayer(1).GetHasDeclared())
}

func TestPageOne_CpuDeclare_Easy_CanForget(t *testing.T) {
	// With Easy difficulty there's a 25% chance of forgetting. Run many iterations.
	forgot := false
	declared := false
	for i := 0; i < 1000 && (!forgot || !declared); i++ {
		g := newTestPageOneWithDifficulty(domain.PageOneCpuDifficultyEasy)
		g.Reset()
		g.GetPlayer(1).Reset()
		g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		g.SetPhase(domain.PageOnePhaseMustDeclare)
		g.SetCurrentPlayerIdx(1)
		cardsBefore := g.GetPlayer(1).GetCardsSize()
		g.CpuDeclare()
		if g.GetPlayer(1).GetHasDeclared() {
			declared = true
		} else if g.GetPlayer(1).GetCardsSize() > cardsBefore {
			forgot = true
		}
	}
	assert.True(t, forgot, "Easy CPU should occasionally forget")
	assert.True(t, declared, "Easy CPU should sometimes declare")
}

func TestPageOne_ScoreRound(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	// Set up: player 0 wins, others have cards.
	for i := 0; i < 4; i++ {
		p := g.GetPlayer(i)
		p.Reset()
	}
	// Player 1 holds K (10 points), player 2 holds Ace (1), player 3 holds 5.
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	g.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
	g.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))

	g.SetPhase(domain.PageOnePhaseRoundEnd)
	g.ScoreRound()

	assert.Equal(t, 16, g.GetPlayer(0).GetCumulativeScore())
}

func TestPageOne_ScoreRound_NoOpIfNotRoundEnd(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.ScoreRound()
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestPageOne_ScoreRound_TriggersGameEnd(t *testing.T) {
	g := newTestPageOne()
	g.SetConfig(domain.PageOneConfig{CpuDifficulty: domain.PageOneCpuDifficultyNormal, PointLimit: 1})
	g.Reset()
	for i := 0; i < 4; i++ {
		g.GetPlayer(i).Reset()
	}
	g.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	g.SetPhase(domain.PageOnePhaseRoundEnd)
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestPageOne_NextRound(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseRoundEnd)
	before := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, before+1, g.GetRoundNumber())
	assert.Equal(t, domain.PageOnePhasePlay, g.GetPhase())
}

func TestPageOne_NextRound_NoOpOutsideRoundEnd(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	before := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, before, g.GetRoundNumber())
}

func TestPageOne_GetValidPlayIndices(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false)) // valid
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))  // invalid
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))  // valid (rank)

	valid := g.GetValidPlayIndices(0)
	assert.Equal(t, []int{0, 2}, valid)
}

func TestPageOne_JSONRoundTrip(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	g.SetPhase(domain.PageOnePhaseMustDeclare)
	g.GetPlayer(0).SetHasDeclared(true)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	g2 := domain.NewPageOne(nil, nil, domain.DefaultPageOneConfig())
	require.NoError(t, json.Unmarshal(data, g2))
	assert.Equal(t, domain.PageOnePhaseMustDeclare, g2.GetPhase())
	assert.True(t, g2.GetPlayer(0).GetHasDeclared())
}

func TestPageOne_UnmarshalJSON_RejectsOversized(t *testing.T) {
	payload := `{"pl":[`
	for i := 0; i < 1001; i++ {
		if i > 0 {
			payload += ","
		}
		payload += `{"gp":null,"rh":null,"hd":false}`
	}
	payload += `]}`
	g := domain.NewPageOne(nil, nil, domain.DefaultPageOneConfig())
	err := json.Unmarshal([]byte(payload), g)
	assert.Error(t, err)
}

func TestPageOne_PlayerDraw_BlockedWhenPlayableExists(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)
	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false)) // playable
	err := g.PlayerDraw()
	assert.Error(t, err)
}

func TestPageOne_DrawEmptyPilesRecyclesThenPasses(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	setupPageOnePlayPhase(g, 0, top)
	g.SetDrawPile([]*domain.Card{})       // empty
	g.SetDiscardPile([]*domain.Card{top}) // only top card → cannot recycle

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // unplayable
	err := g.PlayerDraw()
	assert.NoError(t, err)
	// Turn should advance since no card could be drawn
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestPageOne_RecycleDrawPileFromDiscard(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	top := domain.NewCard(domain.CardDesignSpade, 5, false)
	older := domain.NewCard(domain.CardDesignHeart, 2, false)
	setupPageOnePlayPhase(g, 0, top)
	g.SetDrawPile([]*domain.Card{})
	g.SetDiscardPile([]*domain.Card{older, top})

	p := g.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignClover, 9, false)) // unplayable
	err := g.PlayerDraw()
	assert.NoError(t, err)
	// Drew the recycled card and advanced turn
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestPageOne_UnmarshalJSON_Invalid(t *testing.T) {
	g := domain.NewPageOne(nil, nil, domain.DefaultPageOneConfig())
	err := json.Unmarshal([]byte("not-json"), g)
	var jsonErr *json.SyntaxError
	assert.True(t, errors.As(err, &jsonErr) || err != nil)
}

func TestPageOne_IsValidPlay(t *testing.T) {
	g := newTestPageOne()
	g.Reset()
	setupPageOnePlayPhase(g, 0, domain.NewCard(domain.CardDesignSpade, 5, false))

	assert.True(t, g.IsValidPlay(domain.NewCard(domain.CardDesignSpade, 3, false)), "same suit is playable")
	assert.True(t, g.IsValidPlay(domain.NewCard(domain.CardDesignHeart, 5, false)), "same rank is playable")
	assert.False(t, g.IsValidPlay(domain.NewCard(domain.CardDesignHeart, 7, false)), "different suit and rank is not playable")
}
