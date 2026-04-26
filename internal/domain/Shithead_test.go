//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestShithead returns a Shithead with a deterministic, hand-crafted state.
// The deck is shuffled by Reset, so individual test scenarios manipulate
// the players' cards directly after Reset to set up the precise state under test.
func newTestShithead(cfg ShitheadConfig) *Shithead {
	players := make([]*ShitheadPlayer, ShitheadPlayerCnt)
	players[0] = NewShitheadPlayer(true)
	for i := 1; i < ShitheadPlayerCnt; i++ {
		players[i] = NewShitheadPlayer(false)
	}
	return NewShithead(NewTrumpCards(0), players, cfg)
}

// resetEmpty clears all stocks and returns a fresh Shithead ready to be
// hand-fed cards.
func resetEmpty(s *Shithead) {
	for _, p := range s.players {
		p.Reset()
	}
	s.round = shitheadRoundState{
		discardPile: make([]*Card, 0),
		stockPile:   make([]*Card, 0),
		nextRank:    1,
	}
}

func TestNewDefaultShithead(t *testing.T) {
	s := NewDefaultShithead()
	require.NotNil(t, s)
	assert.Equal(t, ShitheadPlayerCnt, s.GetPlayerCnt())
	assert.True(t, s.GetPlayer(0).GetIsHuman())
	for i := 1; i < ShitheadPlayerCnt; i++ {
		assert.False(t, s.GetPlayer(i).GetIsHuman())
	}
}

func TestShitheadReset_DealsThreeLayers(t *testing.T) {
	s := NewDefaultShithead()
	s.Reset()
	for i := 0; i < ShitheadPlayerCnt; i++ {
		p := s.GetPlayer(i)
		assert.Equal(t, ShitheadInitialFaceDown, p.GetFaceDownSize(), "player %d face-down", i)
		assert.Equal(t, ShitheadInitialFaceUp, p.GetFaceUpSize(), "player %d face-up", i)
		assert.Equal(t, ShitheadInitialHand, p.GetCardsSize(), "player %d hand", i)
	}
	// 52 cards total - (3+3+3)*4 = 52-36 = 16 stock cards
	assert.Equal(t, 16, s.GetStockSize())
	assert.Empty(t, s.GetDiscardPile())
	assert.False(t, s.GetGameEndFlag())
	assert.Equal(t, 0, s.GetCurrentTurn())
}

func TestShitheadReset_DefaultConfigSetsAllMagic(t *testing.T) {
	s := NewDefaultShithead()
	cfg := s.GetConfig()
	assert.True(t, cfg.MagicTwo)
	assert.True(t, cfg.MagicSeven)
	assert.True(t, cfg.MagicEight)
	assert.True(t, cfg.MagicTen)
	assert.True(t, cfg.FourOfAKindBurn)
}

func TestShitheadIsPlayable(t *testing.T) {
	cfg := DefaultShitheadConfig()
	s := newTestShithead(cfg)
	resetEmpty(s)

	// empty pile: anything goes
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignSpade, 5, false)}))

	// top is 5, must be ≥5
	s.round.discardPile = append(s.round.discardPile, NewCard(CardDesignSpade, 5, false))
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 5, false)}))
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 9, false)}))
	assert.False(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 3, false)}))

	// magic 2 always playable
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignClover, 2, false)}))

	// magic 10 always playable
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignClover, 10, false)}))

	// Ace ranks above K
	s.round.discardPile = []*Card{NewCard(CardDesignSpade, 13, false)}
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 1, false)}))

	// 7 active: must be ≤7
	s.round.discardPile = []*Card{NewCard(CardDesignSpade, 7, false)}
	s.round.sevenActive = true
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 5, false)}))
	assert.False(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 9, false)}))
	// magic 2 / 10 still allowed
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 2, false)}))
	assert.True(t, s.isPlayable([]*Card{NewCard(CardDesignHeart, 10, false)}))
}

func TestShitheadIsPlayable_RequiresSameValue(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	// mixed values are not playable
	assert.False(t, s.isPlayable([]*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 6, false),
	}))
	// same value pair is playable
	assert.True(t, s.isPlayable([]*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 5, false),
	}))
	// empty input is not playable
	assert.False(t, s.isPlayable([]*Card{}))
}

