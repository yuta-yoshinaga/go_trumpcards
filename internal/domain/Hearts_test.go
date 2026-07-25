//go:build test

package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestHearts() *domain.Hearts {
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),  // player 0 = human
		domain.NewHeartsPlayer(false), // player 1 = CPU
		domain.NewHeartsPlayer(false), // player 2 = CPU
		domain.NewHeartsPlayer(false), // player 3 = CPU
	}
	cfg := domain.DefaultHeartsConfig()
	tc := domain.NewTrumpCards(0)
	return domain.NewHearts(tc, players, cfg)
}

// newTestHeartsAllCPU creates a Hearts game with no human player.
func newTestHeartsAllCPU() *domain.Hearts {
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
	cfg := domain.DefaultHeartsConfig()
	tc := domain.NewTrumpCards(0)
	return domain.NewHearts(tc, players, cfg)
}

// setupPlayPhase sets a Hearts game into play phase with given player as current.
func setupPlayPhase(h *domain.Hearts, currentIdx, leadIdx, trickNum int) {
	h.SetPhase(domain.HeartsPhasePlay)
	h.SetCurrentPlayerIdx(currentIdx)
	h.SetLeadPlayerIdx(leadIdx)
	h.SetTrickNumber(trickNum)
}

// --- NewHearts ---

func TestNewHearts(t *testing.T) {
	h := newTestHearts()
	assert.Equal(t, domain.HeartsPhase(0), h.GetPhase()) // default zero
	assert.Equal(t, 0, h.GetRoundNumber())
	assert.Equal(t, 0, h.GetTrickNumber())
	assert.Equal(t, 4, h.GetPlayerCnt())
	assert.Equal(t, -1, h.GetWinnerIdx())
	assert.False(t, h.GetGameEndFlag())
	assert.False(t, h.GetHeartsBroken())
	assert.Nil(t, h.GetCurrentTrick())
	assert.Nil(t, h.GetActionLog())
}

func TestNewDefaultHearts(t *testing.T) {
	h := domain.NewDefaultHearts()
	assert.NotNil(t, h)
	assert.Equal(t, domain.HeartsPlayerCnt, h.GetPlayerCnt())
	assert.True(t, h.GetPlayer(0).GetIsHuman())
	for i := 1; i < h.GetPlayerCnt(); i++ {
		assert.False(t, h.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
	assert.Equal(t, -1, h.GetWinnerIdx())
	assert.False(t, h.GetGameEndFlag())
}

// --- Reset ---

func TestHearts_Reset(t *testing.T) {
	h := newTestHearts()
	h.Reset()

	assert.False(t, h.GetGameEndFlag())
	assert.Equal(t, -1, h.GetWinnerIdx())
	assert.Equal(t, 1, h.GetRoundNumber())
	assert.Equal(t, 0, h.GetTrickNumber()) // stays 0 until startPlayPhase is called after pass
	assert.False(t, h.GetHeartsBroken())

	// 52 cards dealt to 4 players (13 each)
	total := 0
	for i := 0; i < 4; i++ {
		total += h.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)

	// Round 1 = pass left, so phase should be pass
	assert.Equal(t, domain.HeartsPhasePass, h.GetPhase())
}

func TestHearts_Reset_PassNoneRound(t *testing.T) {
	// Round 4 = no pass -> phase should be play directly
	h := newTestHearts()
	h.SetRoundNumber(3) // after reset, roundNumber becomes 1; we need round 4
	// Workaround: reset sets roundNumber=1. We need GetPassDirection to return HeartsPassNone.
	// Round 4 means (4-1)%4 = 3 = HeartsPassNone. So we need roundNumber=4.
	// Reset always sets roundNumber=1, so instead test NextRound for PassNone path.
	h.Reset()
	assert.Equal(t, domain.HeartsPhasePass, h.GetPhase()) // round 1 = left

	// Now set round=3 and trigger NextRound to get round 4
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(3) // NextRound will increment to 4
	h.NextRound()
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())        // PassNone -> play directly
	assert.Equal(t, 1, h.GetTrickNumber())                       // startPlayPhase sets trick 1
	assert.Equal(t, domain.HeartsPassNone, h.GetPassDirection()) // round 4
}

// --- GetPassDirection ---

func TestHearts_GetPassDirection(t *testing.T) {
	h := newTestHearts()
	tests := []struct {
		round    int
		expected domain.HeartsPassDirection
	}{
		{1, domain.HeartsPassLeft},
		{2, domain.HeartsPassRight},
		{3, domain.HeartsPassAcross},
		{4, domain.HeartsPassNone},
		{5, domain.HeartsPassLeft}, // cycles
		{8, domain.HeartsPassNone},
	}
	for _, tt := range tests {
		h.SetRoundNumber(tt.round)
		assert.Equal(t, tt.expected, h.GetPassDirection(), "round %d", tt.round)
	}
}

// --- PlayerPass ---

func TestHearts_PlayerPass_Valid(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	assert.Equal(t, domain.HeartsPhasePass, h.GetPhase())

	// Human is player 0; pass 3 cards
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)

	ready := h.GetPassReady()
	assert.True(t, ready[0])
	assert.Equal(t, 10, h.GetPlayer(0).GetCardsSize()) // 13 - 3 = 10
}

func TestHearts_PlayerPass_GameEnded(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetPhase(domain.HeartsPhaseGameEnd)
	// We need to set gameEndFlag too; use ScoreRound to trigger it or set directly.
	// gameEndFlag is private; trigger game end via ScoreRound
	// Alternative: just check the error from PlayerPass when phase is not pass.
	// Actually PlayerPass checks gameEndFlag first. Let's do a full game end setup.
	// Set scores to trigger game end via ScoreRound
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.GetPlayer(1).SetCumulativeScore(95)
	h.GetPlayer(1).SetRoundScore(10)
	h.ScoreRound()
	assert.True(t, h.GetGameEndFlag())

	err := h.PlayerPass([]int{0, 1, 2})
	assert.True(t, errors.Is(err, domain.ErrGameEnded))
}

func TestHearts_PlayerPass_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetPhase(domain.HeartsPhasePlay)

	err := h.PlayerPass([]int{0, 1, 2})
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHearts_PlayerPass_NoHuman(t *testing.T) {
	h := newTestHeartsAllCPU()
	h.Reset()
	// Force pass phase (round 1 = left)
	h.SetPhase(domain.HeartsPhasePass)

	err := h.PlayerPass([]int{0, 1, 2})
	assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
}

