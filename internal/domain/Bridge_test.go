//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test helpers ---

func newTestBridge() *Bridge {
	players := []*BridgePlayer{
		NewBridgePlayer(true, 0),  // North (human, team 0)
		NewBridgePlayer(false, 1), // East (CPU, team 1)
		NewBridgePlayer(false, 0), // South (CPU, team 0)
		NewBridgePlayer(false, 1), // West (CPU, team 1)
	}
	return NewBridge(NewTrumpCards(0), players, DefaultBridgeConfig())
}

func newTestBridgeWithReset() *Bridge {
	b := newTestBridge()
	b.Reset()
	return b
}

func setupBridgePlayPhase(b *Bridge) {
	// 各プレイヤーに13枚ずつ手動配布
	for _, p := range b.players {
		p.ResetRound()
	}

	// スペードの1-13をPlayer0に
	for v := 1; v <= 13; v++ {
		b.players[0].AddCard(NewCard(CardDesignSpade, v, true))
	}
	// クラブの1-13をPlayer1に
	for v := 1; v <= 13; v++ {
		b.players[1].AddCard(NewCard(CardDesignClover, v, true))
	}
	// ハートの1-13をPlayer2に
	for v := 1; v <= 13; v++ {
		b.players[2].AddCard(NewCard(CardDesignHeart, v, true))
	}
	// ダイヤの1-13をPlayer3に
	for v := 1; v <= 13; v++ {
		b.players[3].AddCard(NewCard(CardDesignDiamond, v, true))
	}

	b.phase = BridgePhasePlay
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitSpade
	b.trumpSuit = CardDesignSpade
	b.declarerIdx = 0
	b.dummyIdx = 2
	b.leadPlayerIdx = 1
	b.currentPlayerIdx = 1
	b.trickNumber = 1
	b.openingLeadDone = true
	b.currentTrick = nil
}

// --- Tests ---

func TestNewBridge(t *testing.T) {
	b := newTestBridge()
	assert.Equal(t, 4, b.GetPlayerCnt())
	assert.Equal(t, -1, b.GetWinnerTeam())
	assert.False(t, b.GetGameEndFlag())
}

func TestBridgeReset(t *testing.T) {
	b := newTestBridgeWithReset()

	assert.Equal(t, BridgePhaseBid, b.GetPhase())
	assert.Equal(t, 1, b.GetRoundNumber())
	assert.False(t, b.GetGameEndFlag())
	assert.Equal(t, -1, b.GetWinnerTeam())

	// 各プレイヤーに13枚ずつ配られる
	for i := 0; i < BridgePlayerCnt; i++ {
		assert.Equal(t, BridgeHandSize, b.GetPlayer(i).GetCardsSize())
	}

	// ビッド手番はディーラーの左隣
	assert.Equal(t, 1, b.GetBidPlayerIdx())
}

func TestBridgePlayerBidPass(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0) // human

	err := b.PlayerBid(int(BridgeBidPass), 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, b.bidPlayerIdx)
}

func TestBridgePlayerBidNormal(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0) // human

	err := b.PlayerBid(int(BridgeBidNormal), 1, BridgeBidSuitClub)
	assert.NoError(t, err)
	assert.Equal(t, 1, b.contractLevel)
	assert.Equal(t, BridgeBidSuitClub, b.contractSuit)
	assert.Equal(t, 0, b.lastBidderIdx)
}

func TestBridgePlayerBidInvalidLevel(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)

	err := b.PlayerBid(int(BridgeBidNormal), 0, BridgeBidSuitClub)
	assert.Error(t, err)

	err = b.PlayerBid(int(BridgeBidNormal), 8, BridgeBidSuitClub)
	assert.Error(t, err)
}

func TestBridgePlayerBidInvalidSuit(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)

	err := b.PlayerBid(int(BridgeBidNormal), 1, 0)
	assert.Error(t, err)

	err = b.PlayerBid(int(BridgeBidNormal), 1, 6)
	assert.Error(t, err)
}

func TestBridgePlayerBidMustBeHigher(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitHeart

	// Same level, lower suit
	err := b.PlayerBid(int(BridgeBidNormal), 1, BridgeBidSuitClub)
	assert.Error(t, err)

	// Same level, same suit
	err = b.PlayerBid(int(BridgeBidNormal), 1, BridgeBidSuitHeart)
	assert.Error(t, err)

	// Same level, higher suit
	err = b.PlayerBid(int(BridgeBidNormal), 1, BridgeBidSuitSpade)
	assert.NoError(t, err)
}

func TestBridgePlayerBidDouble(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitClub
	b.lastBidTeam = 0
	b.lastBidderIdx = 0
	b.SetBidPlayerIdx(1) // CPU team 1

	// Team 1 player doubles Team 0's bid
	err := b.executeBid(1, BridgeBidDouble, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, b.doubled)
}

