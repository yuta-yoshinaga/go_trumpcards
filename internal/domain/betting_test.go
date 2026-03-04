package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockBettingPlayer テスト用BettingPlayer実装
type mockBettingPlayer struct {
	chips      int
	currentBet int
	folded     bool
	allIn      bool
	handRank   int
	cards      []*Card
}

func (m *mockBettingPlayer) GetChips() int { return m.chips }
func (m *mockBettingPlayer) SubtractChips(a int) bool {
	if m.chips < a {
		return false
	}
	m.chips -= a
	return true
}
func (m *mockBettingPlayer) AddChips(a int)              { m.chips += a }
func (m *mockBettingPlayer) GetCurrentBet() int          { return m.currentBet }
func (m *mockBettingPlayer) SetCurrentBet(b int)         { m.currentBet = b }
func (m *mockBettingPlayer) GetFolded() bool             { return m.folded }
func (m *mockBettingPlayer) SetFolded(f bool)            { m.folded = f }
func (m *mockBettingPlayer) GetAllIn() bool              { return m.allIn }
func (m *mockBettingPlayer) SetAllIn(a bool)             { m.allIn = a }
func (m *mockBettingPlayer) GetHandRank() int            { return m.handRank }
func (m *mockBettingPlayer) GetComparisonCards() []*Card { return m.cards }

// newMockPlayer テスト用プレイヤー生成ヘルパー
func newMockPlayer(chips int) *mockBettingPlayer {
	return &mockBettingPlayer{chips: chips}
}

// newBettingState テスト用BettingState生成ヘルパー
func newBettingState(n int) *BettingState {
	return &BettingState{
		ActedFlags: make([]bool, n),
		MinRaise:   10,
	}
}

// --- ExecuteBettingAction tests ---

func TestExecuteBettingAction_Fold(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionFold, 0, 10)
	assert.NoError(t, err)
	assert.True(t, players[0].GetFolded())
	assert.True(t, state.ActedFlags[0])
}

func TestExecuteBettingAction_Check_Success(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionCheck, 0, 10)
	assert.NoError(t, err)
	assert.True(t, state.ActedFlags[0])
}

func TestExecuteBettingAction_Check_OutstandingBet(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20

	err := ExecuteBettingAction(players, state, 0, bettingActionCheck, 0, 10)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Contains(t, err.Error(), "outstanding bet")
}

func TestExecuteBettingAction_Call_Success(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20

	err := ExecuteBettingAction(players, state, 0, bettingActionCall, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 80, players[0].GetChips())
	assert.Equal(t, 20, players[0].GetCurrentBet())
	assert.Equal(t, 20, state.Pot)
	assert.True(t, state.ActedFlags[0])
}

func TestExecuteBettingAction_Call_NothingToCall(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionCall, 0, 10)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Contains(t, err.Error(), "Nothing to call")
}

func TestExecuteBettingAction_Call_ShortAllIn(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(10), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 50

	err := ExecuteBettingAction(players, state, 0, bettingActionCall, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, players[0].GetChips())
	assert.Equal(t, 10, players[0].GetCurrentBet())
	assert.Equal(t, 10, state.Pot)
	assert.True(t, players[0].GetAllIn())
	assert.True(t, state.ActedFlags[0])
}

func TestExecuteBettingAction_Bet_Success(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionBet, 20, 10)
	assert.NoError(t, err)
	assert.Equal(t, 80, players[0].GetChips())
	assert.Equal(t, 20, players[0].GetCurrentBet())
	assert.Equal(t, 20, state.Pot)
	assert.Equal(t, 20, state.LastBet)
	assert.Equal(t, 20, state.MinRaise)
	assert.Equal(t, 1, state.RaiseCount)
	// bet player is acted, other is reset
	assert.True(t, state.ActedFlags[0])
	assert.False(t, state.ActedFlags[1])
}

func TestExecuteBettingAction_Bet_AllInOnExact(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(20), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionBet, 20, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, players[0].GetChips())
	assert.True(t, players[0].GetAllIn())
}

func TestExecuteBettingAction_Bet_MaxRaisesReached(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.RaiseCount = bettingMaxRaisesPerRound

	err := ExecuteBettingAction(players, state, 0, bettingActionBet, 20, 10)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Contains(t, err.Error(), "Maximum number of raises")
}