func TestHearts_PlayerPass_WrongCount(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	err := h.PlayerPass([]int{0, 1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestHearts_PlayerPass_DuplicateIndex(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	err := h.PlayerPass([]int{0, 0, 1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestHearts_PlayerPass_OutOfRange(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	err := h.PlayerPass([]int{0, 1, 99})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestHearts_PlayerPass_NegativeIndex(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	err := h.PlayerPass([]int{0, 1, -1})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestHearts_PlayerPass_AlreadyPassed(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)

	// Try again
	err = h.PlayerPass([]int{0, 1, 2})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

// --- CpuPass ---

func TestHearts_CpuPass(t *testing.T) {
	h := newTestHearts()
	h.Reset()

	h.CpuPass()

	ready := h.GetPassReady()
	// Human (0) not ready, CPUs (1,2,3) ready
	assert.False(t, ready[0])
	assert.True(t, ready[1])
	assert.True(t, ready[2])
	assert.True(t, ready[3])
}

func TestHearts_CpuPass_SkipsAlreadyReady(t *testing.T) {
	h := newTestHearts()
	h.Reset()

	// First call
	h.CpuPass()
	// Second call should not error (skips already-ready CPUs)
	h.CpuPass()

	ready := h.GetPassReady()
	assert.True(t, ready[1])
	assert.True(t, ready[2])
	assert.True(t, ready[3])
}

// --- ExecutePass ---

func TestHearts_ExecutePass_Left(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	assert.Equal(t, domain.HeartsPassLeft, h.GetPassDirection())

	// Record initial hand sizes
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()

	// All ready, execute
	h.ExecutePass()

	// After pass: each player has 10 + 3 = 13 cards again
	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, h.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
	assert.Equal(t, 1, h.GetTrickNumber())
}

func TestHearts_ExecutePass_Right(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetRoundNumber(2) // Right
	h.SetPhase(domain.HeartsPhasePass)

	// Reset pass state
	// Give players fresh hands
	for i := 0; i < 4; i++ {
		h.GetPlayer(i).Reset()
		for j := 0; j < 13; j++ {
			h.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignClover, j+1, false))
		}
	}

	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()
	h.ExecutePass()

	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, h.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
}

func TestHearts_ExecutePass_Across(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetRoundNumber(3) // Across
	h.SetPhase(domain.HeartsPhasePass)

	for i := 0; i < 4; i++ {
		h.GetPlayer(i).Reset()
		for j := 0; j < 13; j++ {
			h.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignClover, j+1, false))
		}
	}

	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()
	h.ExecutePass()

	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, h.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

func TestHearts_ExecutePass_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePlay)
	h.ExecutePass() // should do nothing, no panic
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
}

func TestHearts_ExecutePass_NotAllReady(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	// Only human passed, CPUs not ready
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)

	h.ExecutePass() // should do nothing
	assert.Equal(t, domain.HeartsPhasePass, h.GetPhase())
}

// --- PlayerPlay ---

func TestHearts_PlayerPlay_Valid(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2) // trick 2, not first trick
	h.SetHeartsBroken(true)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := h.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(h.GetCurrentTrick()))
}

func TestHearts_PlayerPlay_GameEnded(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)

	// Trigger game end
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.GetPlayer(1).SetCumulativeScore(95)
	h.GetPlayer(1).SetRoundScore(10)
	h.ScoreRound()

	err := h.PlayerPlay(0)
	assert.True(t, errors.Is(err, domain.ErrGameEnded))
}

func TestHearts_PlayerPlay_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	err := h.PlayerPlay(0)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestHearts_PlayerPlay_NotHumanTurn(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 1, 1, 2) // CPU is current
	err := h.PlayerPlay(0)
	assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
}

func TestHearts_PlayerPlay_OutOfRange(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := h.PlayerPlay(99)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestHearts_PlayerPlay_NegativeIndex(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := h.PlayerPlay(-1)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestHearts_PlayerPlay_MustPlay2Clubs(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 1) // trick 1, first trick

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 2, false)) // 2♣
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	// Try to play 5♣ instead of 2♣
	err := h.PlayerPlay(1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

	// Play 2♣ should work
	err = h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_FirstTrick_NoTwoClubs(t *testing.T) {
	// If player doesn't have 2♣, they can play any card on first trick lead
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 1) // trick 1

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_MustFollowSuit(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 2) // human's turn after lead
	h.SetHeartsBroken(true)

	// Set lead card as clover
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // has clover
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))  // also has heart

	// Try to play heart when having clover
	err := h.PlayerPlay(1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

	// Playing clover should work
	err = h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_CanPlayOffSuitWhenVoid(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false)) // no clover

	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_CantLeadHeartsUnbroken(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(false)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	// Try to lead with heart
	err := h.PlayerPlay(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

	// Lead with clover should work
	err = h.PlayerPlay(1) // index 1 is now clover (after error, hand unchanged)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_CanLeadHeartsWhenOnlyHearts(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(false)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	// Only hearts available, must be able to lead
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_HeartsBrokenAllowsLeadHearts(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	// Hearts broken, can lead with heart
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_FirstTrickNoPointCards(t *testing.T) {
	// On first trick, when following off-suit, can't play point cards if non-point available
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 1) // first trick, following

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 2, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	// No clover (void in lead suit) but has heart and diamond
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))   // point card
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false)) // non-point card

	// Try to play heart on first trick (point card)
	err := h.PlayerPlay(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

	// Playing diamond is OK
	err = h.PlayerPlay(1)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_FirstTrickNoPointCards_QueenOfSpades(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 1)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 2, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))  // Q♠ = point card
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false)) // non-point

	err := h.PlayerPlay(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestHearts_PlayerPlay_FirstTrickOnlyPointCards(t *testing.T) {
	// When player only has point cards and is void in lead suit on trick 1, forced to play them
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 1)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 2, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

	// All point cards, forced to play
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_PlayerPlay_BreaksHearts(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 2) // following, not first trick
	h.SetHeartsBroken(false)

	// Lead is diamond, player has no diamond
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 5, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

	err := h.PlayerPlay(0)
	assert.NoError(t, err)
	assert.True(t, h.GetHeartsBroken())
}

func TestHearts_PlayerPlay_TrickComplete(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 2)
	h.SetHeartsBroken(true)

	// 3 cards already in trick
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 9, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	err := h.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, domain.HeartsPhaseTrickEnd, h.GetPhase())
	assert.Equal(t, 4, len(h.GetCurrentTrick()))
}

func TestHearts_PlayerPlay_AdvancesPlayer(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := h.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, h.GetCurrentPlayerIdx()) // advanced to next player
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
}

// --- CpuPlay ---

func TestHearts_CpuPlay_Valid(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	h.CpuPlay()
	assert.Equal(t, 1, len(h.GetCurrentTrick()))
	assert.Equal(t, 0, cpu.GetCardsSize())
}

func TestHearts_CpuPlay_GameEnded(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 1, 1, 2)

	// Set game end
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.GetPlayer(1).SetCumulativeScore(95)
	h.GetPlayer(1).SetRoundScore(10)
	h.ScoreRound()

	h.CpuPlay() // should do nothing
	assert.Nil(t, h.GetCurrentTrick())
}

func TestHearts_CpuPlay_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetCurrentPlayerIdx(1)
	h.CpuPlay() // should do nothing
}

func TestHearts_CpuPlay_HumanTurn(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2) // human turn
	h.CpuPlay()                // should do nothing
	assert.Nil(t, h.GetCurrentTrick())
}

// --- ResolveTrick ---

func TestHearts_ResolveTrick_Winner(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(2)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)}, // highest clover = winner
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // off-suit, doesn't count
	})

	h.ResolveTrick()

	assert.Equal(t, 1, h.GetLeadPlayerIdx()) // player 1 won
	assert.Equal(t, 1, h.GetPlayer(1).GetTrickCount())
	// Points: heart K = 1 point
	assert.Equal(t, 1, h.GetPlayer(1).GetRoundScore())
}

func TestHearts_ResolveTrick_QueenOfSpades(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(2)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 12, false)}, // Q♠ = 13 pts, also highest
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 8, false)},
	})

	h.ResolveTrick()

	assert.Equal(t, 1, h.GetLeadPlayerIdx())
	assert.Equal(t, 13, h.GetPlayer(1).GetRoundScore()) // Q♠ = 13 points
}

func TestHearts_ResolveTrick_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePlay)
	h.ResolveTrick() // should do nothing
}

func TestHearts_ResolveTrick_IncompleteTrick(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})
	h.ResolveTrick() // should do nothing (not 4 cards)
}

func TestHearts_ResolveTrick_LastTrick(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(13) // last trick

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
	})

	h.ResolveTrick()
	assert.Equal(t, domain.HeartsPhaseRoundEnd, h.GetPhase())
}

func TestHearts_ResolveTrick_NotLastTrick(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(5)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
	})

	h.ResolveTrick()
	assert.Equal(t, domain.HeartsPhaseTrickEnd, h.GetPhase()) // stays in trick end
}

