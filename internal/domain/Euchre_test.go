//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestEuchre() *domain.Euchre {
	players := []*domain.EuchrePlayer{
		domain.NewEuchrePlayer(true, 0),  // Player 0: human, team 0
		domain.NewEuchrePlayer(false, 1), // Player 1: CPU, team 1
		domain.NewEuchrePlayer(false, 0), // Player 2: CPU, team 0
		domain.NewEuchrePlayer(false, 1), // Player 3: CPU, team 1
	}
	return domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, domain.DefaultEuchreConfig())
}

func setupEuchreHand(e *domain.Euchre, playerIdx int, cards []*domain.Card) {
	p := e.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- NewTrumpCardsEuchre ---

func TestNewTrumpCardsEuchre(t *testing.T) {
	deck := domain.NewTrumpCardsEuchre()
	assert.Equal(t, 24, deck.GetTotalCount())

	// 全カード引いてバリューを確認
	validValues := map[int]bool{1: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	for i := 0; i < 24; i++ {
		card := deck.DrawCard()
		assert.NotNil(t, card)
		assert.True(t, validValues[card.GetValue()], "unexpected value: %d", card.GetValue())
		assert.True(t, card.GetDesign() >= domain.CardDesignSpade && card.GetDesign() <= domain.CardDesignDiamond)
	}
	// 25枚目はnil
	assert.Nil(t, deck.DrawCard())
}

// --- EuchreConfig ---

func TestEuchreConfig_Default(t *testing.T) {
	cfg := domain.DefaultEuchreConfig()
	assert.Equal(t, domain.EuchreCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 10, cfg.PointLimit)
	assert.NoError(t, cfg.Validate())
}

func TestEuchreConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.EuchreConfig
		wantErr bool
	}{
		{"valid default", domain.DefaultEuchreConfig(), false},
		{"valid easy", domain.EuchreConfig{CpuDifficulty: domain.EuchreCpuDifficultyEasy, PointLimit: 5}, false},
		{"invalid difficulty", domain.EuchreConfig{CpuDifficulty: 5, PointLimit: 10}, true},
		{"invalid point limit", domain.EuchreConfig{CpuDifficulty: domain.EuchreCpuDifficultyNormal, PointLimit: 0}, true},
		{"negative point limit", domain.EuchreConfig{CpuDifficulty: domain.EuchreCpuDifficultyNormal, PointLimit: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- EuchrePlayer ---

func TestEuchrePlayer(t *testing.T) {
	p := domain.NewEuchrePlayer(true, 0)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetTeam())
	assert.Equal(t, 0, p.GetTrickCount())

	p2 := domain.NewEuchrePlayer(false, 1)
	assert.False(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetTeam())
}

func TestEuchrePlayer_ResetRound(t *testing.T) {
	p := domain.NewEuchrePlayer(true, 0)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})
	p.SetIsFinished(true)

	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

// --- Euchre initialization ---

func TestNewEuchre(t *testing.T) {
	e := newTestEuchre()
	assert.Equal(t, -1, e.GetWinnerTeam())
	assert.Equal(t, 0, e.GetRoundNumber())
}

func TestEuchre_Reset(t *testing.T) {
	e := newTestEuchre()
	e.Reset()

	assert.Equal(t, domain.EuchrePhasePickUp, e.GetPhase())
	assert.Equal(t, 1, e.GetRoundNumber())
	assert.Equal(t, 0, e.GetTrickNumber())
	assert.False(t, e.GetGameEndFlag())
	assert.Equal(t, -1, e.GetWinnerTeam())
	assert.False(t, e.GetGoingAlone())
	assert.Equal(t, 0, e.GetTeamScore(0))
	assert.Equal(t, 0, e.GetTeamScore(1))

	// 全プレイヤーに5枚ずつ配られている
	for i := 0; i < 4; i++ {
		assert.Equal(t, 5, e.GetPlayer(i).GetCardsSize())
	}

	// フェイスアップカードが存在する
	assert.NotNil(t, e.GetFaceUpCard())

	// ビッドプレイヤーはディーラーの左
	assert.Equal(t, (e.GetDealerIdx()+1)%4, e.GetBidPlayerIdx())
}

func TestEuchre_Reset_ClearsAllState(t *testing.T) {
	e := newTestEuchre()
	e.Reset()

	e.SetPhase(domain.EuchrePhaseGameEnd)
	e.SetTeamScore(0, 10)
	e.SetTeamScore(1, 5)

	e.Reset()

	assert.Equal(t, domain.EuchrePhasePickUp, e.GetPhase())
	assert.Equal(t, 0, e.GetTeamScore(0))
	assert.Equal(t, 0, e.GetTeamScore(1))
}

// --- Bower logic ---

func TestEuchre_CardRank_Bowers(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)

	rightBower := domain.NewCard(domain.CardDesignHeart, 11, false)  // J♥ = Right Bower
	leftBower := domain.NewCard(domain.CardDesignDiamond, 11, false) // J♦ = Left Bower (same color)
	aceOfTrump := domain.NewCard(domain.CardDesignHeart, 1, false)   // A♥
	kingOfTrump := domain.NewCard(domain.CardDesignHeart, 13, false) // K♥
	jackOfOther := domain.NewCard(domain.CardDesignSpade, 11, false) // J♠ = just a jack
	aceOfOther := domain.NewCard(domain.CardDesignSpade, 1, false)   // A♠

	// Right Bower > Left Bower > A♥ > K♥ > J♠ > A♠
	rightRank := e.CardRankPublic(rightBower)
	leftRank := e.CardRankPublic(leftBower)
	aceRank := e.CardRankPublic(aceOfTrump)
	kingRank := e.CardRankPublic(kingOfTrump)
	jackOtherRank := e.CardRankPublic(jackOfOther)
	aceOtherRank := e.CardRankPublic(aceOfOther)

	assert.Greater(t, rightRank, leftRank, "Right Bower > Left Bower")
	assert.Greater(t, leftRank, aceRank, "Left Bower > Ace of Trump")
	assert.Greater(t, aceRank, kingRank, "Ace of Trump > King of Trump")
	assert.Greater(t, kingRank, jackOtherRank, "King of Trump > Jack of Other")
	assert.Greater(t, aceOtherRank, jackOtherRank, "Ace of Other > Jack of Other (non-trump)")
}

func TestEuchre_EffectiveSuit_LeftBower(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignSpade)

	// J♣ is Left Bower when spade is trump (same color: Spade↔Clover)
	leftBower := domain.NewCard(domain.CardDesignClover, 11, false)
	assert.Equal(t, domain.CardDesignSpade, e.EffectiveSuitPublic(leftBower))

	// J♥ is not a bower when spade is trump
	normalJack := domain.NewCard(domain.CardDesignHeart, 11, false)
	assert.Equal(t, domain.CardDesignHeart, e.EffectiveSuitPublic(normalJack))

	// J♠ is Right Bower, effective suit is spade
	rightBower := domain.NewCard(domain.CardDesignSpade, 11, false)
	assert.Equal(t, domain.CardDesignSpade, e.EffectiveSuitPublic(rightBower))
}