func TestExecuteBettingAction_Bet_OutstandingBet(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 10

	err := ExecuteBettingAction(players, state, 0, bettingActionBet, 20, 10)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Contains(t, err.Error(), "outstanding bet")
}

func TestExecuteBettingAction_Bet_TooSmall(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionBet, 5, 10)
	assert.ErrorIs(t, err, ErrInvalidAmount)
	assert.Contains(t, err.Error(), "minimum bet")
}

func TestExecuteBettingAction_Bet_InsufficientChips(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(5), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionBet, 10, 10)
	assert.ErrorIs(t, err, ErrInsufficientChips)
}

func TestExecuteBettingAction_Raise_Success(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 10

	err := ExecuteBettingAction(players, state, 0, bettingActionRaise, 20, 10)
	assert.NoError(t, err)
	// diff=20, totalNeeded=40
	assert.Equal(t, 60, players[0].GetChips())
	assert.Equal(t, 40, players[0].GetCurrentBet())
	assert.Equal(t, 40, state.Pot)
	assert.Equal(t, 40, state.LastBet)
	assert.Equal(t, 20, state.MinRaise)
	assert.Equal(t, 1, state.RaiseCount)
}

func TestExecuteBettingAction_Raise_MaxRaisesReached(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 10
	state.RaiseCount = bettingMaxRaisesPerRound

	err := ExecuteBettingAction(players, state, 0, bettingActionRaise, 20, 10)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestExecuteBettingAction_Raise_TooSmall(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 20

	err := ExecuteBettingAction(players, state, 0, bettingActionRaise, 10, 10)
	assert.ErrorIs(t, err, ErrInvalidAmount)
	assert.Contains(t, err.Error(), "minimum raise")
}

func TestExecuteBettingAction_Raise_RedirectToAllIn(t *testing.T) {
	// totalNeeded (20+20=40) >= chips (30) → AllIn redirect
	players := []BettingPlayer{newMockPlayer(30), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 20

	err := ExecuteBettingAction(players, state, 0, bettingActionRaise, 20, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, players[0].GetChips())
	assert.True(t, players[0].GetAllIn())
	assert.Equal(t, 30, players[0].GetCurrentBet())
	assert.Equal(t, 30, state.Pot)
}

func TestExecuteBettingAction_Raise_NegativeDiff(t *testing.T) {
	// Player's currentBet > lastBet → diff clamped to 0
	p := newMockPlayer(100)
	p.currentBet = 30
	players := []BettingPlayer{p, newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 10

	err := ExecuteBettingAction(players, state, 0, bettingActionRaise, 10, 10)
	assert.NoError(t, err)
	// diff=0, totalNeeded=10
	assert.Equal(t, 90, players[0].GetChips())
	assert.Equal(t, 40, players[0].GetCurrentBet()) // 30+10
}

func TestExecuteBettingAction_AllIn_Success(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(50), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 10

	err := ExecuteBettingAction(players, state, 0, bettingActionAllIn, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, players[0].GetChips())
	assert.Equal(t, 50, players[0].GetCurrentBet())
	assert.Equal(t, 50, state.Pot)
	assert.True(t, players[0].GetAllIn())
	// newBet (50) > lastBet (20), raiseAmount (30) >= minRaise (10)
	assert.Equal(t, 50, state.LastBet)
	assert.Equal(t, 30, state.MinRaise)
	assert.Equal(t, 1, state.RaiseCount)
}

func TestExecuteBettingAction_AllIn_ShortAllIn(t *testing.T) {
	// newBet > lastBet but raiseAmount < minRaise → short all-in
	players := []BettingPlayer{newMockPlayer(5), newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 20

	err := ExecuteBettingAction(players, state, 0, bettingActionAllIn, 0, 10)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
	assert.True(t, state.ActedFlags[0])
	// raiseAmount=5-20 < 0, wait, newBet=5, lastBet=20, newBet < lastBet → else branch
	// Actually: newBet = 0 + 5 = 5, lastBet = 20, newBet < lastBet → else branch
	assert.Equal(t, 20, state.LastBet) // unchanged
}

func TestExecuteBettingAction_AllIn_NewBetAboveLastBet_ShortRaise(t *testing.T) {
	// newBet > lastBet but raiseAmount < minRaise → short all-in, acted but no resetActed
	p := newMockPlayer(25)
	p.currentBet = 10
	players := []BettingPlayer{p, newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 20
	state.ActedFlags[1] = true

	err := ExecuteBettingAction(players, state, 0, bettingActionAllIn, 0, 10)
	assert.NoError(t, err)
	// newBet = 10 + 25 = 35, lastBet = 20, raiseAmount = 15 < minRaise (20)
	assert.Equal(t, 35, state.LastBet)
	assert.True(t, state.ActedFlags[0]) // short all-in sets acted
	assert.True(t, state.ActedFlags[1]) // not reset (short all-in)
}

func TestExecuteBettingAction_AllIn_NewBetAboveLastBet_FullRaise(t *testing.T) {
	// newBet > lastBet and raiseAmount >= minRaise → full raise, resetActedExcept
	p := newMockPlayer(50)
	p.currentBet = 10
	players := []BettingPlayer{p, newMockPlayer(100)}
	state := newBettingState(2)
	state.LastBet = 20
	state.MinRaise = 20
	state.ActedFlags[1] = true

	err := ExecuteBettingAction(players, state, 0, bettingActionAllIn, 0, 10)
	assert.NoError(t, err)
	// newBet = 10 + 50 = 60, lastBet = 20, raiseAmount = 40 >= minRaise (20)
	assert.Equal(t, 60, state.LastBet)
	assert.Equal(t, 40, state.MinRaise)
	assert.True(t, state.ActedFlags[0])  // exceptIdx
	assert.False(t, state.ActedFlags[1]) // reset
}

func TestExecuteBettingAction_AllIn_NoChips(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(0), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, bettingActionAllIn, 0, 10)
	assert.ErrorIs(t, err, ErrInsufficientChips)
	assert.Contains(t, err.Error(), "No chips to go all-in")
}

func TestExecuteBettingAction_UnknownAction(t *testing.T) {
	players := []BettingPlayer{newMockPlayer(100), newMockPlayer(100)}
	state := newBettingState(2)

	err := ExecuteBettingAction(players, state, 0, 99, 0, 10)
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.Contains(t, err.Error(), "Unknown action")
}

// --- ResetActedExcept tests ---

func TestResetActedExcept_Normal(t *testing.T) {
	players := []BettingPlayer{
		newMockPlayer(100),
		newMockPlayer(100),
		newMockPlayer(100),
	}
	flags := []bool{true, true, true}

	ResetActedExcept(players, flags, 1)
	assert.False(t, flags[0]) // reset
	assert.True(t, flags[1])  // except
	assert.False(t, flags[2]) // reset
}

func TestResetActedExcept_FoldedAndAllIn(t *testing.T) {
	p0 := newMockPlayer(100)
	p0.folded = true
	p1 := newMockPlayer(100)
	p1.allIn = true
	p2 := newMockPlayer(100)
	players := []BettingPlayer{p0, p1, p2}
	flags := []bool{true, true, true}

	ResetActedExcept(players, flags, 2)
	assert.True(t, flags[0]) // folded, not reset
	assert.True(t, flags[1]) // allIn, not reset
	assert.True(t, flags[2]) // except
}

// --- CalculateSidePots tests ---

func TestBetting_CalculateSidePots_NoAllIn(t *testing.T) {
	p0 := newMockPlayer(80) // invested 20
	p1 := newMockPlayer(80) // invested 20
	players := []BettingPlayer{p0, p1}
	startingChips := []int{100, 100}

	pots := CalculateSidePots(players, 40, startingChips)
	assert.Len(t, pots, 1)
	assert.Equal(t, 40, pots[0].Amount)
	assert.Equal(t, []int{0, 1}, pots[0].EligiblePlayers)
}

func TestBetting_CalculateSidePots_NoAllIn_FoldedExcluded(t *testing.T) {
	p0 := newMockPlayer(80)
	p0.folded = true
	p1 := newMockPlayer(80)
	players := []BettingPlayer{p0, p1}
	startingChips := []int{100, 100}

	pots := CalculateSidePots(players, 40, startingChips)
	assert.Len(t, pots, 1)
	assert.Equal(t, []int{1}, pots[0].EligiblePlayers)
}

func TestBetting_CalculateSidePots_SingleAllIn(t *testing.T) {
	p0 := newMockPlayer(0) // invested 50
	p0.allIn = true
	p1 := newMockPlayer(50) // invested 50
	players := []BettingPlayer{p0, p1}
	startingChips := []int{50, 100}

	pots := CalculateSidePots(players, 100, startingChips)
	assert.Len(t, pots, 1)
	assert.Equal(t, 100, pots[0].Amount)
	assert.Equal(t, []int{0, 1}, pots[0].EligiblePlayers)
}

func TestBetting_CalculateSidePots_MultiLevelAllIn(t *testing.T) {
	// 3 players: p0 all-in 30, p1 all-in 60, p2 invested 100
	p0 := newMockPlayer(0) // 30 invested
	p0.allIn = true
	p1 := newMockPlayer(0) // 60 invested
	p1.allIn = true
	p2 := newMockPlayer(0) // 100 invested
	players := []BettingPlayer{p0, p1, p2}
	startingChips := []int{30, 60, 100}

	pots := CalculateSidePots(players, 190, startingChips)
	// Pot 1: level 30 → 30*3=90, eligible [0,1,2]
	// Pot 2: level 60 → 30*2=60, eligible [1,2]
	// Remaining: 190-90-60=40, eligible [2] (non-allin non-folded) → [2]
	assert.Len(t, pots, 3)
	assert.Equal(t, 90, pots[0].Amount)
	assert.Equal(t, []int{0, 1, 2}, pots[0].EligiblePlayers)
	assert.Equal(t, 60, pots[1].Amount)
	assert.Equal(t, []int{1, 2}, pots[1].EligiblePlayers)
	assert.Equal(t, 40, pots[2].Amount)
	assert.Equal(t, []int{2}, pots[2].EligiblePlayers)
}

func TestBetting_CalculateSidePots_DuplicateLevels(t *testing.T) {
	// Two players all-in at same level → skip duplicate
	p0 := newMockPlayer(0)
	p0.allIn = true
	p1 := newMockPlayer(0)
	p1.allIn = true
	p2 := newMockPlayer(50) // invested 50
	players := []BettingPlayer{p0, p1, p2}
	startingChips := []int{30, 30, 100}

	pots := CalculateSidePots(players, 110, startingChips)
	// Level 30: 30*3=90, eligible [0,1,2]
	// Duplicate 30 → skip
	// Remaining: 110-90=20, eligible [2]
	assert.Len(t, pots, 2)
	assert.Equal(t, 90, pots[0].Amount)
	assert.Equal(t, 20, pots[1].Amount)
	assert.Equal(t, []int{2}, pots[1].EligiblePlayers)
}

func TestBetting_CalculateSidePots_AllAllIn_Remaining(t *testing.T) {
	// All players all-in, remaining pot → eligible = all non-folded
	p0 := newMockPlayer(0)
	p0.allIn = true
	p1 := newMockPlayer(0)
	p1.allIn = true
	players := []BettingPlayer{p0, p1}
	startingChips := []int{30, 50}

	pots := CalculateSidePots(players, 80, startingChips)
	// Level 30: 30*2=60, eligible [0,1]
	// Level 50: 20*1=20, eligible [1]
	assert.Len(t, pots, 2)
	assert.Equal(t, 60, pots[0].Amount)
	assert.Equal(t, 20, pots[1].Amount)
	assert.Equal(t, []int{1}, pots[1].EligiblePlayers)
}

func TestBetting_CalculateSidePots_AllAllIn_WithRemainder(t *testing.T) {
	// All players all-in with extra remaining → falls to len(eligible)==0 branch
	p0 := newMockPlayer(0)
	p0.allIn = true
	p1 := newMockPlayer(0)
	p1.allIn = true
	players := []BettingPlayer{p0, p1}
	startingChips := []int{30, 30}

	// pot > sum of investments → remaining > 0 and no non-allin players
	pots := CalculateSidePots(players, 70, startingChips)
	assert.Len(t, pots, 2)
	assert.Equal(t, 60, pots[0].Amount)
	// Remaining: eligible [] → fallback to non-folded [0,1]
	assert.Equal(t, 10, pots[1].Amount)
	assert.Equal(t, []int{0, 1}, pots[1].EligiblePlayers)
}

func TestBetting_CalculateSidePots_FoldedContributor(t *testing.T) {
	// Folded player contributes to pot but is not eligible
	p0 := newMockPlayer(70) // invested 30
	p0.folded = true
	p1 := newMockPlayer(0) // invested 50
	p1.allIn = true
	p2 := newMockPlayer(50) // invested 50
	players := []BettingPlayer{p0, p1, p2}
	startingChips := []int{100, 50, 100}

	pots := CalculateSidePots(players, 130, startingChips)
	// Level 50 (p1 all-in): layer 50
	//   p0 contribution=30 (capped at 50) → 30, not eligible (folded)
	//   p1 contribution=50 → 50, eligible
	//   p2 contribution=50 → 50, eligible
	//   potAmount=130, wait let me recalculate
	// Actually: layer amount = 50-0 = 50
	//   p0: min(30, 50)=30, folded
	//   p1: min(50, 50)=50, eligible
	//   p2: min(50, 50)=50, eligible
	//   potAmount = 30+50+50 = 130
	//   remaining = 130-130 = 0
	assert.Len(t, pots, 1)
	assert.Equal(t, 130, pots[0].Amount)
	assert.Equal(t, []int{1, 2}, pots[0].EligiblePlayers)
}

func TestBetting_CalculateSidePots_NegativeInvested(t *testing.T) {
	// Player has more chips than starting → invested < 0 → clamped to 0
	p0 := newMockPlayer(150) // invested = 100-150 = -50 → 0
	p1 := newMockPlayer(0)
	p1.allIn = true
	players := []BettingPlayer{p0, p1}
	startingChips := []int{100, 50}

	pots := CalculateSidePots(players, 50, startingChips)
	// Level 50: p0 contribution=max(0-0,0)=0, p1 contribution=50
	// potAmount=50, remaining=0
	assert.Len(t, pots, 1)
	assert.Equal(t, 50, pots[0].Amount)
	assert.Equal(t, []int{1}, pots[0].EligiblePlayers)
}

// --- FindPotWinners tests ---

func TestBetting_FindPotWinners_BasicWin(t *testing.T) {
	p0 := &mockBettingPlayer{handRank: 5, cards: []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 1, false),
	}}
	p1 := &mockBettingPlayer{handRank: 2, cards: []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
	}}
	players := []BettingPlayer{p0, p1}

	winners := FindPotWinners(players, []int{0, 1})
	assert.Equal(t, []int{0}, winners)
}

func TestBetting_FindPotWinners_Tie(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 1, false),
	}
	p0 := &mockBettingPlayer{handRank: 4, cards: cards}
	p1 := &mockBettingPlayer{handRank: 4, cards: cards}
	players := []BettingPlayer{p0, p1}

	winners := FindPotWinners(players, []int{0, 1})
	assert.Equal(t, []int{0, 1}, winners)
}