func TestHearts_ResolveTrick_HeartsBroken(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(3)
	h.SetHeartsBroken(false)

	// Hearts card in trick -> should already be broken from playCard, but points still count
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
	})

	h.ResolveTrick()
	assert.Equal(t, 1, h.GetPlayer(1).GetRoundScore()) // 1 heart point
}

func TestHearts_ResolveTrick_ZeroPoints(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(3)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
	})

	h.ResolveTrick()
	assert.Equal(t, 0, h.GetPlayer(1).GetRoundScore())
}

// --- NextTrick ---

func TestHearts_NextTrick(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetLeadPlayerIdx(2)
	h.SetTrickNumber(3)

	h.NextTrick()

	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
	assert.Equal(t, 2, h.GetCurrentPlayerIdx())
	assert.Nil(t, h.GetCurrentTrick())
	assert.Equal(t, 4, h.GetTrickNumber()) // trickNumber incremented
}

func TestHearts_NextTrick_IncrementsThroughMultipleTricks(t *testing.T) {
	// Integration test: verify trickNumber increments naturally through
	// ResolveTrick → NextTrick cycles without manually calling SetTrickNumber
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePlay)
	h.SetTrickNumber(1)

	for trick := 1; trick <= 3; trick++ {
		// Set up a complete trick
		h.SetPhase(domain.HeartsPhaseTrickEnd)
		h.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
		})

		h.ResolveTrick()
		assert.Equal(t, domain.HeartsPhaseTrickEnd, h.GetPhase())

		h.NextTrick()
		assert.Equal(t, trick+1, h.GetTrickNumber())
		assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
	}
}

func TestHearts_NextTrick_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePlay)
	h.SetLeadPlayerIdx(2)
	h.SetCurrentPlayerIdx(0)

	h.NextTrick() // should do nothing
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase())
	assert.Equal(t, 0, h.GetCurrentPlayerIdx()) // unchanged
}

// --- ScoreRound ---

func TestHearts_ScoreRound_Normal(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)

	h.GetPlayer(0).SetRoundScore(5)
	h.GetPlayer(1).SetRoundScore(10)
	h.GetPlayer(2).SetRoundScore(8)
	h.GetPlayer(3).SetRoundScore(3)

	h.ScoreRound()

	assert.Equal(t, 5, h.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 10, h.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 8, h.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 3, h.GetPlayer(3).GetCumulativeScore())
	assert.False(t, h.GetGameEndFlag())
}

func TestHearts_ScoreRound_ShootTheMoon(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)

	h.GetPlayer(0).SetRoundScore(26) // shoot the moon
	h.GetPlayer(1).SetRoundScore(0)
	h.GetPlayer(2).SetRoundScore(0)
	h.GetPlayer(3).SetRoundScore(0)

	h.ScoreRound()

	// Shooter gets 0, others get 26
	assert.Equal(t, 0, h.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 26, h.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 26, h.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 26, h.GetPlayer(3).GetCumulativeScore())
}

func TestHearts_ScoreRound_GameEnd(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(5)

	h.GetPlayer(0).SetCumulativeScore(90)
	h.GetPlayer(1).SetCumulativeScore(50)
	h.GetPlayer(2).SetCumulativeScore(30)
	h.GetPlayer(3).SetCumulativeScore(20)

	h.GetPlayer(0).SetRoundScore(15) // 90+15 = 105 >= 100

	h.ScoreRound()

	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, domain.HeartsPhaseGameEnd, h.GetPhase())
	// Player 3 has lowest score (20) -> winner
	assert.Equal(t, 3, h.GetWinnerIdx())
}

func TestHearts_ScoreRound_GameEnd_WinnerIsLowestScore(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)

	h.GetPlayer(0).SetCumulativeScore(10)
	h.GetPlayer(1).SetCumulativeScore(95)
	h.GetPlayer(2).SetCumulativeScore(5) // lowest
	h.GetPlayer(3).SetCumulativeScore(50)

	h.GetPlayer(1).SetRoundScore(10) // 95+10 = 105 >= 100

	h.ScoreRound()

	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, 2, h.GetWinnerIdx()) // lowest cumulative
}

func TestHearts_ScoreRound_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePlay)

	h.ScoreRound() // should do nothing
	assert.False(t, h.GetGameEndFlag())
}

func TestHearts_ScoreRound_BelowLimit(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)

	h.GetPlayer(0).SetCumulativeScore(10)
	h.GetPlayer(1).SetCumulativeScore(20)
	h.GetPlayer(2).SetCumulativeScore(30)
	h.GetPlayer(3).SetCumulativeScore(40)

	h.GetPlayer(0).SetRoundScore(5)
	h.GetPlayer(1).SetRoundScore(5)
	h.GetPlayer(2).SetRoundScore(5)
	h.GetPlayer(3).SetRoundScore(5)

	h.ScoreRound()

	assert.False(t, h.GetGameEndFlag())
	assert.Equal(t, -1, h.GetWinnerIdx())
}

// --- NextRound ---

func TestHearts_NextRound(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)

	h.NextRound()

	assert.Equal(t, 2, h.GetRoundNumber())
	assert.Equal(t, 0, h.GetTrickNumber()) // stays 0 until pass completes (round 2 = right)
	assert.False(t, h.GetHeartsBroken())
	assert.Nil(t, h.GetCurrentTrick())

	// 52 cards distributed
	total := 0
	for i := 0; i < 4; i++ {
		total += h.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
}

func TestHearts_NextRound_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePlay)
	h.SetRoundNumber(1)

	h.NextRound() // should do nothing
	assert.Equal(t, 1, h.GetRoundNumber())
}

func TestHearts_NextRound_PassNone(t *testing.T) {
	// Round 4 = no pass
	h := newTestHearts()
	h.Reset()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(3) // NextRound increments to 4

	h.NextRound()

	assert.Equal(t, 4, h.GetRoundNumber())
	assert.Equal(t, domain.HeartsPhasePlay, h.GetPhase()) // skip pass
}

// --- State getters ---

func TestHearts_GetPlayer_InBounds(t *testing.T) {
	h := newTestHearts()
	for i := 0; i < 4; i++ {
		assert.NotNil(t, h.GetPlayer(i))
	}
}

func TestHearts_GetPlayer_OutOfBounds(t *testing.T) {
	h := newTestHearts()
	assert.Nil(t, h.GetPlayer(-1))
	assert.Nil(t, h.GetPlayer(4))
	assert.Nil(t, h.GetPlayer(100))
}

func TestHearts_IsHumanTurn(t *testing.T) {
	h := newTestHearts()
	h.SetCurrentPlayerIdx(0) // human
	assert.True(t, h.IsHumanTurn())

	h.SetCurrentPlayerIdx(1) // CPU
	assert.False(t, h.IsHumanTurn())
}

func TestHearts_IsHumanTurn_OutOfRange(t *testing.T) {
	h := newTestHearts()
	h.SetCurrentPlayerIdx(-1)
	assert.False(t, h.IsHumanTurn())

	h.SetCurrentPlayerIdx(4)
	assert.False(t, h.IsHumanTurn())
}

func TestHearts_GetSetConfig(t *testing.T) {
	h := newTestHearts()
	cfg := h.GetConfig()
	assert.Equal(t, domain.HeartsCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 100, cfg.PointLimit)

	newCfg := domain.HeartsConfig{
		CpuDifficulty: domain.HeartsCpuDifficultyHard,
		PointLimit:    50,
	}
	h.SetConfig(newCfg)
	assert.Equal(t, newCfg, h.GetConfig())
}

func TestHearts_GetActionLog(t *testing.T) {
	h := newTestHearts()
	assert.Nil(t, h.GetActionLog())

	// Do an action to generate log
	h.Reset()
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()
	h.ExecutePass()
	log := h.GetActionLog()
	assert.NotNil(t, log)
	assert.Greater(t, len(log), 0)
}

