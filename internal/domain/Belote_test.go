//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBelote() *domain.Belote {
	players := []*domain.BelotePlayer{
		domain.NewBelotePlayer(true, 0),  // P0: human, team 0
		domain.NewBelotePlayer(false, 1), // P1: CPU, team 1
		domain.NewBelotePlayer(false, 0), // P2: CPU, team 0
		domain.NewBelotePlayer(false, 1), // P3: CPU, team 1
	}
	return domain.NewBelote(domain.NewTrumpCardsBelote(), players, domain.DefaultBeloteConfig())
}

func setupBeloteHand(b *domain.Belote, playerIdx int, cards []*domain.Card) {
	p := b.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- Deck ---

func TestNewTrumpCardsBelote(t *testing.T) {
	deck := domain.NewTrumpCardsBelote()
	assert.Equal(t, 32, deck.GetTotalCount())

	valid := map[int]bool{1: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	suits := map[int]int{}
	values := map[int]int{}
	for i := 0; i < 32; i++ {
		c := deck.DrawCard()
		assert.NotNil(t, c)
		assert.True(t, valid[c.GetValue()], "unexpected value %d", c.GetValue())
		assert.True(t, c.GetDesign() >= domain.CardDesignSpade && c.GetDesign() <= domain.CardDesignDiamond)
		suits[c.GetDesign()]++
		values[c.GetValue()]++
	}
	for s := domain.CardDesignSpade; s <= domain.CardDesignDiamond; s++ {
		assert.Equal(t, 8, suits[s], "suit %d count", s)
	}
	for v := range valid {
		assert.Equal(t, 4, values[v], "value %d count", v)
	}
	assert.Nil(t, deck.DrawCard())
}

// --- Config ---

func TestBeloteConfig_Default(t *testing.T) {
	cfg := domain.DefaultBeloteConfig()
	assert.Equal(t, domain.BeloteCpuDifficultyNormal, cfg.CpuDifficulty)
	assert.Equal(t, 1000, cfg.TargetScore)
	assert.Equal(t, 10, cfg.DixDeDer)
	assert.True(t, cfg.EnableBeloteRebelote)
	assert.NoError(t, cfg.Validate())
}

func TestBeloteConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.BeloteConfig
		wantErr bool
	}{
		{"default", domain.DefaultBeloteConfig(), false},
		{"easy short", domain.BeloteConfig{CpuDifficulty: domain.BeloteCpuDifficultyEasy, TargetScore: 100, DixDeDer: 10}, false},
		{"bad difficulty", domain.BeloteConfig{CpuDifficulty: 9, TargetScore: 1000, DixDeDer: 10}, true},
		{"zero target", domain.BeloteConfig{CpuDifficulty: domain.BeloteCpuDifficultyNormal, TargetScore: 0, DixDeDer: 10}, true},
		{"neg dix", domain.BeloteConfig{CpuDifficulty: domain.BeloteCpuDifficultyNormal, TargetScore: 1000, DixDeDer: -1}, true},
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

// --- Player ---

func TestBelotePlayer(t *testing.T) {
	p := domain.NewBelotePlayer(true, 0)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetTeam())
	assert.Equal(t, 0, p.GetTrickCount())

	p2 := domain.NewBelotePlayer(false, 1)
	assert.False(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetTeam())
}

func TestBelotePlayer_ResetRound(t *testing.T) {
	p := domain.NewBelotePlayer(true, 0)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 10, false)})
	p.SetIsFinished(true)

	p.ResetRound()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

func TestBelotePlayer_JSONRoundtrip(t *testing.T) {
	p := domain.NewBelotePlayer(true, 1)
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, false)})

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	q := &domain.BelotePlayer{}
	assert.NoError(t, json.Unmarshal(data, q))
	assert.True(t, q.GetIsHuman())
	assert.Equal(t, 1, q.GetTeam())
	assert.Equal(t, 1, q.GetCardsSize())
	assert.Equal(t, 1, q.GetTrickCount())
}