func TestBridgePlayerBidDoubleOwnTeam(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitClub
	b.lastBidTeam = 0
	b.lastBidderIdx = 0
	b.SetBidPlayerIdx(0)

	err := b.PlayerBid(int(BridgeBidDouble), 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidDoubleNoContract(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)

	err := b.PlayerBid(int(BridgeBidDouble), 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidDoubleAlreadyDoubled(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitClub
	b.lastBidTeam = 0
	b.doubled = 1
	b.SetBidPlayerIdx(1)

	err := b.executeBid(1, BridgeBidDouble, 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidRedouble(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitClub
	b.lastBidTeam = 0
	b.lastBidderIdx = 0
	b.doubled = 1
	b.SetBidPlayerIdx(0) // Team 0 player redoubles

	err := b.PlayerBid(int(BridgeBidRedouble), 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, b.doubled)
}

func TestBridgePlayerBidRedoubleNotDoubled(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.lastBidTeam = 0
	b.doubled = 0
	b.SetBidPlayerIdx(0)

	err := b.PlayerBid(int(BridgeBidRedouble), 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidRedoubleWrongTeam(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.lastBidTeam = 0
	b.doubled = 1
	b.SetBidPlayerIdx(1) // Team 1 (not the bid team)

	err := b.executeBid(1, BridgeBidRedouble, 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidInvalidType(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)

	err := b.PlayerBid(99, 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidWrongPhase(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetPhase(BridgePhasePlay)

	err := b.PlayerBid(int(BridgeBidPass), 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidGameEnded(t *testing.T) {
	b := newTestBridgeWithReset()
	b.gameEndFlag = true

	err := b.PlayerBid(int(BridgeBidPass), 0, 0)
	assert.Error(t, err)
}

func TestBridgePlayerBidNotHumanTurn(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(1) // CPU

	err := b.PlayerBid(int(BridgeBidPass), 0, 0)
	assert.Error(t, err)
}

func TestBridgeFourPassesRedeal(t *testing.T) {
	b := newTestBridgeWithReset()
	initialDealer := b.GetDealerIdx()

	// 4 consecutive passes with no contract
	b.SetBidPlayerIdx(0)
	_ = b.PlayerBid(int(BridgeBidPass), 0, 0)
	b.SetBidPlayerIdx(b.bidPlayerIdx) // advance manually for CPU
	_ = b.executeBid(1, BridgeBidPass, 0, 0)
	_ = b.executeBid(2, BridgeBidPass, 0, 0)
	_ = b.executeBid(3, BridgeBidPass, 0, 0)

	// Dealer should have rotated
	assert.Equal(t, (initialDealer+1)%BridgePlayerCnt, b.GetDealerIdx())
	assert.Equal(t, BridgePhaseBid, b.GetPhase())
}

func TestBridgeAuctionEndsAfterThreePasses(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)

	// Bid 1C
	_ = b.PlayerBid(int(BridgeBidNormal), 1, BridgeBidSuitClub)

	// 3 passes
	_ = b.executeBid(1, BridgeBidPass, 0, 0)
	_ = b.executeBid(2, BridgeBidPass, 0, 0)
	_ = b.executeBid(3, BridgeBidPass, 0, 0)

	assert.Equal(t, BridgePhasePlay, b.GetPhase())
	assert.Equal(t, 1, b.GetContractLevel())
	assert.Equal(t, BridgeBidSuitClub, b.GetContractSuit())
	assert.Equal(t, 0, b.GetDeclarerIdx())
	assert.Equal(t, 2, b.GetDummyIdx()) // partner
}

func TestBridgeFindFirstBidder(t *testing.T) {
	b := newTestBridgeWithReset()

	// Player 2 bids 1H, then player 0 bids 2H
	b.bidHistory = []*BridgeBidEntry{
		{PlayerIdx: 2, BidType: BridgeBidNormal, Level: 1, Suit: BridgeBidSuitHeart},
		{PlayerIdx: 0, BidType: BridgeBidNormal, Level: 2, Suit: BridgeBidSuitHeart},
	}

	// Team 0 first bidder of Hearts should be player 2
	idx := b.findFirstBidder(0, BridgeBidSuitHeart)
	assert.Equal(t, 2, idx)
}

func TestBridgePlayerPlay(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	// Player 1 (East, CPU) leads
	b.SetCurrentPlayerIdx(1)
	// Bid should fail in Play phase
	err := b.PlayerBid(int(BridgeBidPass), 0, 0)
	assert.Error(t, err) // Phase is Play, not Bid

	// CPU plays
	b.CpuPlay()
	assert.Equal(t, 1, len(b.currentTrick))
}

func TestBridgePlayerPlayHuman(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(0) // human

	err := b.PlayerPlay(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(b.currentTrick))
}

func TestBridgePlayerPlayInvalidIndex(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(0)

	err := b.PlayerPlay(-1)
	assert.Error(t, err)

	err = b.PlayerPlay(100)
	assert.Error(t, err)
}

func TestBridgePlayerPlayNotHumanTurn(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(1) // CPU

	err := b.PlayerPlay(0)
	assert.Error(t, err)
}

func TestBridgePlayerPlayDummy(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	// Human (0) is declarer, dummy is player 2
	b.SetCurrentPlayerIdx(2) // dummy's turn

	err := b.PlayerPlay(0) // human plays dummy's card
	assert.NoError(t, err)
}

func TestBridgePlayerPlayWrongPhase(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	// phase is Bid
	err := b.PlayerPlay(0)
	assert.Error(t, err)
}

func TestBridgePlayerPlayGameEnded(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.gameEndFlag = true

	err := b.PlayerPlay(0)
	assert.Error(t, err)
}

func TestBridgeFollowSuit(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	// Lead with a club
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 5, true)},
	}
	b.SetCurrentPlayerIdx(0) // Human has only spades

	// Player 0 has only spades, so can play any
	err := b.PlayerPlay(0)
	assert.NoError(t, err)
}

func TestBridgeFollowSuitViolation(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	// Give player 0 mixed suits
	b.players[0].Reset()
	b.players[0].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[0].AddCard(NewCard(CardDesignSpade, 10, true))

	// Lead with a club
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 3, true)},
	}
	b.SetCurrentPlayerIdx(0)

	// Try to play spade (index 1) when has club
	err := b.PlayerPlay(1) // spade 10
	assert.Error(t, err)

	// Play club (index 0) should work
	err = b.PlayerPlay(0) // club 5
	assert.NoError(t, err)
}

func TestBridgeResolveTrick(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 10, true)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 5, true)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 8, true)},
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 2, true)}, // trump
	}
	b.SetPhase(BridgePhaseTrickEnd)

	b.ResolveTrick()

	// Spade (trump) should win
	assert.Equal(t, 1, b.players[0].GetTrickCount())
	assert.Equal(t, 0, b.GetLeadPlayerIdx())
}

func TestBridgeResolveTrickNoTrump(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.trumpSuit = -1 // NoTrump

	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 10, true)},
		{PlayerIdx: 2, Card: NewCard(CardDesignClover, 1, true)}, // Ace
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 1, true)},
		{PlayerIdx: 0, Card: NewCard(CardDesignClover, 5, true)},
	}
	b.SetPhase(BridgePhaseTrickEnd)

	b.ResolveTrick()

	// Club Ace should win (NT, lead suit is club)
	assert.Equal(t, 1, b.players[2].GetTrickCount())
}

func TestBridgeNextTrick(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetPhase(BridgePhaseTrickEnd)
	b.SetLeadPlayerIdx(2)
	b.trickNumber = 1

	b.NextTrick()

	assert.Equal(t, BridgePhasePlay, b.GetPhase())
	assert.Equal(t, 2, b.GetCurrentPlayerIdx())
	assert.Equal(t, 2, b.GetTrickNumber())
	assert.Nil(t, b.GetCurrentTrick())
}

func TestBridgeNextTrickWrongPhase(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	b.SetPhase(BridgePhasePlay)

	b.NextTrick()
	// Should not change phase
	assert.Equal(t, BridgePhasePlay, b.GetPhase())
}

func TestBridgeScoreRoundMade(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetPhase(BridgePhaseRoundEnd)
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitSpade
	b.declarerIdx = 0

	// Give declarer team (0) 8 tricks (need 7 = 1+6)
	b.players[0].AddTrick(make([]*Card, 0))
	b.players[0].AddTrick(make([]*Card, 0))
	b.players[0].AddTrick(make([]*Card, 0))
	b.players[0].AddTrick(make([]*Card, 0))
	b.players[2].AddTrick(make([]*Card, 0))
	b.players[2].AddTrick(make([]*Card, 0))
	b.players[2].AddTrick(make([]*Card, 0))
	b.players[2].AddTrick(make([]*Card, 0))

	b.ScoreRound()

	assert.Greater(t, b.GetTeamScore(0), 0)
}

func TestBridgeScoreRoundDown(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetPhase(BridgePhaseRoundEnd)
	b.contractLevel = 4
	b.contractSuit = BridgeBidSuitSpade
	b.declarerIdx = 0

	// Give declarer team (0) 5 tricks (need 10 = 4+6): down 5
	for range 3 {
		b.players[0].AddTrick(make([]*Card, 0))
	}
	for range 2 {
		b.players[2].AddTrick(make([]*Card, 0))
	}

	b.ScoreRound()

	// Defenders (team 1) should get penalty points
	assert.Greater(t, b.GetTeamScore(1), 0)
}

func TestBridgeScoreRoundWrongPhase(t *testing.T) {
	b := newTestBridgeWithReset()
	// Phase is Bid, not RoundEnd
	b.ScoreRound()
	assert.Equal(t, 0, b.GetTeamScore(0))
}

