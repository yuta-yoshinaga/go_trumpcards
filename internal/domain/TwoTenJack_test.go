//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestTwoTenJack() *domain.TwoTenJack {
	players := []*domain.TwoTenJackPlayer{
		domain.NewTwoTenJackPlayer(true),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
	}
	return domain.NewTwoTenJack(domain.NewTrumpCards(0), players, domain.DefaultTwoTenJackConfig())
}

func setupTTJDeclarePhase(ttj *domain.TwoTenJack, declarerIdx int) {
	ttj.SetPhase(domain.TwoTenJackPhaseDeclare)
	ttj.SetDeclarerIdx(declarerIdx)
}

func setupTTJPlayPhase(ttj *domain.TwoTenJack, currentIdx, leadIdx, trickNum, trumpSuit int) {
	ttj.SetPhase(domain.TwoTenJackPhasePlay)
	ttj.SetCurrentPlayerIdx(currentIdx)
	ttj.SetLeadPlayerIdx(leadIdx)
	ttj.SetTrickNumber(trickNum)
	ttj.SetTrumpSuit(trumpSuit)
}

func TestNewTwoTenJack(t *testing.T) {
	ttj := newTestTwoTenJack()
	assert.Equal(t, -1, ttj.GetWinnerTeam())
	assert.Equal(t, 1, ttj.GetRoundNumber())
	assert.Equal(t, -1, ttj.GetTrumpSuit())
}

func TestTwoTenJack_Reset(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()

	assert.Equal(t, domain.TwoTenJackPhaseDeclare, ttj.GetPhase())
	assert.Equal(t, 1, ttj.GetRoundNumber())
	assert.Equal(t, 0, ttj.GetTrickNumber())
	assert.False(t, ttj.GetGameEndFlag())
	assert.Equal(t, -1, ttj.GetWinnerTeam())
	assert.Equal(t, 0, ttj.GetDeclarerIdx())
	assert.Equal(t, -1, ttj.GetTrumpSuit())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, ttj.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, ttj.GetPlayer(i).GetCumulativeScore())
	}
}

func TestTwoTenJack_PlayerDeclareTrump(t *testing.T) {
	t.Run("valid declare moves to play phase", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJDeclarePhase(ttj, 0)
		err := ttj.PlayerDeclareTrump(domain.CardDesignSpade)
		assert.NoError(t, err)
		assert.Equal(t, domain.CardDesignSpade, ttj.GetTrumpSuit())
		assert.Equal(t, domain.TwoTenJackPhasePlay, ttj.GetPhase())
		assert.Equal(t, 0, ttj.GetCurrentPlayerIdx())
		assert.Equal(t, 1, ttj.GetTrickNumber())
	})

	t.Run("wrong phase", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		ttj.SetPhase(domain.TwoTenJackPhasePlay)
		err := ttj.PlayerDeclareTrump(domain.CardDesignSpade)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJDeclarePhase(ttj, 1) // CPU declarer
		err := ttj.PlayerDeclareTrump(domain.CardDesignSpade)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid suit", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJDeclarePhase(ttj, 0)
		err := ttj.PlayerDeclareTrump(99)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	})

	t.Run("joker suit rejected", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJDeclarePhase(ttj, 0)
		err := ttj.PlayerDeclareTrump(domain.CardDesignJoker)
		assert.Error(t, err)
	})
}

func TestTwoTenJack_CpuDeclareTrump(t *testing.T) {
	t.Run("cpu declares when its turn", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJDeclarePhase(ttj, 1)
		ttj.CpuDeclareTrump()
		assert.NotEqual(t, -1, ttj.GetTrumpSuit())
		assert.Equal(t, domain.TwoTenJackPhasePlay, ttj.GetPhase())
	})

	t.Run("cpu does not declare when human turn", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJDeclarePhase(ttj, 0)
		ttj.CpuDeclareTrump()
		assert.Equal(t, -1, ttj.GetTrumpSuit())
	})

	t.Run("cpu does not declare when wrong phase", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		ttj.SetPhase(domain.TwoTenJackPhasePlay)
		ttj.SetDeclarerIdx(1)
		ttj.CpuDeclareTrump()
		assert.Equal(t, -1, ttj.GetTrumpSuit())
	})
}

func TestTwoTenJack_PlayerPlay_Errors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		ttj.SetPhase(domain.TwoTenJackPhaseDeclare)
		err := ttj.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 1, 1, 1, domain.CardDesignSpade)
		err := ttj.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("card index out of range", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 0, 0, 1, domain.CardDesignSpade)
		err := ttj.PlayerPlay(-1)
		assert.Error(t, err)
		err = ttj.PlayerPlay(100)
		assert.Error(t, err)
	})
}