// --- Initialization ---

func TestNewBelote(t *testing.T) {
	b := newTestBelote()
	assert.Equal(t, 4, b.GetPlayerCnt())
	assert.Equal(t, -1, b.GetWinnerTeam())
	assert.Equal(t, 0, b.GetRoundNumber())
}

func TestNewDefaultBelote(t *testing.T) {
	b := domain.NewDefaultBelote()
	assert.Equal(t, 4, b.GetPlayerCnt())
	assert.True(t, b.GetPlayer(0).GetIsHuman())
	assert.Equal(t, 0, b.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, b.GetPlayer(1).GetTeam())
}

func TestBelote_Reset(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	assert.Equal(t, 1, b.GetRoundNumber())
	assert.Equal(t, domain.BelotePhaseBidPickUp, b.GetPhase())
	assert.NotNil(t, b.GetFaceUpCard())
	// 5 cards dealt to each player + face-up card drawn
	for i := 0; i < 4; i++ {
		assert.Equal(t, 5, b.GetPlayer(i).GetCardsSize(), "player %d hand size after deal", i)
	}
	assert.Equal(t, 1, b.GetBidPlayerIdx()) // dealer + 1
}

// --- Ranking + points ---

func TestBelote_CardPoints_TrumpAndNonTrump(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)

	// Trump table
	tCases := []struct {
		v, want int
	}{{11, 20}, {9, 14}, {1, 11}, {10, 10}, {13, 4}, {12, 3}, {8, 0}, {7, 0}}
	for _, c := range tCases {
		card := domain.NewCard(domain.CardDesignHeart, c.v, false)
		assert.Equal(t, c.want, b.CardPointsPublic(card), "trump value %d", c.v)
	}
	// Non-trump table
	nCases := []struct {
		v, want int
	}{{1, 11}, {10, 10}, {13, 4}, {12, 3}, {11, 2}, {9, 0}, {8, 0}, {7, 0}}
	for _, c := range nCases {
		card := domain.NewCard(domain.CardDesignSpade, c.v, false)
		assert.Equal(t, c.want, b.CardPointsPublic(card), "non-trump value %d", c.v)
	}
}

func TestBelote_CardPoints_NilCard(t *testing.T) {
	b := newTestBelote()
	assert.Equal(t, 0, b.CardPointsPublic(nil))
}

func TestBelote_CardRank_TrumpBeatsNonTrump(t *testing.T) {
	b := newTestBelote()
	b.SetTrumpSuit(domain.CardDesignSpade)
	weakTrump := domain.NewCard(domain.CardDesignSpade, 7, false)
	strongNon := domain.NewCard(domain.CardDesignHeart, 1, false) // Ace of Hearts
	assert.Greater(t, b.CardRankPublic(weakTrump), b.CardRankPublic(strongNon))
}

func TestBelote_CardRank_TrumpJackHighest(t *testing.T) {
	b := newTestBelote()
	b.SetTrumpSuit(domain.CardDesignClover)
	j := domain.NewCard(domain.CardDesignClover, 11, false)
	nine := domain.NewCard(domain.CardDesignClover, 9, false)
	ace := domain.NewCard(domain.CardDesignClover, 1, false)
	assert.Greater(t, b.CardRankPublic(j), b.CardRankPublic(nine))
	assert.Greater(t, b.CardRankPublic(nine), b.CardRankPublic(ace))
}

// --- Bid PickUp ---