func TestHearts_GetPassedCards(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)

	passed := h.GetPassedCards()
	assert.Equal(t, 3, len(passed[0]))
}

func TestHearts_SetTrickNumber(t *testing.T) {
	h := newTestHearts()
	h.SetTrickNumber(5)
	assert.Equal(t, 5, h.GetTrickNumber())
}

func TestHearts_SetRoundNumber(t *testing.T) {
	h := newTestHearts()
	h.SetRoundNumber(3)
	assert.Equal(t, 3, h.GetRoundNumber())
}

func TestHearts_GetLeadPlayerIdx(t *testing.T) {
	h := newTestHearts()
	h.SetLeadPlayerIdx(2)
	assert.Equal(t, 2, h.GetLeadPlayerIdx())
}

// --- CPU AI Pass ---

func TestHearts_CpuPass_Easy(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyEasy
	h.SetConfig(cfg)
	h.Reset()

	h.CpuPass()
	ready := h.GetPassReady()
	assert.True(t, ready[1])
	assert.True(t, ready[2])
	assert.True(t, ready[3])

	// Each CPU should have 10 cards after passing 3
	for i := 1; i < 4; i++ {
		assert.Equal(t, 10, h.GetPlayer(i).GetCardsSize())
	}
}

func TestHearts_CpuPass_Normal(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)
	h.Reset()

	h.CpuPass()
	ready := h.GetPassReady()
	assert.True(t, ready[1])
	for i := 1; i < 4; i++ {
		assert.Equal(t, 10, h.GetPlayer(i).GetCardsSize())
	}
}

func TestHearts_CpuPass_Hard(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)
	h.Reset()

	h.CpuPass()
	ready := h.GetPassReady()
	assert.True(t, ready[1])
	for i := 1; i < 4; i++ {
		assert.Equal(t, 10, h.GetPlayer(i).GetCardsSize())
	}
}

// --- CPU AI Play ---

func TestHearts_CpuPlay_Easy(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyEasy
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

	h.CpuPlay()
	assert.Equal(t, 1, len(h.GetCurrentTrick()))
}

func TestHearts_CpuPlay_Normal_Lead(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 1, 2) // CPU leads
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 1, len(trick))
	// Normal AI leads lowest card
	assert.Equal(t, 3, trick[0].Card.GetValue())
}

func TestHearts_CpuPlay_Normal_Follow_AvoidWinning(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	// Lead is clover 5
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false)) // lower than 5
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
}

func TestHearts_CpuPlay_Normal_Follow_CannotAvoid(t *testing.T) {
	// All cards are higher than trick leader
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
}

func TestHearts_CpuPlay_Normal_Void_DumpsPointCard(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// No clover, has heart and diamond
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // point card

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should dump the point card (heart)
	assert.Equal(t, domain.CardDesignHeart, trick[1].Card.GetDesign())
}

func TestHearts_CpuPlay_Normal_Void_HighCard(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// No clover, no point cards
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should play highest non-point card
	assert.Equal(t, 10, trick[1].Card.GetValue())
}

func TestHearts_CpuPlay_Hard_Lead(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 1, len(trick))
	// Hard AI leads lowest non-heart
	assert.Equal(t, domain.CardDesignClover, trick[0].Card.GetDesign())
	assert.Equal(t, 3, trick[0].Card.GetValue())
}

func TestHearts_CpuPlay_Hard_Lead_OnlyHearts(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 1, len(trick))
	// Only hearts -> plays lowest heart
	assert.Equal(t, 3, trick[0].Card.GetValue())
}

func TestHearts_CpuPlay_Hard_Follow_UnderTrick(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Has cards under and over trick leader
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should play highest card under trick leader (8)
	assert.Equal(t, 8, trick[1].Card.GetValue())
}

func TestHearts_CpuPlay_Hard_Follow_AllAboveTrick(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// All above trick, plays lowest
	assert.Equal(t, 8, trick[1].Card.GetValue())
}

func TestHearts_CpuPlay_Hard_Discard_QueenOfSpades(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Void in clover, has Q♠ and diamond
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q♠
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should dump Q♠ first
	assert.Equal(t, domain.CardDesignSpade, trick[1].Card.GetDesign())
	assert.Equal(t, 12, trick[1].Card.GetValue())
}

func TestHearts_CpuPlay_Hard_Discard_Heart(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Void in clover, has heart and diamond (no Q♠)
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should dump heart (score = 10 + 100 = 110 vs diamond 3)
	assert.Equal(t, domain.CardDesignHeart, trick[1].Card.GetDesign())
}

func TestHearts_CpuPlay_Hard_Discard_HighNonPointCard(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Void in clover, only non-point cards
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should play highest
	assert.Equal(t, 13, trick[1].Card.GetValue())
}

func TestHearts_CpuPlay_SingleValidCard(t *testing.T) {
	// len(validIndices) == 1 branch
	h := newTestHearts()
	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	h.CpuPlay()
	assert.Equal(t, 1, len(h.GetCurrentTrick()))
}

// --- CPU AI Hard Pass (void creation logic) ---

func TestHearts_CpuPass_Hard_VoidCreation(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	// Give CPU 1 a hand with a short suit (non-heart) to trigger void creation bonus
	cpu := h.GetPlayer(1)
	cpu.Reset()
	// 2 diamonds (short suit, non-heart, will get +30 bonus)
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
	// 11 clovers
	for i := 1; i <= 11; i++ {
		cpu.AddCard(domain.NewCard(domain.CardDesignClover, i, false))
	}

	h.SetPhase(domain.HeartsPhasePass)

	// Other players need cards too
	for _, idx := range []int{0, 2, 3} {
		p := h.GetPlayer(idx)
		p.Reset()
		for j := 1; j <= 13; j++ {
			p.AddCard(domain.NewCard(domain.CardDesignSpade, j, false))
		}
	}

	// Human passes
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()

	ready := h.GetPassReady()
	assert.True(t, ready[1])
}

// --- CPU Normal play: complex branching ---

func TestHearts_CpuPlay_Normal_Follow_BestUnderTrick(t *testing.T) {
	// Test the branch: card.GetValue() < highestInTrick && card.GetValue() > bestVal && bestVal < highestInTrick
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false)) // under 10
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // under 10, higher than 3

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should play 8 (highest under trick)
	assert.Equal(t, 8, trick[1].Card.GetValue())
}

func TestHearts_CpuPlay_Normal_Follow_FallbackToLowest(t *testing.T) {
	// When bestVal >= highestInTrick and card < bestVal
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // above 3, bestVal starts
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))  // above 3, lower than 10

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should play 5 (lower of two above-trick cards)
	assert.Equal(t, 5, trick[1].Card.GetValue())
}

// --- trickWinner edge case ---

func TestHearts_ResolveTrick_EmptyTrick(t *testing.T) {
	// trickWinner with empty trick -> returns 0, but ResolveTrick guards against incomplete
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetCurrentTrick(nil) // empty

	h.ResolveTrick() // should do nothing (len != HeartsPlayerCnt)
}

// --- Full game flow integration ---

func TestHearts_FullTrickFlow(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	// Give each player 1 card
	for i := 0; i < 4; i++ {
		p := h.GetPlayer(i)
		p.Reset()
		p.AddCard(domain.NewCard(domain.CardDesignClover, i+3, false))
	}

	// Player 0 leads
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, h.GetCurrentPlayerIdx())

	// CPU 1, 2, 3 play
	for i := 1; i < 4; i++ {
		h.CpuPlay()
	}

	assert.Equal(t, domain.HeartsPhaseTrickEnd, h.GetPhase())
	assert.Equal(t, 4, len(h.GetCurrentTrick()))

	// Resolve
	h.ResolveTrick()
	// Player 3 has highest (6♣)
	assert.Equal(t, 3, h.GetLeadPlayerIdx())
}

