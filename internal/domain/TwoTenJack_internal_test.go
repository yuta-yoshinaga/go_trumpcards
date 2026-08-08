//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInternalTTJ() *TwoTenJack {
	players := []*TwoTenJackPlayer{
		NewTwoTenJackPlayer(true),
		NewTwoTenJackPlayer(false),
		NewTwoTenJackPlayer(false),
		NewTwoTenJackPlayer(false),
	}
	return NewTwoTenJack(NewTrumpCards(0), players, DefaultTwoTenJackConfig())
}

func TestTTJInternal_EffectiveValue(t *testing.T) {
	assert.Equal(t, 14, ttjEffectiveValue(NewCard(CardDesignSpade, 1, false)))
	assert.Equal(t, 13, ttjEffectiveValue(NewCard(CardDesignSpade, 13, false)))
	assert.Equal(t, 2, ttjEffectiveValue(NewCard(CardDesignSpade, 2, false)))
}

func TestTTJInternal_IsValidSuit(t *testing.T) {
	assert.True(t, isValidTwoTenJackSuit(CardDesignSpade))
	assert.True(t, isValidTwoTenJackSuit(CardDesignHeart))
	assert.True(t, isValidTwoTenJackSuit(CardDesignDiamond))
	assert.True(t, isValidTwoTenJackSuit(CardDesignClover))
	assert.False(t, isValidTwoTenJackSuit(CardDesignJoker))
	assert.False(t, isValidTwoTenJackSuit(99))
}

func TestTTJInternal_SuitName(t *testing.T) {
	assert.Equal(t, "Spade", twoTenJackSuitName(CardDesignSpade))
	assert.Equal(t, "Heart", twoTenJackSuitName(CardDesignHeart))
	assert.Equal(t, "Diamond", twoTenJackSuitName(CardDesignDiamond))
	assert.Equal(t, "Club", twoTenJackSuitName(CardDesignClover))
	assert.Equal(t, "?", twoTenJackSuitName(99))
}

func TestTTJInternal_CpuTrumpNormal_ChoosesMost(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	for v := 2; v <= 10; v++ {
		p.AddCard(NewCard(CardDesignHeart, v, false))
	}
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignClover, 4, false))
	p.AddCard(NewCard(CardDesignDiamond, 5, false))
	p.AddCard(NewCard(CardDesignSpade, 6, false))
	suit := ttj.cpuTrumpNormal(1)
	assert.Equal(t, CardDesignHeart, suit)
}

func TestTTJInternal_CpuTrumpHard_PrefersHonors(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	// 4 spades with all honors (A, 10, J, K)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 10, false))
	p.AddCard(NewCard(CardDesignSpade, 11, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))
	// 5 clubs with no honors
	for _, v := range []int{2, 3, 4, 5, 6} {
		p.AddCard(NewCard(CardDesignClover, v, false))
	}
	suit := ttj.cpuTrumpHard(1)
	// honors * 3 (9) + count*2 (8) = 17 for spades
	// clubs: 0 + 10 = 10 → spades wins
	assert.Equal(t, CardDesignSpade, suit)
}

func TestTTJInternal_CpuSelectTrump_AllDifficulties(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	for v := 2; v <= 10; v++ {
		p.AddCard(NewCard(CardDesignHeart, v, false))
	}
	for _, diff := range []TwoTenJackCpuDifficulty{TwoTenJackCpuDifficultyEasy, TwoTenJackCpuDifficultyNormal, TwoTenJackCpuDifficultyHard} {
		cfg := ttj.config
		cfg.CpuDifficulty = diff
		ttj.SetConfig(cfg)
		s := ttj.cpuSelectTrump(1)
		assert.True(t, isValidTwoTenJackSuit(s))
	}
}