func TestBridgeContractPoints(t *testing.T) {
	tests := []struct {
		name  string
		level int
		suit  int
		want  int
	}{
		{"1C", 1, BridgeBidSuitClub, 20},
		{"2D", 2, BridgeBidSuitDiamond, 40},
		{"1H", 1, BridgeBidSuitHeart, 30},
		{"3S", 3, BridgeBidSuitSpade, 90},
		{"1NT", 1, BridgeBidSuitNT, 40},
		{"3NT", 3, BridgeBidSuitNT, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBridge()
			b.contractLevel = tt.level
			b.contractSuit = tt.suit
			assert.Equal(t, tt.want, b.calcContractPoints())
		})
	}
}

func TestBridgeUndertrickPenalty(t *testing.T) {
	tests := []struct {
		name        string
		vul         bool
		doubled     int
		undertricks int
		want        int
	}{
		{"1 NV undoubled", false, 0, 1, 50},
		{"3 NV undoubled", false, 0, 3, 150},
		{"1 V undoubled", true, 0, 1, 100},
		{"1 NV doubled", false, 1, 1, 100},
		{"2 NV doubled", false, 1, 2, 300},
		{"1 V doubled", true, 1, 1, 200},
		{"2 V doubled", true, 1, 2, 500},
		{"1 NV redoubled", false, 2, 1, 200},
		{"1 V redoubled", true, 2, 1, 400},
		{"2 V redoubled", true, 2, 2, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBridge()
			b.doubled = tt.doubled
			b.vulnerability[0] = tt.vul
			got := b.calcUndertrickPenalty(0, tt.undertricks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBridgeGameEnd(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	b.gamesWon[0] = 1

	// Simulate winning another game
	b.SetPhase(BridgePhaseRoundEnd)
	b.contractLevel = 3
	b.contractSuit = BridgeBidSuitNT
	b.declarerIdx = 0
	b.belowLine = [BridgeTeamCnt]int{}

	// Give declarer team 10 tricks (need 9)
	for range 6 {
		b.players[0].AddTrick(make([]*Card, 0))
	}
	for range 4 {
		b.players[2].AddTrick(make([]*Card, 0))
	}

	b.ScoreRound()

	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, 0, b.GetWinnerTeam())
}

func TestBridgeNextRound(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetPhase(BridgePhaseRoundEnd)
	initialDealer := b.GetDealerIdx()

	b.NextRound()

	assert.Equal(t, BridgePhaseBid, b.GetPhase())
	assert.Equal(t, (initialDealer+1)%BridgePlayerCnt, b.GetDealerIdx())
	assert.Equal(t, 2, b.GetRoundNumber())
}

func TestBridgeNextRoundWrongPhase(t *testing.T) {
	b := newTestBridgeWithReset()
	// Phase is Bid

	b.NextRound()
	// Should not change
	assert.Equal(t, BridgePhaseBid, b.GetPhase())
}

func TestBridgeCpuBid(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(1) // CPU

	b.CpuBid()

	// Should have added a bid entry
	assert.GreaterOrEqual(t, len(b.bidHistory), 1)
}

func TestBridgeCpuBidHumanTurn(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0) // Human

	b.CpuBid()
	// Should not do anything
	assert.Equal(t, 0, len(b.bidHistory))
}

func TestBridgeCpuBidGameEnded(t *testing.T) {
	b := newTestBridgeWithReset()
	b.gameEndFlag = true
	b.SetBidPlayerIdx(1)

	b.CpuBid()
	assert.Equal(t, 0, len(b.bidHistory))
}

func TestBridgeCpuBidWrongPhase(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetPhase(BridgePhasePlay)
	b.SetBidPlayerIdx(1)

	b.CpuBid()
	assert.Equal(t, 0, len(b.bidHistory))
}

func TestBridgeCpuPlay(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(1) // CPU

	b.CpuPlay()
	assert.Equal(t, 1, len(b.currentTrick))
}

func TestBridgeCpuPlayGameEnded(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.gameEndFlag = true

	b.CpuPlay()
	assert.Equal(t, 0, len(b.currentTrick))
}

func TestBridgeCpuPlayDummy(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	// Player 2 is dummy, declarer (0) is human
	b.SetCurrentPlayerIdx(2)

	b.CpuPlay()
	// Should not play since human declarer controls dummy
	assert.Equal(t, 0, len(b.currentTrick))
}

func TestBridgeCpuPlayDummyCpuDeclarer(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.declarerIdx = 1 // CPU declarer
	b.dummyIdx = 3    // CPU dummy
	b.SetCurrentPlayerIdx(3)

	b.CpuPlay()
	// CPU declarer plays dummy's card
	assert.Equal(t, 1, len(b.currentTrick))
}

func TestBridgeIsHumanTurn(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	b.SetCurrentPlayerIdx(0)
	assert.True(t, b.IsHumanTurn())

	b.SetCurrentPlayerIdx(1)
	assert.False(t, b.IsHumanTurn())

	// Dummy turn with human declarer
	b.SetCurrentPlayerIdx(2) // dummy
	assert.True(t, b.IsHumanTurn())
}

func TestBridgeIsHumanBidTurn(t *testing.T) {
	b := newTestBridgeWithReset()

	b.SetBidPlayerIdx(0)
	assert.True(t, b.IsHumanBidTurn())

	b.SetBidPlayerIdx(1)
	assert.False(t, b.IsHumanBidTurn())
}

func TestBridgeGetDummyHand(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	hand := b.GetDummyHand()
	assert.Equal(t, 13, len(hand))

	// Before opening lead
	b.openingLeadDone = false
	hand = b.GetDummyHand()
	assert.Nil(t, hand)
}

func TestBridgeGetValidPlayIndices(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(0)

	indices := b.GetValidPlayIndices(0)
	assert.Equal(t, 13, len(indices)) // All cards valid when leading
}

func TestBridgeGetHintBid(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)

	hint := b.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.BidType)
}