func TestBetting_FindPotWinners_KickerComparison(t *testing.T) {
	// Same rank, different kicker
	p0 := &mockBettingPlayer{handRank: 1, cards: []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 5, false),
	}}
	p1 := &mockBettingPlayer{handRank: 1, cards: []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignSpade, 1, false), // Ace kicker
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 2, false),
	}}
	players := []BettingPlayer{p0, p1}

	winners := FindPotWinners(players, []int{0, 1})
	assert.Equal(t, []int{1}, winners)
}

func TestBetting_FindPotWinners_FoldedExcluded(t *testing.T) {
	p0 := &mockBettingPlayer{handRank: 9, folded: true, cards: []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 1, false),
	}}
	p1 := &mockBettingPlayer{handRank: 0, cards: []*Card{
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 10, false),
	}}
	players := []BettingPlayer{p0, p1}

	winners := FindPotWinners(players, []int{0, 1})
	assert.Equal(t, []int{1}, winners)
}

func TestBetting_FindPotWinners_SameRankLowerKicker(t *testing.T) {
	// Same hand rank, p0 has higher kicker; tests the cmp < 0 branch for p1
	p0 := &mockBettingPlayer{handRank: 1, cards: []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 11, false),
	}}
	p1 := &mockBettingPlayer{handRank: 1, cards: []*Card{
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 3, false),
	}}
	players := []BettingPlayer{p0, p1}

	winners := FindPotWinners(players, []int{0, 1})
	assert.Equal(t, []int{0}, winners)
}

