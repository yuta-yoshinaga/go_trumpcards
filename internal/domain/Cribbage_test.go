//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCribbage() *Cribbage {
	config := DefaultCribbageConfig()
	players := []*CribbagePlayer{
		NewCribbagePlayer(true),  // human (index 0)
		NewCribbagePlayer(false), // CPU (index 1)
	}
	return NewCribbage(NewTrumpCards(0), players, config)
}

func newTestCribbageWithDifficulty(d CribbageCpuDifficulty) *Cribbage {
	config := CribbageConfig{CpuDifficulty: d, PointLimit: 121}
	players := []*CribbagePlayer{
		NewCribbagePlayer(true),
		NewCribbagePlayer(false),
	}
	return NewCribbage(NewTrumpCards(0), players, config)
}

// setupDiscardPhase sets up the game in discard phase with human's turn
func setupDiscardPhase(g *Cribbage) {
	g.SetPhase(CribbagePhaseDiscard)
	g.SetCurrentPlayerIdx(0) // human
	g.SetDealerIdx(1)        // CPU is dealer
	// Give human 6 cards
	g.players[0].Reset()
	for i := 1; i <= 6; i++ {
		g.players[0].AddCard(cCard(CardDesignSpade, i))
	}
	// Give CPU 6 cards
	g.players[1].Reset()
	for i := 7; i <= 12; i++ {
		g.players[1].AddCard(cCard(CardDesignHeart, i))
	}
}

// setupPeggingPhase sets up the game in pegging phase
func setupPeggingPhase(g *Cribbage) {
	g.SetPhase(CribbagePhasePegging)
	g.SetCurrentPlayerIdx(0) // human
	g.SetDealerIdx(1)        // CPU is dealer
	g.SetPegCount(0)
	g.SetPegPlayedCards(nil)
	// Give human 4 cards
	g.players[0].Reset()
	g.players[0].AddCard(cCard(CardDesignSpade, 1))  // A
	g.players[0].AddCard(cCard(CardDesignSpade, 2))  // 2
	g.players[0].AddCard(cCard(CardDesignSpade, 3))  // 3
	g.players[0].AddCard(cCard(CardDesignHeart, 10)) // 10
	// Give CPU 4 cards
	g.players[1].Reset()
	g.players[1].AddCard(cCard(CardDesignHeart, 4))    // 4
	g.players[1].AddCard(cCard(CardDesignHeart, 5))    // 5
	g.players[1].AddCard(cCard(CardDesignHeart, 6))    // 6
	g.players[1].AddCard(cCard(CardDesignDiamond, 10)) // 10
}

// ---- Constructor/Config Tests ----

func TestNewCribbage(t *testing.T) {
	g := newTestCribbage()
	assert.NotNil(t, g)
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, CribbageCpuDifficultyNormal, g.GetConfig().CpuDifficulty)
	assert.Equal(t, 121, g.GetConfig().PointLimit)
}

func TestCribbageConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  CribbageConfig
		wantErr bool
	}{
		{"valid default", DefaultCribbageConfig(), false},
		{"valid easy", CribbageConfig{CpuDifficulty: CribbageCpuDifficultyEasy, PointLimit: 61}, false},
		{"invalid difficulty", CribbageConfig{CpuDifficulty: 5, PointLimit: 121}, true},
		{"invalid point limit", CribbageConfig{CpuDifficulty: 0, PointLimit: 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- Reset Tests ----

func TestCribbage_Reset(t *testing.T) {
	g := newTestCribbage()
	g.Reset()

	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, CribbagePhaseDiscard, g.GetPhase())
	// Human is index 0, CPU is index 1, dealerIdx=1 → non-dealer first → human
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	// Each player should have 6 cards
	assert.Equal(t, CribbageDealSize, g.players[0].GetCardsSize())
	assert.Equal(t, CribbageDealSize, g.players[1].GetCardsSize())
}

// ---- Discard Tests ----

func TestCribbage_PlayerDiscard_Success(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)

	err := g.PlayerDiscard([]int{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, 4, g.players[0].GetCardsSize()) // 6-2=4
	assert.Equal(t, 2, len(g.GetCrib()))
}

func TestCribbage_PlayerDiscard_WrongPhase(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhasePegging)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerDiscard([]int{0, 1})
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestCribbage_PlayerDiscard_GameEnded(t *testing.T) {
	g := newTestCribbage()
	g.SetGameEndFlag(true)
	err := g.PlayerDiscard([]int{0, 1})
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestCribbage_PlayerDiscard_NotHumanTurn(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)
	g.SetCurrentPlayerIdx(1) // CPU
	err := g.PlayerDiscard([]int{0, 1})
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestCribbage_PlayerDiscard_InvalidIndices(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)

	// Wrong count
	err := g.PlayerDiscard([]int{0})
	assert.ErrorIs(t, err, ErrInvalidIndices)

	// Out of range
	err = g.PlayerDiscard([]int{0, 10})
	assert.ErrorIs(t, err, ErrInvalidCard)

	// Same index
	err = g.PlayerDiscard([]int{2, 2})
	assert.ErrorIs(t, err, ErrInvalidIndices)
}

func TestCribbage_BothDiscard_TransitionsToCut(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)
	// Ensure draw pile has cards for the cut
	g.drawPile = []*Card{cCard(CardDesignSpade, 10), cCard(CardDesignHeart, 7)}

	// Human discards
	err := g.PlayerDiscard([]int{0, 1})
	require.NoError(t, err)
	// Now CPU's turn
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, CribbagePhaseDiscard, g.GetPhase())

	// CPU discards → dealer is CPU(1), so the non-dealer human(0) is the cutter.
	// The game must STOP at the cut phase (starter not yet revealed) and wait
	// for the human's explicit cut.
	g.CpuPlay()
	assert.Equal(t, CribbagePhaseCut, g.GetPhase())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx()) // human is the cutter
	assert.Nil(t, g.GetStarter())
	assert.Equal(t, 4, len(g.GetCrib()))

	// Human cuts the deck → starter revealed and pegging begins.
	err = g.PlayerCut()
	require.NoError(t, err)
	assert.Equal(t, CribbagePhasePegging, g.GetPhase())
	assert.NotNil(t, g.GetStarter())
}

func TestCribbage_PlayerCut_WrongPhase(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerCut()
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestCribbage_PlayerCut_GameEnded(t *testing.T) {
	g := newTestCribbage()
	g.SetGameEndFlag(true)
	err := g.PlayerCut()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestCribbage_PlayerCut_NotHumanTurn(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhaseCut)
	g.SetCurrentPlayerIdx(1) // CPU is the cutter
	err := g.PlayerCut()
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestCribbage_CpuPlay_Cut_Auto(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhaseCut)
	g.SetDealerIdx(0)        // human is dealer → CPU(1) is the non-dealer cutter
	g.SetCurrentPlayerIdx(1) // CPU's cut
	g.drawPile = []*Card{cCard(CardDesignSpade, 10)}

	g.CpuPlay()
	// CPU auto-cuts: starter revealed and pegging begins.
	assert.Equal(t, CribbagePhasePegging, g.GetPhase())
	assert.NotNil(t, g.GetStarter())
}

// ---- Pegging Tests ----

func TestCribbage_PlayerPeg_Success(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)

	// Play A♠ (value 1)
	err := g.PlayerPeg(0)
	assert.NoError(t, err)
	assert.Equal(t, 3, g.players[0].GetCardsSize())
}