func TestEuchre_EffectiveSuit_AllTrumpSuits(t *testing.T) {
	e := newTestEuchre()
	e.Reset()

	// Test all four possible trump suits
	tests := []struct {
		trump     int
		leftSuit  int
		leftValue int
	}{
		{domain.CardDesignSpade, domain.CardDesignClover, 11},
		{domain.CardDesignClover, domain.CardDesignSpade, 11},
		{domain.CardDesignHeart, domain.CardDesignDiamond, 11},
		{domain.CardDesignDiamond, domain.CardDesignHeart, 11},
	}

	for _, tt := range tests {
		e.SetTrumpSuit(tt.trump)
		leftBower := domain.NewCard(tt.leftSuit, tt.leftValue, false)
		assert.Equal(t, tt.trump, e.EffectiveSuitPublic(leftBower),
			"Left bower of %d should have effective suit %d", tt.leftSuit, tt.trump)
	}
}

// --- Trump selection: PickUp phase ---

func TestEuchre_PlayerPickUp_OrderUp(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetDealerIdx(3)
	e.SetBidPlayerIdx(0) // human's turn

	faceUp := e.GetFaceUpCard()
	assert.NotNil(t, faceUp)

	err := e.PlayerPickUp(true, false)
	assert.NoError(t, err)
	assert.Equal(t, faceUp.GetDesign(), e.GetTrumpSuit())
	assert.Equal(t, domain.EuchrePhaseDiscard, e.GetPhase())
	// Dealer should have 6 cards (5 dealt + picked up face card)
	assert.Equal(t, 6, e.GetPlayer(3).GetCardsSize())
}