// --- DistributePots tests ---

func TestDistributePots_SinglePot(t *testing.T) {
	p0 := &mockBettingPlayer{handRank: 5, chips: 0, cards: []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 1, false),
	}}
	p1 := &mockBettingPlayer{handRank: 2, chips: 0, cards: []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
	}}
	players := []BettingPlayer{p0, p1}
	sidePots := []SidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}}

	won := DistributePots(players, sidePots)
	assert.Equal(t, 100, won[0])
	assert.Equal(t, 0, won[1])
	assert.Equal(t, 100, players[0].GetChips())
}

func TestDistributePots_SplitWithRemainder(t *testing.T) {
	cards := []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 11, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 1, false),
	}
	p0 := &mockBettingPlayer{handRank: 4, chips: 0, cards: cards}
	p1 := &mockBettingPlayer{handRank: 4, chips: 0, cards: cards}
	p2 := &mockBettingPlayer{handRank: 4, chips: 0, cards: cards}
	players := []BettingPlayer{p0, p1, p2}
	sidePots := []SidePot{{Amount: 100, EligiblePlayers: []int{0, 1, 2}}}

	won := DistributePots(players, sidePots)
	// 100/3=33, remainder=1 → first winner gets 34
	assert.Equal(t, 34, won[0])
	assert.Equal(t, 33, won[1])
	assert.Equal(t, 33, won[2])
}