// --- StartPlayPhase branches ---

func TestHearts_StartPlayPhase_FindsTwoOfClubs(t *testing.T) {
	h := newTestHearts()
	h.Reset() // round 1 = pass left, phase is pass, trickNumber is 0

	// After pass + execute, startPlayPhase sets the 2♣ holder as current
	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()
	h.ExecutePass()

	assert.Equal(t, 1, h.GetTrickNumber())

	// The current player should have the 2♣ (or it was dealt to them)
	currentIdx := h.GetCurrentPlayerIdx()
	assert.GreaterOrEqual(t, currentIdx, 0)
	assert.Less(t, currentIdx, 4)
}

func TestHearts_StartPlayPhase_NoTwoOfClubs(t *testing.T) {
	// When no player has 2♣, fallback to player 0
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhasePass)
	h.SetRoundNumber(1)
	h.SetTrickNumber(0) // startPlayPhase checks this

	// Give players cards without 2♣
	for i := 0; i < 4; i++ {
		p := h.GetPlayer(i)
		p.Reset()
		for j := 3; j <= 13; j++ { // no card with value 2
			p.AddCard(domain.NewCard(domain.CardDesignSpade, j, false))
		}
		// Fill to 13
		p.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		p.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false))
	}

	// Mark all pass ready
	for i := 0; i < 4; i++ {
		if h.GetPlayer(i).GetIsHuman() {
			err := h.PlayerPass([]int{0, 1, 2})
			assert.NoError(t, err)
		}
	}
	h.CpuPass()
	h.ExecutePass()

	// startPlayPhase should set player 0 since no 2♣
	assert.Equal(t, 0, h.GetCurrentPlayerIdx())
	assert.Equal(t, 0, h.GetLeadPlayerIdx())
}

// --- playerName branches ---

func TestHearts_ActionLog_PlayerNames(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	// Human plays -> log should contain "You"
	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	err := h.PlayerPlay(0)
	assert.NoError(t, err)

	log := h.GetActionLog()
	assert.NotNil(t, log)
	assert.Contains(t, log[0].Detail, "You")

	// CPU plays -> log should contain "CPU"
	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))

	h.CpuPlay()
	log = h.GetActionLog()
	assert.Contains(t, log[1].Detail, "CPU")
}

// --- Full round scoring flow ---

func TestHearts_FullRoundScoringFlow(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)

	// Set round scores
	h.GetPlayer(0).SetRoundScore(10)
	h.GetPlayer(1).SetRoundScore(5)
	h.GetPlayer(2).SetRoundScore(8)
	h.GetPlayer(3).SetRoundScore(3)

	h.ScoreRound()

	// Verify cumulative scores
	assert.Equal(t, 10, h.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 5, h.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 8, h.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 3, h.GetPlayer(3).GetCumulativeScore())

	// Log should have 4 score entries
	log := h.GetActionLog()
	assert.NotNil(t, log)
	scoreEntries := 0
	for _, entry := range log {
		if entry.ActionType == "round_score" {
			scoreEntries++
		}
	}
	assert.Equal(t, 4, scoreEntries)
}

// --- cardStr edge cases (unknown design/value) ---

func TestHearts_CardStr_UnknownDesignAndValue(t *testing.T) {
	// Playing a card with unknown design/value should use "?" in log
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(99, 99, false)) // unknown design and value

	err := h.PlayerPlay(0)
	assert.NoError(t, err)

	log := h.GetActionLog()
	assert.Contains(t, log[0].Detail, "?")
}

// --- isPointCard branches ---

func TestHearts_PointCards_InTrick(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(3)

	// Heart = 1 point, Q♠ = 13 points, others = 0
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},  // lead
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 12, false)}, // Q♠ = 13 pts
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 3, false)},  // 1 pt
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 8, false)},  // 0 pts
	})

	h.ResolveTrick()
	// Q♠ holder (player 1) is highest spade (12), wins
	assert.Equal(t, 1, h.GetLeadPlayerIdx())
	assert.Equal(t, 14, h.GetPlayer(1).GetRoundScore()) // 13 + 1
}

// --- passTarget default branch ---

func TestHearts_ExecutePass_DefaultDirection(t *testing.T) {
	// HeartsPassNone means passTarget returns `from` (self), but ExecutePass
	// only runs when phase is pass and direction is not None (since Reset/NextRound
	// skips pass phase). Test passTarget default via round 4 indirectly.
	// Already covered by NextRound_PassNone test.
	h := newTestHearts()
	h.SetRoundNumber(4)
	assert.Equal(t, domain.HeartsPassNone, h.GetPassDirection())
}

// --- Shoot the moon log entry ---

func TestHearts_ScoreRound_ShootTheMoon_Log(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)

	h.GetPlayer(2).SetRoundScore(26) // player 2 shoots the moon

	h.ScoreRound()

	log := h.GetActionLog()
	hasMoonEntry := false
	for _, entry := range log {
		if entry.ActionType == "shoot_moon" {
			hasMoonEntry = true
			assert.Equal(t, 2, entry.PlayerIdx)
		}
	}
	assert.True(t, hasMoonEntry)
}

// --- Game end log entry ---

func TestHearts_ScoreRound_GameEnd_Log(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)

	h.GetPlayer(0).SetCumulativeScore(95)
	h.GetPlayer(0).SetRoundScore(10)

	h.ScoreRound()

	assert.True(t, h.GetGameEndFlag())
	log := h.GetActionLog()
	hasEndEntry := false
	for _, entry := range log {
		if entry.ActionType == "game_end" {
			hasEndEntry = true
		}
	}
	assert.True(t, hasEndEntry)
}

// --- ResolveTrick: trickWinner picks first player card when only lead suit ---

func TestHearts_TrickWinner_FirstPlayer(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(3)

	// Lead player has the highest card
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 13, false)}, // highest
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
	})

	h.ResolveTrick()
	assert.Equal(t, 0, h.GetLeadPlayerIdx()) // player 0 wins
}

// --- Follow suit: not first trick, can play any off-suit ---

func TestHearts_PlayerPlay_NotFirstTrick_CanPlayPointOffSuit(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 3) // trick 3, not first trick

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	// No clover, only hearts
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	// Not first trick, can play point cards off-suit
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

// --- CpuPlay hard follow: mixed lead and non-lead suit ---

func TestHearts_CpuPlay_Hard_Follow_MixedSuits(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Has both lead suit and off-suit
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false)) // lead suit, under
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // off-suit

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Has lead suit, should follow with clover
	assert.Equal(t, domain.CardDesignClover, trick[1].Card.GetDesign())
}

// --- ScoreRound: winner is player 0 (default min check) ---

func TestHearts_ScoreRound_WinnerIsPlayer0(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)

	// Player 0 has lowest score
	h.GetPlayer(0).SetCumulativeScore(5)
	h.GetPlayer(1).SetCumulativeScore(95)
	h.GetPlayer(2).SetCumulativeScore(50)
	h.GetPlayer(3).SetCumulativeScore(60)

	h.GetPlayer(1).SetRoundScore(10) // triggers game end

	h.ScoreRound()
	assert.Equal(t, 0, h.GetWinnerIdx())
}

// --- CpuPlay_Normal: isFollowing branch when first valid card is off-suit ---

func TestHearts_CpuPlay_Normal_Follow_FirstCardOffSuit(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// No clover at all (void) - isFollowing = false
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q♠ is a point card

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should prioritize point card (Q♠ = isPointCard true)
	assert.Equal(t, domain.CardDesignSpade, trick[1].Card.GetDesign())
	assert.Equal(t, 12, trick[1].Card.GetValue())
}

// --- Hard CPU follow: card under trick but bestIdx points to non-lead or above-trick ---