func TestShitheadPlayerPlay_WrongTurnAndEndedGame(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	s.round.gameEndFlag = true
	err := s.PlayerPlay([]int{0})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrGameEnded))

	s.round.gameEndFlag = false
	s.round.currentTurn = 1 // CPU turn
	err = s.PlayerPlay([]int{0})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotHumanTurn))
}

func TestShitheadPlayerPlay_FromHand(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 5, false))
	human.AddCard(NewCard(CardDesignHeart, 9, false))
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	assert.Equal(t, 1, len(s.GetDiscardPile()))
	assert.Equal(t, 5, s.GetTopCard().GetValue())
	require.NotNil(t, s.GetHumanAction())
	assert.Equal(t, ShitheadSourceHand, s.GetHumanAction().Source)
}

func TestShitheadPlayerPlay_PickupEmptyIndices(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	s.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 4, false))
	s.round.discardPile = []*Card{
		NewCard(CardDesignClover, 9, false),
	}
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{})
	require.NoError(t, err)
	assert.Empty(t, s.GetDiscardPile())
	assert.Equal(t, 2, s.GetPlayer(0).GetCardsSize())
	require.NotNil(t, s.GetHumanAction())
	assert.True(t, s.GetHumanAction().Pickup)
}

func TestShitheadApplyPlay_TenBurnsPile(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 10, false))
	// Backup card so the human doesn't immediately finish (which would end the game).
	human.AddCard(NewCard(CardDesignDiamond, 4, false))
	s.round.discardPile = []*Card{
		NewCard(CardDesignClover, 5, false),
	}
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	assert.Empty(t, s.GetDiscardPile(), "10 burns the pile")
	assert.True(t, s.GetHumanAction().Burned)
	// same player goes again on burn (still human)
	assert.Equal(t, 0, s.GetCurrentTurn())
}

func TestShitheadApplyPlay_EightSkipsNext(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 8, false))
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	assert.True(t, s.GetHumanAction().Skipped)
	// from 0, with skip, advance step=2 → CPU 2
	assert.Equal(t, 2, s.GetCurrentTurn())
}

func TestShitheadApplyPlay_SevenLocksLow(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 7, false))
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	assert.True(t, s.GetSevenActive())
	// CPU player 1 must play ≤7 next; if they don't have one they pickup.
	// Not asserting CPU side here — covered by integration via CpuPlay tests.
}

func TestShitheadApplyPlay_FourOfAKindBurns(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 5, false))
	// Backup card so the human doesn't immediately finish.
	human.AddCard(NewCard(CardDesignDiamond, 4, false))
	// Three 5s already on the discard pile
	s.round.discardPile = []*Card{
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 5, false),
	}
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	assert.Empty(t, s.GetDiscardPile())
	assert.True(t, s.GetHumanAction().Burned)
	// Same player goes again on burn
	assert.Equal(t, 0, s.GetCurrentTurn())
}

func TestShitheadApplyPlay_MultipleSameValue(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 6, false))
	human.AddCard(NewCard(CardDesignHeart, 6, false))
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0, 1})
	require.NoError(t, err)
	assert.Equal(t, 2, len(s.GetDiscardPile()))
	assert.Equal(t, 0, s.GetPlayer(0).GetCardsSize())
}

func TestShitheadApplyPlay_RefillFromStock(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 5, false))
	// stock has cards available for refill (need to fill to 3)
	s.round.stockPile = []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	// hand should be refilled to 3 (still had 0 left; refill 3 from stock)
	assert.Equal(t, 3, s.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 0, s.GetStockSize())
}

func TestShitheadApplyPlay_InvalidPlayMixedValues(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 5, false))
	human.AddCard(NewCard(CardDesignHeart, 9, false))
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0, 1})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPlay))
}

func TestShitheadApplyPlay_InvalidPlayBelowTop(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 3, false))
	s.round.discardPile = []*Card{NewCard(CardDesignClover, 9, false)}
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidPlay))
}

func TestShitheadApplyPlay_InvalidIndexOutOfRange(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	s.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 5, false))
	s.round.currentTurn = 0
	err := s.PlayerPlay([]int{99})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}

func TestShitheadAdvanceTurn_SkipsFinishedPlayers(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	s.GetPlayer(1).SetIsFinished(true)
	s.round.currentTurn = 0
	s.advanceTurn()
	assert.Equal(t, 2, s.GetCurrentTurn())
}