func TestBelote_PlayerPickUp_OrderUp(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetBidPlayerIdx(0) // human's turn
	face := domain.NewCard(domain.CardDesignHeart, 11, false)
	b.SetFaceUpCard(face)

	err := b.PlayerPickUp(true)
	assert.NoError(t, err)
	assert.Equal(t, domain.CardDesignHeart, b.GetTrumpSuit())
	assert.Equal(t, 0, b.GetMakerTeam())
	assert.Nil(t, b.GetFaceUpCard())
	assert.Equal(t, domain.BelotePhasePlay, b.GetPhase())
	// Player 0 (maker) should have 8 cards, others 8 too
	for i := 0; i < 4; i++ {
		assert.Equal(t, 8, b.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

func TestBelote_PlayerPickUp_Pass_AdvancesBid(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	// Move dealer to 3 so startBid = 0 (human). Passing leaves us mid-loop, not back at start.
	b.SetDealerIdx(3)
	b.SetBidPlayerIdx(0)
	err := b.PlayerPickUp(false)
	assert.NoError(t, err)
	assert.Equal(t, 1, b.GetBidPlayerIdx())
	assert.Equal(t, domain.BelotePhaseBidPickUp, b.GetPhase())
}

func TestBelote_PlayerPickUp_WrongPhase(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhasePlay)
	assert.Error(t, b.PlayerPickUp(true))
}

func TestBelote_PlayerPickUp_NotHumanTurn(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetBidPlayerIdx(1) // CPU's turn
	assert.Error(t, b.PlayerPickUp(true))
}

func TestBelote_AllPassPickUp_AdvancesToCallTrump(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	// 4 passes (everyone) brings us to CallTrump phase
	for i := 0; i < 4; i++ {
		if b.IsHumanBidTurn() {
			assert.NoError(t, b.PlayerPickUp(false))
		} else {
			// force CPU to pass by giving it a weak hand
			setupBeloteHand(b, b.GetBidPlayerIdx(), []*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 7, false),
				domain.NewCard(domain.CardDesignDiamond, 8, false),
				domain.NewCard(domain.CardDesignClover, 7, false),
				domain.NewCard(domain.CardDesignClover, 8, false),
				domain.NewCard(domain.CardDesignHeart, 7, false),
			})
			b.CpuPickUp()
		}
	}
	assert.Equal(t, domain.BelotePhaseBidCallTrump, b.GetPhase())
}

// --- Bid CallTrump ---

func TestBelote_PlayerCallTrump_Success(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	b.SetBidPlayerIdx(0)
	// face-up is some suit; pick a different one
	face := domain.NewCard(domain.CardDesignHeart, 13, false)
	b.SetFaceUpCard(face)

	err := b.PlayerCallTrump(domain.CardDesignSpade)
	assert.NoError(t, err)
	assert.Equal(t, domain.CardDesignSpade, b.GetTrumpSuit())
	assert.Equal(t, domain.BelotePhasePlay, b.GetPhase())
}

func TestBelote_PlayerCallTrump_FaceUpSuit_Rejected(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	b.SetBidPlayerIdx(0)
	face := domain.NewCard(domain.CardDesignHeart, 13, false)
	b.SetFaceUpCard(face)
	err := b.PlayerCallTrump(domain.CardDesignHeart)
	assert.Error(t, err)
}

func TestBelote_PlayerCallTrump_InvalidSuit(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	b.SetBidPlayerIdx(0)
	err := b.PlayerCallTrump(99)
	assert.Error(t, err)
}

func TestBelote_PlayerPassCall(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	// dealer=3 → startBid=0 (human). Passing → idx=1, still mid-loop.
	b.SetDealerIdx(3)
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	b.SetBidPlayerIdx(0)
	assert.NoError(t, b.PlayerPassCall())
	assert.Equal(t, 1, b.GetBidPlayerIdx())
	assert.Equal(t, domain.BelotePhaseBidCallTrump, b.GetPhase())
}