func TestBridgeGetHintPlay(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(0)

	hint := b.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestBridgeGetHintNoHuman(t *testing.T) {
	players := []*BridgePlayer{
		NewBridgePlayer(false, 0),
		NewBridgePlayer(false, 1),
		NewBridgePlayer(false, 0),
		NewBridgePlayer(false, 1),
	}
	b := NewBridge(NewTrumpCards(0), players, DefaultBridgeConfig())
	b.Reset()

	hint := b.GetHint()
	assert.Nil(t, hint)
}

func TestBridgeGetHintBidNotHumanTurn(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(1) // CPU

	hint := b.GetHint()
	assert.Nil(t, hint)
}

func TestBridgeGetHintPlayNotHumanTurn(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(1) // CPU

	hint := b.GetHint()
	assert.Nil(t, hint)
}

func TestBridgeStateGetters(t *testing.T) {
	b := newTestBridgeWithReset()

	assert.NotNil(t, b.GetConfig())
	assert.Equal(t, 0, b.GetDealerIdx())
	assert.Equal(t, -1, b.GetTrumpSuit())
	assert.Equal(t, 0, b.GetContractLevel())
	assert.Equal(t, 0, b.GetContractSuit())
	assert.Equal(t, 0, b.GetDoubled())
	assert.Equal(t, -1, b.GetDeclarerIdx())
	assert.Equal(t, -1, b.GetDummyIdx())
	assert.False(t, b.GetVulnerability(0))
	assert.False(t, b.GetVulnerability(1))
	assert.Equal(t, 0, b.GetTeamScore(0))
	assert.Equal(t, 0, b.GetGamesWon(0))
	assert.Equal(t, 0, b.GetBelowLine(0))
	assert.False(t, b.IsOpeningLeadDone())
	assert.Empty(t, b.GetBidHistory())

	// Out of range
	assert.Nil(t, b.GetPlayer(-1))
	assert.Nil(t, b.GetPlayer(10))
	assert.False(t, b.GetVulnerability(-1))
	assert.Equal(t, 0, b.GetTeamScore(-1))
	assert.Equal(t, 0, b.GetGamesWon(-1))
	assert.Equal(t, 0, b.GetBelowLine(-1))
}

func TestBridgeSetters(t *testing.T) {
	b := newTestBridgeWithReset()

	b.SetConfig(BridgeConfig{CpuDifficulty: BridgeCpuDifficultyHard})
	assert.Equal(t, BridgeCpuDifficultyHard, b.GetConfig().CpuDifficulty)

	b.SetVulnerability(0, true)
	assert.True(t, b.GetVulnerability(0))

	b.SetTeamScore(0, 100)
	assert.Equal(t, 100, b.GetTeamScore(0))

	b.SetGamesWon(0, 1)
	assert.Equal(t, 1, b.GetGamesWon(0))

	b.SetBelowLine(0, 50)
	assert.Equal(t, 50, b.GetBelowLine(0))

	// Out of range setters (should not panic)
	b.SetVulnerability(-1, true)
	b.SetTeamScore(-1, 100)
	b.SetGamesWon(-1, 1)
	b.SetBelowLine(-1, 50)
}

func TestBridgeJSON(t *testing.T) {
	b := newTestBridgeWithReset()
	b.SetBidPlayerIdx(0)
	_ = b.PlayerBid(int(BridgeBidNormal), 1, BridgeBidSuitClub)

	data, err := json.Marshal(b)
	require.NoError(t, err)

	b2 := &Bridge{}
	err = json.Unmarshal(data, b2)
	require.NoError(t, err)

	assert.Equal(t, b.GetPhase(), b2.GetPhase())
	assert.Equal(t, b.GetRoundNumber(), b2.GetRoundNumber())
	assert.Equal(t, b.GetContractLevel(), b2.GetContractLevel())
	assert.Equal(t, b.GetContractSuit(), b2.GetContractSuit())
	assert.Equal(t, b.GetPlayerCnt(), b2.GetPlayerCnt())
}

func TestBridgeBidSuitName(t *testing.T) {
	b := newTestBridge()
	assert.Equal(t, "C", b.bidSuitName(BridgeBidSuitClub))
	assert.Equal(t, "D", b.bidSuitName(BridgeBidSuitDiamond))
	assert.Equal(t, "H", b.bidSuitName(BridgeBidSuitHeart))
	assert.Equal(t, "S", b.bidSuitName(BridgeBidSuitSpade))
	assert.Equal(t, "NT", b.bidSuitName(BridgeBidSuitNT))
	assert.Equal(t, "?", b.bidSuitName(99))
}

func TestBridgeBidSuitToCardDesign(t *testing.T) {
	b := newTestBridge()
	assert.Equal(t, CardDesignClover, b.bidSuitToCardDesign(BridgeBidSuitClub))
	assert.Equal(t, CardDesignDiamond, b.bidSuitToCardDesign(BridgeBidSuitDiamond))
	assert.Equal(t, CardDesignHeart, b.bidSuitToCardDesign(BridgeBidSuitHeart))
	assert.Equal(t, CardDesignSpade, b.bidSuitToCardDesign(BridgeBidSuitSpade))
	assert.Equal(t, -1, b.bidSuitToCardDesign(BridgeBidSuitNT))
	assert.Equal(t, -1, b.bidSuitToCardDesign(99))
}

func TestBridgeCalcHCP(t *testing.T) {
	b := newTestBridge()
	b.Reset()

	// Clear player's hand and add specific cards
	b.players[0].Reset()
	b.players[0].AddCard(NewCard(CardDesignSpade, 1, true))  // Ace = 4
	b.players[0].AddCard(NewCard(CardDesignSpade, 13, true)) // King = 3
	b.players[0].AddCard(NewCard(CardDesignSpade, 12, true)) // Queen = 2
	b.players[0].AddCard(NewCard(CardDesignSpade, 11, true)) // Jack = 1
	b.players[0].AddCard(NewCard(CardDesignSpade, 10, true)) // 10 = 0

	assert.Equal(t, 10, b.calcHCP(0))
}

func TestBridgeIsBalancedHand(t *testing.T) {
	b := newTestBridge()
	player := NewBridgePlayer(false, 0)

	// Give balanced hand: 4-3-3-3
	for v := 1; v <= 4; v++ {
		player.AddCard(NewCard(CardDesignSpade, v, true))
	}
	for v := 1; v <= 3; v++ {
		player.AddCard(NewCard(CardDesignClover, v, true))
	}
	for v := 1; v <= 3; v++ {
		player.AddCard(NewCard(CardDesignHeart, v, true))
	}
	for v := 1; v <= 3; v++ {
		player.AddCard(NewCard(CardDesignDiamond, v, true))
	}

	assert.True(t, b.isBalancedHand(player))

	// Unbalanced: 7-0-3-3
	player2 := NewBridgePlayer(false, 0)
	for v := 1; v <= 7; v++ {
		player2.AddCard(NewCard(CardDesignSpade, v, true))
	}
	for v := 1; v <= 3; v++ {
		player2.AddCard(NewCard(CardDesignHeart, v, true))
	}
	for v := 1; v <= 3; v++ {
		player2.AddCard(NewCard(CardDesignDiamond, v, true))
	}

	assert.False(t, b.isBalancedHand(player2))
}

func TestBridgeCpuBidEasy(t *testing.T) {
	b := newTestBridgeWithReset()
	b.config.CpuDifficulty = BridgeCpuDifficultyEasy

	// Run multiple times to cover random branches
	for range 100 {
		bt, _, _ := b.cpuBidEasy(1)
		assert.True(t, bt == BridgeBidPass || bt == BridgeBidNormal)
	}
}

func TestBridgeCpuBidNormal(t *testing.T) {
	b := newTestBridgeWithReset()

	// Give player 1 a strong hand
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))  // A
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true)) // K
	b.players[1].AddCard(NewCard(CardDesignSpade, 12, true)) // Q
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true)) // K
	b.players[1].AddCard(NewCard(CardDesignClover, 1, true)) // A
	// HCP = 4+3+2+4+3+4 = 20

	bt, level, _ := b.cpuBidNormal(1)
	assert.Equal(t, BridgeBidNormal, bt)
	assert.GreaterOrEqual(t, level, 1)
}

func TestBridgeCpuBidNormalWeak(t *testing.T) {
	b := newTestBridgeWithReset()

	// Give player 1 a weak hand
	b.players[1].Reset()
	for v := 2; v <= 8; v++ {
		b.players[1].AddCard(NewCard(CardDesignSpade, v, true))
	}
	// HCP = 0

	bt, _, _ := b.cpuBidNormal(1)
	assert.Equal(t, BridgeBidPass, bt)
}