func TestCribbage_PlayerPeg_WrongPhase(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhaseDiscard)
	g.SetCurrentPlayerIdx(0)
	err := g.PlayerPeg(0)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestCribbage_PlayerPeg_ExceedsLimit(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	g.SetPegCount(25) // 25 + 10(card at idx 3) = 35 > 31

	err := g.PlayerPeg(3) // 10 of hearts
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestCribbage_PlayerPeg_InvalidIndex(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	err := g.PlayerPeg(10)
	assert.ErrorIs(t, err, ErrInvalidCard)
}

func TestCribbage_PlayerGo_Success(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	g.SetPegCount(28) // Only cards with value <= 3 can be played
	// Remove all low cards from human, leave only 10♥
	g.players[0].Reset()
	g.players[0].AddCard(cCard(CardDesignHeart, 10)) // can't play (28+10=38>31)

	err := g.PlayerGo()
	assert.NoError(t, err)
}

func TestCribbage_PlayerGo_CanStillPlay(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	// pegCount=0, human has A which is playable
	err := g.PlayerGo()
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

// ---- Show Phase Tests ----

func TestCribbage_ShowNext(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhaseShow)
	g.SetDealerIdx(1)
	g.SetStarter(cCard(CardDesignSpade, 10))
	// Non-dealer (0) hand: 5♠ 5♥ 5♦ J♠ → good scoring hand
	g.SetOriginalHand(0, []*Card{
		cCard(CardDesignSpade, 5),
		cCard(CardDesignHeart, 5),
		cCard(CardDesignDiamond, 5),
		cCard(CardDesignSpade, 11),
	})
	// Dealer (1) hand: A♠ 2♠ 3♠ 4♠
	g.SetOriginalHand(1, []*Card{
		cCard(CardDesignSpade, 1),
		cCard(CardDesignSpade, 2),
		cCard(CardDesignSpade, 3),
		cCard(CardDesignSpade, 4),
	})
	// Crib: 7♥ 8♥ 9♥ K♥
	g.SetCrib([]*Card{
		cCard(CardDesignHeart, 7),
		cCard(CardDesignHeart, 8),
		cCard(CardDesignHeart, 9),
		cCard(CardDesignHeart, 13),
	})

	// Step 0: non-dealer hand score
	err := g.ShowNext()
	assert.NoError(t, err)
	assert.NotNil(t, g.GetHandScoreDetails()[0])
	assert.Greater(t, g.GetHandScoreDetails()[0].Total, 0)

	// Step 1: dealer hand score
	err = g.ShowNext()
	assert.NoError(t, err)
	assert.NotNil(t, g.GetHandScoreDetails()[1])

	// Step 2: crib score → transitions to RoundEnd
	err = g.ShowNext()
	assert.NoError(t, err)
	assert.NotNil(t, g.GetHandScoreDetails()[2])
	assert.Equal(t, CribbagePhaseRoundEnd, g.GetPhase())
}

func TestCribbage_ShowNext_WrongPhase(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhasePegging)
	err := g.ShowNext()
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestCribbage_ShowNext_WinDuringShow(t *testing.T) {
	g := newTestCribbage()
	g.SetPhase(CribbagePhaseShow)
	g.SetDealerIdx(1)
	g.SetStarter(cCard(CardDesignHeart, 5))
	// Non-dealer at 119 points, high scoring hand should push over 121
	g.players[0].SetCumulativeScore(119)
	g.SetOriginalHand(0, []*Card{
		cCard(CardDesignSpade, 5),
		cCard(CardDesignHeart, 5),
		cCard(CardDesignDiamond, 5),
		cCard(CardDesignClover, 11),
	})
	g.SetOriginalHand(1, []*Card{
		cCard(CardDesignSpade, 1),
		cCard(CardDesignSpade, 2),
		cCard(CardDesignSpade, 3),
		cCard(CardDesignSpade, 4),
	})

	err := g.ShowNext()
	assert.NoError(t, err)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

// ---- NextRound Tests ----

func TestCribbage_NextRound(t *testing.T) {
	g := newTestCribbage()
	g.Reset()
	g.SetPhase(CribbagePhaseRoundEnd)
	initialDealer := g.GetDealerIdx()

	g.NextRound()

	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, 1-initialDealer, g.GetDealerIdx()) // dealer swapped
	assert.Equal(t, CribbagePhaseDiscard, g.GetPhase())
}

func TestCribbage_NextRound_WrongPhase(t *testing.T) {
	g := newTestCribbage()
	g.Reset()
	// Phase is Discard, not RoundEnd
	roundBefore := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, roundBefore, g.GetRoundNumber()) // no change
}

// ---- CPU Play Tests ----

func TestCribbage_CpuPlay_Discard_Easy(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyEasy)
	setupDiscardPhase(g)
	g.SetCurrentPlayerIdx(1) // CPU

	g.CpuPlay()
	assert.Equal(t, 4, g.players[1].GetCardsSize()) // discarded 2
}

func TestCribbage_CpuPlay_Discard_Normal(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyNormal)
	setupDiscardPhase(g)
	g.SetCurrentPlayerIdx(1)

	g.CpuPlay()
	assert.Equal(t, 4, g.players[1].GetCardsSize())
}