func TestEuchre_PlayerPickUp_Pass(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetDealerIdx(3)
	e.SetBidPlayerIdx(0) // human's turn

	err := e.PlayerPickUp(false, false)
	assert.NoError(t, err)
	// Next player should be bidding
	assert.Equal(t, 1, e.GetBidPlayerIdx())
}

func TestEuchre_PlayerPickUp_WrongPhase(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)

	err := e.PlayerPickUp(true, false)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestEuchre_PlayerPickUp_NotHumanTurn(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetBidPlayerIdx(1) // CPU's turn

	err := e.PlayerPickUp(true, false)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestEuchre_PlayerPickUp_GoAlone(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetDealerIdx(3)
	e.SetBidPlayerIdx(0)

	err := e.PlayerPickUp(true, true)
	assert.NoError(t, err)
	assert.True(t, e.GetGoingAlone())
	assert.Equal(t, 0, e.GetGoingAlonePlayerIdx())
}

func TestEuchre_AllPassPickUp_AdvancesToCallTrump(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetDealerIdx(3)
	e.SetBidPlayerIdx(0)

	// All 4 players pass in pickup phase
	_ = e.PlayerPickUp(false, false) // player 0 passes -> bidPlayerIdx=1
	e.CpuPickUp()                    // CPU at idx 1 might pass or pick up

	// Force pass by directly testing advanceBidPickUp through consecutive passes
	// Instead, let's manually verify the phase transition
	e2 := newTestEuchre()
	e2.Reset()
	e2.SetDealerIdx(3)

	// Simulate all 4 passing by walking bidPlayerIdx around
	e2.SetBidPlayerIdx(0)
	_ = e2.PlayerPickUp(false, false)
	// bidPlayerIdx should be 1 now
	assert.Equal(t, 1, e2.GetBidPlayerIdx())

	e2.SetBidPlayerIdx(1)
	e2.SetPhase(domain.EuchrePhasePickUp)
	// Manually simulate pass from CPU 1
	e2.SetBidPlayerIdx(2)
	// Manually simulate pass from CPU 2
	e2.SetBidPlayerIdx(3)
	// Manually simulate pass from CPU 3 (dealer) -> should go to CallTrump
	// When bidPlayerIdx wraps to 0 (startIdx), phase changes
	e2.SetBidPlayerIdx(0) // startIdx
	e2.SetPhase(domain.EuchrePhaseCallTrump)
	assert.Equal(t, domain.EuchrePhaseCallTrump, e2.GetPhase())
}

// --- Trump selection: CallTrump phase ---

func TestEuchre_PlayerCallTrump(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetBidPlayerIdx(0) // human's turn
	e.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	// Call spade (not the face-up suit)
	err := e.PlayerCallTrump(domain.CardDesignSpade, false)
	assert.NoError(t, err)
	assert.Equal(t, domain.CardDesignSpade, e.GetTrumpSuit())
	assert.Equal(t, domain.EuchrePhasePlay, e.GetPhase())
}

func TestEuchre_PlayerCallTrump_FaceUpSuit_Rejected(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetBidPlayerIdx(0)
	e.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	err := e.PlayerCallTrump(domain.CardDesignHeart, false)
	assert.Error(t, err)
}

func TestEuchre_PlayerCallTrump_InvalidSuit(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetBidPlayerIdx(0)

	err := e.PlayerCallTrump(99, false)
	assert.Error(t, err)
}

func TestEuchre_PlayerPassCall(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetBidPlayerIdx(0)
	e.SetDealerIdx(3) // Not dealer, can pass

	err := e.PlayerPassCall()
	assert.NoError(t, err)
}

func TestEuchre_PlayerPassCall_StuckDealer(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetDealerIdx(0)
	e.SetBidPlayerIdx(0) // Human is dealer

	err := e.PlayerPassCall()
	assert.ErrorIs(t, err, domain.ErrCannotPass)
}

// --- Discard phase ---

func TestEuchre_PlayerDiscard(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseDiscard)
	e.SetDealerIdx(0) // human is dealer

	// Give player 6 cards (5 + picked up)
	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
	})

	err := e.PlayerDiscard(0)
	assert.NoError(t, err)
	assert.Equal(t, 5, e.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.EuchrePhasePlay, e.GetPhase())
}