func TestTTJInternal_CpuSelectPlayCard_AllDifficulties(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignHeart, 10, false))
	p.AddCard(NewCard(CardDesignClover, 7, false))
	ttj.trumpSuit = CardDesignSpade
	ttj.currentPlayerIdx = 1
	ttj.leadPlayerIdx = 1
	ttj.phase = TwoTenJackPhasePlay
	ttj.trickNumber = 1

	for _, diff := range []TwoTenJackCpuDifficulty{TwoTenJackCpuDifficultyEasy, TwoTenJackCpuDifficultyNormal, TwoTenJackCpuDifficultyHard} {
		cfg := ttj.config
		cfg.CpuDifficulty = diff
		ttj.SetConfig(cfg)
		idx := ttj.cpuSelectPlayCard(1)
		assert.True(t, idx >= 0 && idx < p.GetCardsSize())
	}
}

func TestTTJInternal_CpuSelectPlayCard_SingleOption(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	p.AddCard(NewCard(CardDesignClover, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	ttj.trumpSuit = CardDesignSpade
	ttj.currentPlayerIdx = 1
	ttj.leadPlayerIdx = 0
	ttj.phase = TwoTenJackPhasePlay
	ttj.trickNumber = 1
	// lead clover, so only clover card is valid
	ttj.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 10, false)},
	}
	idx := ttj.cpuSelectPlayCard(1)
	assert.Equal(t, 0, idx)
}

func TestTTJInternal_CurrentTrickWinnerIdx_Empty(t *testing.T) {
	ttj := newInternalTTJ()
	assert.Equal(t, -1, ttj.currentTrickWinnerIdx())
}

func TestTTJInternal_TeamOf(t *testing.T) {
	ttj := newInternalTTJ()
	assert.Equal(t, 0, ttj.teamOf(0))
	assert.Equal(t, 1, ttj.teamOf(1))
	assert.Equal(t, 0, ttj.teamOf(2))
	assert.Equal(t, 1, ttj.teamOf(3))
	assert.Equal(t, -1, ttj.teamOf(-1))
}

func TestTTJInternal_PickLowestWinning_NoWinners(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	p.AddCard(NewCard(CardDesignClover, 2, false))
	ttj.trumpSuit = CardDesignSpade
	ttj.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 13, false)},
	}
	_, ok := pickLowestWinning(ttj, p, []int{0})
	assert.False(t, ok)
}

func TestTTJInternal_PickLowestWinning_TrumpWins(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[1]
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	ttj.trumpSuit = CardDesignSpade
	ttj.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 13, false)},
	}
	idx, ok := pickLowestWinning(ttj, p, []int{0})
	assert.True(t, ok)
	assert.Equal(t, 0, idx)
}

func TestTTJInternal_StartPlayPhase(t *testing.T) {
	ttj := newInternalTTJ()
	ttj.declarerIdx = 2
	ttj.startPlayPhase()
	assert.Equal(t, TwoTenJackPhasePlay, ttj.phase)
	assert.Equal(t, 2, ttj.leadPlayerIdx)
	assert.Equal(t, 2, ttj.currentPlayerIdx)
	assert.Equal(t, 1, ttj.trickNumber)
}

func TestTTJInternal_PlayerName(t *testing.T) {
	ttj := newInternalTTJ()
	assert.Equal(t, "You", playerName(ttj.players, 0))
	assert.Equal(t, "CPU 1", playerName(ttj.players, 1))
	assert.Equal(t, "Player -1", playerName(ttj.players, -1))
	assert.Equal(t, "Player 99", playerName(ttj.players, 99))
}

func TestTTJInternal_SortAllHands_StableOrder(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[0]
	p.Reset()
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	p.AddCard(NewCard(CardDesignHeart, 1, false))
	ttj.sortAllHands()
	assert.Equal(t, CardDesignSpade, p.GetCard(0).GetDesign())
	assert.Equal(t, CardDesignHeart, p.GetCard(1).GetDesign())
}

func TestTTJInternal_ValidatePlay_LeadAnyCard(t *testing.T) {
	ttj := newInternalTTJ()
	p := ttj.players[0]
	p.Reset()
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	ttj.currentTrick = nil
	assert.NoError(t, ttj.validatePlay(0, p.GetCard(0)))
}