func TestCribbage_CpuPlay_Discard_Hard(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyHard)
	setupDiscardPhase(g)
	g.SetCurrentPlayerIdx(1)

	g.CpuPlay()
	assert.Equal(t, 4, g.players[1].GetCardsSize())
}

func TestCribbage_CpuPlay_Peg_Easy(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyEasy)
	setupPeggingPhase(g)
	g.SetCurrentPlayerIdx(1) // CPU
	initialCards := g.players[1].GetCardsSize()

	g.CpuPlay()
	assert.Less(t, g.players[1].GetCardsSize(), initialCards)
}

func TestCribbage_CpuPlay_Peg_Normal(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyNormal)
	setupPeggingPhase(g)
	g.SetCurrentPlayerIdx(1)
	initialCards := g.players[1].GetCardsSize()

	g.CpuPlay()
	assert.Less(t, g.players[1].GetCardsSize(), initialCards)
}

func TestCribbage_CpuPlay_Peg_Hard(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyHard)
	setupPeggingPhase(g)
	g.SetCurrentPlayerIdx(1)
	initialCards := g.players[1].GetCardsSize()

	g.CpuPlay()
	assert.Less(t, g.players[1].GetCardsSize(), initialCards)
}

func TestCribbage_CpuPlay_Go(t *testing.T) {
	g := newTestCribbageWithDifficulty(CribbageCpuDifficultyEasy)
	setupPeggingPhase(g)
	g.SetCurrentPlayerIdx(1)
	g.SetPegCount(28)
	// CPU has only 10-value cards
	g.players[1].Reset()
	g.players[1].AddCard(cCard(CardDesignHeart, 10))

	g.CpuPlay() // Should auto-Go
}

func TestCribbage_CpuPlay_HumanTurn_NoOp(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	g.SetCurrentPlayerIdx(0) // human
	initialCards := g.players[0].GetCardsSize()

	g.CpuPlay() // should do nothing
	assert.Equal(t, initialCards, g.players[0].GetCardsSize())
}

// ---- His Heels Tests ----

func TestCribbage_HisHeels(t *testing.T) {
	g := newTestCribbage()
	g.Reset()
	// Simulate both discards complete, then set up a J starter
	g.SetPhase(CribbagePhaseDiscard)
	g.SetDealerIdx(1)
	g.SetCurrentPlayerIdx(0)

	// Manually complete discards
	setupDiscardPhase(g)
	g.players[0].Reset()
	for i := 1; i <= 4; i++ {
		g.players[0].AddCard(cCard(CardDesignSpade, i))
	}
	g.players[1].Reset()
	for i := 5; i <= 8; i++ {
		g.players[1].AddCard(cCard(CardDesignHeart, i))
	}
	g.SetDiscardDone(0, true)
	g.SetDiscardDone(1, true)

	// Put J at top of draw pile
	g.drawPile = []*Card{cCard(CardDesignSpade, 11)}

	// Set original hands
	for i := range CribbagePlayerCnt {
		cards := make([]*Card, g.players[i].GetCardsSize())
		for j := range cards {
			cards[j] = g.players[i].GetCard(j)
		}
		g.SetOriginalHand(i, cards)
	}

	// Trigger cut
	g.doCut()

	// Dealer (1) should get 2 points for His Heels
	assert.Equal(t, 2, g.players[1].GetCumulativeScore())
}

// ---- Getter Tests ----

func TestCribbage_Getters(t *testing.T) {
	g := newTestCribbage()
	g.Reset()

	assert.Equal(t, 1, g.GetDealerIdx())
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetPlayer(1))
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(2))
	assert.Equal(t, 0, g.GetPegCount())
	assert.Nil(t, g.GetPegPlayedCards())
	assert.Equal(t, 0, g.GetShowPhaseStep())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetOriginalHand(-1))
	assert.Nil(t, g.GetPlayerPeggedCards(-1))
}

func TestCribbage_SetConfig(t *testing.T) {
	g := newTestCribbage()
	newConfig := CribbageConfig{CpuDifficulty: CribbageCpuDifficultyHard, PointLimit: 61}
	g.SetConfig(newConfig)
	assert.Equal(t, CribbageCpuDifficultyHard, g.GetConfig().CpuDifficulty)
	assert.Equal(t, 61, g.GetConfig().PointLimit)
}