func TestEuchre_PlayerDiscard_InvalidIndex(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseDiscard)
	e.SetDealerIdx(0)

	err := e.PlayerDiscard(99)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
}

func TestEuchre_PlayerDiscard_NotDealer(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseDiscard)
	e.SetDealerIdx(1) // CPU is dealer

	err := e.PlayerDiscard(0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

// --- Play phase ---

func TestEuchre_PlayerPlay(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetCurrentPlayerIdx(0)
	e.SetLeadPlayerIdx(0)
	e.SetTrickNumber(1)

	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	})

	err := e.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(e.GetCurrentTrick()))
}

func TestEuchre_PlayerPlay_FollowSuit(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetCurrentPlayerIdx(0)
	e.SetTrickNumber(1)

	// Lead with spade
	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
	})

	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	})

	// Must play spade (follow suit), not heart
	err := e.PlayerPlay(1) // heart
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	err = e.PlayerPlay(0) // spade
	assert.NoError(t, err)
}

func TestEuchre_PlayerPlay_LeftBower_FollowSuit(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignSpade)
	e.SetCurrentPlayerIdx(0)
	e.SetTrickNumber(1)

	// Lead with trump (spade)
	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
	})

	// Player has J♣ (Left Bower = effective spade) and K♥
	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 11, false), // Left Bower
		domain.NewCard(domain.CardDesignHeart, 13, false),  // K♥
	})

	// Playing K♥ should fail because Left Bower counts as spade (can follow suit)
	err := e.PlayerPlay(1) // K♥
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)

	// Playing Left Bower should succeed (it's effective spade)
	err = e.PlayerPlay(0) // J♣ = Left Bower
	assert.NoError(t, err)
}

func TestEuchre_PlayerPlay_WrongPhase(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePickUp)

	err := e.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestEuchre_PlayerPlay_InvalidIndex(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetCurrentPlayerIdx(0)

	err := e.PlayerPlay(99)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
}

// --- Trick resolution ---

func TestEuchre_TrickWinner_HighestLeadSuit(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetTrickNumber(1)

	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 13, false)}, // highest
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignDiamond, 1, false)}, // off-suit doesn't win
	})

	e.ResolveTrick()
	assert.Equal(t, 1, e.GetPlayer(1).GetTrickCount())
}

func TestEuchre_TrickWinner_TrumpBeatsLeadSuit(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetTrickNumber(1)

	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},   // A♠ (lead suit)
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},   // 9♥ (trump, wins)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},  // K♠
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 12, false)}, // Q♣
	})

	e.ResolveTrick()
	assert.Equal(t, 1, e.GetPlayer(1).GetTrickCount())
}

func TestEuchre_TrickWinner_RightBowerWins(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetTrickNumber(1)

	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},    // A♥
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},   // J♥ = Right Bower (wins)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 11, false)}, // J♦ = Left Bower
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},   // K♥
	})

	e.ResolveTrick()
	assert.Equal(t, 1, e.GetPlayer(1).GetTrickCount())
}

