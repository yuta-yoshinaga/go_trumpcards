package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dramahaFullyPopulated builds a table with every persisted field set to a
// distinct, non-zero value, so a round trip that silently drops one is visible.
func dramahaFullyPopulated() *Dramaha {
	o := newTestDramaha()
	for i, p := range o.players {
		p.SetChips(1000 + i)
		p.SetCurrentBet(10 + i)
		dealDramahaHole(p, diamondHole()...)
		p.EvalDrawHand()
		p.SetBestHand([]*Card{NewCard(CardDesignSpade, 13, false)})
		for j := 0; j <= i; j++ {
			p.IncrementTotalHands()
			p.IncrementVPIP()
			p.IncrementPFR()
			p.IncrementThreeBet()
			p.IncrementThreeBetOpportunity()
			p.IncrementPostFlopBetRaise()
			p.IncrementPostFlopCall()
		}
	}
	o.players[3].SetFolded(true)
	o.players[2].SetAllIn(true)

	o.communityCards = splitBoard()
	o.pot = 275
	o.sidePots = []SidePot{{Amount: 275, EligiblePlayers: []int{0, 1, 2}}}
	o.dealerIdx = 2
	o.currentTurn = 1
	o.phase = DramahaPhaseRiver
	o.gameEndFlag = true
	o.lastBet = 60
	o.minRaise = 20
	o.raiseCount = 3
	o.actedFlags = []bool{true, false, true, false}
	o.roundResults = []HoldemResult{{
		PlayerIdx:    1,
		HandRank:     PokerHandThreeOfAKind,
		HandName:     "Three of a Kind",
		WonAmount:    138,
		HiWonAmount:  138,
		LowWonAmount: 0,
		LowQualifies: true,
		LowBestHand:  diamondHole(),
	}}
	o.cpuActions = []HoldemCpuAction{{PlayerIdx: 2, Action: DramahaActionRaise, Amount: 40}}
	o.startingChips = []int{900, 901, 902, 903}
	o.vpipTracked = []bool{true, false, true, false}
	o.pfrTracked = []bool{false, true, false, true}
	o.threeBetTracked = []bool{true, true, false, false}
	o.handCount = 12
	o.rebuyCounts = []int{1, 2, 0, 3}
	o.addonUsed = []bool{true, false, true, false}
	o.rebuyPhaseType = DramahaRebuyPhaseAddon
	o.lastHumanPlayMs = 1234
	o.humanProfile = &BettingHumanProfile{GamesPlayed: 7}
	o.appendLog(0, "bet", "bet 60", nil)

	cfg := o.config
	cfg.SmallBlind = 25
	cfg.BigBlind = 50
	cfg.TournamentMode = true
	cfg.CpuMetaAI = true
	o.config = cfg
	return o
}