// ---- Win Condition Tests ----

func TestCribbage_WinDuringPegging(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	g.players[0].SetCumulativeScore(120) // one more point wins

	// Play card that scores at least 1 point (make a 15)
	g.SetPegCount(14)     // 14 + 1(A) = 15 → 2 pts
	err := g.PlayerPeg(0) // A♠ (value 1)
	assert.NoError(t, err)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

// ---- Action Log Tests ----

func TestCribbage_ActionLog(t *testing.T) {
	g := newTestCribbage()
	g.Reset()
	assert.Nil(t, g.GetActionLog()) // reset clears log

	setupDiscardPhase(g)
	_ = g.PlayerDiscard([]int{0, 1})
	assert.Greater(t, len(g.GetActionLog()), 0)
}

// ---- estimateCribValue Tests ----

func TestCribbage_EstimateCribValue(t *testing.T) {
	g := newTestCribbage()
	// 15 combo
	score := g.estimateCribValue([]*Card{cCard(1, 5), cCard(2, 10)})
	assert.Equal(t, 3, score) // 15=2 + five=1

	// Pair
	score = g.estimateCribValue([]*Card{cCard(1, 7), cCard(2, 7)})
	assert.Equal(t, 2, score) // pair=2

	// Single card
	score = g.estimateCribValue([]*Card{cCard(1, 3)})
	assert.Equal(t, 0, score)
}

// ---- Player Tests ----

func TestCribbagePlayer_ResetRound(t *testing.T) {
	p := NewCribbagePlayer(true)
	p.AddCard(cCard(1, 5))
	p.SetRoundScore(10)
	p.SetCumulativeScore(50)

	p.ResetRound()
	assert.Equal(t, 0, p.GetRoundScore())
	assert.Equal(t, 0, p.GetCardsSize())
}

// ---- GetHint Tests ----

func TestCribbage_GetHint_DiscardPhase(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)
	// Hand ♠A..♠6: keeping {3,4,5,6} scores best (15×1 + run of 4 + 4-flush = 10),
	// so the recommended discard is indices 0 and 1 (A and 2).
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "discard", hint.Type)
	assert.Equal(t, []int{0, 1}, hint.Indices)
}

func TestCribbage_GetHint_DiscardPhase_NotHumanTurn(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

func TestCribbage_GetHint_DiscardPhase_AlreadyDiscarded(t *testing.T) {
	g := newTestCribbage()
	setupDiscardPhase(g)
	// A 4-card hand means the human already gave 2 cards to the crib.
	g.players[0].Reset()
	for i := 1; i <= 4; i++ {
		g.players[0].AddCard(cCard(CardDesignSpade, i))
	}
	assert.Nil(t, g.GetHint())
}

func TestCribbage_GetHint_PeggingPhase_PrefersFifteen(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	// CPU led a 5: playing the ♥10 (index 3) hits 15 for 2 points.
	g.SetPegPlayedCards([]*Card{cCard(CardDesignHeart, 5)})
	g.SetPegCount(5)
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "play", hint.Type)
	assert.Equal(t, []int{3}, hint.Indices)
}

func TestCribbage_GetHint_PeggingPhase_FirstLegalOnTie(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	// Empty sequence: no play scores, so the first legal card (index 0) is suggested.
	hint := g.GetHint()
	assert.NotNil(t, hint)
	assert.Equal(t, "play", hint.Type)
	assert.Equal(t, []int{0}, hint.Indices)
}

func TestCribbage_GetHint_PeggingPhase_NoLegalPlay(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	g.players[0].Reset()
	g.players[0].AddCard(cCard(CardDesignHeart, 10))
	g.SetPegCount(30) // 30 + 10 > 31 → the human can only call Go
	assert.Nil(t, g.GetHint())
}

func TestCribbage_GetHint_PeggingPhase_NotHumanTurn(t *testing.T) {
	g := newTestCribbage()
	setupPeggingPhase(g)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

func TestCribbage_GetHint_OtherPhases(t *testing.T) {
	g := newTestCribbage()
	for _, phase := range []CribbagePhase{CribbagePhaseCut, CribbagePhaseShow, CribbagePhaseRoundEnd, CribbagePhaseGameEnd} {
		g.SetPhase(phase)
		assert.Nil(t, g.GetHint())
	}
}