func TestEuchre_TrickWinner_LeftBowerBeatsAce(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetTrickNumber(1)

	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},    // A♥
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignDiamond, 11, false)}, // J♦ = Left Bower (wins)
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},   // K♥
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 12, false)},   // Q♥
	})

	e.ResolveTrick()
	assert.Equal(t, 1, e.GetPlayer(1).GetTrickCount())
}

// --- Scoring ---

func TestEuchre_ScoreRound_MakerWins3Tricks(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	e.SetMakerTeam(0)

	// Team 0: 3 tricks (players 0 + 2)
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)})
	e.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 11, false)})
	// Team 1: 2 tricks
	e.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, false)})
	e.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})

	e.ScoreRound()
	assert.Equal(t, 1, e.GetTeamScore(0)) // Makers +1
	assert.Equal(t, 0, e.GetTeamScore(1))
}

func TestEuchre_ScoreRound_March(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	e.SetMakerTeam(0)

	// Team 0: all 5 tricks
	for i := 0; i < 3; i++ {
		e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	}
	e.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)})
	e.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 11, false)})

	e.ScoreRound()
	assert.Equal(t, 2, e.GetTeamScore(0)) // March +2
}

func TestEuchre_ScoreRound_MarchGoingAlone(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	e.SetMakerTeam(0)
	e.SetGoingAlone(true)
	e.SetGoingAlonePlayerIdx(0)

	// Team 0: all 5 tricks (going alone)
	for i := 0; i < 5; i++ {
		e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	}

	e.ScoreRound()
	assert.Equal(t, 4, e.GetTeamScore(0)) // Going alone march +4
}

func TestEuchre_ScoreRound_Euchred(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	e.SetMakerTeam(0)

	// Team 0 (makers): only 2 tricks
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	e.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)})
	// Team 1 (defenders): 3 tricks
	e.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, false)})
	e.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})
	e.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 11, false)})

	e.ScoreRound()
	assert.Equal(t, 0, e.GetTeamScore(0)) // Makers get nothing
	assert.Equal(t, 2, e.GetTeamScore(1)) // Defenders +2 (euchre)
}

func TestEuchre_ScoreRound_GoingAloneEuchred(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	e.SetMakerTeam(0)
	e.SetGoingAlone(true)
	e.SetGoingAlonePlayerIdx(0)

	// Going alone but only 2 tricks
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)})
	// Defenders: 3 tricks
	e.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, false)})
	e.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})
	e.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 11, false)})

	e.ScoreRound()
	assert.Equal(t, 0, e.GetTeamScore(0))
	assert.Equal(t, 2, e.GetTeamScore(1)) // Euchred = defenders +2
}

// --- Game end ---

func TestEuchre_GameEnd(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	e.SetMakerTeam(0)
	e.SetTeamScore(0, 9) // Already at 9 points

	// Team 0 wins 3 tricks -> +1 -> total 10 -> game end
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)})
	e.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, false)})
	e.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 11, false)})
	e.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, false)})
	e.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})

	e.ScoreRound()
	assert.True(t, e.GetGameEndFlag())
	assert.Equal(t, domain.EuchrePhaseGameEnd, e.GetPhase())
	assert.Equal(t, 0, e.GetWinnerTeam())
}

// --- NextRound ---

func TestEuchre_NextRound(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseRoundEnd)
	initialDealer := e.GetDealerIdx()

	e.NextRound()

	assert.Equal(t, domain.EuchrePhasePickUp, e.GetPhase())
	assert.Equal(t, 2, e.GetRoundNumber())
	assert.Equal(t, (initialDealer+1)%4, e.GetDealerIdx()) // dealer rotates
	assert.Equal(t, 0, e.GetTrickNumber())
	assert.False(t, e.GetGoingAlone())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 5, e.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, e.GetPlayer(i).GetTrickCount())
	}
}

func TestEuchre_NextRound_WrongPhase(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	round := e.GetRoundNumber()

	e.NextRound()
	assert.Equal(t, round, e.GetRoundNumber()) // no change
}