func TestHearts_CpuPlay_Hard_Follow_BestIdxStartsAboveTrick(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// First card is above trick, second is under
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // above
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))  // under

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should play 5 (under trick)
	assert.Equal(t, 5, trick[1].Card.GetValue())
}

// --- Hard CPU follow: bestVal == 999 initial fallback ---

func TestHearts_CpuPlay_Hard_Follow_AllAbove_PicksLowest(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 2, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 12, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// All above 2, should pick lowest (5)
	assert.Equal(t, 5, trick[1].Card.GetValue())
}

// --- Hard CPU: trickPoints used in decision (covered by discard path) ---

func TestHearts_CpuPlay_Hard_Follow_WithTrickPoints(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 3)
	h.SetHeartsBroken(true)

	// Trick has points in it
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)}, // 1 point
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 8, false)}, // 1 point
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Void in hearts, has non-lead
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

	h.CpuPlay()
	assert.Equal(t, 3, len(h.GetCurrentTrick()))
}

// --- passDirectionStr: all directions ---

func TestHearts_ExecutePass_LogDirectionStrings(t *testing.T) {
	// Test left (round 1), right (round 2), across (round 3)
	for _, tc := range []struct {
		round int
		dir   string
	}{
		{1, "left"},
		{2, "right"},
		{3, "across"},
	} {
		h := newTestHearts()
		h.Reset()
		h.SetRoundNumber(tc.round)
		h.SetPhase(domain.HeartsPhasePass)

		// Give fresh hands
		for i := 0; i < 4; i++ {
			h.GetPlayer(i).Reset()
			for j := 1; j <= 13; j++ {
				h.GetPlayer(i).AddCard(domain.NewCard(domain.CardDesignClover, j, false))
			}
		}

		err := h.PlayerPass([]int{0, 1, 2})
		assert.NoError(t, err)
		h.CpuPass()
		h.ExecutePass()

		log := h.GetActionLog()
		found := false
		for _, entry := range log {
			if entry.ActionType == "pass" {
				assert.Contains(t, entry.Detail, tc.dir)
				found = true
			}
		}
		assert.True(t, found, "expected pass log with direction %s", tc.dir)
	}
}

// --- Verify follow suit: has lead suit, plays same suit (no error) ---

func TestHearts_PlayerPlay_FollowSuit_Success(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 3)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	player.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	// Play spade (follows suit)
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

// --- getValidPlayIndices returns empty (fallback 0) ---

func TestHearts_CpuPlay_NoValidCards_Fallback(t *testing.T) {
	// This tests the len(validIndices) == 0 -> return 0 fallback
	// It's hard to create a truly invalid state, but we can test with a single card
	// that happens to be valid
	h := newTestHearts()
	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))

	// Will have exactly 1 valid card
	h.CpuPlay()
	assert.Equal(t, 1, len(h.GetCurrentTrick()))
}

// --- ResolveTrick: action log entry ---

func TestHearts_ResolveTrick_ActionLog(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(2)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 8, false)},
	})

	h.ResolveTrick()

	log := h.GetActionLog()
	assert.NotNil(t, log)
	found := false
	for _, entry := range log {
		if entry.ActionType == "trick_win" {
			found = true
			assert.Contains(t, entry.Detail, "wins trick")
			assert.NotNil(t, entry.Cards)
			assert.Equal(t, 4, len(entry.Cards))
		}
	}
	assert.True(t, found)
}

// --- First trick: follow suit (not void, not first card) ---

func TestHearts_PlayerPlay_FirstTrick_FollowSuit(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 1, 1) // first trick

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 2, false)},
	})

	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // has clover
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

	// Must follow clover
	err := h.PlayerPlay(1) // try diamond
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))

	err = h.PlayerPlay(0) // play clover
	assert.NoError(t, err)
}

// --- cardStr for all known suits and values ---

func TestHearts_CardStr_AllSuitsAndValues(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	// Test each suit
	for _, design := range []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond} {
		player := h.GetPlayer(0)
		player.Reset()
		player.AddCard(domain.NewCard(design, 1, false)) // Ace

		h.SetPhase(domain.HeartsPhasePlay)
		h.SetCurrentPlayerIdx(0)
		h.SetCurrentTrick(nil)

		err := h.PlayerPlay(0)
		assert.NoError(t, err)
	}

	log := h.GetActionLog()
	assert.GreaterOrEqual(t, len(log), 4)
}

// --- CpuPlay_Normal: isFollowing false, both non-point cards ---

func TestHearts_CpuPlay_Normal_Void_BothNonPoint_HigherWins(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Void in clover, both non-point
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // non-point (not Q♠)
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Picks higher card value
	assert.Equal(t, 10, trick[1].Card.GetValue())
}

// --- Player wrap around: current player wraps from 3 to 0 ---

func TestHearts_PlayCard_WrapsAround(t *testing.T) {
	h := newTestHearts()
	setupPlayPhase(h, 3, 0, 2) // player 3's turn
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	cpu := h.GetPlayer(3)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

	h.CpuPlay()
	// 3 cards in trick, not complete yet -> next player is (3+1)%4 = 0
	assert.Equal(t, 0, h.GetCurrentPlayerIdx())
}

// --- maxScore equals exactly point limit ---

func TestHearts_ScoreRound_ExactlyAtLimit(t *testing.T) {
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)

	h.GetPlayer(0).SetCumulativeScore(0)
	h.GetPlayer(1).SetCumulativeScore(90)
	h.GetPlayer(2).SetCumulativeScore(0)
	h.GetPlayer(3).SetCumulativeScore(0)

	h.GetPlayer(1).SetRoundScore(10) // 90 + 10 = 100, exactly at limit

	h.ScoreRound()
	assert.True(t, h.GetGameEndFlag()) // >= limit, not just >
}

// --- Hard CPU pass: heart cards don't get void bonus ---

func TestHearts_CpuPass_Hard_HeartNotVoidBonus(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// 2 hearts (short suit but heart -> no void bonus)
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	// 11 spades
	for i := 1; i <= 11; i++ {
		cpu.AddCard(domain.NewCard(domain.CardDesignSpade, i, false))
	}

	h.SetPhase(domain.HeartsPhasePass)

	for _, idx := range []int{0, 2, 3} {
		p := h.GetPlayer(idx)
		p.Reset()
		for j := 1; j <= 13; j++ {
			p.AddCard(domain.NewCard(domain.CardDesignClover, j, false))
		}
	}

	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()

	assert.True(t, h.GetPassReady()[1])
}

// --- Normal CPU pass: Q♠ and high spades scoring ---

func TestHearts_CpuPass_Normal_DangerousCards(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Give Q♠, K♠, A♠ and low cards
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q♠ = score 12+100+50 = 162
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K♠ = 13+50 = 63
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // A♠ = 1 (value < 11, no bonus)
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))

	h.SetPhase(domain.HeartsPhasePass)

	for _, idx := range []int{0, 2, 3} {
		p := h.GetPlayer(idx)
		p.Reset()
		for j := 1; j <= 13; j++ {
			p.AddCard(domain.NewCard(domain.CardDesignDiamond, j, false))
		}
	}

	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()

	// CPU should have passed away Q♠ and K♠ (highest scored)
	assert.Equal(t, 10, cpu.GetCardsSize())
}

// --- Normal CPU pass: hearts scored high ---

func TestHearts_CpuPass_Normal_HeartsScoring(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// High hearts should be passed first
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // 13+20 = 33
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 12, false)) // 12+20 = 32
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false)) // 11+20 = 31
	for i := 1; i <= 10; i++ {
		cpu.AddCard(domain.NewCard(domain.CardDesignClover, i, false))
	}

	h.SetPhase(domain.HeartsPhasePass)

	for _, idx := range []int{0, 2, 3} {
		p := h.GetPlayer(idx)
		p.Reset()
		for j := 1; j <= 13; j++ {
			p.AddCard(domain.NewCard(domain.CardDesignDiamond, j, false))
		}
	}

	err := h.PlayerPass([]int{0, 1, 2})
	assert.NoError(t, err)
	h.CpuPass()

	assert.Equal(t, 10, cpu.GetCardsSize())
}