func TestBridgeOvertrickBonus(t *testing.T) {
	tests := []struct {
		name       string
		vul        bool
		doubled    int
		suit       int
		overtricks int
		want       int
	}{
		{"NV undoubled minor", false, 0, BridgeBidSuitClub, 2, 40},
		{"NV undoubled major", false, 0, BridgeBidSuitHeart, 2, 60},
		{"NV doubled", false, 1, BridgeBidSuitHeart, 2, 200},
		{"V doubled", true, 1, BridgeBidSuitHeart, 2, 400},
		{"NV redoubled", false, 2, BridgeBidSuitHeart, 1, 200},
		{"V redoubled", true, 2, BridgeBidSuitHeart, 1, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBridge()
			b.contractSuit = tt.suit
			b.doubled = tt.doubled
			b.vulnerability[0] = tt.vul
			got := b.calcOvertrickBonus(0, tt.overtricks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBridgeOpeningLeadRevealssDummy(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.openingLeadDone = false
	b.SetCurrentPlayerIdx(1) // opening leader

	b.CpuPlay()

	assert.True(t, b.IsOpeningLeadDone())
}

func TestBridgePlayHintReasons(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	// Lead - trump
	reason := b.playHintReason(0, 0) // spade card
	assert.Equal(t, "lead_trump", reason)

	// Lead - non-trump
	b.trumpSuit = -1
	reason = b.playHintReason(0, 0)
	assert.Equal(t, "lead_strong", reason)

	// Follow suit
	b.trumpSuit = CardDesignSpade
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 5, true)},
	}
	reason = b.playHintReason(0, 0)
	assert.Equal(t, "follow_suit", reason)

	// Trump cut
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 5, true)},
	}
	reason = b.playHintReason(0, 0) // spade card when lead is club
	assert.Equal(t, "trump_cut", reason)

	// Discard weak
	b.trumpSuit = CardDesignHeart // player 0 has only spades
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 5, true)},
	}
	reason = b.playHintReason(0, 0)
	assert.Equal(t, "discard_weak", reason)
}

func TestBridgeCardDesignToBidSuit(t *testing.T) {
	b := newTestBridge()
	assert.Equal(t, BridgeBidSuitClub, b.cardDesignToBidSuit(CardDesignClover))
	assert.Equal(t, BridgeBidSuitDiamond, b.cardDesignToBidSuit(CardDesignDiamond))
	assert.Equal(t, BridgeBidSuitHeart, b.cardDesignToBidSuit(CardDesignHeart))
	assert.Equal(t, BridgeBidSuitSpade, b.cardDesignToBidSuit(CardDesignSpade))
	assert.Equal(t, BridgeBidSuitClub, b.cardDesignToBidSuit(99)) // default
}

func TestBridgePlayerName(t *testing.T) {
	b := newTestBridge()
	assert.Contains(t, b.playerName(0), "You")
	assert.Contains(t, b.playerName(1), "CPU")
	assert.Contains(t, b.playerName(-1), "Player")
	assert.Contains(t, b.playerName(10), "Player")
}

func TestBridgeResolveTrickWrongPhase(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	b.SetPhase(BridgePhasePlay)
	b.ResolveTrick()
	// No crash, no effect
}

func TestBridgeResolveTrickWrongCardCount(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	b.SetPhase(BridgePhaseTrickEnd)
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 5, true)},
	}
	b.ResolveTrick()
	// Should not resolve with only 1 card
}

func TestBridgeTrickEndToRoundEnd(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.trickNumber = BridgeTotalTricks
	b.SetPhase(BridgePhaseTrickEnd)
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, true)},
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 1, true)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 1, true)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 1, true)},
	}

	b.ResolveTrick()
	assert.Equal(t, BridgePhaseRoundEnd, b.GetPhase())
}

func TestBridgeCpuPlayEasy(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.config.CpuDifficulty = BridgeCpuDifficultyEasy
	b.SetCurrentPlayerIdx(1)

	b.CpuPlay()
	assert.Equal(t, 1, len(b.currentTrick))
}

func TestBridgeCpuPlayHard(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.config.CpuDifficulty = BridgeCpuDifficultyHard
	b.SetCurrentPlayerIdx(1)

	b.CpuPlay()
	assert.Equal(t, 1, len(b.currentTrick))
}

func TestBridgeCpuPlayHardPartnerWinning(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.config.CpuDifficulty = BridgeCpuDifficultyHard

	// Player 1 (team 1) leads high card, player 3 (team 1) follows
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 1, true)}, // Ace
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 2, true)},
	}
	b.SetCurrentPlayerIdx(3) // Team 1

	b.CpuPlay()
	// Should play weak card since partner is winning
	assert.Equal(t, 3, len(b.currentTrick))
}

func TestBridgeSlamBonus(t *testing.T) {
	b := newTestBridge()
	b.contractLevel = 6
	b.contractSuit = BridgeBidSuitNT

	// Non-vulnerable small slam
	b.vulnerability[0] = false
	score := b.calcMadeContractScore(0, 0)
	// Should include 500 slam bonus + contract points + game bonus
	assert.GreaterOrEqual(t, score, 500)

	// Vulnerable grand slam
	b.contractLevel = 7
	b.vulnerability[0] = true
	score = b.calcMadeContractScore(0, 0)
	assert.GreaterOrEqual(t, score, 1500)
}

func TestBridgeDoubledMadeBonus(t *testing.T) {
	b := newTestBridge()
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitSpade

	b.doubled = 1
	score := b.calcMadeContractScore(0, 0)
	// Should include 50 bonus for making doubled contract
	assert.GreaterOrEqual(t, score, 50)

	b.doubled = 2
	score = b.calcMadeContractScore(0, 0)
	// Should include 100 bonus for making redoubled contract
	assert.GreaterOrEqual(t, score, 100)
}

func TestBridgeRubberBonus(t *testing.T) {
	b := newTestBridge()
	b.Reset()

	// Team 0 wins 2-0
	b.gamesWon[0] = 1
	b.SetPhase(BridgePhaseRoundEnd)
	b.contractLevel = 3
	b.contractSuit = BridgeBidSuitNT
	b.declarerIdx = 0

	for range 10 {
		b.players[0].AddTrick(make([]*Card, 0))
	}
	for range 3 {
		b.players[2].AddTrick(make([]*Card, 0))
	}

	b.ScoreRound()
	// 700 rubber bonus for 2-0
	assert.True(t, b.GetGameEndFlag())
	assert.Contains(t, b.actionLog[len(b.actionLog)-1].Detail, "rubber")
}

func TestBridgeCpuBidNormalDouble(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 2
	b.contractSuit = BridgeBidSuitSpade
	b.lastBidTeam = 0
	b.lastBidderIdx = 0

	// Give player 1 (team 1) HCP >= 15
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))    // A = 4
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))    // A = 4
	b.players[1].AddCard(NewCard(CardDesignClover, 1, true))   // A = 4
	b.players[1].AddCard(NewCard(CardDesignDiamond, 13, true)) // K = 3

	bt, _, _ := b.cpuBidNormal(1)
	assert.Equal(t, BridgeBidDouble, bt)
}