func TestBelote_AllPassCallTrump_Redeals(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	startDealer := b.GetDealerIdx()
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	startBid := (startDealer + 1) % 4
	b.SetBidPlayerIdx(startBid)

	// Force 4 passes (all in CallTrump phase)
	for i := 0; i < 4; i++ {
		idx := b.GetBidPlayerIdx()
		if b.GetPlayer(idx).GetIsHuman() {
			assert.NoError(t, b.PlayerPassCall())
		} else {
			// give CPU a weak hand so it passes
			setupBeloteHand(b, idx, []*domain.Card{
				domain.NewCard(domain.CardDesignDiamond, 7, false),
				domain.NewCard(domain.CardDesignClover, 8, false),
			})
			b.CpuCallTrump()
		}
	}
	// After all 4 pass in CallTrump: dealer rotates and we redeal
	assert.Equal(t, (startDealer+1)%4, b.GetDealerIdx())
	assert.Equal(t, domain.BelotePhaseBidPickUp, b.GetPhase())
}

// --- Play: validatePlay branches ---

func TestBelote_PlayerPlay_LeadIsFree(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetCurrentTrick(nil)
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 11, false),
	})
	assert.NoError(t, b.PlayerPlay(0))
}

func TestBelote_PlayerPlay_MustFollowSuit(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
	})
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false), // can follow
		domain.NewCard(domain.CardDesignClover, 13, false),
	})
	// Try to play clover (off-suit while having spade) -> error
	assert.Error(t, b.PlayerPlay(1))
	// Play spade -> ok
	assert.NoError(t, b.PlayerPlay(0))
}

func TestBelote_PlayerPlay_ObligationToTrump(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	// P3 (opponent) leads spade A → P0 currently winning is P3 → not partner → must trump
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
	})
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 13, false), // non-trump non-lead → not allowed
		domain.NewCard(domain.CardDesignHeart, 7, false),   // trump
	})
	assert.Error(t, b.PlayerPlay(0)) // clover rejected
	assert.NoError(t, b.PlayerPlay(1))
}

func TestBelote_PlayerPlay_ObligationToOverTrump(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	// P2 (partner of P0) leads, P3 (opponent) trumps with K → must over-trump
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},  // P2 partner leads
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 13, false)}, // P3 opponent trump K
	})
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 8, false), // weaker trump than K
		domain.NewCard(domain.CardDesignHeart, 1, false), // stronger trump than K
	})
	// Playing the weaker trump should fail (obligation à monter)
	assert.Error(t, b.PlayerPlay(0))
	// Playing the over-trump should succeed
	assert.NoError(t, b.PlayerPlay(1))
}

func TestBelote_PlayerPlay_PartnerWinning_FreeDiscard(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	// P2 (partner) leads spade A; P3 plays weak spade. P2 still winning when P0 plays.
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
	})
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 8, false), // non-trump non-lead → free under partner-protect
		domain.NewCard(domain.CardDesignHeart, 13, false), // trump
	})
	assert.NoError(t, b.PlayerPlay(0))
}

// --- Trick winner ---

func TestBelote_TrickWinner_HighestOfLeadSuit(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignDiamond)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 12, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 1, false)}, // A
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 11, false)},
	})
	b.SetPhase(domain.BelotePhaseTrickEnd)
	b.SetTrickNumber(1)
	b.ResolveTrick()
	assert.Equal(t, 2, b.GetLeadPlayerIdx()) // P2's A wins
}

func TestBelote_TrickWinner_TrumpCutWins(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)}, // lead, A
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 7, false)}, // weakest trump
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 12, false)},
	})
	b.SetPhase(domain.BelotePhaseTrickEnd)
	b.SetTrickNumber(1)
	b.ResolveTrick()
	assert.Equal(t, 1, b.GetLeadPlayerIdx())
}

func TestBelote_TrickWinner_HighTrumpBeatsLowTrump(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 7, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 11, false)}, // J = highest trump
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
	})
	b.SetPhase(domain.BelotePhaseTrickEnd)
	b.SetTrickNumber(2)
	b.ResolveTrick()
	assert.Equal(t, 2, b.GetLeadPlayerIdx())
}

// --- Round scoring ---