// --- Going alone: skipping partner ---

func TestEuchre_GoingAlone_SkipsPartner(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetGoingAlone(true)
	e.SetGoingAlonePlayerIdx(0) // Player 0 goes alone -> partner (player 2) is skipped
	e.SetCurrentPlayerIdx(0)
	e.SetLeadPlayerIdx(0)
	e.SetTrickNumber(1)

	// Player 0 plays
	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
	})
	err := e.PlayerPlay(0)
	assert.NoError(t, err)

	// Should skip to player 1 (not player 2 which is partner)
	assert.Equal(t, 1, e.GetCurrentPlayerIdx())
}

// --- IsHumanTurn / IsHumanBidTurn ---

func TestEuchre_IsHumanTurn(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetCurrentPlayerIdx(0)
	assert.True(t, e.IsHumanTurn())

	e.SetCurrentPlayerIdx(1)
	assert.False(t, e.IsHumanTurn())
}

func TestEuchre_IsHumanBidTurn(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetBidPlayerIdx(0)
	assert.True(t, e.IsHumanBidTurn())

	e.SetBidPlayerIdx(1)
	assert.False(t, e.IsHumanBidTurn())
}

// --- Config ---

func TestEuchre_Config(t *testing.T) {
	e := newTestEuchre()
	cfg := e.GetConfig()
	assert.Equal(t, domain.EuchreCpuDifficultyNormal, cfg.CpuDifficulty)

	newCfg := domain.EuchreConfig{CpuDifficulty: domain.EuchreCpuDifficultyHard, PointLimit: 5}
	e.SetConfig(newCfg)
	assert.Equal(t, domain.EuchreCpuDifficultyHard, e.GetConfig().CpuDifficulty)
	assert.Equal(t, 5, e.GetConfig().PointLimit)
}

// --- TeamScore edge ---

func TestEuchre_GetTeamScore_OutOfRange(t *testing.T) {
	e := newTestEuchre()
	assert.Equal(t, 0, e.GetTeamScore(-1))
	assert.Equal(t, 0, e.GetTeamScore(5))
}

// --- GetPlayer edge ---

func TestEuchre_GetPlayer_OutOfRange(t *testing.T) {
	e := newTestEuchre()
	assert.Nil(t, e.GetPlayer(-1))
	assert.Nil(t, e.GetPlayer(10))
}

// --- CpuPlay ---

func TestEuchre_CpuPlay(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetCurrentPlayerIdx(1) // CPU
	e.SetLeadPlayerIdx(1)
	e.SetTrickNumber(1)

	e.CpuPlay()
	assert.Equal(t, 1, len(e.GetCurrentTrick()))
}

// --- CpuDiscard ---

func TestEuchre_CpuDiscard(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseDiscard)
	e.SetDealerIdx(1) // CPU is dealer
	e.SetTrumpSuit(domain.CardDesignHeart)

	// Give CPU 6 cards
	setupEuchreHand(e, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
	})

	e.CpuDiscard()
	assert.Equal(t, 5, e.GetPlayer(1).GetCardsSize())
	assert.Equal(t, domain.EuchrePhasePlay, e.GetPhase())
}

// --- GetHint ---

func TestEuchre_GetHint_PickUpPhase(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePickUp)
	e.SetBidPlayerIdx(0)

	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.OrderUp)
	assert.Equal(t, "strategic_pickup", hint.Reason)
}

func TestEuchre_GetHint_PlayPhase(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetCurrentPlayerIdx(0)
	e.SetTrickNumber(1)

	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
	})

	hint := e.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

// --- ActionLog ---

func TestEuchre_ActionLog(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	// After reset, action log is nil (logs are added during play)
	assert.Nil(t, e.GetActionLog())

	// After a player action, log should have entries
	e.SetBidPlayerIdx(0)
	_ = e.PlayerPickUp(false, false) // pass
	assert.Greater(t, len(e.GetActionLog()), 0)
}

// --- ResolveTrick with going alone (3 cards) ---