func TestDistributePots_MultiplePots(t *testing.T) {
	p0 := &mockBettingPlayer{handRank: 5, chips: 0, cards: []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignSpade, 1, false),
	}}
	p1 := &mockBettingPlayer{handRank: 2, chips: 0, cards: []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
	}}
	players := []BettingPlayer{p0, p1}
	sidePots := []SidePot{
		{Amount: 60, EligiblePlayers: []int{0, 1}},
		{Amount: 40, EligiblePlayers: []int{1}},
	}

	won := DistributePots(players, sidePots)
	assert.Equal(t, 60, won[0])
	assert.Equal(t, 40, won[1])
}

func TestDistributePots_EmptyWinners(t *testing.T) {
	// All eligible players are folded → no winners → skip
	p0 := &mockBettingPlayer{folded: true, chips: 0}
	players := []BettingPlayer{p0}
	sidePots := []SidePot{{Amount: 100, EligiblePlayers: []int{0}}}

	won := DistributePots(players, sidePots)
	assert.Equal(t, 0, won[0])
}

// --- CPU helper tests ---

func TestCpuFoldOrCheck(t *testing.T) {
	t.Run("fold when callAmount > 0", func(t *testing.T) {
		action, amount := CpuFoldOrCheck(10)
		assert.Equal(t, bettingActionFold, action)
		assert.Equal(t, 0, amount)
	})
	t.Run("check when callAmount == 0", func(t *testing.T) {
		action, amount := CpuFoldOrCheck(0)
		assert.Equal(t, bettingActionCheck, action)
		assert.Equal(t, 0, amount)
	})
}