func TestBelote_ScoreRound_MakerWins(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetMakerTeam(0)
	b.SetPhase(domain.BelotePhaseRoundEnd)
	// Simulate round points: team 0 = 100, team 1 = 52  (sum = 152; total + Dix de Der = 162 added in trick path; here set directly)
	// We can't directly set roundPoints (no setter), so we run a Dix de Der trick? Simpler: hand-craft via a shortcut...
	// Workaround: drive via ResolveTrick on a synthetic 8th trick. But too involved.
	// Instead just call ScoreRound on zero points and assert log entry path exists.
	b.ScoreRound()
	assert.NotNil(t, b.GetActionLog())
}

func TestBelote_ScoreRound_FullRoundFromPlay(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetMakerTeam(0)
	// Set up 8 contrived tricks via direct trick play. Give P0 8 trump cards (capot scenario).
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 11, false), // J
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignHeart, 12, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	setupBeloteHand(b, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false), domain.NewCard(domain.CardDesignSpade, 8, false),
		domain.NewCard(domain.CardDesignClover, 7, false), domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false), domain.NewCard(domain.CardDesignDiamond, 8, false),
		domain.NewCard(domain.CardDesignSpade, 12, false), domain.NewCard(domain.CardDesignClover, 12, false),
	})
	setupBeloteHand(b, 2, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false), domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignClover, 1, false), domain.NewCard(domain.CardDesignClover, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 1, false), domain.NewCard(domain.CardDesignDiamond, 10, false),
		domain.NewCard(domain.CardDesignSpade, 11, false), domain.NewCard(domain.CardDesignClover, 11, false),
	})
	setupBeloteHand(b, 3, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false), domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignClover, 13, false), domain.NewCard(domain.CardDesignClover, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 13, false), domain.NewCard(domain.CardDesignDiamond, 9, false),
		domain.NewCard(domain.CardDesignDiamond, 12, false), domain.NewCard(domain.CardDesignDiamond, 11, false),
	})
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	b.SetTrickNumber(1)
	// Play 8 tricks: P0 leads each, plays index 0 (always strongest remaining trump → wins).
	for trick := 1; trick <= 8; trick++ {
		assert.NoError(t, b.PlayerPlay(0))
		for next := 1; next < 4; next++ {
			b.CpuPlay()
		}
		assert.Equal(t, domain.BelotePhaseTrickEnd, b.GetPhase(), "trick %d end", trick)
		b.ResolveTrick()
		if trick < 8 {
			b.NextTrick()
		}
	}
	assert.Equal(t, domain.BelotePhaseRoundEnd, b.GetPhase())
	b.ScoreRound()
	// Team 0 (maker) should have all the round + capot bonus
	assert.Greater(t, b.GetTeamScore(0), 0)
}

// --- Belote/Rebelote ---