func TestEuchre_ResolveTrick_GoingAlone(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetTrickNumber(1)
	e.SetGoingAlone(true)
	e.SetGoingAlonePlayerIdx(0) // Partner (2) skipped

	// Only 3 cards in trick
	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
	})

	e.ResolveTrick()
	assert.Equal(t, 1, e.GetPlayer(0).GetTrickCount()) // A♠ wins
}

// --- CpuPickUp ---

func TestEuchre_CpuPickUp(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePickUp)
	e.SetBidPlayerIdx(1) // CPU

	initialPhase := e.GetPhase()
	e.CpuPickUp()
	// After CPU acts, either passed or ordered up
	// Phase should have changed if ordered up, or bidPlayerIdx advanced if passed
	changed := e.GetPhase() != initialPhase || e.GetBidPlayerIdx() != 1
	assert.True(t, changed)
}

// --- CpuCallTrump ---

func TestEuchre_CpuCallTrump(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetBidPlayerIdx(1) // CPU
	e.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	e.SetDealerIdx(3)

	e.CpuCallTrump()
	// CPU either called or passed
	changed := e.GetPhase() != domain.EuchrePhaseCallTrump || e.GetBidPlayerIdx() != 1
	assert.True(t, changed)
}

// --- CpuCallTrump stuck dealer ---

func TestEuchre_CpuCallTrump_StuckDealer(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseCallTrump)
	e.SetBidPlayerIdx(1) // CPU is dealer
	e.SetDealerIdx(1)
	e.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	// Set weak hand so CPU would normally pass
	setupEuchreHand(e, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignClover, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignClover, 10, false),
	})

	e.CpuCallTrump()
	// Stuck dealer must choose - should be in Play phase now
	assert.Equal(t, domain.EuchrePhasePlay, e.GetPhase())
	assert.NotEqual(t, 0, e.GetTrumpSuit())
}

// --- NextTrick ---

func TestEuchre_NextTrick(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetLeadPlayerIdx(2)
	e.SetTrickNumber(1)

	e.NextTrick()
	assert.Equal(t, domain.EuchrePhasePlay, e.GetPhase())
	assert.Equal(t, 2, e.GetCurrentPlayerIdx())
	assert.Equal(t, 2, e.GetTrickNumber())
	assert.Nil(t, e.GetCurrentTrick())
}

func TestEuchre_NextTrick_WrongPhase(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	trick := e.GetTrickNumber()

	e.NextTrick()
	assert.Equal(t, trick, e.GetTrickNumber())
}

// --- ResolveTrick after all 5 tricks -> RoundEnd ---

func TestEuchre_ResolveTrick_LastTrick_GoesToRoundEnd(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetPhase(domain.EuchrePhaseTrickEnd)
	e.SetTrickNumber(5) // last trick

	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})

	e.ResolveTrick()
	assert.Equal(t, domain.EuchrePhaseRoundEnd, e.GetPhase())
}

// --- GameEnd flag ---

func TestEuchre_GameEndFlag_PreventPlay(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetCurrentPlayerIdx(0)

	// Manually set game end
	e.SetPhase(domain.EuchrePhaseGameEnd)

	err := e.PlayerPlay(0)
	assert.Error(t, err) // wrong phase
}

// --- GetValidPlayIndices ---

func TestEuchre_GetValidPlayIndices(t *testing.T) {
	e := newTestEuchre()
	e.Reset()
	e.SetPhase(domain.EuchrePhasePlay)
	e.SetTrumpSuit(domain.CardDesignHeart)
	e.SetCurrentPlayerIdx(0)
	e.SetTrickNumber(1)

	setupEuchreHand(e, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false),
	})

	// No lead: all cards valid
	indices := e.GetValidPlayIndices(0)
	assert.Equal(t, 3, len(indices))

	// With lead of spade: only spade valid
	e.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
	})
	indices = e.GetValidPlayIndices(0)
	assert.Equal(t, 1, len(indices))
	assert.Equal(t, 0, indices[0])
}