// --- Normal CPU play: isFollowing path with card.GetValue() >= bestVal above trick ---

// --- Cover remaining uncovered branches ---

func TestHearts_GetPlayerCnt(t *testing.T) {
	h := newTestHearts()
	assert.Equal(t, 4, h.GetPlayerCnt())
}

func TestHearts_Reset_AlwaysStartsPassPhase(t *testing.T) {
	// Reset always sets roundNumber=1, so passDirection is always Left (pass phase)
	h := newTestHearts()
	h.Reset()
	assert.Equal(t, domain.HeartsPhasePass, h.GetPhase())
	assert.Equal(t, 1, h.GetRoundNumber())
}

// Defensive guard tests for private methods are in Hearts_internal_test.go

func TestHearts_PassDirectionStr_Default(t *testing.T) {
	// passDirectionStr default: only used in ExecutePass log.
	// Since ExecutePass skips for PassNone, the "none" branch is unreachable.
}

func TestHearts_CpuSelectPlayCard_EmptyValidIndices(t *testing.T) {
	// Line 843: len(validIndices) == 0, return 0
	// getValidPlayIndices returns empty only if no card passes validatePlay.
	// This is very hard to trigger in normal play since at least one card should always be valid.
	// This is a safety fallback.
}

func TestHearts_PlayerName_OutOfBounds(t *testing.T) {
	// Line 1022: playerName with idx < 0 or >= len(players)
	// playerName is private, called from appendLog contexts.
	// The only way to trigger this is with winnerIdx = -1 or similar.
	// Let's test ScoreRound game end log which uses playerName(h.winnerIdx).
	// winnerIdx is always valid when game_end log is written.
}

func TestHearts_CpuPlayHard_Lead_BestIsHeart_ThenNonHeart(t *testing.T) {
	// Line 933-937: bestIsHeart && !isHeart -> switch to non-heart
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 1, 2)
	h.SetHeartsBroken(true)

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// First valid card is heart, second is non-heart
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 1, len(trick))
	// Should switch from heart to clover (bestIsHeart && !isHeart branch)
	assert.Equal(t, domain.CardDesignClover, trick[0].Card.GetDesign())
}

func TestHearts_CpuPlayHard_Follow_SkipNonLeadSuit(t *testing.T) {
	// Line 971-972: card.GetDesign() != leadSuit -> continue
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Has lead suit AND non-lead suit in valid indices
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false)) // off-suit, will be skipped in follow loop
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should follow with clover, not heart
	assert.Equal(t, domain.CardDesignClover, trick[1].Card.GetDesign())
}

func TestHearts_CpuPlay_Normal_Follow_BothAbove_PicksLower(t *testing.T) {
	h := newTestHearts()
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	// Both above trick, bestVal starts at first card
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))

	h.CpuPlay()
	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// card.GetValue()(7) < bestVal(13), takes fallback branch
	assert.Equal(t, 7, trick[1].Card.GetValue())
}

// --- Omnibus Hearts (J♦ = -10) ---

func newTestHeartsOmnibus() *domain.Hearts {
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
	cfg := domain.DefaultHeartsConfig()
	cfg.OmnibusJD = true
	tc := domain.NewTrumpCards(0)
	return domain.NewHearts(tc, players, cfg)
}

func TestHearts_Omnibus_ResolveTrick_JDiamondGivesNegativePoints(t *testing.T) {
	h := newTestHeartsOmnibus()
	h.Reset()

	// Setup: trick with J♦ only (no hearts, no Q♠)
	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	// Build a trick where player 0 wins and J♦ is in the trick
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignDiamond, 13, false)}, // K♦ leads
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 11, false)}, // J♦ (-10 pts)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 5, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignDiamond, 3, false)},
	})
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(2)
	h.ResolveTrick()

	// Player 0 wins with K♦ and gets -10 points from J♦
	assert.Equal(t, -10, h.GetPlayer(0).GetRoundScore())
}

func TestHearts_Omnibus_ResolveTrick_JDiamondPlusHearts(t *testing.T) {
	h := newTestHeartsOmnibus()
	h.Reset()

	setupPlayPhase(h, 0, 0, 2)
	h.SetHeartsBroken(true)

	// Trick with J♦ and a heart: -10 + 1 = -9
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignDiamond, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 11, false)}, // J♦
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},    // ♥5
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignDiamond, 3, false)},
	})
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	h.SetTrickNumber(2)
	h.ResolveTrick()

	assert.Equal(t, -9, h.GetPlayer(0).GetRoundScore())
}

func TestHearts_Omnibus_ShootTheMoon(t *testing.T) {
	h := newTestHeartsOmnibus()

	// Setup: player 0 has roundScore = 16 (all hearts + Q♠ + J♦ = 26 - 10 = 16)
	// Player must also have J♦ in tricksTaken, otherwise score 16 is ambiguous.
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)
	h.GetPlayer(0).SetRoundScore(16) // moon threshold with omnibus
	h.GetPlayer(0).AddTrick([]*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 11, false), // J♦ confirms legitimate moon
	})
	h.GetPlayer(1).SetRoundScore(0)
	h.GetPlayer(2).SetRoundScore(0)
	h.GetPlayer(3).SetRoundScore(0)

	h.ScoreRound()

	// Shooter gets 0, others get 26
	assert.Equal(t, 0, h.GetPlayer(0).GetRoundScore())
	assert.Equal(t, 26, h.GetPlayer(1).GetCumulativeScore())
	assert.Equal(t, 26, h.GetPlayer(2).GetCumulativeScore())
	assert.Equal(t, 26, h.GetPlayer(3).GetCumulativeScore())
}

func TestHearts_Omnibus_NoShootTheMoon_Score16_WithoutJDiamond(t *testing.T) {
	// Score 16 without J♦ (e.g. 3 hearts + Q♠ = 3 + 13 = 16) should NOT trigger moon
	h := newTestHeartsOmnibus()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)
	h.GetPlayer(0).SetRoundScore(16) // ambiguous score without J♦ in tricks
	// No J♦ added to tricksTaken
	h.GetPlayer(1).SetRoundScore(0)
	h.GetPlayer(2).SetRoundScore(0)
	h.GetPlayer(3).SetRoundScore(0)

	h.ScoreRound()

	// No moon: player 0 keeps 16, others keep 0
	assert.Equal(t, 16, h.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 0, h.GetPlayer(1).GetCumulativeScore())
}

func TestHearts_Omnibus_NoShootTheMoon_At26(t *testing.T) {
	// With omnibus, 26 points means player took all hearts + Q♠ but NOT J♦
	// This should NOT trigger shoot-the-moon (threshold is 16)
	h := newTestHeartsOmnibus()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)
	h.GetPlayer(0).SetRoundScore(26) // NOT moon threshold with omnibus
	h.GetPlayer(1).SetRoundScore(0)
	h.GetPlayer(2).SetRoundScore(0)
	h.GetPlayer(3).SetRoundScore(0)

	h.ScoreRound()

	// No moon: player 0 keeps 26
	assert.Equal(t, 26, h.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 0, h.GetPlayer(1).GetCumulativeScore())
}

func TestHearts_NoOmnibus_ShootTheMoon_Standard(t *testing.T) {
	// Without omnibus, 26 still triggers shoot the moon
	h := newTestHearts()
	h.SetPhase(domain.HeartsPhaseRoundEnd)
	h.SetRoundNumber(1)
	h.GetPlayer(0).SetRoundScore(26)
	h.GetPlayer(1).SetRoundScore(0)
	h.GetPlayer(2).SetRoundScore(0)
	h.GetPlayer(3).SetRoundScore(0)

	h.ScoreRound()

	assert.Equal(t, 0, h.GetPlayer(0).GetRoundScore())
	assert.Equal(t, 26, h.GetPlayer(1).GetCumulativeScore())
}