func TestBridgeCpuBidNormalRedouble(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitClub
	b.lastBidTeam = 0
	b.lastBidderIdx = 0
	b.doubled = 1

	// Give player 0 (team 0, bid team) HCP >= 10
	b.players[0].Reset()
	b.players[0].AddCard(NewCard(CardDesignSpade, 1, true))   // A = 4
	b.players[0].AddCard(NewCard(CardDesignHeart, 1, true))   // A = 4
	b.players[0].AddCard(NewCard(CardDesignClover, 12, true)) // Q = 2

	bt, _, _ := b.cpuBidNormal(0)
	assert.Equal(t, BridgeBidRedouble, bt)
}

func TestBridgeFindLongestSuit(t *testing.T) {
	b := newTestBridge()
	player := NewBridgePlayer(false, 0)
	for v := 1; v <= 7; v++ {
		player.AddCard(NewCard(CardDesignSpade, v, true))
	}
	for v := 1; v <= 3; v++ {
		player.AddCard(NewCard(CardDesignHeart, v, true))
	}

	suit, length := b.findLongestSuit(player)
	assert.Equal(t, CardDesignSpade, suit)
	assert.Equal(t, 7, length)
}

func TestBridgePlayCompleteTrick(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)
	b.SetCurrentPlayerIdx(1)

	// Play 4 cards
	for range 4 {
		b.CpuPlay()
		if b.GetPhase() == BridgePhaseTrickEnd {
			break
		}
		if b.IsHumanTurn() {
			_ = b.PlayerPlay(0) // Human or dummy
		}
	}

	assert.Equal(t, BridgePhaseTrickEnd, b.GetPhase())
}

func TestBridgeCpuBidHighContract(t *testing.T) {
	b := newTestBridgeWithReset()
	b.contractLevel = 6
	b.contractSuit = BridgeBidSuitNT
	b.lastBidTeam = 1 // Same team as player 1, so no double attempt

	// Even with high HCP, should not bid over 7NT
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 1, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 1, true))
	// HCP = 16

	bt, _, _ := b.cpuBidNormal(1)
	// Should pass since bidding 7-level with only 16 HCP is too aggressive
	assert.Equal(t, BridgeBidPass, bt)
}

func TestBridgeSmartFollowTrumpCut(t *testing.T) {
	b := newTestBridge()
	b.Reset()
	setupBridgePlayPhase(b)

	// Player 1 has clubs only, lead is hearts, trump is spades
	// Give player 1 mix of suits
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 2, true))  // Trump
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true)) // Non-trump

	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, true)},
	}

	idx := b.cpuPlayNormal(1, []int{0, 1})
	// Should play trump to cut
	card := b.players[1].GetCard(idx)
	assert.Equal(t, CardDesignSpade, card.GetDesign())
}

// --- Hard difficulty helper tests ---

func newTestBridgeHard() *Bridge {
	players := []*BridgePlayer{
		NewBridgePlayer(true, 0),  // North (human, team 0)
		NewBridgePlayer(false, 1), // East (CPU, team 1)
		NewBridgePlayer(false, 0), // South (CPU, team 0)
		NewBridgePlayer(false, 1), // West (CPU, team 1)
	}
	return NewBridge(NewTrumpCards(0), players, BridgeConfig{
		CpuDifficulty: BridgeCpuDifficultyHard,
	})
}

func TestBridge_calcDistributionPoints(t *testing.T) {
	tests := []struct {
		name     string
		cards    []struct{ design, value int }
		expected int
	}{
		{
			name: "void in one suit gives 3 points",
			cards: []struct{ design, value int }{
				// 5 spades, 4 hearts, 4 clubs, 0 diamonds = void
				{CardDesignSpade, 1}, {CardDesignSpade, 2}, {CardDesignSpade, 3}, {CardDesignSpade, 4}, {CardDesignSpade, 5},
				{CardDesignHeart, 1}, {CardDesignHeart, 2}, {CardDesignHeart, 3}, {CardDesignHeart, 4},
				{CardDesignClover, 1}, {CardDesignClover, 2}, {CardDesignClover, 3}, {CardDesignClover, 4},
			},
			expected: 3,
		},
		{
			name: "singleton gives 2 points",
			cards: []struct{ design, value int }{
				// 5 spades, 4 hearts, 3 clubs, 1 diamond = singleton
				{CardDesignSpade, 1}, {CardDesignSpade, 2}, {CardDesignSpade, 3}, {CardDesignSpade, 4}, {CardDesignSpade, 5},
				{CardDesignHeart, 1}, {CardDesignHeart, 2}, {CardDesignHeart, 3}, {CardDesignHeart, 4},
				{CardDesignClover, 1}, {CardDesignClover, 2}, {CardDesignClover, 3},
				{CardDesignDiamond, 1},
			},
			expected: 2,
		},
		{
			name: "doubleton gives 1 point",
			cards: []struct{ design, value int }{
				// 4 spades, 4 hearts, 3 clubs, 2 diamonds = doubleton
				{CardDesignSpade, 1}, {CardDesignSpade, 2}, {CardDesignSpade, 3}, {CardDesignSpade, 4},
				{CardDesignHeart, 1}, {CardDesignHeart, 2}, {CardDesignHeart, 3}, {CardDesignHeart, 4},
				{CardDesignClover, 1}, {CardDesignClover, 2}, {CardDesignClover, 3},
				{CardDesignDiamond, 1}, {CardDesignDiamond, 2},
			},
			expected: 1,
		},
		{
			name: "balanced hand gives 0 points",
			cards: []struct{ design, value int }{
				// 4-3-3-3
				{CardDesignSpade, 1}, {CardDesignSpade, 2}, {CardDesignSpade, 3}, {CardDesignSpade, 4},
				{CardDesignHeart, 1}, {CardDesignHeart, 2}, {CardDesignHeart, 3},
				{CardDesignClover, 1}, {CardDesignClover, 2}, {CardDesignClover, 3},
				{CardDesignDiamond, 1}, {CardDesignDiamond, 2}, {CardDesignDiamond, 3},
			},
			expected: 0,
		},
		{
			name: "multiple short suits accumulate",
			cards: []struct{ design, value int }{
				// 6 spades, 5 hearts, 1 club, 1 diamond = 2+2=4
				{CardDesignSpade, 1}, {CardDesignSpade, 2}, {CardDesignSpade, 3}, {CardDesignSpade, 4}, {CardDesignSpade, 5}, {CardDesignSpade, 6},
				{CardDesignHeart, 1}, {CardDesignHeart, 2}, {CardDesignHeart, 3}, {CardDesignHeart, 4}, {CardDesignHeart, 5},
				{CardDesignClover, 1},
				{CardDesignDiamond, 1},
			},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTestBridgeHard()
			b.players[1].Reset()
			for _, c := range tt.cards {
				b.players[1].AddCard(NewCard(c.design, c.value, true))
			}
			assert.Equal(t, tt.expected, b.calcDistributionPoints(1))
		})
	}
}

func TestBridge_countSuitCards(t *testing.T) {
	b := newTestBridgeHard()
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 2, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 3, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))

	assert.Equal(t, 3, b.countSuitCards(1, CardDesignSpade))
	assert.Equal(t, 1, b.countSuitCards(1, CardDesignHeart))
	assert.Equal(t, 0, b.countSuitCards(1, CardDesignDiamond))
}