func TestCpuCallOrCheck(t *testing.T) {
	t.Run("call when callAmount > 0", func(t *testing.T) {
		action, amount := CpuCallOrCheck(10)
		assert.Equal(t, bettingActionCall, action)
		assert.Equal(t, 0, amount)
	})
	t.Run("check when callAmount == 0", func(t *testing.T) {
		action, amount := CpuCallOrCheck(0)
		assert.Equal(t, bettingActionCheck, action)
		assert.Equal(t, 0, amount)
	})
}

func TestCpuRaiseOrBet(t *testing.T) {
	t.Run("allIn when raiseAmt > chips", func(t *testing.T) {
		action, amount := CpuRaiseOrBet(10, 5, 20)
		assert.Equal(t, bettingActionAllIn, action)
		assert.Equal(t, 0, amount)
	})
	t.Run("allIn when raise+call > chips", func(t *testing.T) {
		action, amount := CpuRaiseOrBet(20, 15, 10)
		assert.Equal(t, bettingActionAllIn, action)
		assert.Equal(t, 0, amount)
	})
	t.Run("raise when callAmount > 0 and affordable", func(t *testing.T) {
		action, amount := CpuRaiseOrBet(100, 10, 20)
		assert.Equal(t, bettingActionRaise, action)
		assert.Equal(t, 20, amount)
	})
	t.Run("bet when callAmount == 0", func(t *testing.T) {
		action, amount := CpuRaiseOrBet(100, 0, 20)
		assert.Equal(t, bettingActionBet, action)
		assert.Equal(t, 20, amount)
	})
}