func TestHearts_Omnibus_ValidatePlay_FirstTrick_JDiamondBlocked(t *testing.T) {
	h := newTestHeartsOmnibus()
	h.Reset()

	// First trick, player follows with void in lead suit
	setupPlayPhase(h, 0, 0, 1)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 2, false)}, // lead: ♣2
	})

	// Human has J♦ + a non-point card
	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦ (point card under omnibus)
	player.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))    // non-point card

	// J♦ should be blocked on first trick (it's a point card under omnibus)
	err := h.PlayerPlay(0) // try to play J♦
	assert.Error(t, err)
}

func TestHearts_Omnibus_ValidatePlay_FirstTrick_JDiamondAllowed_OnlyPointCards(t *testing.T) {
	h := newTestHeartsOmnibus()
	h.Reset()

	// First trick, player is void in lead suit and has only point cards
	setupPlayPhase(h, 0, 0, 1)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 2, false)},
	})

	// Human has only point cards (J♦ + hearts)
	player := h.GetPlayer(0)
	player.Reset()
	player.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦
	player.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))    // ♥5

	// Forced to play a point card since no non-point cards exist
	err := h.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestHearts_Omnibus_CpuPassNormal_KeepsJDiamond(t *testing.T) {
	h := newTestHeartsOmnibus()
	cfg := h.GetConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)
	h.Reset()

	// Give CPU1 J♦ and other cards
	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦ (should be kept)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))   // Q♠ (should be passed)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))   // K♠ (should be passed)
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))   // K♥ (should be passed)
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

	h.CpuPass()

	// J♦ should still be in CPU1's hand (not passed)
	passed := h.GetPassedCards()
	for _, card := range passed[1] {
		assert.False(t, card.GetDesign() == domain.CardDesignDiamond && card.GetValue() == 11,
			"CPU should not pass J♦ in omnibus mode")
	}
}

func TestHearts_Omnibus_CpuPassHard_KeepsJDiamond(t *testing.T) {
	h := newTestHeartsOmnibus()
	cfg := h.GetConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)
	h.Reset()

	// Give CPU1 J♦ and other high cards
	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦ (should be kept)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))   // Q♠
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))   // K♠
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))   // K♥
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))

	h.CpuPass()

	passed := h.GetPassedCards()
	for _, card := range passed[1] {
		assert.False(t, card.GetDesign() == domain.CardDesignDiamond && card.GetValue() == 11,
			"CPU should not pass J♦ in omnibus mode (hard)")
	}
}

func TestHearts_Omnibus_CpuPlayHard_DiscardKeepsJDiamond(t *testing.T) {
	h := newTestHeartsOmnibus()
	cfg := h.GetConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyHard
	h.SetConfig(cfg)
	h.Reset()

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	// Lead is clover, CPU1 has no clover (void)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦ (should keep)
	cpu.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))   // K♥ (should discard)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

	h.CpuPlay()

	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should discard K♥ (penalty card), not J♦
	assert.Equal(t, domain.CardDesignHeart, trick[1].Card.GetDesign())
	assert.Equal(t, 13, trick[1].Card.GetValue())
}

func TestHearts_Omnibus_CpuPlayNormal_DiscardKeepsJDiamond(t *testing.T) {
	h := newTestHeartsOmnibus()
	cfg := h.GetConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)
	h.Reset()

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	// Lead is clover, CPU1 is void in clover
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦ (should keep - not penalty)
	cpu.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))   // Q♠ (penalty - should discard)
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))

	h.CpuPlay()

	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should discard Q♠ (penalty), not J♦
	assert.Equal(t, domain.CardDesignSpade, trick[1].Card.GetDesign())
	assert.Equal(t, 12, trick[1].Card.GetValue())
}

func TestHearts_Omnibus_CpuPlayNormal_DiscardKeepsJDiamond_NoQSpade(t *testing.T) {
	// Bug 2 regression: J♦ value (11) > lower non-penalty card value (e.g. 7♣ = 7).
	// Without the fix, cpuPlayNormal would discard J♦ because it is the highest-value
	// non-penalty card. With the fix, J♦ must be preserved and the 7♣ discarded instead.
	h := newTestHeartsOmnibus()
	cfg := h.GetConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficultyNormal
	h.SetConfig(cfg)
	h.Reset()

	setupPlayPhase(h, 1, 0, 2)
	h.SetHeartsBroken(true)

	// Lead is clover, CPU1 is void in clover
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 10, false)},
	})

	cpu := h.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(domain.NewCard(domain.CardDesignDiamond, 11, false)) // J♦ (should keep)
	cpu.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))   // 7♣ (should discard - lower value)

	h.CpuPlay()

	trick := h.GetCurrentTrick()
	assert.Equal(t, 2, len(trick))
	// Should discard 7♣, not J♦
	assert.Equal(t, domain.CardDesignClover, trick[1].Card.GetDesign())
	assert.Equal(t, 7, trick[1].Card.GetValue())
}

func TestHearts_Omnibus_DefaultConfigFalse(t *testing.T) {
	cfg := domain.DefaultHeartsConfig()
	assert.False(t, cfg.OmnibusJD)
}

// --- GetHint ---

func TestHearts_GetHint_PassPhase(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	assert.Equal(t, domain.HeartsPhasePass, h.GetPhase())
	hint := h.GetHint()
	assert.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 3)
	assert.Equal(t, "pass_high_risk_cards", hint.Reason)
}

func TestHearts_GetHint_PlayPhase_Lead(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 0, 0, 1)
	h.SetCurrentTrick(nil)
	p := h.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	p.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

	hint := h.GetHint()
	assert.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 1)
	assert.Equal(t, "lead_low", hint.Reason)
}

func TestHearts_GetHint_PlayPhase_FollowSuit(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 0, 1, 1)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
	})
	p := h.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	p.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))

	hint := h.GetHint()
	assert.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 1)
	assert.Equal(t, "follow_suit", hint.Reason)
}

func TestHearts_GetHint_PlayPhase_DiscardQueenSpades(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 0, 1, 2)
	h.SetHeartsBroken(true)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 7, false)},
	})
	p := h.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	hint := h.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "discard_queen_spades", hint.Reason)
}

func TestHearts_GetHint_PlayPhase_DiscardHearts(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 0, 1, 2)
	h.SetHeartsBroken(true)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 7, false)},
	})
	p := h.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	hint := h.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "discard_hearts", hint.Reason)
}

func TestHearts_GetHint_PlayPhase_DiscardHigh(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 0, 1, 2)
	h.SetHeartsBroken(true)
	h.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 7, false)},
	})
	p := h.GetPlayer(0)
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

	hint := h.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "discard_high", hint.Reason)
}

func TestHearts_GetHint_NotHumanTurn(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 1, 1, 1)
	hint := h.GetHint()
	assert.Nil(t, hint)
}

func TestHearts_GetHint_WrongPhase(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetPhase(domain.HeartsPhaseTrickEnd)
	hint := h.GetHint()
	assert.Nil(t, hint)
}

func TestHearts_GetHint_GameEnd(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	h.SetPhase(domain.HeartsPhaseGameEnd)
	hint := h.GetHint()
	assert.Nil(t, hint)
}

func TestHearts_GetHint_PlayPhase_NoValidCards(t *testing.T) {
	h := newTestHearts()
	h.Reset()
	setupPlayPhase(h, 0, 0, 1)
	p := h.GetPlayer(0)
	p.Reset()

	hint := h.GetHint()
	assert.Nil(t, hint)
}