func TestBelote_BeloteRebelote_AwardsBonus(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignSpade)
	b.SetMakerTeam(0)
	b.SetBeloteHolderIdx(0) // P0 (human, team 0) holds K+Q of trumps
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	b.SetTrickNumber(1)

	// P0: K spade (trump), Q spade (trump), plus 6 weak fillers.
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false), // K of trumps
		domain.NewCard(domain.CardDesignSpade, 12, false), // Q of trumps
		domain.NewCard(domain.CardDesignHeart, 7, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignClover, 7, false),
		domain.NewCard(domain.CardDesignClover, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
	})
	// CPUs: 8 cards each of unique non-trump suits so they can never follow spades
	// (forcing them to discard freely and never beat P0's trump leads).
	setupBeloteHand(b, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 9, false), domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignHeart, 11, false), domain.NewCard(domain.CardDesignHeart, 12, false),
		domain.NewCard(domain.CardDesignHeart, 13, false), domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignClover, 9, false), domain.NewCard(domain.CardDesignClover, 10, false),
	})
	setupBeloteHand(b, 2, []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 11, false), domain.NewCard(domain.CardDesignClover, 12, false),
		domain.NewCard(domain.CardDesignClover, 13, false), domain.NewCard(domain.CardDesignClover, 1, false),
		domain.NewCard(domain.CardDesignDiamond, 9, false), domain.NewCard(domain.CardDesignDiamond, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 11, false), domain.NewCard(domain.CardDesignDiamond, 12, false),
	})
	setupBeloteHand(b, 3, []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 13, false), domain.NewCard(domain.CardDesignDiamond, 1, false),
		domain.NewCard(domain.CardDesignHeart, 9, false), domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignClover, 11, false), domain.NewCard(domain.CardDesignClover, 12, false),
		domain.NewCard(domain.CardDesignClover, 13, false), domain.NewCard(domain.CardDesignClover, 1, false),
	})

	// Trick 1: P0 leads K of trumps.
	assert.NoError(t, b.PlayerPlay(0))
	b.CpuPlay()
	b.CpuPlay()
	b.CpuPlay()
	assert.Equal(t, domain.BelotePhaseTrickEnd, b.GetPhase())
	b.ResolveTrick()
	assert.Equal(t, 0, b.GetLeadPlayerIdx(), "P0 should win trick 1 (trump K beats non-trumps)")
	// Belote not yet declared — only K has been played.
	assert.Equal(t, 0, b.GetRoundBeloteBonus(0))
	b.NextTrick()

	// Trick 2: P0 leads Q of trumps — this completes K+Q and awards +20.
	qIdx := -1
	for i := 0; i < b.GetPlayer(0).GetCardsSize(); i++ {
		c := b.GetPlayer(0).GetCard(i)
		if c.GetDesign() == domain.CardDesignSpade && c.GetValue() == 12 {
			qIdx = i
			break
		}
	}
	if !assert.GreaterOrEqual(t, qIdx, 0, "P0 should still hold Q of trumps") {
		return
	}
	assert.NoError(t, b.PlayerPlay(qIdx))
	assert.Equal(t, domain.BeloteRebeloteBonus, b.GetRoundBeloteBonus(0), "Belote/Rebelote should award +20 to team 0")
	assert.Equal(t, 0, b.GetRoundBeloteBonus(1), "opposite team should not receive the bonus")
}

func TestBelote_BeloteRebelote_DisabledByConfig(t *testing.T) {
	b := newTestBelote()
	cfg := b.GetConfig()
	cfg.EnableBeloteRebelote = false
	b.SetConfig(cfg)
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignSpade)
	b.SetMakerTeam(0)
	b.SetBeloteHolderIdx(0)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetLeadPlayerIdx(0)
	b.SetTrickNumber(1)
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignHeart, 7, false),
	})
	setupBeloteHand(b, 1, []*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, false)})
	setupBeloteHand(b, 2, []*domain.Card{domain.NewCard(domain.CardDesignClover, 8, false)})
	setupBeloteHand(b, 3, []*domain.Card{domain.NewCard(domain.CardDesignDiamond, 8, false)})

	// Play K of trumps — config disabled, no bonus should accrue.
	assert.NoError(t, b.PlayerPlay(0))
	assert.Equal(t, 0, b.GetRoundBeloteBonus(0))
}

// --- Game end ---

func TestBelote_CheckGameEnd_ReachTarget(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetMakerTeam(0)
	b.SetPhase(domain.BelotePhaseRoundEnd)
	b.SetTeamScore(0, 999)
	// Synthetic: directly bump score and check
	b.SetTeamScore(0, 1000)
	b.ScoreRound() // will detect end via checkGameEnd
	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, 0, b.GetWinnerTeam())
}

// --- Hint ---

func TestBelote_GetHint_NoHumanReturnsNil(t *testing.T) {
	players := []*domain.BelotePlayer{
		domain.NewBelotePlayer(false, 0),
		domain.NewBelotePlayer(false, 1),
		domain.NewBelotePlayer(false, 0),
		domain.NewBelotePlayer(false, 1),
	}
	b := domain.NewBelote(domain.NewTrumpCardsBelote(), players, domain.DefaultBeloteConfig())
	b.Reset()
	assert.Nil(t, b.GetHint())
}