func TestShitheadCheckGameEnd_AssignsLastRank(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	for i := 0; i < 3; i++ {
		s.GetPlayer(i).SetIsFinished(true)
		s.GetPlayer(i).SetRank(i + 1)
	}
	s.round.nextRank = 4
	s.checkGameEnd()
	assert.True(t, s.GetGameEndFlag())
	assert.Equal(t, 4, s.GetPlayer(3).GetRank())
}

func TestShitheadCpuPlay_PicksValidCard(t *testing.T) {
	cfg := DefaultShitheadConfig()
	s := newTestShithead(cfg)
	resetEmpty(s)
	cpu := s.GetPlayer(1)
	cpu.AddCard(NewCard(CardDesignSpade, 9, false))
	cpu.AddCard(NewCard(CardDesignHeart, 4, false))
	s.round.currentTurn = 1
	s.round.discardPile = []*Card{NewCard(CardDesignClover, 5, false)}

	s.CpuPlay()
	// CPU should play the 9 (only valid play) — but normal/hard prefer non-magic;
	// 9 is non-magic so it should play.
	require.Len(t, s.GetCpuActions(), 1)
	action := s.GetCpuActions()[0]
	require.False(t, action.Pickup)
	require.Len(t, action.PlayedCards, 1)
	assert.Equal(t, 9, action.PlayedCards[0].GetValue())
}

func TestShitheadCpuPlay_PickupWhenNothingValid(t *testing.T) {
	cfg := DefaultShitheadConfig()
	cfg.MagicTwo = false
	cfg.MagicTen = false
	s := newTestShithead(cfg)
	resetEmpty(s)
	cpu := s.GetPlayer(1)
	cpu.AddCard(NewCard(CardDesignSpade, 3, false))
	cpu.AddCard(NewCard(CardDesignHeart, 4, false))
	s.round.currentTurn = 1
	s.round.discardPile = []*Card{NewCard(CardDesignClover, 13, false)} // K, can't beat

	s.CpuPlay()
	require.Len(t, s.GetCpuActions(), 1)
	assert.True(t, s.GetCpuActions()[0].Pickup)
}

func TestShitheadCpuPlay_FromFaceUpWhenHandEmpty(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	cpu := s.GetPlayer(1)
	cpu.AddFaceUp(NewCard(CardDesignSpade, 4, false))
	s.round.currentTurn = 1

	s.CpuPlay()
	require.Len(t, s.GetCpuActions(), 1)
	assert.Equal(t, ShitheadSourceFaceUp, s.GetCpuActions()[0].Source)
	assert.Equal(t, 0, cpu.GetFaceUpSize())
}

func TestShitheadCpuPlay_FromFaceDownWhenAllUpper(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	cpu := s.GetPlayer(1)
	cpu.AddFaceDown(NewCard(CardDesignSpade, 4, false))
	s.round.currentTurn = 1

	s.CpuPlay()
	require.Len(t, s.GetCpuActions(), 1)
	assert.Equal(t, ShitheadSourceFaceDown, s.GetCpuActions()[0].Source)
}

func TestShitheadFaceDownPlay_PickupOnInvalid(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddFaceDown(NewCard(CardDesignSpade, 3, false))
	s.round.discardPile = []*Card{NewCard(CardDesignClover, 13, false)} // K
	s.round.currentTurn = 0

	err := s.PlayerPlay([]int{0})
	require.NoError(t, err)
	require.NotNil(t, s.GetHumanAction())
	assert.True(t, s.GetHumanAction().Pickup)
	// Discard moved into hand (3 was taken into hand because not playable)
	assert.True(t, human.GetCardsSize() >= 1)
	assert.Empty(t, s.GetDiscardPile())
}

func TestShitheadCurrentSourceTransitions(t *testing.T) {
	s := newTestShithead(DefaultShitheadConfig())
	resetEmpty(s)
	human := s.GetPlayer(0)
	human.AddCard(NewCard(CardDesignSpade, 5, false))
	human.AddFaceUp(NewCard(CardDesignHeart, 9, false))
	human.AddFaceDown(NewCard(CardDesignDiamond, 3, false))

	assert.Equal(t, ShitheadSourceHand, s.CurrentSource())
	human.RemoveCard(0)
	assert.Equal(t, ShitheadSourceFaceUp, s.CurrentSource())
	human.RemoveFaceUpCard(0)
	assert.Equal(t, ShitheadSourceFaceDown, s.CurrentSource())
}