func TestBridge_partnerBidSuit(t *testing.T) {
	b := newTestBridgeHard()

	// No bids yet
	assert.Equal(t, 0, b.partnerBidSuit(1))

	// Partner of player 1 (team 1) is player 3
	b.bidHistory = []*BridgeBidEntry{
		{PlayerIdx: 0, BidType: BridgeBidNormal, Level: 1, Suit: BridgeBidSuitHeart},
		{PlayerIdx: 3, BidType: BridgeBidNormal, Level: 1, Suit: BridgeBidSuitSpade},
	}
	assert.Equal(t, BridgeBidSuitSpade, b.partnerBidSuit(1))

	// Partner passed only
	b.bidHistory = []*BridgeBidEntry{
		{PlayerIdx: 3, BidType: BridgeBidPass, Level: 0, Suit: 0},
	}
	assert.Equal(t, 0, b.partnerBidSuit(1))
}

func TestBridge_countTrumpsRemaining(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhase(b)

	// Player 1 has all clubs by default, trump is spades
	assert.Equal(t, 0, b.countTrumpsRemaining(1))

	// Player 0 has all spades, trump is spades
	assert.Equal(t, 13, b.countTrumpsRemaining(0))

	// Give player 1 some spades
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 2, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	assert.Equal(t, 2, b.countTrumpsRemaining(1))

	// NoTrump game
	b.trumpSuit = -1
	assert.Equal(t, 0, b.countTrumpsRemaining(1))
}

// --- cpuBidHard tests ---

func TestBridgeCpuBidHard_WeakHand(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 0
	b.doubled = 0

	// Give player 1 a weak hand (HCP < 12, no distribution)
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 2, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 3, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 4, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 2, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 3, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 4, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 2, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))

	bt, _, _ := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidPass, bt)
}

func TestBridgeCpuBidHard_StrongHand(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 0
	b.doubled = 0

	// Give player 1 a strong hand (HCP=20 + distribution)
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignSpade, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignSpade, 11, true)) // J=1
	b.players[1].AddCard(NewCard(CardDesignSpade, 10, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignHeart, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignClover, 1, true)) // A=4 -> HCP=23 total, singleton diamond? no, need 13 cards
	b.players[1].AddCard(NewCard(CardDesignClover, 2, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 1, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	// HCP = 4+3+2+1+4+3+2+4 = 23, distPoints = 0 (3-2-5-3), totalPts = 23

	bt, level, _ := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidNormal, bt)
	assert.GreaterOrEqual(t, level, 2)
}

func TestBridgeCpuBidHard_BalancedNT(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 0
	b.doubled = 0

	// Give player 1 a balanced hand with HCP=16
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 11, true)) // J=1
	b.players[1].AddCard(NewCard(CardDesignClover, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))
	// HCP = 4+3+4+2+1+2 = 16, balanced (3-4-3-3), totalPts = 16

	bt, level, suit := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidNormal, bt)
	assert.Equal(t, 1, level)
	assert.Equal(t, BridgeBidSuitNT, suit)
}

func TestBridgeCpuBidHard_PartnerFit(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitHeart
	b.doubled = 0
	b.lastBidTeam = 1 // Partner's team bid
	b.lastBidderIdx = 3

	// Partner (player 3) bid hearts
	b.bidHistory = []*BridgeBidEntry{
		{PlayerIdx: 3, BidType: BridgeBidNormal, Level: 1, Suit: BridgeBidSuitHeart},
	}

	// Player 1 has 4 hearts and moderate HCP
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true)) // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 6, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 7, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))
	// HCP = 4+3+2 = 9, distPts = 0 (balanced), partnerFit = +3 -> totalPts = 12

	bt, _, suit := b.cpuBidHard(1)
	// Should raise partner's hearts
	assert.Equal(t, BridgeBidNormal, bt)
	assert.Equal(t, BridgeBidSuitHeart, suit)
}

func TestBridgeCpuBidHard_Double(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 2
	b.contractSuit = BridgeBidSuitHeart
	b.doubled = 0
	b.lastBidTeam = 0 // Opponent's team (player 1 is team 1)
	b.lastBidderIdx = 0

	// Player 1 has strong hand and 3+ cards in opponent's suit (hearts)
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 2, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))
	// HCP = 4+3+4+3 = 14, 3 hearts (can defend)

	bt, _, _ := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidDouble, bt)
}

func TestBridgeCpuBidHard_Redouble(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 1
	b.contractSuit = BridgeBidSuitSpade
	b.doubled = 1
	b.lastBidTeam = 1 // Own team bid
	b.lastBidderIdx = 3

	// Player 1 has totalPts >= 12
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 11, true)) // J=1
	b.players[1].AddCard(NewCard(CardDesignHeart, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 2, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))
	// HCP = 4+3+4+1 = 12, distPts = 0 -> totalPts = 12

	bt, _, _ := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidRedouble, bt)
}

func TestBridgeCpuBidHard_PreferMajor(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 0
	b.doubled = 0

	// Player 1 has equal-length hearts and clubs (4 each), moderate HCP
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignClover, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 2, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))
	// HCP = 4+3+4+2 = 13, distPts = 0, totalPts = 13
	// Hearts and clubs both 4 cards, should prefer hearts (major)

	bt, _, suit := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidNormal, bt)
	assert.Equal(t, BridgeBidSuitHeart, suit)
}

func TestBridgeCpuBidHard_SafetyCap(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 4
	b.contractSuit = BridgeBidSuitSpade
	b.doubled = 0
	b.lastBidTeam = 0
	b.lastBidderIdx = 0

	// Player 1 has moderate hand, shouldn't overbid at level 5+
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A=4
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true)) // K=3
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 4, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 3, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 11, true)) // J=1
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 4, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 12, true)) // Q=2
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))
	// HCP = 4+3+1+2 = 10, distPts = 0 -> totalPts = 10 -> should pass (< 12)

	bt, _, _ := b.cpuBidHard(1)
	assert.Equal(t, BridgeBidPass, bt)
}

// --- cpuPlayHard enhanced tests ---

func setupBridgePlayPhaseHard(b *Bridge) {
	for _, p := range b.players {
		p.ResetRound()
	}
	b.phase = BridgePhasePlay
	b.contractLevel = 2
	b.contractSuit = BridgeBidSuitSpade
	b.trumpSuit = CardDesignSpade
	b.declarerIdx = 0
	b.dummyIdx = 2
	b.leadPlayerIdx = 1
	b.currentPlayerIdx = 1
	b.trickNumber = 1
	b.openingLeadDone = true
	b.currentTrick = nil
}

func TestBridgeCpuPlayHard_LeadTrumpDraw(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)
	b.declarerIdx = 1 // Player 1 is declarer (declaring team)
	b.dummyIdx = 3
	b.trickNumber = 2

	// Player 1 has 4 trumps + non-trump cards
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 5, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 8, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 10, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 3, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))

	idx := b.cpuPlayHard(1, []int{0, 1, 2, 3, 4, 5})
	card := b.players[1].GetCard(idx)
	assert.Equal(t, CardDesignSpade, card.GetDesign())
}

func TestBridgeCpuPlayHard_LeadPartnerSuit(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)
	b.declarerIdx = 0 // Player 0 is declarer, so player 1 is defending
	b.dummyIdx = 2

	// Partner (player 3) bid hearts
	b.bidHistory = []*BridgeBidEntry{
		{PlayerIdx: 3, BidType: BridgeBidNormal, Level: 1, Suit: BridgeBidSuitHeart},
	}

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 8, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 4, true))

	idx := b.cpuPlayHard(1, []int{0, 1, 2, 3})
	card := b.players[1].GetCard(idx)
	assert.Equal(t, CardDesignHeart, card.GetDesign())
}