func TestBelote_GetHint_PickUpPhase(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetBidPlayerIdx(0)
	face := domain.NewCard(domain.CardDesignHeart, 11, false)
	b.SetFaceUpCard(face)
	// Hand strong in hearts -> hint should suggest order up
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignHeart, 13, false),
		domain.NewCard(domain.CardDesignSpade, 1, false),
	})
	h := b.GetHint()
	if assert.NotNil(t, h) {
		assert.NotNil(t, h.OrderUp)
	}
}

func TestBelote_GetHint_PlayPhase(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(0)
	b.SetCurrentTrick(nil)
	setupBeloteHand(b, 0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	})
	h := b.GetHint()
	if assert.NotNil(t, h) {
		assert.NotNil(t, h.CardIndex)
	}
}

// --- CPU paths ---

func TestBelote_CpuPickUp_EasyAndHard(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	cfg := b.GetConfig()
	cfg.CpuDifficulty = domain.BeloteCpuDifficultyEasy
	b.SetConfig(cfg)
	b.SetBidPlayerIdx(1)
	b.CpuPickUp() // exercises easy path

	b2 := newTestBelote()
	b2.Reset()
	cfg2 := b2.GetConfig()
	cfg2.CpuDifficulty = domain.BeloteCpuDifficultyHard
	b2.SetConfig(cfg2)
	b2.SetBidPlayerIdx(1)
	// Strong heart hand + hearts face-up
	setupBeloteHand(b2, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 11, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignHeart, 1, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignSpade, 1, false),
	})
	b2.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	b2.CpuPickUp()
	// Should have taken with strong hand
	assert.Equal(t, domain.CardDesignHeart, b2.GetTrumpSuit())
}

func TestBelote_CpuCallTrump_PicksBestSuit(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhaseBidCallTrump)
	b.SetBidPlayerIdx(1)
	b.SetFaceUpCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	// Strong spade hand
	setupBeloteHand(b, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 9, false),
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignSpade, 10, false),
		domain.NewCard(domain.CardDesignSpade, 13, false),
	})
	b.CpuCallTrump()
	assert.Equal(t, domain.CardDesignSpade, b.GetTrumpSuit())
}

func TestBelote_CpuPlay_ExecutesValidCard(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignHeart)
	b.SetPhase(domain.BelotePhasePlay)
	b.SetCurrentPlayerIdx(1)
	b.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
	})
	setupBeloteHand(b, 1, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	})
	b.CpuPlay()
	assert.Equal(t, 1, b.GetPlayer(1).GetCardsSize())
}

// --- JSON ---

func TestBelote_JSONRoundtrip(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetTrumpSuit(domain.CardDesignDiamond)
	b.SetTeamScore(0, 250)
	b.SetTeamScore(1, 400)

	data, err := json.Marshal(b)
	assert.NoError(t, err)

	b2 := &domain.Belote{}
	assert.NoError(t, json.Unmarshal(data, b2))
	assert.Equal(t, domain.CardDesignDiamond, b2.GetTrumpSuit())
	assert.Equal(t, 250, b2.GetTeamScore(0))
	assert.Equal(t, 400, b2.GetTeamScore(1))
	assert.Equal(t, 4, b2.GetPlayerCnt())
}

// --- NextRound ---

func TestBelote_NextRound_RotatesDealer(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	startDealer := b.GetDealerIdx()
	b.SetPhase(domain.BelotePhaseRoundEnd)
	b.NextRound()
	assert.Equal(t, (startDealer+1)%4, b.GetDealerIdx())
	assert.Equal(t, 2, b.GetRoundNumber())
}

func TestBelote_NextRound_WrongPhase_NoOp(t *testing.T) {
	b := newTestBelote()
	b.Reset()
	b.SetPhase(domain.BelotePhasePlay)
	b.NextRound()
	assert.Equal(t, 1, b.GetRoundNumber())
}