// TestDramahaJSONRoundTrip walks every persisted field across MarshalJSON /
// UnmarshalJSON.
//
// NOTE: the draw round's own state (drawnFlags) is deliberately absent from
// these assertions because the wire format does not carry it -- a table saved
// during the draw round comes back with an empty slice and Draw() panics on
// it. That is a production bug, reported rather than papered over here; when it
// is fixed, add drawnFlags to dramahaFullyPopulated and to the assertions
// below.
func TestDramahaJSONRoundTrip(t *testing.T) {
	src := dramahaFullyPopulated()

	data, err := src.MarshalJSON()
	require.NoError(t, err)
	require.NotEqual(t, "{}", string(data),
		"a struct of unexported fields serialises to {} without a MarshalJSON")
	require.Greater(t, len(data), 500, "the payload must carry real state, not a stub")

	var got Dramaha
	require.NoError(t, got.UnmarshalJSON(data))

	assert.Equal(t, src.pot, got.pot)
	assert.Equal(t, src.dealerIdx, got.dealerIdx)
	assert.Equal(t, src.currentTurn, got.currentTurn)
	assert.Equal(t, src.phase, got.phase)
	assert.Equal(t, src.gameEndFlag, got.gameEndFlag)
	assert.Equal(t, src.lastBet, got.lastBet)
	assert.Equal(t, src.minRaise, got.minRaise)
	assert.Equal(t, src.raiseCount, got.raiseCount)
	assert.Equal(t, src.actedFlags, got.actedFlags)
	assert.Equal(t, src.sidePots, got.sidePots)
	assert.Equal(t, src.cpuActions, got.cpuActions)
	assert.Equal(t, src.startingChips, got.startingChips)
	assert.Equal(t, src.vpipTracked, got.vpipTracked)
	assert.Equal(t, src.pfrTracked, got.pfrTracked)
	assert.Equal(t, src.threeBetTracked, got.threeBetTracked)
	assert.Equal(t, src.handCount, got.handCount)
	assert.Equal(t, src.rebuyCounts, got.rebuyCounts)
	assert.Equal(t, src.addonUsed, got.addonUsed)
	assert.Equal(t, src.rebuyPhaseType, got.rebuyPhaseType)
	assert.Equal(t, src.lastHumanPlayMs, got.lastHumanPlayMs)
	assert.Equal(t, src.preflopCommunity, got.preflopCommunity)
	assert.Equal(t, src.config, got.config)
	assert.Equal(t, DramahaHoleCards, got.GetHoleCardCount())

	require.NotNil(t, got.humanProfile)
	assert.Equal(t, src.humanProfile.GamesPlayed, got.humanProfile.GamesPlayed)

	require.Len(t, got.actionLog, len(src.actionLog))
	assert.Equal(t, src.actionLog[0].ActionType, got.actionLog[0].ActionType)
	assert.Equal(t, src.actionLog[0].Detail, got.actionLog[0].Detail)

	// Board and deck
	require.Len(t, got.communityCards, len(src.communityCards))
	for i := range src.communityCards {
		assert.Equal(t, dramahaCardID(src.communityCards[i]), dramahaCardID(got.communityCards[i]), "board card %d", i)
	}
	require.NotNil(t, got.trumpCards)
	assert.Equal(t, src.trumpCards.GetRemainingCount(), got.trumpCards.GetRemainingCount())

	// Round results, including the draw side of the split.
	require.Len(t, got.roundResults, 1)
	assert.Equal(t, src.roundResults[0].PlayerIdx, got.roundResults[0].PlayerIdx)
	assert.Equal(t, src.roundResults[0].HandRank, got.roundResults[0].HandRank)
	assert.Equal(t, src.roundResults[0].WonAmount, got.roundResults[0].WonAmount)
	assert.Equal(t, src.roundResults[0].HiWonAmount, got.roundResults[0].HiWonAmount)
	assert.Equal(t, src.roundResults[0].LowWonAmount, got.roundResults[0].LowWonAmount)
	assert.True(t, got.roundResults[0].LowQualifies)
	require.Len(t, got.roundResults[0].LowBestHand, DramahaHoleCards)

	// Per-seat state, including the draw hand each seat carries into showdown.
	require.Len(t, got.players, len(src.players))
	for i := range src.players {
		s, g := src.players[i], got.players[i]
		assert.Equal(t, s.GetChips(), g.GetChips(), "seat %d chips", i)
		assert.Equal(t, s.GetCurrentBet(), g.GetCurrentBet(), "seat %d bet", i)
		assert.Equal(t, s.GetIsHuman(), g.GetIsHuman(), "seat %d human flag", i)
		assert.Equal(t, s.GetFolded(), g.GetFolded(), "seat %d folded", i)
		assert.Equal(t, s.GetAllIn(), g.GetAllIn(), "seat %d all-in", i)
		assert.Equal(t, s.GetPlayStyle(), g.GetPlayStyle(), "seat %d style", i)
		assert.Equal(t, s.GetTotalHands(), g.GetTotalHands(), "seat %d hands", i)
		assert.Equal(t, s.GetVPIPCount(), g.GetVPIPCount(), "seat %d vpip", i)
		assert.Equal(t, s.GetPFRCount(), g.GetPFRCount(), "seat %d pfr", i)
		assert.Equal(t, s.GetThreeBetCount(), g.GetThreeBetCount(), "seat %d 3bet", i)
		assert.Equal(t, s.GetThreeBetOpportunity(), g.GetThreeBetOpportunity(), "seat %d 3bet opps", i)
		assert.Equal(t, s.GetPostFlopBetRaise(), g.GetPostFlopBetRaise(), "seat %d bet/raise", i)
		assert.Equal(t, s.GetPostFlopCall(), g.GetPostFlopCall(), "seat %d calls", i)
		assert.Equal(t, dramahaHandIDs(s), dramahaHandIDs(g), "seat %d hole cards", i)
		assert.Equal(t, s.GetDrawRank(), g.GetDrawRank(), "seat %d draw rank", i)
		require.Len(t, g.GetDrawBestHand(), DramahaHoleCards, "seat %d draw hand", i)
		for j := range s.GetDrawBestHand() {
			assert.Equal(t, dramahaCardID(s.GetDrawBestHand()[j]), dramahaCardID(g.GetDrawBestHand()[j]),
				"seat %d draw card %d", i, j)
		}
		require.Len(t, g.GetBestHand(), 1, "seat %d best hand", i)
	}
}

// TestDramahaJSONRoundTrip_RestoredTableKeepsPlaying is the check a field-by-
// field comparison cannot make: the restored table has to be usable, not just
// equal. A saved showdown must resolve to the same split on the other side.
func TestDramahaJSONRoundTrip_RestoredTableKeepsPlaying(t *testing.T) {
	src := newDramahaAtShowdown(200)
	dealDramahaHole(src.players[0], heartFlushHole()...)
	dealDramahaHole(src.players[1], twoKingsHole()...)

	data, err := src.MarshalJSON()
	require.NoError(t, err)
	var got Dramaha
	require.NoError(t, got.UnmarshalJSON(data))

	before := dramahaTotalChips(&got)
	got.resolveShowdown()

	assert.Equal(t, before, dramahaTotalChips(&got), "chips are conserved on the restored table too")
	assert.Equal(t, 1100, got.players[0].GetChips(), "the draw winner still takes half after a restore")
	assert.Equal(t, 1100, got.players[1].GetChips(), "the Omaha winner still takes half after a restore")
}

func TestDramahaPlayerJSONRoundTrip_IsNotEmpty(t *testing.T) {
	p := NewDramahaPlayer(true, HoldemStyleLAG)
	dealDramahaHole(p, heartFlushHole()...)
	p.SetChips(777)
	require.Equal(t, PokerHandFlush, p.EvalDrawHand())

	data, err := p.MarshalJSON()
	require.NoError(t, err)
	require.NotEqual(t, "{}", string(data),
		"DramahaPlayer is all unexported fields; without MarshalJSON it would ship as two bytes")

	var got DramahaPlayer
	require.NoError(t, got.UnmarshalJSON(data))
	assert.Equal(t, 777, got.GetChips())
	assert.True(t, got.GetIsHuman())
	assert.Equal(t, HoldemPlayStyle(HoldemStyleLAG), got.GetPlayStyle())
	assert.Equal(t, PokerHandFlush, got.GetDrawRank())
	assert.Equal(t, dramahaHandIDs(p), dramahaHandIDs(&got))
	assert.Equal(t, PokerHandFlush, got.EvalDrawHand(), "the restored hole cards still rank as a flush")
}