func TestTwoTenJack_ValidatePlay_FollowSuit(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()

	p := ttj.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	setupTTJPlayPhase(ttj, 0, 1, 2, domain.CardDesignSpade)
	ttj.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	err := ttj.PlayerPlay(0) // heart - must follow clover
	assert.Error(t, err)
	err = ttj.PlayerPlay(1) // clover ok
	assert.NoError(t, err)
}

func TestTwoTenJack_ValidatePlay_NoLeadSuit(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()

	p := ttj.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

	setupTTJPlayPhase(ttj, 0, 1, 2, domain.CardDesignSpade)
	ttj.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	err := ttj.PlayerPlay(0) // any is allowed
	assert.NoError(t, err)
}

func TestTwoTenJack_CpuPlay(t *testing.T) {
	t.Run("cpu plays when its turn", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 1, 1, 1, domain.CardDesignSpade)
		before := ttj.GetPlayer(1).GetCardsSize()
		ttj.CpuPlay()
		assert.Equal(t, before-1, ttj.GetPlayer(1).GetCardsSize())
	})
	t.Run("cpu does not play on human turn", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 0, 0, 1, domain.CardDesignSpade)
		before := ttj.GetPlayer(0).GetCardsSize()
		ttj.CpuPlay()
		assert.Equal(t, before, ttj.GetPlayer(0).GetCardsSize())
	})
	t.Run("cpu does not play on wrong phase", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		ttj.SetPhase(domain.TwoTenJackPhaseDeclare)
		ttj.SetCurrentPlayerIdx(1)
		before := ttj.GetPlayer(1).GetCardsSize()
		ttj.CpuPlay()
		assert.Equal(t, before, ttj.GetPlayer(1).GetCardsSize())
	})
}

func TestTwoTenJack_TrickWinner(t *testing.T) {
	t.Run("highest of lead suit when no trump", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 0, 0, 2, domain.CardDesignSpade)
		ttj.SetPhase(domain.TwoTenJackPhaseTrickEnd)
		ttj.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 1, false)}, // A=14
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 13, false)},
		})
		ttj.ResolveTrick()
		assert.Equal(t, 2, ttj.GetLeadPlayerIdx())
	})

	t.Run("trump beats lead suit", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 0, 0, 2, domain.CardDesignSpade)
		ttj.SetPhase(domain.TwoTenJackPhaseTrickEnd)
		ttj.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 2, false)}, // trump
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		ttj.ResolveTrick()
		assert.Equal(t, 2, ttj.GetLeadPlayerIdx())
	})

	t.Run("highest trump wins", func(t *testing.T) {
		ttj := newTestTwoTenJack()
		ttj.Reset()
		setupTTJPlayPhase(ttj, 0, 0, 2, domain.CardDesignSpade)
		ttj.SetPhase(domain.TwoTenJackPhaseTrickEnd)
		ttj.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)}, // A=14
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
		})
		ttj.ResolveTrick()
		assert.Equal(t, 2, ttj.GetLeadPlayerIdx())
	})
}

func TestTwoTenJack_ResolveTrick_WrongPhaseOrIncomplete(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetPhase(domain.TwoTenJackPhasePlay)
	ttj.ResolveTrick() // no-op

	ttj.SetPhase(domain.TwoTenJackPhaseTrickEnd)
	ttj.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})
	ttj.ResolveTrick() // no-op
}

func TestTwoTenJack_NextTrick(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetPhase(domain.TwoTenJackPhaseTrickEnd)
	ttj.SetLeadPlayerIdx(2)
	ttj.SetTrickNumber(3)
	ttj.NextTrick()
	assert.Equal(t, domain.TwoTenJackPhasePlay, ttj.GetPhase())
	assert.Equal(t, 2, ttj.GetCurrentPlayerIdx())
	assert.Equal(t, 4, ttj.GetTrickNumber())
	assert.Nil(t, ttj.GetCurrentTrick())

	ttj.SetPhase(domain.TwoTenJackPhasePlay)
	ttj.NextTrick() // no-op
}

// assignTricks assigns tricks with the given point value totals to players to
// produce the desired scoring outcome. Each player receives a single trick made
// of arbitrary cards summing to the requested point value.
func assignTrickPoints(ttj *domain.TwoTenJack, playerIdx, points int) {
	// Build a trick whose GetCapturedPointCards yields `points`.
	// Use 10s (worth 10) and Aces (worth 1) and Jacks (worth 1) and plain 2s (worth 0).
	cards := []*domain.Card{}
	for points >= 10 {
		cards = append(cards, domain.NewCard(domain.CardDesignSpade, 10, false))
		points -= 10
	}
	for points > 0 {
		cards = append(cards, domain.NewCard(domain.CardDesignSpade, 1, false))
		points--
	}
	// pad with 0-point filler
	cards = append(cards, domain.NewCard(domain.CardDesignSpade, 2, false))
	ttj.GetPlayer(playerIdx).AddTrick(cards)
}