func TestShitheadConfigValidate(t *testing.T) {
	cfg := DefaultShitheadConfig()
	assert.NoError(t, cfg.Validate())
	cfg.CpuDifficulty = -1
	assert.Error(t, cfg.Validate())
	cfg.CpuDifficulty = 99
	assert.Error(t, cfg.Validate())
}

func TestShitheadJSONRoundTrip(t *testing.T) {
	s := NewDefaultShithead()
	s.Reset()
	data, err := json.Marshal(s)
	require.NoError(t, err)

	s2 := NewDefaultShithead()
	err = json.Unmarshal(data, s2)
	require.NoError(t, err)
	assert.Equal(t, s.GetCurrentTurn(), s2.GetCurrentTurn())
	assert.Equal(t, s.GetStockSize(), s2.GetStockSize())
	assert.Equal(t, s.GetPlayerCnt(), s2.GetPlayerCnt())
	for i := 0; i < s.GetPlayerCnt(); i++ {
		assert.Equal(t, s.GetPlayer(i).GetCardsSize(), s2.GetPlayer(i).GetCardsSize())
		assert.Equal(t, s.GetPlayer(i).GetFaceDownSize(), s2.GetPlayer(i).GetFaceDownSize())
		assert.Equal(t, s.GetPlayer(i).GetFaceUpSize(), s2.GetPlayer(i).GetFaceUpSize())
	}
}

func TestShitheadConfigJSONRoundTrip(t *testing.T) {
	cfg := DefaultShitheadConfig()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var cfg2 ShitheadConfig
	require.NoError(t, json.Unmarshal(data, &cfg2))
	assert.Equal(t, cfg, cfg2)
}

func TestShitheadPlayer_FaceDownAndUpAccessors(t *testing.T) {
	p := NewShitheadPlayer(false)
	c1 := NewCard(CardDesignSpade, 5, false)
	c2 := NewCard(CardDesignHeart, 9, false)
	p.AddFaceDown(c1)
	p.AddFaceUp(c2)
	assert.Equal(t, c1, p.GetFaceDownCard(0))
	assert.Equal(t, c2, p.GetFaceUpCard(0))
	assert.Nil(t, p.GetFaceDownCard(99))
	assert.Nil(t, p.GetFaceUpCard(99))
	assert.Equal(t, []*Card{c1}, p.GetFaceDownCards())
	assert.Equal(t, []*Card{c2}, p.GetFaceUpCards())
}

func TestShitheadPlayer_HasAnyCardsAndReset(t *testing.T) {
	p := NewShitheadPlayer(false)
	assert.False(t, p.HasAnyCards())
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	assert.True(t, p.HasAnyCards())
	assert.True(t, p.HasHandCards())
	p.Reset()
	assert.False(t, p.HasAnyCards())
}

func TestShitheadGetPlayer_OutOfRange(t *testing.T) {
	s := NewDefaultShithead()
	assert.Nil(t, s.GetPlayer(-1))
	assert.Nil(t, s.GetPlayer(99))
}

func TestShitheadAccessors(t *testing.T) {
	s := NewDefaultShithead()
	assert.False(t, s.GetSkipNext())
	assert.Empty(t, s.GetActionLog())
	cfg := DefaultShitheadConfig()
	cfg.MagicSeven = false
	s.SetConfig(cfg)
	assert.False(t, s.GetConfig().MagicSeven)
}

func TestShitheadGameLoopEndsEventually(t *testing.T) {
	// Smoke test: with the default deck, drive CPU/human until game ends.
	// The human always picks up to make sure the game progresses.
	s := NewDefaultShithead()
	s.Reset()
	for steps := 0; steps < 5000 && !s.GetGameEndFlag(); steps++ {
		if s.IsHumanTurn() {
			// Try to play first hand card if possible, else pickup.
			human := s.GetPlayer(0)
			if human.GetCardsSize() > 0 {
				if err := s.PlayerPlay([]int{0}); err != nil {
					_ = s.PlayerPlay(nil)
				}
			} else if human.GetFaceUpSize() > 0 {
				if err := s.PlayerPlay([]int{0}); err != nil {
					_ = s.PlayerPlay(nil)
				}
			} else {
				_ = s.PlayerPlay([]int{0})
			}
		} else {
			s.CpuPlay()
		}
	}
	assert.True(t, s.GetGameEndFlag(), "game should end within step budget")
}