func TestBridgeCpuPlayHard_LeadLateGame(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)
	b.trickNumber = 11

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 2, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true)) // Ace

	idx := b.cpuPlayHard(1, []int{0, 1})
	card := b.players[1].GetCard(idx)
	// Late game: should lead strongest
	assert.Equal(t, 1, card.GetValue()) // Ace
}

func TestBridgeCpuPlayHard_PartnerWinning(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)

	// Player 3 (team 1, partner of player 1) is winning
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 5, true)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 1, true)}, // Partner winning with Ace
	}
	b.currentPlayerIdx = 1

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true)) // K
	b.players[1].AddCard(NewCard(CardDesignHeart, 2, true))  // 2

	idx := b.cpuPlayHard(1, []int{0, 1})
	card := b.players[1].GetCard(idx)
	// Partner winning, should play weakest
	assert.Equal(t, 2, card.GetValue())
}

func TestBridgeCpuPlayHard_4thSeatMinWin(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)

	// Player 1 (team 1) is 4th to play. Opponent (team 0) is winning.
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 5, true)},  // team 1 (partner, low)
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, true)}, // team 0 (opponent, winning)
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 8, true)},  // team 0
	}

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // Ace
	b.players[1].AddCard(NewCard(CardDesignHeart, 11, true)) // Jack (just above 10)
	b.players[1].AddCard(NewCard(CardDesignHeart, 3, true))  // Can't win

	idx := b.cpuPlayHard(1, []int{0, 1, 2})
	card := b.players[1].GetCard(idx)
	// Should play Jack (minimum winning card), not Ace
	assert.Equal(t, 11, card.GetValue())
}

func TestBridgeCpuPlayHard_TrumpCutLowest(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)

	// Lead is hearts, player 1 has no hearts but has multiple trumps (spades)
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, true)},
	}

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 2, true))  // Low trump
	b.players[1].AddCard(NewCard(CardDesignSpade, 10, true)) // High trump
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true)) // Non-trump

	idx := b.cpuPlayHard(1, []int{0, 1, 2})
	card := b.players[1].GetCard(idx)
	// Should trump with lowest trump (spade 2)
	assert.Equal(t, CardDesignSpade, card.GetDesign())
	assert.Equal(t, 2, card.GetValue())
}

func TestBridgeCpuPlayHard_DiscardShortSuit(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)
	b.trumpSuit = -1 // NoTrump

	// Lead is hearts, player 1 has no hearts and no trumps
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, true)},
	}

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignClover, 3, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true)) // Shortest suit (1 card)

	idx := b.cpuPlayHard(1, []int{0, 1, 2})
	card := b.players[1].GetCard(idx)
	// Should discard from shortest suit (diamond)
	assert.Equal(t, CardDesignDiamond, card.GetDesign())
}

func TestBridgeCpuPlayHard_OpponentWinningSmartFollow(t *testing.T) {
	b := newTestBridgeHard()
	setupBridgePlayPhaseHard(b)

	// Lead is hearts, opponent winning, player 1 can follow
	b.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 10, true)},
	}

	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignHeart, 11, true)) // J (wins)
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))  // A (wins)
	b.players[1].AddCard(NewCard(CardDesignHeart, 3, true))  // Loses

	idx := b.cpuPlayHard(1, []int{0, 1, 2})
	card := b.players[1].GetCard(idx)
	// Should play lowest winning card (Jack)
	assert.Equal(t, 11, card.GetValue())
}

func TestBridgeCpuSelectBid_HardDispatch(t *testing.T) {
	b := newTestBridgeHard()
	b.Reset()
	b.phase = BridgePhaseBid
	b.contractLevel = 0
	b.doubled = 0

	// Strong hand for a clear bid
	b.players[1].Reset()
	b.players[1].AddCard(NewCard(CardDesignSpade, 1, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 13, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 12, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 11, true))
	b.players[1].AddCard(NewCard(CardDesignSpade, 10, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 1, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 13, true))
	b.players[1].AddCard(NewCard(CardDesignHeart, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 1, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 5, true))
	b.players[1].AddCard(NewCard(CardDesignClover, 4, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 2, true))
	b.players[1].AddCard(NewCard(CardDesignDiamond, 3, true))

	bt, _, _ := b.cpuSelectBid(1)
	assert.NotEqual(t, BridgeBidPass, bt)
}

// **どこから上回れるか・ダブルできるかを、拒否条件の裏返しで出す** (#4903)。
func TestBridge_MinLegalBidAndDoubleRules(t *testing.T) {
	g := newTestBridgeWithReset()

	// まだ誰もビッドしていなければ 1♣ から。
	g.SetContractLevel(0)
	lv, st, ok := g.BridgeMinLegalBid()
	assert.True(t, ok)
	assert.Equal(t, 1, lv)
	assert.Equal(t, BridgeBidSuitClub, st)

	// **同レベルの上位スートが先。**レベルを上げるのは NT まで埋まってから。
	g.SetContractLevel(3)
	g.SetContractSuit(BridgeBidSuitClub)
	lv, st, ok = g.BridgeMinLegalBid()
	assert.True(t, ok)
	assert.Equal(t, 3, lv)
	assert.Equal(t, BridgeBidSuitClub+1, st)

	// 3NT の次は 4♣。
	g.SetContractSuit(BridgeBidSuitNT)
	lv, st, ok = g.BridgeMinLegalBid()
	assert.True(t, ok)
	assert.Equal(t, 4, lv)
	assert.Equal(t, BridgeBidSuitClub, st)

	// 7NT まで埋まれば上は無い。
	g.SetContractLevel(7)
	g.SetContractSuit(BridgeBidSuitNT)
	_, _, ok = g.BridgeMinLegalBid()
	assert.False(t, ok)

	// --- ダブル / リダブル ---
	g.SetContractLevel(3)
	g.SetContractSuit(BridgeBidSuitClub)
	g.SetDoubled(0)
	g.lastBidTeam = g.GetPlayer(0).GetTeam()

	// 相手チームだけがダブルできる。
	assert.False(t, g.BridgeCanDouble(0), "own team's bid cannot be doubled")
	assert.True(t, g.BridgeCanDouble(1))
	// ビッドが無ければ不可。
	g.SetContractLevel(0)
	assert.False(t, g.BridgeCanDouble(1))
	g.SetContractLevel(3)

	// 既にダブル済みなら不可。代わりに、ダブルされた側だけがリダブルできる。
	g.SetDoubled(1)
	assert.False(t, g.BridgeCanDouble(1))
	assert.True(t, g.BridgeCanRedouble(0))
	assert.False(t, g.BridgeCanRedouble(1), "only the doubled team may redouble")

	// リダブル済みならどちらも不可。
	g.SetDoubled(2)
	assert.False(t, g.BridgeCanDouble(1))
	assert.False(t, g.BridgeCanRedouble(0))

	// 範囲外の席。
	g.SetDoubled(0)
	assert.False(t, g.BridgeCanDouble(99))
	assert.False(t, g.BridgeCanRedouble(-1))
}