func TestTwoTenJack_ScoreRound_DeclarerMakesContract(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetDeclarerIdx(0) // team 0 declares
	ttj.SetPhase(domain.TwoTenJackPhaseRoundEnd)
	ttj.SetLeadPlayerIdx(0) // team 0 wins last trick (+2)

	// team 0 captures 30 points from cards (seats 0 and 2)
	assignTrickPoints(ttj, 0, 20)
	assignTrickPoints(ttj, 2, 10)
	// team 1 captures 18 points
	assignTrickPoints(ttj, 1, 10)
	assignTrickPoints(ttj, 3, 8)
	// total pool = 48 cards + 2 last trick = 50
	// team 0 = 30 + 2 (last trick) = 32, team 1 = 18
	// team 0 >= 26 => round score = 32 - 25 = 7

	ttj.ScoreRound()

	assert.Equal(t, 7, ttj.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 7, ttj.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 0, ttj.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 0, ttj.GetPlayer(3).GetCumulativeScore())
}

func TestTwoTenJack_ScoreRound_DeclarerFails(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetDeclarerIdx(0) // team 0 declares
	ttj.SetPhase(domain.TwoTenJackPhaseRoundEnd)
	ttj.SetLeadPlayerIdx(1) // team 1 wins last trick

	// team 0 captures 20 cards, team 1 captures 28 cards
	assignTrickPoints(ttj, 0, 10)
	assignTrickPoints(ttj, 2, 10)
	assignTrickPoints(ttj, 1, 18)
	assignTrickPoints(ttj, 3, 10)
	// last trick bonus 2 to player 1 => team 1 total = 30, team 0 = 20
	// declarer team 0 pts = 20 < 26 => defenders score 26 - 20 = 6

	ttj.ScoreRound()

	assert.Equal(t, 0, ttj.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 0, ttj.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 6, ttj.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 6, ttj.GetPlayer(3).GetCumulativeScore())
}

func TestTwoTenJack_ScoreRound_WrongPhase(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetPhase(domain.TwoTenJackPhasePlay)
	ttj.ScoreRound() // no-op
}

func TestTwoTenJack_GameEnd_PointLimit(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetDeclarerIdx(0)
	ttj.SetPhase(domain.TwoTenJackPhaseRoundEnd)
	ttj.SetLeadPlayerIdx(0)

	// Preload team 0 cumulative
	ttj.GetPlayer(0).SetCumulativeScore(45)
	ttj.GetPlayer(2).SetCumulativeScore(0)

	// team 0 captures everything (46 cards +2 = 48 points) — make contract easily
	assignTrickPoints(ttj, 0, 30)
	assignTrickPoints(ttj, 2, 18)

	ttj.ScoreRound()
	assert.True(t, ttj.GetGameEndFlag())
	assert.Equal(t, domain.TwoTenJackPhaseGameEnd, ttj.GetPhase())
	assert.Equal(t, 0, ttj.GetWinnerTeam())
}

func TestTwoTenJack_NextRound(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetDeclarerIdx(0)
	ttj.SetPhase(domain.TwoTenJackPhaseRoundEnd)
	ttj.NextRound()
	assert.Equal(t, domain.TwoTenJackPhaseDeclare, ttj.GetPhase())
	assert.Equal(t, 2, ttj.GetRoundNumber())
	assert.Equal(t, 1, ttj.GetDeclarerIdx()) // rotated
	assert.Equal(t, -1, ttj.GetTrumpSuit())
	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, ttj.GetPlayer(i).GetCardsSize())
	}
}

func TestTwoTenJack_NextRound_WrongPhase(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetPhase(domain.TwoTenJackPhasePlay)
	ttj.NextRound() // no-op
	assert.Equal(t, 1, ttj.GetRoundNumber())
}

func TestTwoTenJack_IsHumanTurn(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetCurrentPlayerIdx(0)
	assert.True(t, ttj.IsHumanTurn())
	ttj.SetCurrentPlayerIdx(1)
	assert.False(t, ttj.IsHumanTurn())
	ttj.SetCurrentPlayerIdx(-1)
	assert.False(t, ttj.IsHumanTurn())
}

func TestTwoTenJack_IsHumanDeclareTurn(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetDeclarerIdx(0)
	assert.True(t, ttj.IsHumanDeclareTurn())
	ttj.SetDeclarerIdx(1)
	assert.False(t, ttj.IsHumanDeclareTurn())
	ttj.SetDeclarerIdx(-1)
	assert.False(t, ttj.IsHumanDeclareTurn())
}

func TestTwoTenJack_GetConfig_SetConfig(t *testing.T) {
	ttj := newTestTwoTenJack()
	cfg := ttj.GetConfig()
	cfg.PointLimit = 100
	ttj.SetConfig(cfg)
	assert.Equal(t, 100, ttj.GetConfig().PointLimit)
}

func TestTwoTenJack_GetHint_Declare(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	setupTTJDeclarePhase(ttj, 0)
	hint := ttj.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.TrumpSuit)
	assert.Nil(t, hint.CardIndex)
}

func TestTwoTenJack_GetHint_Play(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	setupTTJPlayPhase(ttj, 0, 0, 1, domain.CardDesignSpade)
	hint := ttj.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestTwoTenJack_GetHint_NotHumanTurn(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	setupTTJPlayPhase(ttj, 1, 1, 1, domain.CardDesignSpade)
	assert.Nil(t, ttj.GetHint())
}

func TestTwoTenJack_GetValidPlayIndices(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	p := ttj.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	setupTTJPlayPhase(ttj, 0, 1, 2, domain.CardDesignSpade)
	ttj.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})
	valid := ttj.GetValidPlayIndices(0)
	assert.Equal(t, []int{1}, valid)
}

func TestTwoTenJack_PlayerPlay_FullTrick(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	// Setup all players with 1 card each
	for i := 0; i < 4; i++ {
		p := ttj.GetPlayer(i)
		p.Reset()
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 2+i, false))
	}
	setupTTJPlayPhase(ttj, 0, 0, 1, domain.CardDesignSpade)

	err := ttj.PlayerPlay(0)
	assert.NoError(t, err)
	// Human played, now CPU 1's turn
	assert.Equal(t, 1, ttj.GetCurrentPlayerIdx())
}

func TestTwoTenJack_GameEndedGuards(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	ttj.SetPhase(domain.TwoTenJackPhaseDeclare)
	// Force game ended manually
	ttj.SetDeclarerIdx(0)
	ttj.SetPhase(domain.TwoTenJackPhaseRoundEnd)
	ttj.SetLeadPlayerIdx(0)
	ttj.GetPlayer(0).SetCumulativeScore(45)
	assignTrickPoints(ttj, 0, 48)
	ttj.ScoreRound()
	assert.True(t, ttj.GetGameEndFlag())

	err := ttj.PlayerDeclareTrump(domain.CardDesignSpade)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
	err = ttj.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestTwoTenJack_GetActionLog(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	setupTTJDeclarePhase(ttj, 0)
	_ = ttj.PlayerDeclareTrump(domain.CardDesignSpade)
	log := ttj.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestTwoTenJack_JSONRoundTrip(t *testing.T) {
	ttj := newTestTwoTenJack()
	ttj.Reset()
	data, err := ttj.MarshalJSON()
	assert.NoError(t, err)
	ttj2 := &domain.TwoTenJack{}
	err = ttj2.UnmarshalJSON(data)
	assert.NoError(t, err)
}

func TestTwoTenJack_UnmarshalJSON_WrongPlayerCount(t *testing.T) {
	// Fewer than TwoTenJackPlayerCnt players must be rejected.
	ttj := newTestTwoTenJack()
	ttj.Reset()
	data, err := ttj.MarshalJSON()
	require.NoError(t, err)

	// Decode, remove one player, re-encode.
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	players := raw["ps"].([]interface{})
	raw["ps"] = players[:3] // 3 instead of 4
	badData, err := json.Marshal(raw)
	require.NoError(t, err)

	var ttj2 domain.TwoTenJack
	err = json.Unmarshal(badData, &ttj2)
	assert.ErrorContains(t, err, "invalid")
}

func TestTwoTenJack_UnmarshalJSON_TooManyTrickCards(t *testing.T) {
	// More than TwoTenJackPlayerCnt trick entries must be rejected.
	ttj := newTestTwoTenJack()
	ttj.Reset()
	data, err := ttj.MarshalJSON()
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["ct"] = []interface{}{
		map[string]interface{}{"pi": 0, "c": map[string]interface{}{"d": 1, "v": 1, "w": false}},
		map[string]interface{}{"pi": 1, "c": map[string]interface{}{"d": 2, "v": 2, "w": false}},
		map[string]interface{}{"pi": 2, "c": map[string]interface{}{"d": 3, "v": 3, "w": false}},
		map[string]interface{}{"pi": 3, "c": map[string]interface{}{"d": 4, "v": 4, "w": false}},
		map[string]interface{}{"pi": 0, "c": map[string]interface{}{"d": 1, "v": 5, "w": false}},
	}
	badData, err := json.Marshal(raw)
	require.NoError(t, err)

	var ttj2 domain.TwoTenJack
	err = json.Unmarshal(badData, &ttj2)
	assert.ErrorContains(t, err, "invalid")
}
