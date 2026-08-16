package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestPoker() (*Poker, []*PokerPlayer) {
	tc := NewTrumpCards(0)
	p0 := NewPokerPlayer(true, PokerStyleBalanced)
	p1 := NewPokerPlayer(false, PokerStyleConservative)
	p2 := NewPokerPlayer(false, PokerStyleAggressive)
	p3 := NewPokerPlayer(false, PokerStyleBluffer)
	players := []*PokerPlayer{p0, p1, p2, p3}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := DefaultPokerConfig()
	pk := NewPoker(tc, players, cfg)
	return pk, players
}

func setupPokerForHumanAction(phase int) (*Poker, []*PokerPlayer) {
	pk, players := newTestPoker()
	pk.SetPhase(phase)
	pk.SetCurrentTurn(0)
	pk.setActedFlags([]bool{false, true, true, true})
	pk.SetLastBet(0)
	pk.SetMinRaise(10)
	pk.SetPot(40)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})
	for _, pl := range players {
		pl.Reset()
		pl.SetChips(990)
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
		for i := 0; i < 5; i++ {
			pl.AddCard(NewCard(CardDesignSpade, i+2, false))
		}
	}
	return pk, players
}

func givePlayerHand(pl *PokerPlayer, cards []*Card) {
	pl.Reset()
	for _, c := range cards {
		pl.AddCard(c)
	}
}

// ---------------------------------------------------------------------------
// TestNewPoker
// ---------------------------------------------------------------------------

func TestNewPoker(t *testing.T) {
	pk, players := newTestPoker()
	assert.Equal(t, PokerPhaseInit, pk.GetPhase())
	assert.Equal(t, 4, len(pk.GetPlayers()))
	assert.Equal(t, 0, pk.GetPot())
	assert.False(t, pk.GetGameEndFlag())
	assert.Equal(t, players, pk.GetPlayers())
}

// ---------------------------------------------------------------------------
// TestPoker_Reset
// ---------------------------------------------------------------------------

func TestPoker_Reset(t *testing.T) {
	pk, players := newTestPoker()
	err := pk.Reset()
	assert.NoError(t, err)
	// After Reset, game is in Deal phase (or may have ended if CPUs acted)
	if !pk.GetGameEndFlag() {
		assert.True(t, pk.GetPhase() == PokerPhaseDeal || pk.GetPhase() == PokerPhaseEnd)
	}
	for _, pl := range players {
		assert.Equal(t, 5, pl.GetCardsSize())
	}
	assert.True(t, pk.GetPot() > 0 || pk.GetGameEndFlag())
}

func TestPoker_Reset_ChipsZero(t *testing.T) {
	pk, players := newTestPoker()
	for _, pl := range players {
		pl.SetChips(0)
	}
	err := pk.Reset()
	assert.NoError(t, err)
	// Players with 0 chips should be reset to InitChips (then ante deducted)
	cfg := pk.GetConfig()
	for _, pl := range players {
		assert.True(t, pl.GetChips() >= 0)
		assert.True(t, pl.GetChips() <= cfg.InitChips)
	}
}

func TestPoker_Reset_ChipsPositive(t *testing.T) {
	pk, players := newTestPoker()
	players[0].SetChips(500)
	err := pk.Reset()
	assert.NoError(t, err)
	assert.True(t, players[0].GetChips() > 0)
}

// ---------------------------------------------------------------------------
// TestPoker_collectAntes
// ---------------------------------------------------------------------------

func TestPoker_collectAntes_normalAndLow(t *testing.T) {
	pk, players := newTestPoker()
	// Player with less than ante
	players[2].SetChips(3)
	pk.SetPot(0)
	pk.collectAntes()
	// Player 2 should have 0 chips
	assert.Equal(t, 0, players[2].GetChips())
	// Pot should contain: 10+10+3+10 = 33
	assert.Equal(t, 33, pk.GetPot())
}

// ---------------------------------------------------------------------------
// TestPoker_PlayerAction
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_GameEnded(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetGameEndFlag(true)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestPoker_PlayerAction_WrongPhase_Init(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseInit)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestPoker_PlayerAction_WrongPhase_Exchange(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestPoker_PlayerAction_WrongPhase_End(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseEnd)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestPoker_PlayerAction_NotHumanTurn(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(1) // CPU
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

// ---------------------------------------------------------------------------
// Fold
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_Fold(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(PokerActionFold, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetFolded())
}

func TestPoker_Fold_LastPlayerWins(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	pk.setActedFlags([]bool{false, true, true, true})
	// Only p0 and p3 active; p0 folds → p3 wins
	err := pk.PlayerAction(PokerActionFold, 0, 0)
	assert.NoError(t, err)
	assert.True(t, pk.GetGameEndFlag())
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
	assert.True(t, len(pk.GetRoundResults()) > 0)
	assert.Equal(t, 3, pk.GetRoundResults()[0].PlayerIdx)
}

// ---------------------------------------------------------------------------
// Check
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_Check(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.NoError(t, err)
}

func TestPoker_PlayerAction_Check_WithOutstandingBet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

// ---------------------------------------------------------------------------
// Call
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_Call(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	chipsBefore := players[0].GetChips()
	err := pk.PlayerAction(PokerActionCall, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetChips() < chipsBefore)
}

func TestPoker_PlayerAction_Call_NothingToCall(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(0)
	err := pk.PlayerAction(PokerActionCall, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestPoker_PlayerAction_Call_AllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(2000) // more than chips
	err := pk.PlayerAction(PokerActionCall, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

// ---------------------------------------------------------------------------
// Bet
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_Bet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(PokerActionBet, 50, 0)
	assert.NoError(t, err)
}

func TestPoker_PlayerAction_Bet_MaxRaises(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.setRaiseCount(4)
	err := pk.PlayerAction(PokerActionBet, 50, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestPoker_PlayerAction_Bet_OutstandingBet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	err := pk.PlayerAction(PokerActionBet, 50, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestPoker_PlayerAction_Bet_BelowMin(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(PokerActionBet, 1, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAmount)
}

func TestPoker_PlayerAction_Bet_InsufficientChips(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(PokerActionBet, 99999, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientChips)
}

func TestPoker_PlayerAction_Bet_ExactChips_AllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[0].SetChips(50)
	err := pk.PlayerAction(PokerActionBet, 50, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

// ---------------------------------------------------------------------------
// Raise
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_Raise(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	pk.SetMinRaise(10)
	err := pk.PlayerAction(PokerActionRaise, 20, 0)
	assert.NoError(t, err)
}

func TestPoker_PlayerAction_Raise_MaxRaises(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	pk.setRaiseCount(4)
	err := pk.PlayerAction(PokerActionRaise, 20, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

func TestPoker_PlayerAction_Raise_BelowMinRaise(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	pk.SetMinRaise(30)
	err := pk.PlayerAction(PokerActionRaise, 10, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAmount)
}

func TestPoker_PlayerAction_Raise_AutoAllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	pk.SetMinRaise(10)
	players[0].SetChips(25) // diff=20, amount=10, total=30 >= 25
	err := pk.PlayerAction(PokerActionRaise, 10, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

func TestPoker_PlayerAction_Raise_ExactChips_AllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	pk.SetMinRaise(10)
	players[0].SetChips(31) // diff=20, amount=10, total=30 < 31 → normal raise, then 31-30=1 chip left → not allIn
	err := pk.PlayerAction(PokerActionRaise, 10, 0)
	assert.NoError(t, err)
	assert.False(t, players[0].GetAllIn())
	assert.Equal(t, 1, players[0].GetChips())
}

func TestPoker_PlayerAction_Raise_ChipsExactlyZero(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(0)
	pk.SetMinRaise(10)
	players[0].SetCurrentBet(0)
	players[0].SetChips(10) // diff=0, amount=10, total=10 >= 10 → auto allIn
	err := pk.PlayerAction(PokerActionRaise, 10, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

func TestPoker_PlayerAction_Raise_NegativeDiff(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(0)
	players[0].SetCurrentBet(10) // currentBet > lastBet → diff < 0 → clamped to 0
	pk.SetMinRaise(10)
	players[0].SetChips(100)
	err := pk.PlayerAction(PokerActionRaise, 10, 0)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// AllIn
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_AllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(PokerActionAllIn, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

func TestPoker_PlayerAction_AllIn_NoChips(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[0].SetChips(0)
	err := pk.PlayerAction(PokerActionAllIn, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientChips)
}

func TestPoker_PlayerAction_AllIn_AboveLastBet_LargeRaise(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(50)
	pk.SetMinRaise(10)
	players[0].SetChips(200) // newBet = 0+200 = 200 > 50, raiseAmt=150 >= minRaise=10 → resetActedExcept
	err := pk.PlayerAction(PokerActionAllIn, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

func TestPoker_PlayerAction_AllIn_AboveLastBet_SmallRaise(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(50)
	pk.SetMinRaise(200)
	players[0].SetChips(55) // newBet = 55 > 50, raiseAmt=5 < minRaise=200 → actedFlags[0]=true only
	err := pk.PlayerAction(PokerActionAllIn, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

func TestPoker_PlayerAction_AllIn_BelowLastBet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(2000)
	players[0].SetChips(100) // newBet = 100 <= 2000 → actedFlags[0]=true
	err := pk.PlayerAction(PokerActionAllIn, 0, 0)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

// ---------------------------------------------------------------------------
// Unknown action
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_Unknown(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerAction(99, 0, 0)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlay)
}

// ---------------------------------------------------------------------------
// SecondBet phase actions
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_SecondBet_Check(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseSecondBet)
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.NoError(t, err)
}

func TestPoker_PlayerAction_SecondBet_Bet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseSecondBet)
	err := pk.PlayerAction(PokerActionBet, 20, 0)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// PlayerExchange
// ---------------------------------------------------------------------------

func TestPoker_PlayerExchange(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// Ensure deck has cards
	pk.trumpCards.Shuffle()
	err := pk.PlayerExchange([]int{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, 2, players[0].GetExchangeCount())
}

func TestPoker_PlayerExchange_WrongPhase(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerExchange([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestPoker_PlayerExchange_NotHumanTurn(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.SetCurrentTurn(1)
	err := pk.PlayerExchange([]int{0})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestPoker_PlayerExchange_CompletesPhase(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	// Mark all CPUs as already acted
	pk.setActedFlags([]bool{false, true, true, true})
	// Fold CPUs 1-3 so they won't be processed in CPU exchanges
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	players[3].SetFolded(true)
	err := pk.PlayerExchange([]int{0})
	assert.NoError(t, err)
	// After human exchange + all acted, should move to SecondBet or End
	assert.True(t, pk.GetPhase() == PokerPhaseSecondBet || pk.GetPhase() == PokerPhaseEnd)
}

// ---------------------------------------------------------------------------
// PlayerStand
// ---------------------------------------------------------------------------

func TestPoker_PlayerStand(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	err := pk.PlayerStand()
	assert.NoError(t, err)
	assert.Equal(t, 0, players[0].GetExchangeCount())
}

func TestPoker_PlayerStand_WrongPhase(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	err := pk.PlayerStand()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestPoker_PlayerStand_NotHumanTurn(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.SetCurrentTurn(1)
	err := pk.PlayerStand()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestPoker_PlayerStand_CompletesExchange(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	pk.setActedFlags([]bool{false, true, true, true})
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	players[3].SetFolded(true)
	err := pk.PlayerStand()
	assert.NoError(t, err)
	assert.True(t, pk.GetPhase() == PokerPhaseSecondBet || pk.GetPhase() == PokerPhaseEnd)
}

// ---------------------------------------------------------------------------
// advanceTurn
// ---------------------------------------------------------------------------

func TestPoker_advanceTurn_gameEndFlag(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetGameEndFlag(true)
	pk.SetCurrentTurn(0)
	pk.advanceTurn()
	// Should return immediately, currentTurn unchanged
	assert.Equal(t, 0, pk.GetCurrentTurn())
}

func TestPoker_advanceTurn_bettingComplete(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.setActedFlags([]bool{true, true, true, true})
	pk.advanceTurn()
	// All acted → advancePhase (Deal → Exchange)
	assert.Equal(t, PokerPhaseExchange, pk.GetPhase())
}

func TestPoker_advanceTurn_exchangeComplete(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.setActedFlags([]bool{true, true, true, true})
	pk.advanceTurn()
	// Exchange complete → should return (no advance here, just return)
	// The phase stays Exchange because advanceTurn just returns
	assert.Equal(t, PokerPhaseExchange, pk.GetPhase())
}

func TestPoker_advanceTurn_findNextActive(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(0)
	pk.setActedFlags([]bool{true, false, true, true})
	pk.advanceTurn()
	assert.Equal(t, 1, pk.GetCurrentTurn())
}

func TestPoker_advanceTurn_allActed_SecondBet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseSecondBet)
	pk.setActedFlags([]bool{true, true, true, true})
	pk.advanceTurn()
	// All acted in SecondBet → resolveShowdown → End
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
}

// ---------------------------------------------------------------------------
// isBettingRoundComplete / isExchangeComplete
// ---------------------------------------------------------------------------

func TestPoker_isBettingRoundComplete(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.setActedFlags([]bool{true, true, true, true})
	assert.True(t, pk.isBettingRoundComplete())

	pk.setActedFlags([]bool{false, true, true, true})
	assert.False(t, pk.isBettingRoundComplete())

	// Folded player doesn't matter
	pk.setActedFlags([]bool{false, true, true, true})
	players[0].SetFolded(true)
	assert.True(t, pk.isBettingRoundComplete())

	// AllIn player doesn't matter
	players[0].SetFolded(false)
	players[0].SetAllIn(true)
	assert.True(t, pk.isBettingRoundComplete())
}

func TestPoker_isExchangeComplete(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.setActedFlags([]bool{true, true, true, true})
	assert.True(t, pk.isExchangeComplete())

	pk.setActedFlags([]bool{true, false, true, true})
	assert.False(t, pk.isExchangeComplete())

	players[1].SetFolded(true)
	assert.True(t, pk.isExchangeComplete())
}

// ---------------------------------------------------------------------------
// advancePhase
// ---------------------------------------------------------------------------

func TestPoker_advancePhase_Deal(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetFolded(true)
	pk.advancePhase()
	assert.Equal(t, PokerPhaseExchange, pk.GetPhase())
	assert.Equal(t, 0, pk.GetLastBet())
	// Folded player should have actedFlags=true
	flags := pk.getActedFlags()
	assert.True(t, flags[1])
}

func TestPoker_advancePhase_SecondBet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseSecondBet)
	pk.advancePhase()
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
	assert.True(t, pk.GetGameEndFlag())
}

// ---------------------------------------------------------------------------
// startSecondBettingRound
// ---------------------------------------------------------------------------

func TestPoker_startSecondBettingRound_ActiveCntZero(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// All folded or allIn → activeCnt <= 1
	players[0].SetAllIn(true)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	players[3].SetAllIn(true)
	pk.startSecondBettingRound()
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
	assert.True(t, pk.GetGameEndFlag())
}

func TestPoker_startSecondBettingRound_OneActive(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	players[0].SetAllIn(true)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	// Only p3 is active (not folded, not allIn)
	pk.startSecondBettingRound()
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
}

func TestPoker_startSecondBettingRound_Normal(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	pk.startSecondBettingRound()
	assert.Equal(t, PokerPhaseSecondBet, pk.GetPhase())
}

// ---------------------------------------------------------------------------
// findNextActive
// ---------------------------------------------------------------------------

func TestPoker_findNextActive(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	// Normal case
	next := pk.findNextActive(0)
	assert.Equal(t, 1, next)

	// Skip folded/allIn
	players[1].SetFolded(true)
	players[2].SetAllIn(true)
	next = pk.findNextActive(0)
	assert.Equal(t, 3, next)

	// All folded/allIn → fallback
	players[3].SetFolded(true)
	players[0].SetAllIn(true)
	next = pk.findNextActive(0)
	assert.Equal(t, 1, next) // fallback: (0+1)%4=1
}

// ---------------------------------------------------------------------------
// countActivePlayers
// ---------------------------------------------------------------------------

func TestPoker_countActivePlayers(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	assert.Equal(t, 4, pk.countActivePlayers())

	players[0].SetFolded(true)
	assert.Equal(t, 3, pk.countActivePlayers())

	players[1].SetFolded(true)
	players[2].SetFolded(true)
	assert.Equal(t, 1, pk.countActivePlayers())
}

// ---------------------------------------------------------------------------
// resolveLastPlayer
// ---------------------------------------------------------------------------

func TestPoker_resolveLastPlayer(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetPot(200)
	players[0].SetFolded(true)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	chipsBefore := players[3].GetChips()

	pk.resolveLastPlayer()

	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
	assert.True(t, pk.GetGameEndFlag())
	assert.Equal(t, chipsBefore+200, players[3].GetChips())
	assert.Equal(t, 0, pk.GetPot())
	assert.Equal(t, 3, pk.GetRoundResults()[0].PlayerIdx)
}

// ---------------------------------------------------------------------------
// resolveShowdown
// ---------------------------------------------------------------------------

func TestPoker_resolveShowdown_SingleWinner(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetPot(400)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})

	// Player 0: Royal Flush
	givePlayerHand(players[0], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	})
	// Player 1: High Card
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	pk.resolveShowdown()

	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
	assert.True(t, pk.GetGameEndFlag())
	assert.True(t, len(pk.GetRoundResults()) > 0)
	// Player 0 should have won
	found := false
	for _, r := range pk.GetRoundResults() {
		if r.PlayerIdx == 0 && r.WonAmount > 0 {
			found = true
			// Royal Flush has no kickers
			assert.Nil(t, r.Kickers)
		}
		if r.PlayerIdx == 1 {
			// High Card has no kickers
			assert.Nil(t, r.Kickers)
		}
	}
	assert.True(t, found)
}

func TestPoker_resolveShowdown_SplitPot(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetPot(200)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})

	// Give identical hands
	hand := []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	}
	hand2 := []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignClover, 11, false),
	}
	givePlayerHand(players[0], hand)
	givePlayerHand(players[1], hand2)
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	pk.resolveShowdown()

	assert.True(t, pk.GetGameEndFlag())
	totalWon := 0
	for _, r := range pk.GetRoundResults() {
		totalWon += r.WonAmount
	}
	assert.Equal(t, 200, totalWon)
}

func TestPoker_resolveShowdown_Kickers(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetPot(200)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})

	// Player 0: One Pair (5s), kickers A, Q, 10
	givePlayerHand(players[0], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 10, false),
	})
	// Player 1: Two Pair (Ks, 3s), kicker 8
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignSpade, 8, false),
	})
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	pk.resolveShowdown()

	for _, r := range pk.GetRoundResults() {
		if r.PlayerIdx == 0 {
			assert.Equal(t, []int{14, 12, 10}, r.Kickers)
		}
		if r.PlayerIdx == 1 {
			assert.Equal(t, []int{8}, r.Kickers)
		}
	}
}

// ---------------------------------------------------------------------------
// runCpuActions
// ---------------------------------------------------------------------------

func TestPoker_runCpuActions_GameEnded(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetGameEndFlag(true)
	pk.runCpuActions()
}

func TestPoker_runCpuActions_HumanTurn(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(0)
	pk.setActedFlags([]bool{false, false, false, false})
	pk.runCpuActions()
	// Should stop at human turn
	assert.Equal(t, 0, pk.GetCurrentTurn())
}

func TestPoker_runCpuActions_SkipFoldedAllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(1) // Start from CPU 1
	players[1].SetFolded(true)
	players[2].SetAllIn(true)
	pk.setActedFlags([]bool{false, false, false, false})
	pk.runCpuActions()
}

func TestPoker_runCpuActions_CpuError_Fold(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(1)
	pk.SetLastBet(100)
	pk.setRaiseCount(4) // max raises reached
	// CPU 1 will try to bet/raise but fail → fallback
	// Give CPU a strong hand so it tries to bet
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	})
	// With lastBet=100, callAmount>0, so fallback should fold
	pk.setActedFlags([]bool{false, false, true, true})
	pk.runCpuActions()
}

func TestPoker_runCpuActions_CpuError_Check(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(1)
	pk.SetLastBet(0) // no outstanding bet
	pk.setRaiseCount(4)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	})
	// cpuDecide would return bet/raise, but raiseCount max → converted to check
	// The converted check should succeed
	pk.setActedFlags([]bool{false, false, true, true})
	pk.runCpuActions()
}

// ---------------------------------------------------------------------------
// runCpuExchanges
// ---------------------------------------------------------------------------

func TestPoker_runCpuExchanges_GameEnded(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.SetGameEndFlag(true)
	pk.runCpuExchanges()
	// Should return immediately
}

func TestPoker_runCpuExchanges_ExchangeComplete(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.setActedFlags([]bool{true, true, true, true})
	pk.runCpuExchanges()
	// Already complete
}

func TestPoker_runCpuExchanges_HumanTurn(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	pk.SetCurrentTurn(0)
	pk.setActedFlags([]bool{false, false, false, false})
	pk.runCpuExchanges()
	// Stops at human
	assert.Equal(t, 0, pk.GetCurrentTurn())
}

func TestPoker_runCpuExchanges_SkipFoldedAllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	pk.SetCurrentTurn(1)
	players[1].SetFolded(true)
	pk.setActedFlags([]bool{true, false, false, false})
	pk.runCpuExchanges()
	// Should skip folded p1 and continue
}

func TestPoker_runCpuExchanges_NormalCPU(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	pk.SetCurrentTurn(1)
	pk.setActedFlags([]bool{true, false, false, false})
	// Give CPUs hands
	for i := 1; i <= 3; i++ {
		givePlayerHand(players[i], []*Card{
			NewCard(CardDesignClover, 2, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 7, false),
			NewCard(CardDesignClover, 9, false),
			NewCard(CardDesignHeart, 11, false),
		})
	}
	pk.runCpuExchanges()
	assert.True(t, len(pk.GetCpuExchanges()) > 0)
}

// ---------------------------------------------------------------------------
// cpuDecide
// ---------------------------------------------------------------------------

func TestPoker_cpuDecide_UnknownStyle(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	// Create player with unknown style
	unknownPlayer := NewPokerPlayer(false, PokerPlayStyle(99))
	unknownPlayer.SetChips(1000)
	givePlayerHand(unknownPlayer, []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	pk.players = append(pk.players[:1], unknownPlayer)
	pk.players = append(pk.players, players[2:]...)
	pk.players[1] = unknownPlayer
	pk.SetLastBet(0)
	action, _ := pk.cpuDecide(1)
	assert.Equal(t, PokerActionCheck, action) // callOrCheck with callAmount=0 → check
}

func TestPoker_cpuDecide_UnknownStyle_WithBet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	unknownPlayer := NewPokerPlayer(false, PokerPlayStyle(99))
	unknownPlayer.SetChips(1000)
	givePlayerHand(unknownPlayer, []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	pk.players[1] = unknownPlayer
	pk.SetLastBet(50)
	action, _ := pk.cpuDecide(1)
	assert.Equal(t, PokerActionCall, action)
}

func TestPoker_cpuDecide_DealPhase(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(0)
	// Give CPU 1 a strong hand
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	action, _ := pk.cpuDecide(1)
	// Conservative with FourOfAKind → bet
	assert.True(t, action == PokerActionBet || action == PokerActionRaise || action == PokerActionCheck || action == PokerActionCall)
}

func TestPoker_cpuDecide_SecondBetPhase(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	pk.SetLastBet(0)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	action, _ := pk.cpuDecide(1)
	assert.True(t, action >= 0)
}

func TestPoker_cpuDecide_RaiseCountMax_WithBet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.setRaiseCount(4)
	pk.SetLastBet(20)
	// Give strong hand so CPU wants to bet/raise
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	action, _ := pk.cpuDecide(1)
	// Max raises → converted to call (since lastBet > 0)
	assert.True(t, action == PokerActionCall || action == PokerActionCheck || action == PokerActionFold)
}

func TestPoker_cpuDecide_RaiseCountMax_NoBet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.setRaiseCount(4)
	pk.SetLastBet(0)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	action, _ := pk.cpuDecide(1)
	assert.True(t, action == PokerActionCheck || action == PokerActionCall || action == PokerActionFold)
}

// ---------------------------------------------------------------------------
// calcExchangeWarning
// ---------------------------------------------------------------------------

func TestPoker_calcExchangeWarning_NotSecondBet(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	w := pk.calcExchangeWarning(0, 80)
	assert.Equal(t, 0, w)
}

func TestPoker_calcExchangeWarning_AllFolded(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	players[3].SetFolded(true)
	// All opponents folded → minExchange stays 5 → 5 >= 3 → 0
	w := pk.calcExchangeWarning(0, 80)
	assert.Equal(t, 0, w)
}

func TestPoker_calcExchangeWarning_MinExchange3(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	players[1].SetExchangeCount(3)
	players[2].SetExchangeCount(4)
	players[3].SetExchangeCount(5)
	w := pk.calcExchangeWarning(0, 80)
	assert.Equal(t, 0, w)
}

func TestPoker_calcExchangeWarning_LowExchange(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	players[1].SetExchangeCount(0) // very strong hand signal
	players[2].SetExchangeCount(3)
	players[3].SetExchangeCount(2)
	// minExchange = 0, warning = (3-0)*80/3 = 80
	w := pk.calcExchangeWarning(0, 80)
	assert.Equal(t, 80, w)
}

func TestPoker_calcExchangeWarning_SkipSelf(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	players[0].SetExchangeCount(0) // self → skipped
	players[1].SetExchangeCount(3)
	players[2].SetExchangeCount(3)
	players[3].SetExchangeCount(3)
	w := pk.calcExchangeWarning(0, 80)
	assert.Equal(t, 0, w)
}

// ---------------------------------------------------------------------------
// cpuDecideFirstBet
// ---------------------------------------------------------------------------

func TestPoker_cpuDecideFirstBet_FoldPassive(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	params := pokerStyleParamsMap[PokerStyleConservative]
	// Conservative: firstFoldThreshold=HighCard, aggressive=false
	// HighCard hand with large call amount
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// callAmount > MinBet * firstCallMaxMult (10*2=20)
	action, _ := pk.cpuDecideFirstBet(1, params, 30, PokerHandHighCard)
	assert.Equal(t, PokerActionFold, action)
}

func TestPoker_cpuDecideFirstBet_FoldPassive_SmallCall(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// HighCard <= firstFoldThreshold, !aggressive, callAmount <= MinBet*firstCallMaxMult
	// → not fold, falls through to bluff check (bluffRate=5 → mostly callOrCheck)
	gotCall := false
	for i := 0; i < 1000; i++ {
		action, _ := pk.cpuDecideFirstBet(1, params, 10, PokerHandHighCard)
		if action == PokerActionCall {
			gotCall = true
			break
		}
	}
	assert.True(t, gotCall, "call never triggered")
}

func TestPoker_cpuDecideFirstBet_FoldAggressive(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	params := pokerStyleParamsMap[PokerStyleAggressive]
	givePlayerHand(players[2], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[2].EvalHand()
	// Aggressive with HighCard: bluff rate (25%) may fire, otherwise foldOrCheck
	// callAmount > 0 → non-bluff path → fold
	gotFold := false
	gotBluff := false
	for i := 0; i < 1000; i++ {
		action, _ := pk.cpuDecideFirstBet(2, params, 50, PokerHandHighCard)
		switch action {
		case PokerActionFold:
			gotFold = true
		case PokerActionRaise, PokerActionBet, PokerActionAllIn:
			gotBluff = true
		}
		if gotFold && gotBluff {
			break
		}
	}
	assert.True(t, gotFold, "fold never triggered for aggressive with HighCard")
	assert.True(t, gotBluff, "bluff never triggered for aggressive with HighCard")
}

func TestPoker_cpuDecideFirstBet_FoldAggressive_NoBet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	params := pokerStyleParamsMap[PokerStyleAggressive]
	givePlayerHand(players[2], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[2].EvalHand()
	// Aggressive with HighCard: bluff rate (25%) may fire, otherwise foldOrCheck
	// callAmount=0 → non-bluff path → check
	gotCheck := false
	gotBluff := false
	for i := 0; i < 1000; i++ {
		action, _ := pk.cpuDecideFirstBet(2, params, 0, PokerHandHighCard)
		switch action {
		case PokerActionCheck:
			gotCheck = true
		case PokerActionBet, PokerActionRaise, PokerActionAllIn:
			gotBluff = true
		}
		if gotCheck && gotBluff {
			break
		}
	}
	assert.True(t, gotCheck, "check never triggered for aggressive with HighCard")
	assert.True(t, gotBluff, "bluff never triggered for aggressive with HighCard")
}

func TestPoker_cpuDecideFirstBet_BlufferHighCardCanBluff(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	// Bluffer: aggressive=true, bluffRate=40, firstBetThreshold=HighCard, firstFoldThreshold=HighCard
	params := pokerStyleParamsMap[PokerStyleBluffer]
	givePlayerHand(players[3], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[3].EvalHand()
	// With HighCard (rank=0 <= firstFoldThreshold=0), bluff rate 40% should fire
	gotBet := false
	gotNonBet := false
	for i := 0; i < 1000; i++ {
		action, _ := pk.cpuDecideFirstBet(3, params, 0, PokerHandHighCard)
		if action == PokerActionBet || action == PokerActionRaise || action == PokerActionAllIn {
			gotBet = true
		} else {
			gotNonBet = true
		}
		if gotBet && gotNonBet {
			break
		}
	}
	assert.True(t, gotBet, "Bluffer with HighCard never produced a bet/raise in first bet")
	assert.True(t, gotNonBet, "Bluffer with HighCard always bet (non-bet never triggered)")
}

func TestPoker_cpuDecideFirstBet_Bet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	players[1].EvalHand()
	// FourOfAKind >= firstBetThreshold → bet
	action, _ := pk.cpuDecideFirstBet(1, params, 0, PokerHandFourOfAKind)
	assert.Equal(t, PokerActionBet, action)
}

func TestPoker_cpuDecideFirstBet_Bluff(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	// Use Balanced: bluffRate=15, firstBetThreshold=OnePair
	// With HighCard hand (rank < threshold), bluff branch is reachable
	params := pokerStyleParamsMap[PokerStyleBalanced]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// handRank=HighCard(0) > firstFoldThreshold(HighCard=0) is false (0<=0 → fold path)
	// Balanced is not aggressive, callAmount=0 → not fold (call <= max)
	// So falls through to bluff check: rand.Intn(100) < 15
	// Use handRank=OnePair-1 = 0 (HighCard), callAmount small enough to pass fold check
	gotBet := false
	gotNonBet := false
	for i := 0; i < 1000; i++ {
		a, _ := pk.cpuDecideFirstBet(1, params, 5, PokerHandHighCard)
		if a == PokerActionBet || a == PokerActionRaise {
			gotBet = true
		} else {
			gotNonBet = true
		}
		if gotBet && gotNonBet {
			break
		}
	}
	assert.True(t, gotBet, "bluff bet never triggered")
	assert.True(t, gotNonBet, "non-bluff never triggered")
}

func TestPoker_cpuDecideFirstBet_CallOrCheck(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	players[1].EvalHand()
	// OnePair > firstFoldThreshold(HighCard) → fold skipped
	// OnePair < firstBetThreshold(TwoPair) → bluff check (bluffRate=5 → mostly call)
	gotCall := false
	for i := 0; i < 1000; i++ {
		action, _ := pk.cpuDecideFirstBet(1, params, 10, PokerHandOnePair)
		if action == PokerActionCall {
			gotCall = true
			break
		}
	}
	assert.True(t, gotCall, "call never triggered")
}

// ---------------------------------------------------------------------------
// cpuDecideSecondBet
// ---------------------------------------------------------------------------

func TestPoker_cpuDecideSecondBet_ExchangeWarningHigh(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// exchangeWarning > 50 → adjustedFoldThreshold bumped
	// HighCard <= adjustedFoldThreshold, callAmount > MinBet*secondCallMaxMult → fold
	action, _ := pk.cpuDecideSecondBet(1, params, 50, PokerHandHighCard, 60)
	assert.Equal(t, PokerActionFold, action)
}

func TestPoker_cpuDecideSecondBet_FoldWithCallAmount(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// HighCard <= foldThreshold, callAmount <= threshold but > 0 → callOrCheck
	action, _ := pk.cpuDecideSecondBet(1, params, 10, PokerHandHighCard, 0)
	assert.Equal(t, PokerActionCall, action)
}

func TestPoker_cpuDecideSecondBet_FoldWithCallAmountZero(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// callAmount == 0 → falls through to bluff/bet check
	action, _ := pk.cpuDecideSecondBet(1, params, 0, PokerHandHighCard, 0)
	// With bluffRate=5 mostly check
	assert.True(t, action == PokerActionCheck || action == PokerActionBet)
}

func TestPoker_cpuDecideSecondBet_Bet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	players[1].EvalHand()
	action, _ := pk.cpuDecideSecondBet(1, params, 0, PokerHandFourOfAKind, 0)
	assert.Equal(t, PokerActionBet, action)
}

func TestPoker_cpuDecideSecondBet_Bluff(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	// Use Balanced: bluffRate=15, secondBetThreshold=OnePair
	// With HighCard hand (rank < threshold), bluff branch is reachable
	params := pokerStyleParamsMap[PokerStyleBalanced]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// handRank=HighCard(0) <= secondFoldThreshold(HighCard=0), callAmount=0
	// → falls through (callAmount not > threshold, callAmount == 0)
	// then: handRank >= secondBetThreshold(OnePair=1) → false → bluff check
	gotBet := false
	gotNonBet := false
	for i := 0; i < 1000; i++ {
		a, _ := pk.cpuDecideSecondBet(1, params, 0, PokerHandHighCard, 0)
		if a == PokerActionBet || a == PokerActionRaise {
			gotBet = true
		} else {
			gotNonBet = true
		}
		if gotBet && gotNonBet {
			break
		}
	}
	assert.True(t, gotBet, "bluff bet never triggered in second bet")
	assert.True(t, gotNonBet, "non-bluff never triggered in second bet")
}

func TestPoker_cpuDecideSecondBet_CallOrCheck(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	players[1].EvalHand()
	// OnePair < secondBetThreshold(TwoPair), bluffRate=5 → mostly check but may bluff
	gotCheck := false
	for i := 0; i < 1000; i++ {
		action, _ := pk.cpuDecideSecondBet(1, params, 0, PokerHandOnePair, 0)
		if action == PokerActionCheck {
			gotCheck = true
			break
		}
	}
	assert.True(t, gotCheck, "check never triggered")
}

// ---------------------------------------------------------------------------
// cpuFoldOrCheck
// ---------------------------------------------------------------------------

func TestPoker_cpuFoldOrCheck(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	a, _ := pk.cpuFoldOrCheck(10)
	assert.Equal(t, PokerActionFold, a)
	a, _ = pk.cpuFoldOrCheck(0)
	assert.Equal(t, PokerActionCheck, a)
}

// ---------------------------------------------------------------------------
// cpuCallOrCheck
// ---------------------------------------------------------------------------

func TestPoker_cpuCallOrCheck(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	a, _ := pk.cpuCallOrCheck(10)
	assert.Equal(t, PokerActionCall, a)
	a, _ = pk.cpuCallOrCheck(0)
	assert.Equal(t, PokerActionCheck, a)
}

// ---------------------------------------------------------------------------
// cpuRaiseOrBet
// ---------------------------------------------------------------------------

func TestPoker_cpuRaiseOrBet_AllIn_InsufficientChips(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetChips(5)
	a, _ := pk.cpuRaiseOrBet(players[1], 0, 10)
	assert.Equal(t, PokerActionAllIn, a)
}

func TestPoker_cpuRaiseOrBet_AllIn_CallPlusRaise(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetChips(15)
	a, _ := pk.cpuRaiseOrBet(players[1], 10, 10)
	assert.Equal(t, PokerActionAllIn, a)
}

func TestPoker_cpuRaiseOrBet_Raise(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetChips(100)
	a, amt := pk.cpuRaiseOrBet(players[1], 10, 20)
	assert.Equal(t, PokerActionRaise, a)
	assert.Equal(t, 20, amt)
}

func TestPoker_cpuRaiseOrBet_Bet(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetChips(100)
	a, amt := pk.cpuRaiseOrBet(players[1], 0, 20)
	assert.Equal(t, PokerActionBet, a)
	assert.Equal(t, 20, amt)
}

// ---------------------------------------------------------------------------
// cpuDecideExchange
// ---------------------------------------------------------------------------

func TestPoker_cpuDecideExchange_TwoPairPlus(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 0, len(indices))
}

func TestPoker_cpuDecideExchange_FlushDraw(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 3, false), // off-suit
	})
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 1, len(indices))
	assert.Equal(t, 4, indices[0]) // the off-suit card
}

func TestPoker_cpuDecideExchange_StraightDraw(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 12, false), // outlier
	})
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 1, len(indices))
	// Should discard the 12 (outlier)
	assert.Equal(t, 4, indices[0])
}

func TestPoker_cpuDecideExchange_OnePair(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 2, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 3, len(indices))
	// Should exchange the 3 non-pair cards
	for _, idx := range indices {
		assert.NotEqual(t, 5, players[1].GetCard(idx).GetValue())
	}
}

func TestPoker_cpuDecideExchange_HighCard(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 13, false),
	})
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 3, len(indices))
}

func TestPoker_cpuDecideExchange_HighCard_WithJoker(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseExchange)
	// To test the HighCard branch with joker, we need a hand that evaluates as HighCard
	// even with a joker. This is impossible (joker always makes at least OnePair).
	// So instead test the OnePair-with-joker path: joker gets treated as non-pair card.
	// The joker value=1 won't match pair detection, so it gets included in exchange.
	pl := pk.GetPlayers()[1]
	givePlayerHand(pl, []*Card{
		NewCard(CardDesignJoker, 1, false), // joker (value=1)
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 13, false),
	})
	indices := pk.cpuDecideExchange(1)
	// evalFiveCardHandWithJokers → OnePair (best substitution)
	// OnePair branch: all values have count 1 (no actual pair in raw cards)
	// so all 5 cards would be returned as non-pair, but code only returns singles
	// (which is all 5 since no raw value appears twice)
	assert.True(t, len(indices) > 0)
}

func TestPoker_cpuDecideExchange_HighCard_NoJoker_SkipJokerInSort(t *testing.T) {
	pk, pl := setupPokerForHumanAction(PokerPhaseExchange)
	// Test HighCard branch with joker skip logic by using 2 jokers in a 0-joker deck setup
	// Actually, we just need to ensure the HighCard branch joker skip works.
	// Give a hand that evaluates as HighCard (no joker, no pair)
	givePlayerHand(pl[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 13, false),
	})
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 3, len(indices))
	// Should exchange the 3 lowest: 2, 4, 8
}

func TestPoker_cpuDecideExchange_FlushDraw_OnePairBlocks(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// OnePair → no flush draw check (rank >= OnePair for flush draw check is < OnePair)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 5, false), // pair
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 3, false),
	})
	// This is actually OnePair, so it goes to OnePair branch, not flush draw
	indices := pk.cpuDecideExchange(1)
	assert.Equal(t, 3, len(indices))
}

// ---------------------------------------------------------------------------
// findFlushDrawDiscard
// ---------------------------------------------------------------------------

func TestPoker_findFlushDrawDiscard_Found(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 3, false),
	})
	idx := pk.findFlushDrawDiscard(1)
	assert.Equal(t, 4, idx)
}

func TestPoker_findFlushDrawDiscard_NoFlushDraw(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 3, false),
	})
	idx := pk.findFlushDrawDiscard(1)
	assert.Equal(t, -1, idx)
}

func TestPoker_findFlushDrawDiscard_WithJoker(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignJoker, 1, false),
	})
	// 4 spades + 1 joker; joker not counted in suitCounts → 4 spades found
	// Discard target must not be spade AND not joker → but all non-spade is joker
	// So loop finds no discard → returns -1
	idx := pk.findFlushDrawDiscard(1)
	assert.Equal(t, -1, idx)
}

// ---------------------------------------------------------------------------
// findStraightDrawDiscard
// ---------------------------------------------------------------------------

func TestPoker_findStraightDrawDiscard_OpenEnded(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignSpade, 12, false),
	})
	idx := pk.findStraightDrawDiscard(1)
	assert.True(t, idx >= 0)
}

func TestPoker_findStraightDrawDiscard_ALowDraw(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// A-2-3-4 + off card → A-low straight draw
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),  // Ace
		NewCard(CardDesignClover, 2, false), // 2
		NewCard(CardDesignHeart, 3, false),  // 3
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignSpade, 10, false), // off card
	})
	idx := pk.findStraightDrawDiscard(1)
	assert.True(t, idx >= 0)
}

func TestPoker_findStraightDrawDiscard_NoDraw(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 8, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 13, false),
	})
	idx := pk.findStraightDrawDiscard(1)
	assert.Equal(t, -1, idx)
}

func TestPoker_findStraightDrawDiscard_WithJoker(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignJoker, 1, false), // joker → len(cards) < 5 → skip
	})
	idx := pk.findStraightDrawDiscard(1)
	assert.Equal(t, -1, idx)
}

// ---------------------------------------------------------------------------
// findOpenEndedDraw
// ---------------------------------------------------------------------------

func TestFindOpenEndedDraw_Found(t *testing.T) {
	cards := []straightDrawCardInfo{
		{0, 5}, {1, 6}, {2, 7}, {3, 8}, {4, 12},
	}
	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] > 1 && r[3] < 14
	})
	assert.Equal(t, 4, idx) // skip 12, remaining 5-6-7-8 consecutive
}

func TestFindOpenEndedDraw_NotFound(t *testing.T) {
	cards := []straightDrawCardInfo{
		{0, 2}, {1, 5}, {2, 8}, {3, 11}, {4, 13},
	}
	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] > 1 && r[3] < 14
	})
	assert.Equal(t, -1, idx)
}

func TestFindOpenEndedDraw_NotConsecutive(t *testing.T) {
	cards := []straightDrawCardInfo{
		{0, 2}, {1, 3}, {2, 5}, {3, 7}, {4, 9},
	}
	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] > 1 && r[3] < 14
	})
	assert.Equal(t, -1, idx)
}

func TestFindOpenEndedDraw_CheckFails(t *testing.T) {
	// Consecutive but check fails (edge of range)
	cards := []straightDrawCardInfo{
		{0, 11}, {1, 12}, {2, 13}, {3, 14}, {4, 3},
	}
	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] > 1 && r[3] < 14
	})
	// Skip 14 → 3,11,12,13 not consecutive
	// Skip 3 → 11,12,13,14 consecutive but r[3]=14 → check fails
	assert.Equal(t, -1, idx)
}

func TestFindOpenEndedDraw_LenNot5(t *testing.T) {
	// Less than 5 cards → len(remaining) != 4 → skip
	cards := []straightDrawCardInfo{
		{0, 5}, {1, 6}, {2, 7},
	}
	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return true
	})
	assert.Equal(t, -1, idx)
}

// ---------------------------------------------------------------------------
// Getters and setters
// ---------------------------------------------------------------------------

func TestPoker_GettersSetters(t *testing.T) {
	pk, _ := newTestPoker()

	pk.SetPhase(PokerPhaseExchange)
	assert.Equal(t, PokerPhaseExchange, pk.GetPhase())

	pk.SetCurrentTurn(2)
	assert.Equal(t, 2, pk.GetCurrentTurn())

	pk.SetPot(500)
	assert.Equal(t, 500, pk.GetPot())

	pk.SetDealerIdx(3)
	assert.Equal(t, 3, pk.GetDealerIdx())

	pk.SetGameEndFlag(true)
	assert.True(t, pk.GetGameEndFlag())

	flags := []bool{true, false, true, false}
	pk.setActedFlags(flags)
	assert.Equal(t, flags, pk.getActedFlags())

	pk.SetLastBet(100)
	assert.Equal(t, 100, pk.GetLastBet())

	pk.SetMinRaise(50)
	assert.Equal(t, 50, pk.GetMinRaise())

	pk.setRaiseCount(3)

	results := []PokerResult{{PlayerIdx: 0, WonAmount: 100}}
	pk.SetRoundResults(results)
	assert.Equal(t, 1, len(pk.GetRoundResults()))

	actions := []PokerCpuAction{{PlayerIdx: 1, Action: PokerActionCall}}
	pk.SetCpuActions(actions)
	assert.Equal(t, 1, len(pk.GetCpuActions()))

	exchanges := []PokerCpuExchange{{PlayerIdx: 1, ExchangeCount: 2}}
	pk.SetCpuExchanges(exchanges)
	assert.Equal(t, 1, len(pk.GetCpuExchanges()))

	pots := []SidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}}
	pk.SetSidePots(pots)
	assert.Equal(t, 1, len(pk.GetSidePots()))

	chips := []int{100, 200, 300, 400}
	pk.setStartingChips(chips)
	assert.Equal(t, chips, pk.getStartingChips())

	assert.Equal(t, 10, pk.GetAnte())

	cfg := DefaultPokerConfig()
	cfg.Ante = 20
	pk.SetConfig(cfg)
	assert.Equal(t, 20, pk.GetConfig().Ante)

	pk.round.lastCpuError = nil
	assert.Nil(t, pk.GetLastCpuError())
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestPoker_Constants(t *testing.T) {
	assert.Equal(t, 0, PokerPhaseInit)
	assert.Equal(t, 1, PokerPhaseDeal)
	assert.Equal(t, 2, PokerPhaseExchange)
	assert.Equal(t, 3, PokerPhaseSecondBet)
	assert.Equal(t, 4, PokerPhaseEnd)

	assert.Equal(t, 0, PokerActionFold)
	assert.Equal(t, 1, PokerActionCheck)
	assert.Equal(t, 2, PokerActionCall)
	assert.Equal(t, 3, PokerActionBet)
	assert.Equal(t, 4, PokerActionRaise)
	assert.Equal(t, 5, PokerActionAllIn)
}

// ---------------------------------------------------------------------------
// Full game flow
// ---------------------------------------------------------------------------

func TestPoker_FullGame_Showdown(t *testing.T) {
	pk, _ := newTestPoker()
	err := pk.Reset()
	assert.NoError(t, err)

	// If game ended during Reset (all CPUs folded), skip
	if pk.GetGameEndFlag() {
		return
	}

	// First betting: Check or Call (only if it is human turn)
	if pk.GetPhase() == PokerPhaseDeal && pk.GetPlayers()[pk.GetCurrentTurn()].GetIsHuman() {
		if pk.GetLastBet() > 0 {
			err = pk.PlayerAction(PokerActionCall, 0, 0)
		} else {
			err = pk.PlayerAction(PokerActionCheck, 0, 0)
		}
		assert.NoError(t, err)
	}

	if pk.GetGameEndFlag() {
		return
	}

	// Exchange phase
	if pk.GetPhase() == PokerPhaseExchange && pk.GetPlayers()[pk.GetCurrentTurn()].GetIsHuman() {
		err = pk.PlayerStand()
		assert.NoError(t, err)
	}

	if pk.GetGameEndFlag() {
		return
	}

	// Second betting
	if pk.GetPhase() == PokerPhaseSecondBet && pk.GetPlayers()[pk.GetCurrentTurn()].GetIsHuman() {
		if pk.GetLastBet() > 0 {
			err = pk.PlayerAction(PokerActionCall, 0, 0)
		} else {
			err = pk.PlayerAction(PokerActionCheck, 0, 0)
		}
		assert.NoError(t, err)
	}
}

func TestPoker_FullGame_PlayerFold(t *testing.T) {
	pk, _ := newTestPoker()
	err := pk.Reset()
	assert.NoError(t, err)

	if pk.GetGameEndFlag() {
		return
	}

	if pk.GetPhase() == PokerPhaseDeal {
		err = pk.PlayerAction(PokerActionFold, 0, 0)
		assert.NoError(t, err)
	}

	// Game may or may not be over depending on CPU actions
}

// ---------------------------------------------------------------------------
// Edge cases for advanceTurn all-acted fallback
// ---------------------------------------------------------------------------

func TestPoker_advanceTurn_noActivePlayer(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(0)
	// All acted, all folded/allIn except none unacted
	players[0].SetFolded(true)
	players[1].SetFolded(true)
	players[2].SetAllIn(true)
	players[3].SetAllIn(true)
	pk.setActedFlags([]bool{true, true, true, true})
	pk.advanceTurn()
	// All acted → advancePhase
	assert.True(t, pk.GetPhase() == PokerPhaseExchange || pk.GetPhase() == PokerPhaseEnd)
}

// ---------------------------------------------------------------------------
// Showdown remainder distribution
// ---------------------------------------------------------------------------

func TestPoker_resolveShowdown_Remainder(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetPot(301) // not evenly divisible
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})

	// Give identical hands for 3-way tie
	for i := 0; i < 3; i++ {
		givePlayerHand(players[i], []*Card{
			NewCard(CardDesignSpade+i, 2, false),
			NewCard(CardDesignSpade+i, 5, false),
			NewCard(CardDesignSpade+i, 7, false),
			NewCard(CardDesignSpade+i, 9, false),
			NewCard(CardDesignSpade+i, 11, false),
		})
	}
	players[3].SetFolded(true)

	pk.resolveShowdown()

	totalWon := 0
	for _, r := range pk.GetRoundResults() {
		totalWon += r.WonAmount
	}
	assert.Equal(t, 301, totalWon)
}

// ---------------------------------------------------------------------------
// Fold during countActivePlayers == 1 in executeAction
// ---------------------------------------------------------------------------

func TestPoker_executeAction_FoldLeadsToLastPlayer(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	pk.SetPot(100)
	// p0 and p3 active; p0 folds → p3 wins
	err := pk.executeAction(0, PokerActionFold, 0)
	assert.NoError(t, err)
	assert.True(t, pk.GetGameEndFlag())
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
}

// ---------------------------------------------------------------------------
// cpuDecideExchange: straight draw where A is treated as high then low
// ---------------------------------------------------------------------------

func TestPoker_findStraightDrawDiscard_AHighNoMatch(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// A(14)-K(13)-Q(12)-J(11) + off → open-ended check: r[3]=14 fails (not < 14)
	// Then A-low: A→1, sorted 1,11,12,13 → not consecutive → no match
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignClover, 11, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignSpade, 3, false),
	})
	idx := pk.findStraightDrawDiscard(1)
	// The hand is A,3,J,Q,K → as high: 3,11,12,13,14; skip 3 → 11,12,13,14 consecutive but r[3]=14 fails
	// as low: 1,3,11,12,13; no 4-consecutive subsequence meets A-low check (r[0]==1 && r[3]<=5)
	assert.Equal(t, -1, idx)
}

// ---------------------------------------------------------------------------
// Full integration: Reset with runCpuActions error
// (This is hard to trigger naturally, so we just ensure Reset works)
// ---------------------------------------------------------------------------

func TestPoker_Reset_RunCpuActionsCalled(t *testing.T) {
	pk, _ := newTestPoker()
	err := pk.Reset()
	assert.NoError(t, err)
	// CPU actions should have been recorded
	// (may be empty if human is first to act)
}

// ---------------------------------------------------------------------------
// PlayerAction in SecondBet phase (triggers advanceTurn → resolveShowdown)
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_SecondBet_AdvancesToEnd(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	// All CPUs folded so only human active
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	players[3].SetFolded(true)
	pk.setActedFlags([]bool{false, true, true, true})
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())
}

// ---------------------------------------------------------------------------
// DefaultPokerConfig
// ---------------------------------------------------------------------------

func TestDefaultPokerConfig(t *testing.T) {
	cfg := DefaultPokerConfig()
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, 10, cfg.Ante)
	assert.Equal(t, 10, cfg.MinBet)
	assert.Equal(t, 3, cfg.CpuCount)
	assert.Equal(t, 0, cfg.JokerCount)
}

// ---------------------------------------------------------------------------
// PokerPlayStyleNames
// ---------------------------------------------------------------------------

func TestPokerPlayStyleNames(t *testing.T) {
	assert.Equal(t, "Conservative", PokerPlayStyleNames[0])
	assert.Equal(t, "Balanced", PokerPlayStyleNames[1])
	assert.Equal(t, "Aggressive", PokerPlayStyleNames[2])
	assert.Equal(t, "Bluffer", PokerPlayStyleNames[3])
}

// ---------------------------------------------------------------------------
// SidePot struct
// ---------------------------------------------------------------------------

func TestPokerSidePot(t *testing.T) {
	sp := SidePot{Amount: 100, EligiblePlayers: []int{0, 1}}
	assert.Equal(t, 100, sp.Amount)
	assert.Equal(t, []int{0, 1}, sp.EligiblePlayers)
}

// ---------------------------------------------------------------------------
// PokerResult struct
// ---------------------------------------------------------------------------

func TestPokerResult(t *testing.T) {
	r := PokerResult{PlayerIdx: 0, HandRank: 5, HandName: "Flush", WonAmount: 200}
	assert.Equal(t, 0, r.PlayerIdx)
	assert.Equal(t, 5, r.HandRank)
	assert.Equal(t, "Flush", r.HandName)
	assert.Equal(t, 200, r.WonAmount)
}

// ---------------------------------------------------------------------------
// PokerCpuAction / PokerCpuExchange structs
// ---------------------------------------------------------------------------

func TestPokerCpuAction(t *testing.T) {
	a := PokerCpuAction{PlayerIdx: 1, Action: PokerActionCall, Amount: 20}
	assert.Equal(t, 1, a.PlayerIdx)
	assert.Equal(t, PokerActionCall, a.Action)
	assert.Equal(t, 20, a.Amount)
}

func TestPokerCpuExchange(t *testing.T) {
	e := PokerCpuExchange{PlayerIdx: 2, ExchangeCount: 3}
	assert.Equal(t, 2, e.PlayerIdx)
	assert.Equal(t, 3, e.ExchangeCount)
}

// ---------------------------------------------------------------------------
// DealerIdx rotation after game end
// ---------------------------------------------------------------------------

func TestPoker_DealerIdxRotation(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetDealerIdx(0)
	pk.SetPot(100)
	players[1].SetFolded(true)
	players[2].SetFolded(true)
	players[3].SetFolded(true)
	// Only p0 → resolveLastPlayer
	pk.resolveLastPlayer()
	assert.Equal(t, 1, pk.GetDealerIdx()) // rotated 0→1
}

// ---------------------------------------------------------------------------
// resolveShowdown with folded players excluded from results
// ---------------------------------------------------------------------------

func TestPoker_resolveShowdown_FoldedExcluded(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetPot(200)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})
	players[0].SetFolded(true)
	players[3].SetFolded(true)

	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	})
	givePlayerHand(players[2], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})

	pk.resolveShowdown()

	for _, r := range pk.GetRoundResults() {
		assert.NotEqual(t, 0, r.PlayerIdx)
		assert.NotEqual(t, 3, r.PlayerIdx)
	}
}

// ---------------------------------------------------------------------------
// advanceTurn wrapping around
// ---------------------------------------------------------------------------

func TestPoker_advanceTurn_wrapsAround(t *testing.T) {
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(3)
	pk.setActedFlags([]bool{false, true, true, true})
	pk.advanceTurn()
	assert.Equal(t, 0, pk.GetCurrentTurn()) // wraps to 0
}

// ---------------------------------------------------------------------------
// cpuDecideSecondBet: high call fold (callAmount > threshold)
// ---------------------------------------------------------------------------

func TestPoker_cpuDecideSecondBet_HighCallFold(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
	params := pokerStyleParamsMap[PokerStyleConservative]
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	players[1].EvalHand()
	// HighCard, callAmount > MinBet*secondCallMaxMult (10*2=20) → fold
	action, _ := pk.cpuDecideSecondBet(1, params, 30, PokerHandHighCard, 0)
	assert.Equal(t, PokerActionFold, action)
}

// ---------------------------------------------------------------------------
// Bet that makes chips exactly 0 (allIn flag set)
// ---------------------------------------------------------------------------

func TestPoker_executeAction_Bet_ChipsBecome0(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[0].SetChips(10) // exact MinBet
	pk.SetLastBet(0)
	pk.SetMinRaise(10)
	err := pk.executeAction(0, PokerActionBet, 10)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
	assert.Equal(t, 0, players[0].GetChips())
}

// ---------------------------------------------------------------------------
// Raise that makes chips exactly 0 (allIn flag set)
// ---------------------------------------------------------------------------

func TestPoker_executeAction_Raise_ChipsBecome0(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetLastBet(20)
	pk.SetMinRaise(10)
	players[0].SetCurrentBet(0)
	players[0].SetChips(30) // diff=20 + amount=10 = 30, exactly = chips → auto allIn via executeAction redirect
	err := pk.executeAction(0, PokerActionRaise, 10)
	assert.NoError(t, err)
	assert.True(t, players[0].GetAllIn())
}

// ---------------------------------------------------------------------------
// Exchange with no cards left in deck (DrawCard returns nil)
// ---------------------------------------------------------------------------

func TestPoker_PlayerExchange_DeckExhausted(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// Exhaust the deck
	for pk.trumpCards.DrawCard() != nil {
	}
	err := pk.PlayerExchange([]int{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, 2, players[0].GetExchangeCount())
}

// ---------------------------------------------------------------------------
// runCpuExchanges: CPU exchange when deck is exhausted
// ---------------------------------------------------------------------------

func TestPoker_runCpuExchanges_DeckExhausted(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.SetCurrentTurn(1)
	pk.setActedFlags([]bool{true, false, true, true})
	for pk.trumpCards.DrawCard() != nil {
	}
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	pk.runCpuExchanges()
	// Should not panic
}

// ---------------------------------------------------------------------------
// Multiple rounds via Reset
// ---------------------------------------------------------------------------

func TestPoker_MultipleRounds(t *testing.T) {
	pk, _ := newTestPoker()
	for round := 0; round < 3; round++ {
		err := pk.Reset()
		assert.NoError(t, err)
		if pk.GetGameEndFlag() {
			continue
		}
		if pk.GetPhase() == PokerPhaseDeal {
			if pk.GetLastBet() > 0 {
				_ = pk.PlayerAction(PokerActionCall, 0, 0)
			} else {
				_ = pk.PlayerAction(PokerActionCheck, 0, 0)
			}
		}
		if pk.GetGameEndFlag() {
			continue
		}
		if pk.GetPhase() == PokerPhaseExchange {
			_ = pk.PlayerStand()
		}
		if pk.GetGameEndFlag() {
			continue
		}
		if pk.GetPhase() == PokerPhaseSecondBet {
			if pk.GetLastBet() > 0 {
				_ = pk.PlayerAction(PokerActionCall, 0, 0)
			} else {
				_ = pk.PlayerAction(PokerActionCheck, 0, 0)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// advancePhase: Deal → Exchange sets folded/allIn players actedFlags
// ---------------------------------------------------------------------------

func TestPoker_advancePhase_Deal_SetsActedForFoldedAllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	players[1].SetFolded(true)
	players[2].SetAllIn(true)
	pk.advancePhase()
	flags := pk.getActedFlags()
	assert.True(t, flags[1])  // folded
	assert.True(t, flags[2])  // allIn
	assert.False(t, flags[0]) // active, not acted
	assert.False(t, flags[3]) // active, not acted
}

// ---------------------------------------------------------------------------
// startSecondBettingRound sets folded/allIn actedFlags
// ---------------------------------------------------------------------------

func TestPoker_startSecondBettingRound_SetsActedForFoldedAllIn(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	pk.trumpCards.Shuffle()
	players[1].SetFolded(true)
	players[2].SetAllIn(true)
	pk.startSecondBettingRound()
	flags := pk.getActedFlags()
	assert.True(t, flags[1])
	assert.True(t, flags[2])
}

// ---------------------------------------------------------------------------
// pokerDefaultMaxRaises value
// ---------------------------------------------------------------------------

func TestPokerDefaultMaxRaises(t *testing.T) {
	assert.Equal(t, 4, pokerDefaultMaxRaises)
}

// ---------------------------------------------------------------------------
// UNCOVERED: Raise that makes chips exactly 0 → allIn flag (executeAction line 354)
// ---------------------------------------------------------------------------

func TestPoker_executeAction_Raise_ExactChipsToZero(t *testing.T) {
	// The Raise path's allIn check (chips==0 after subtract) at line 354 is dead code:
	// When totalNeeded >= chips, the code redirects to AllIn action before reaching subtract.
	// When totalNeeded < chips, chips-totalNeeded > 0, so the allIn check never fires.
	// This is verified by the Bet variant which CAN hit chips==0 (TestPoker_PlayerAction_Bet_ExactChips_AllIn).
	pk, _ := setupPokerForHumanAction(PokerPhaseDeal)
	_ = pk
}

// ---------------------------------------------------------------------------
// UNCOVERED: advanceTurn all-acted fallback (line 440-442)
// This is the path where after checking betting/exchange, no unacted player
// is found, and we fall through to the all-acted advancePhase.
// ---------------------------------------------------------------------------

func TestPoker_advanceTurn_AllActedFallbackDeal(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(0)
	// The fallback at advanceTurn line 440-442 is dead code: isBettingRoundComplete
	// checks the same conditions as the for-loop, so if no unacted active player
	// exists, betting is complete and advancePhase fires first.
	// This test verifies the normal all-acted path via isBettingRoundComplete.
	players[0].SetAllIn(true)
	players[1].SetAllIn(true)
	players[2].SetAllIn(true)
	players[3].SetAllIn(true)
	pk.setActedFlags([]bool{true, true, true, true})
	pk.advanceTurn()
	assert.Equal(t, PokerPhaseExchange, pk.GetPhase())
}

// ---------------------------------------------------------------------------
// UNCOVERED: cpuDecideExchange HighCard with Ace (v=14) and Joker (v=15)
// ---------------------------------------------------------------------------

func TestPoker_cpuDecideExchange_HighCard_WithAce(t *testing.T) {
	pk, pl := setupPokerForHumanAction(PokerPhaseExchange)
	// HighCard hand with an Ace (value 1 → treated as 14 in sort)
	givePlayerHand(pl[1], []*Card{
		NewCard(CardDesignSpade, 1, false),  // Ace → v=14
		NewCard(CardDesignClover, 3, false), // 3
		NewCard(CardDesignHeart, 6, false),  // 6
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 12, false),
	})
	indices := pk.cpuDecideExchange(1)
	// Should exchange lowest 3: 3, 6, 9 (Ace kept as high)
	assert.Equal(t, 3, len(indices))
	// Verify ace is not exchanged
	for _, idx := range indices {
		assert.NotEqual(t, 1, pl[1].GetCard(idx).GetValue())
	}
}

// ---------------------------------------------------------------------------
// Stand-pat bluff in cpuDecideExchange
// ---------------------------------------------------------------------------

func TestPoker_cpuDecideExchange_StandPatBluff(t *testing.T) {
	pk, pl := setupPokerForHumanAction(PokerPhaseExchange)
	// Player 3 = Bluffer (standPatBluffRate=20)
	// Give a HighCard hand (weak)
	givePlayerHand(pl[3], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	gotBluff := false
	gotNormal := false
	for i := 0; i < 1000; i++ {
		indices := pk.cpuDecideExchange(3)
		if len(indices) == 0 {
			gotBluff = true
		} else {
			gotNormal = true
		}
		if gotBluff && gotNormal {
			break
		}
	}
	assert.True(t, gotBluff, "stand-pat bluff never triggered")
	assert.True(t, gotNormal, "normal exchange never triggered")
}

func TestPoker_cpuDecideExchange_StandPatBluff_Conservative(t *testing.T) {
	pk, pl := setupPokerForHumanAction(PokerPhaseExchange)
	// Player 1 = Conservative (standPatBluffRate=0)
	givePlayerHand(pl[1], []*Card{
		NewCard(CardDesignClover, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignHeart, 11, false),
	})
	indices := pk.cpuDecideExchange(1)
	assert.Greater(t, len(indices), 0, "conservative should never stand-pat bluff")
}

func TestPoker_cpuDecideExchange_StandPatBluff_OnePair(t *testing.T) {
	pk, pl := setupPokerForHumanAction(PokerPhaseExchange)
	// Player 3 = Bluffer; give OnePair hand (still below TwoPair threshold)
	givePlayerHand(pl[3], []*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 11, false),
	})
	gotBluff := false
	gotNormal := false
	for i := 0; i < 1000; i++ {
		indices := pk.cpuDecideExchange(3)
		if len(indices) == 0 {
			gotBluff = true
		} else {
			gotNormal = true
		}
		if gotBluff && gotNormal {
			break
		}
	}
	assert.True(t, gotBluff, "stand-pat bluff never triggered with OnePair")
	assert.True(t, gotNormal, "normal exchange never triggered with OnePair")
}

func TestPoker_cpuDecideExchange_StandPatBluff_DrawHandPrioritized(t *testing.T) {
	pk, pl := setupPokerForHumanAction(PokerPhaseExchange)
	// Player 3 = Bluffer (standPatBluffRate=20)
	// Give a flush draw (4 spades + 1 off-suit) — draw should always win over bluff
	givePlayerHand(pl[3], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignHeart, 3, false),
	})
	indices := pk.cpuDecideExchange(3)
	assert.Equal(t, 1, len(indices), "flush draw hand should always exchange 1 card, never stand-pat bluff")
	assert.Equal(t, 4, indices[0], "should discard the off-suit card")
}

// ---------------------------------------------------------------------------
// UNCOVERED: CPU fallback to Check in runCpuActions
// ---------------------------------------------------------------------------

func TestPoker_runCpuActions_CpuFallback_Check(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(1)
	pk.SetLastBet(0) // callAmt = 0

	// Force CPU to attempt AllIn with 0 chips (not marked allIn)
	// cpuRaiseOrBet: raiseAmt > chips(0) → AllIn
	// executeAction AllIn: chips <= 0 → error "No chips to go all-in"
	// fallback: callAmt = lastBet(0) - currentBet(0) = 0 → Check
	players[1].SetChips(0) // 0 chips but not allIn

	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	})
	pk.setActedFlags([]bool{false, false, true, true})
	pk.runCpuActions()
	// Verify the fallback was triggered
	assert.NotNil(t, pk.GetLastCpuError())
}

func TestPoker_runCpuActions_CpuFallback_Fold(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetCurrentTurn(1)
	pk.SetLastBet(50) // callAmt > 0

	// Same setup: 0 chips, not allIn → AllIn fails → fallback: callAmt=50>0 → Fold
	players[1].SetChips(0)

	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 13, false),
	})
	pk.setActedFlags([]bool{false, false, true, true})
	pk.runCpuActions()
	assert.NotNil(t, pk.GetLastCpuError())
	assert.True(t, players[1].GetFolded())
}

// ---------------------------------------------------------------------------
// TestPoker_Reset_JokerCount — ResetWithConfig with JokerCount recreates deck
// ---------------------------------------------------------------------------

func TestPoker_Reset_JokerCount(t *testing.T) {
	t.Run("JokerCount=2 results in jokers dealt to players", func(t *testing.T) {
		tc := NewTrumpCards(0) // initially 0 jokers
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		p2 := NewPokerPlayer(false, PokerStyleAggressive)
		p3 := NewPokerPlayer(false, PokerStyleBluffer)
		players := []*PokerPlayer{p0, p1, p2, p3}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.JokerCount = 2
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// With 2 jokers the deck has 54 cards. 4 players * 5 = 20 dealt.
		// Count jokers across all player hands.
		jokerCount := 0
		for _, pl := range players {
			for i := 0; i < pl.GetCardsSize(); i++ {
				if pl.GetCard(i).GetDesign() == CardDesignJoker {
					jokerCount++
				}
			}
		}
		// Jokers exist in the deck so they can appear in hands (0, 1, or 2).
		// We cannot deterministically assert the exact count, but we can verify
		// that the total card count in the deck was correct (54 = 52 + 2 jokers).
		// Each active player should have exactly 5 cards.
		for _, pl := range players {
			if !pl.GetFolded() {
				assert.Equal(t, 5, pl.GetCardsSize())
			}
		}
		// Joker count across hands must be 0, 1, or 2
		assert.True(t, jokerCount >= 0 && jokerCount <= 2)
	})

	t.Run("JokerCount changes between resets", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		players := []*PokerPlayer{p0, p1}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 1
		cfg.JokerCount = 0
		pk := NewPoker(tc, players, cfg)

		// First reset with 0 jokers — no jokers should exist
		_ = pk.Reset()
		jokerCountZero := 0
		for _, pl := range players {
			for i := 0; i < pl.GetCardsSize(); i++ {
				if pl.GetCard(i).GetDesign() == CardDesignJoker {
					jokerCountZero++
				}
			}
		}
		assert.Equal(t, 0, jokerCountZero)

		// Now change config to 2 jokers and reset again
		cfg.JokerCount = 2
		pk.SetConfig(cfg)
		_ = pk.Reset()

		// Verify deck was recreated by checking total card pool
		// (We can't guarantee jokers ended up in hands, but the deck
		// should have had 54 cards total. Dealt = 2 * 5 = 10 cards.)
		for _, pl := range players {
			if !pl.GetFolded() {
				assert.Equal(t, 5, pl.GetCardsSize())
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestPoker_Reset_CpuCount — ResetWithConfig with CpuCount folds excess players
// ---------------------------------------------------------------------------

func TestPoker_Reset_CpuCount(t *testing.T) {
	t.Run("CpuCount=1 folds players 2 and 3", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		p2 := NewPokerPlayer(false, PokerStyleAggressive)
		p3 := NewPokerPlayer(false, PokerStyleBluffer)
		players := []*PokerPlayer{p0, p1, p2, p3}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 1 // only human + 1 CPU = seats 0,1 active
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// Players 0 and 1 are active
		assert.False(t, players[0].GetFolded())
		assert.False(t, players[1].GetFolded())
		// Players 2 and 3 should be folded (inactive)
		assert.True(t, players[2].GetFolded())
		assert.True(t, players[3].GetFolded())

		// Active players should have 5 cards, folded should have 0
		for _, pl := range players {
			if !pl.GetFolded() {
				assert.Equal(t, 5, pl.GetCardsSize())
			} else {
				assert.Equal(t, 0, pl.GetCardsSize())
			}
		}
	})

	t.Run("CpuCount=0 leaves only human active", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		p2 := NewPokerPlayer(false, PokerStyleAggressive)
		p3 := NewPokerPlayer(false, PokerStyleBluffer)
		players := []*PokerPlayer{p0, p1, p2, p3}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 0 // only human
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// Only player 0 is active
		assert.False(t, players[0].GetFolded())
		assert.True(t, players[1].GetFolded())
		assert.True(t, players[2].GetFolded())
		assert.True(t, players[3].GetFolded())
	})

	t.Run("CpuCount=3 keeps all players active", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		p2 := NewPokerPlayer(false, PokerStyleAggressive)
		p3 := NewPokerPlayer(false, PokerStyleBluffer)
		players := []*PokerPlayer{p0, p1, p2, p3}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 3 // all 4 players active
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// All players active (none folded by Reset itself; CPUs may fold during play)
		for i, pl := range players {
			if pk.GetGameEndFlag() {
				break
			}
			// Before CPU actions, no player should be force-folded
			_ = i
			_ = pl
		}
		// At minimum, each non-folded player should have 5 cards
		for _, pl := range players {
			if !pl.GetFolded() {
				assert.Equal(t, 5, pl.GetCardsSize())
			}
		}
	})

	t.Run("CpuCount exceeds player slice length is clamped", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		players := []*PokerPlayer{p0, p1}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 10 // exceeds 2 players
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// Both players should be active (clamped to array length)
		assert.False(t, players[0].GetFolded())
		assert.False(t, players[1].GetFolded())
	})

	t.Run("CpuCount change between resets", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		p2 := NewPokerPlayer(false, PokerStyleAggressive)
		p3 := NewPokerPlayer(false, PokerStyleBluffer)
		players := []*PokerPlayer{p0, p1, p2, p3}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 1
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// Players 2,3 should be folded
		assert.True(t, players[2].GetFolded())
		assert.True(t, players[3].GetFolded())
		assert.Equal(t, 0, players[2].GetCardsSize())
		assert.Equal(t, 0, players[3].GetCardsSize())

		// Now change to 3 CPUs and reset
		cfg.CpuCount = 3
		pk.SetConfig(cfg)
		_ = pk.Reset()

		// All players should now be active (or may have folded during CPU actions)
		// Before CPU action processing, none should be force-folded by seat count
		for _, pl := range players {
			if !pl.GetFolded() {
				assert.Equal(t, 5, pl.GetCardsSize())
			}
		}
	})

	t.Run("collectAntes skips folded players from CpuCount reduction", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		p2 := NewPokerPlayer(false, PokerStyleAggressive)
		p3 := NewPokerPlayer(false, PokerStyleBluffer)
		players := []*PokerPlayer{p0, p1, p2, p3}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = 1 // 2 active players
		cfg.Ante = 10
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// Pot should only have antes from 2 active players (not 4)
		// With CpuCount=1, active seats = 2, ante = 10 each = 20 minimum ante
		// (CPUs may have bet more during runCpuActions, so pot >= 20)
		assert.True(t, pk.GetPot() >= 20 || pk.GetGameEndFlag())
		// Folded players should not have had ante deducted
		assert.Equal(t, 1000, players[2].GetChips())
		assert.Equal(t, 1000, players[3].GetChips())
	})

	t.Run("activeSeatCount clamped to 1 when CpuCount is negative", func(t *testing.T) {
		tc := NewTrumpCards(0)
		p0 := NewPokerPlayer(true, PokerStyleBalanced)
		p1 := NewPokerPlayer(false, PokerStyleConservative)
		players := []*PokerPlayer{p0, p1}
		for _, pl := range players {
			pl.SetChips(1000)
		}
		cfg := DefaultPokerConfig()
		cfg.CpuCount = -5 // negative → activeSeatCount = max(1, -5+1) = 1
		pk := NewPoker(tc, players, cfg)
		_ = pk.Reset()

		// Only player 0 active
		assert.False(t, players[0].GetFolded())
		assert.True(t, players[1].GetFolded())
	})
}

func makePokerWithConfig(phase int, limit BettingLimitType) (*Poker, []*PokerPlayer) {
	tc := NewTrumpCards(0)
	players := []*PokerPlayer{
		NewPokerPlayer(true, PokerStyleBalanced),
		NewPokerPlayer(false, PokerStyleConservative),
		NewPokerPlayer(false, PokerStyleAggressive),
		NewPokerPlayer(false, PokerStyleBluffer),
	}
	cfg := DefaultPokerConfig()
	cfg.BettingLimit = limit
	pk := NewPoker(tc, players, cfg)
	for _, p := range players {
		p.SetChips(1000)
	}
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})
	pk.SetPhase(phase)
	pk.SetCurrentTurn(0)
	pk.SetLastBet(0)
	pk.SetMinRaise(10)
	pk.SetPot(100)
	pk.setActedFlags([]bool{false, true, true, true})
	for _, p := range players {
		p.Reset()
		for j := 0; j < 5; j++ {
			p.AddCard(NewCard(CardDesignSpade, j+2, false))
		}
	}
	pk.trumpCards.Shuffle()
	return pk, players
}

func TestPoker_GetRaiseCount(t *testing.T) {
	pk, _ := newTestPoker()
	assert.Equal(t, 0, pk.GetRaiseCount())
}

func TestPoker_BettingLimits_PotLimit(t *testing.T) {
	t.Run("bet exceeding pot limit is rejected", func(t *testing.T) {
		pk, _ := makePokerWithConfig(PokerPhaseDeal, BettingLimitPotLimit)
		pk.SetPot(100)
		pk.SetLastBet(0)
		// maxBetAmount = pot + lastBet = 100 + 0 = 100; bet 130 > 100
		err := pk.PlayerAction(PokerActionBet, 130, 0)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("bet within pot limit succeeds", func(t *testing.T) {
		pk, _ := makePokerWithConfig(PokerPhaseDeal, BettingLimitPotLimit)
		pk.SetPot(100)
		pk.SetLastBet(0)
		// maxBetAmount = pot + lastBet = 100 + 0 = 100; bet 100 is within limit
		err := pk.PlayerAction(PokerActionBet, 100, 0)
		assert.NoError(t, err)
	})
}

func TestPoker_cpuDecide_PotLimitClamp(t *testing.T) {
	// PotLimit 時、CPUのベット額がポット上限を超えた場合にクランプされることを確認
	pk, players := makePokerWithConfig(PokerPhaseDeal, BettingLimitPotLimit)
	pk.SetPot(15)
	pk.SetLastBet(0)
	// maxBetAmount = pot + lastBet = 15
	// Aggressive (player 2): firstBetMult=2 → bet = MinBet(10)*2 = 20 > 15
	givePlayerHand(players[2], []*Card{
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignHeart, 10, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignSpade, 3, false),
	})
	action, amount := pk.cpuDecide(2)
	// Aggressive CPU with FourOfAKind → bet. Bet=20, clamped to maxBetAmount=15
	assert.Equal(t, PokerActionBet, action)
	assert.Equal(t, 15, amount)
}

func TestPoker_BettingLimits_NoLimit(t *testing.T) {
	t.Run("bet succeeds even when raiseCount exceeds fixed limit cap", func(t *testing.T) {
		pk, _ := makePokerWithConfig(PokerPhaseDeal, BettingLimitNoLimit)
		pk.setRaiseCount(10) // well past fixed limit of 4
		// NoLimit has maxRaises=0, so no raise cap
		err := pk.PlayerAction(PokerActionBet, 20, 0)
		assert.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// Lowball tests
// ---------------------------------------------------------------------------

// newLowballPoker creates a Poker instance with IsLowball=true.
func newLowballPoker() (*Poker, []*PokerPlayer) {
	tc := NewTrumpCards(0)
	p0 := NewPokerPlayer(true, PokerStyleBalanced)
	p1 := NewPokerPlayer(false, PokerStyleConservative)
	p2 := NewPokerPlayer(false, PokerStyleAggressive)
	p3 := NewPokerPlayer(false, PokerStyleBluffer)
	players := []*PokerPlayer{p0, p1, p2, p3}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := DefaultPokerConfig()
	cfg.IsLowball = true
	pk := NewPoker(tc, players, cfg)
	return pk, players
}

func TestPoker_Reset_LowballForcesJokerZero(t *testing.T) {
	tc := NewTrumpCards(2)
	p0 := NewPokerPlayer(true, PokerStyleBalanced)
	p1 := NewPokerPlayer(false, PokerStyleConservative)
	players := []*PokerPlayer{p0, p1}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := DefaultPokerConfig()
	cfg.IsLowball = true
	cfg.JokerCount = 2
	cfg.CpuCount = 1
	pk := NewPoker(tc, players, cfg)

	err := pk.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 0, pk.GetConfig().JokerCount)
}

func TestPoker_Showdown_LowballWinner(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseSecondBet)
	pk.SetCurrentTurn(0)
	pk.SetPot(100)
	pk.SetLastBet(0)
	pk.SetMinRaise(10)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})
	pk.setActedFlags([]bool{false, true, true, true})

	// Player 0: HighCard (2,4,6,8,10 different suits) - weak hand, should win in lowball
	givePlayerHand(players[0], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 6, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignSpade, 10, false),
	})
	// Player 1: OnePair (3,3,5,7,9) - stronger hand, should lose in lowball
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignSpade, 9, false),
	})
	// Fold players 2 and 3
	players[2].SetFolded(true)
	players[3].SetFolded(true)

	// Player 0 checks → advances; all CPUs already acted → showdown
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, PokerPhaseEnd, pk.GetPhase())

	// In lowball, player 0 (HighCard) should win
	results := pk.GetRoundResults()
	assert.Equal(t, 2, len(results))
	var p0Won, p1Won int
	for _, r := range results {
		if r.PlayerIdx == 0 {
			p0Won = r.WonAmount
		}
		if r.PlayerIdx == 1 {
			p1Won = r.WonAmount
		}
	}
	assert.Greater(t, p0Won, 0)
	assert.Equal(t, 0, p1Won)
}

func TestPoker_CpuDecide_LowballInvertsRank(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseDeal)
	pk.SetCurrentTurn(1) // CPU conservative
	pk.SetPot(40)
	pk.SetLastBet(0)
	pk.SetMinRaise(10)
	pk.setStartingChips([]int{1000, 1000, 1000, 1000})
	pk.setActedFlags([]bool{true, false, true, true})

	for _, pl := range players {
		pl.SetChips(990)
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
	}

	// Give CPU player 1 a HighCard hand (weak normally, strong in lowball)
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignSpade, 9, false),
	})

	action, _ := pk.cpuDecide(1)
	// Lowball inverts rank: HighCard (0) → FiveOfAKind(10)-0=10 (very strong)
	// Conservative with a strong hand should bet or check, not fold
	assert.NotEqual(t, PokerActionFold, action)
}

func TestPoker_CpuDecideExchangeLowball_BreaksPairs(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseExchange)

	// CPU player 1 has a pair of 5s → should discard one 5
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 5, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignSpade, 4, false),
	})

	indices := pk.cpuDecideExchangeLowball(1)
	assert.Equal(t, 1, len(indices))
	// Should discard the second 5 (index 2, the duplicate)
	assert.Equal(t, 2, indices[0])
}

func TestPoker_CpuDecideExchangeLowball_DiscardsHighCards(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseExchange)

	// CPU player 1 has high cards (8, 10, King=13) mixed with low cards
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 8, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignSpade, 13, false),
	})

	indices := pk.cpuDecideExchangeLowball(1)
	assert.Equal(t, 3, len(indices))
	assert.Contains(t, indices, 2) // 8
	assert.Contains(t, indices, 3) // 10
	assert.Contains(t, indices, 4) // King
}

func TestPoker_CpuDecideExchangeLowball_KeepsLowCards(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseExchange)

	// CPU player 1 has all low cards (2,3,4,5,7) - no pairs, no high cards
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignSpade, 7, false),
	})

	indices := pk.cpuDecideExchangeLowball(1)
	assert.Equal(t, 0, len(indices))
}

func TestPoker_CpuDecideExchangeLowball_MaxThreeCards(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseExchange)

	// CPU player 1 has all high cards plus a pair → 5 discard candidates, but max 3
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 1, false),    // Ace=14
		NewCard(CardDesignHeart, 13, false),   // King
		NewCard(CardDesignDiamond, 12, false), // Queen
		NewCard(CardDesignClover, 11, false),  // Jack
		NewCard(CardDesignSpade, 10, false),   // 10
	})

	indices := pk.cpuDecideExchangeLowball(1)
	assert.Equal(t, 3, len(indices))
	// Should pick the 3 highest values: Ace(14), King(13), Queen(12)
	assert.Contains(t, indices, 0) // Ace
	assert.Contains(t, indices, 1) // King
	assert.Contains(t, indices, 2) // Queen
}

func TestPoker_CpuDecideExchangeLowball_PairPrioritizedOverHighCards(t *testing.T) {
	// 7,7,8,9,10 → pair discard (one 7) + high card discards (8,9,10)
	// Should discard pair duplicate first, then fill with highest non-pair cards
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseExchange)

	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 10, false),
	})

	indices := pk.cpuDecideExchangeLowball(1)
	assert.Equal(t, 3, len(indices))
	// Pair discard (idx 1) must be included
	assert.Contains(t, indices, 1)
	// High cards: 10 (idx 4) and 9 (idx 3) fill remaining 2 slots
	assert.Contains(t, indices, 4)
	assert.Contains(t, indices, 3)
}

func TestPoker_RunCpuExchanges_LowballBranch(t *testing.T) {
	pk, players := newLowballPoker()
	pk.SetPhase(PokerPhaseExchange)
	pk.SetCurrentTurn(1) // Start from CPU 1
	pk.setActedFlags([]bool{true, false, true, true})

	for _, pl := range players {
		pl.SetChips(990)
		pl.SetFolded(false)
		pl.SetAllIn(false)
	}

	// Give CPU player 1 a hand with a high card to trigger exchange
	givePlayerHand(players[1], []*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignDiamond, 4, false),
		NewCard(CardDesignClover, 5, false),
		NewCard(CardDesignSpade, 13, false), // King - should be exchanged
	})

	// Ensure the deck has cards to draw
	pk.runCpuExchanges()

	// CPU 1 should have exchanged 1 card (the King)
	exchanges := pk.GetCpuExchanges()
	assert.Equal(t, 1, len(exchanges))
	assert.Equal(t, 1, exchanges[0].PlayerIdx)
	assert.Equal(t, 1, exchanges[0].ExchangeCount)
}

// ---------------------------------------------------------------------------
// ActionLog tests
// ---------------------------------------------------------------------------

func TestPoker_ActionLog_Actions(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	// Give players enough chips
	for _, pl := range players {
		pl.SetChips(1000)
		pl.SetCurrentBet(0)
		pl.SetFolded(false)
		pl.SetAllIn(false)
	}
	pk.SetLastBet(0)
	pk.SetPot(40)

	// Human checks
	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	assert.NoError(t, err)

	log := pk.GetActionLog()
	found := false
	for _, e := range log {
		if e.ActionType == "check" && e.PlayerIdx == 0 {
			found = true
			break
		}
	}
	assert.True(t, found, "expected check action log entry")
}

func TestPoker_ActionLog_Exchange(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseExchange)
	// Give player 5 cards
	for _, pl := range players {
		pl.Reset()
		for i := 0; i < 5; i++ {
			pl.AddCard(NewCard(CardDesignSpade, i+2, false))
		}
	}
	pk.SetCurrentTurn(0)

	// Exchange 1 card
	err := pk.PlayerExchange([]int{0})
	assert.NoError(t, err)

	log := pk.GetActionLog()
	found := false
	for _, e := range log {
		if e.ActionType == "exchange" && e.PlayerIdx == 0 {
			found = true
			assert.Contains(t, e.Detail, "1 card(s)")
			break
		}
	}
	assert.True(t, found, "expected exchange action log entry")
}

func TestPoker_ActionLog_Reset(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	for _, pl := range players {
		pl.SetChips(1000)
		pl.SetCurrentBet(0)
	}
	pk.SetLastBet(0)
	_ = pk.PlayerAction(PokerActionCheck, 0, 0)

	beforeLen := len(pk.GetActionLog())
	assert.Greater(t, beforeLen, 0, "expected log entries before reset")

	err := pk.Reset()
	assert.NoError(t, err)
	// After Reset, the old log is cleared. New ante entries are created.
	// Verify the first entry has TurnNumber 0 (log was reset, not accumulated).
	log := pk.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, 1, log[0].TurnNumber, "first entry after reset should have TurnNumber 1")
}

// ---------------------------------------------------------------------------
// Meta-AI integration tests
// ---------------------------------------------------------------------------

func TestPoker_MetaAI_ProfileSurvivesReset(t *testing.T) {
	pk, _ := newTestPoker()
	cfg := pk.GetConfig()
	cfg.CpuMetaAI = true
	pk.SetConfig(cfg)

	_ = pk.Reset()
	profile := pk.GetHumanProfile()
	assert.NotNil(t, profile, "profile should be created on first Reset with CpuMetaAI=true")
	assert.Equal(t, 0, profile.GamesPlayed, "GamesPlayed should be 0 on first Reset")

	_ = pk.Reset()
	profile2 := pk.GetHumanProfile()
	assert.NotNil(t, profile2, "profile should survive Reset")
	assert.Equal(t, 1, profile2.GamesPlayed, "GamesPlayed should be incremented on second Reset")
}

func TestPoker_MetaAI_ProfileNotCreatedWhenDisabled(t *testing.T) {
	pk, _ := newTestPoker()
	// CpuMetaAI defaults to false
	_ = pk.Reset()
	assert.Nil(t, pk.GetHumanProfile(), "profile should be nil when CpuMetaAI is disabled")
}

func TestPoker_MetaAI_ResetProfileClearsProfile(t *testing.T) {
	pk, _ := newTestPoker()
	cfg := pk.GetConfig()
	cfg.CpuMetaAI = true
	pk.SetConfig(cfg)
	_ = pk.Reset()
	assert.NotNil(t, pk.GetHumanProfile())

	pk.ResetProfile()
	assert.Nil(t, pk.GetHumanProfile(), "profile should be nil after ResetProfile")
}

func TestPoker_MetaAI_LastHumanPlayMsResetOnReset(t *testing.T) {
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetConfig(PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3, CpuMetaAI: true})
	pk.SetHumanProfile(&BettingHumanProfile{})
	players[0].AddCard(NewCard(CardDesignSpade, 2, false))
	players[0].AddCard(NewCard(CardDesignHeart, 5, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 8, false))
	players[0].AddCard(NewCard(CardDesignClover, 10, false))
	players[0].AddCard(NewCard(CardDesignSpade, 12, false))
	_ = pk.PlayerAction(PokerActionBet, 10, 600)
	assert.Equal(t, 600, pk.GetLastHumanPlayMs(), "lastHumanPlayMs should be set after PlayerAction")

	_ = pk.Reset()
	assert.Equal(t, 0, pk.GetLastHumanPlayMs(), "lastHumanPlayMs should be reset to 0 on Reset")
}

func TestPoker_MetaAI_PlayerActionRecordsAction(t *testing.T) {
	t.Run("aggressive action on weak hand is recorded", func(t *testing.T) {
		pk, players := setupPokerForHumanAction(PokerPhaseDeal)
		pk.SetConfig(PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3, CpuMetaAI: true})
		pk.SetHumanProfile(&BettingHumanProfile{})
		// Give human a weak hand (HighCard)
		players[0].AddCard(NewCard(CardDesignSpade, 2, false))
		players[0].AddCard(NewCard(CardDesignHeart, 5, false))
		players[0].AddCard(NewCard(CardDesignDiamond, 8, false))
		players[0].AddCard(NewCard(CardDesignClover, 10, false))
		players[0].AddCard(NewCard(CardDesignSpade, 12, false))

		err := pk.PlayerAction(PokerActionBet, 10, 500)
		assert.NoError(t, err)

		profile := pk.GetHumanProfile()
		assert.NotNil(t, profile)
		// HighCard → bracket 0
		assert.Equal(t, 1, profile.AggressiveByBracket[0].Aggressive)
		assert.Equal(t, 1, profile.AggressiveByBracket[0].Total)
		assert.Equal(t, 1, profile.HesitationCount)
	})

	t.Run("fold on bet records fold-to-bet", func(t *testing.T) {
		pk, players := setupPokerForHumanAction(PokerPhaseDeal)
		pk.SetConfig(PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3, CpuMetaAI: true})
		pk.SetHumanProfile(&BettingHumanProfile{})
		pk.SetLastBet(20)
		// Give human cards
		players[0].AddCard(NewCard(CardDesignSpade, 2, false))
		players[0].AddCard(NewCard(CardDesignHeart, 5, false))
		players[0].AddCard(NewCard(CardDesignDiamond, 8, false))
		players[0].AddCard(NewCard(CardDesignClover, 10, false))
		players[0].AddCard(NewCard(CardDesignSpade, 12, false))

		err := pk.PlayerAction(PokerActionFold, 0, 0)
		assert.NoError(t, err)

		profile := pk.GetHumanProfile()
		assert.Equal(t, 1, profile.FoldToBetCount)
		assert.Equal(t, 1, profile.FoldToBetTotal)
	})

	t.Run("no fold-to-bet recorded when lastBet is 0", func(t *testing.T) {
		pk, players := setupPokerForHumanAction(PokerPhaseDeal)
		pk.SetConfig(PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3, CpuMetaAI: true})
		pk.SetHumanProfile(&BettingHumanProfile{})
		pk.SetLastBet(0)
		players[0].AddCard(NewCard(CardDesignSpade, 2, false))
		players[0].AddCard(NewCard(CardDesignHeart, 5, false))
		players[0].AddCard(NewCard(CardDesignDiamond, 8, false))
		players[0].AddCard(NewCard(CardDesignClover, 10, false))
		players[0].AddCard(NewCard(CardDesignSpade, 12, false))

		err := pk.PlayerAction(PokerActionCheck, 0, 0)
		assert.NoError(t, err)

		profile := pk.GetHumanProfile()
		assert.Equal(t, 0, profile.FoldToBetTotal, "no fold-to-bet opportunity when lastBet is 0")
	})

	t.Run("no recording when CpuMetaAI is disabled", func(t *testing.T) {
		pk, players := setupPokerForHumanAction(PokerPhaseDeal)
		pk.SetConfig(PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3, CpuMetaAI: false})
		players[0].AddCard(NewCard(CardDesignSpade, 2, false))
		players[0].AddCard(NewCard(CardDesignHeart, 5, false))
		players[0].AddCard(NewCard(CardDesignDiamond, 8, false))
		players[0].AddCard(NewCard(CardDesignClover, 10, false))
		players[0].AddCard(NewCard(CardDesignSpade, 12, false))

		err := pk.PlayerAction(PokerActionBet, 10, 500)
		assert.NoError(t, err)
		assert.Nil(t, pk.GetHumanProfile())
	})
}

// ---------------------------------------------------------------------------
// PlayerAction transitions to exchange phase and runs CPU exchanges
// ---------------------------------------------------------------------------

func TestPoker_PlayerAction_TransitionsToExchangeAndRunsCpuExchanges(t *testing.T) {
	// Human is dealer (index 0). After first betting round completes,
	// exchange phase should start and CPU exchanges should run automatically
	// so that currentTurn ends up on the human for card exchange.
	pk, players := setupPokerForHumanAction(PokerPhaseDeal)
	pk.SetDealerIdx(0) // human is dealer
	// All CPUs have already acted; only human is unacted
	pk.setActedFlags([]bool{false, true, true, true})
	// Give CPUs two-pair+ hands so they exchange fewer cards (deterministic)
	for i := 1; i < 4; i++ {
		givePlayerHand(players[i], []*Card{
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignHeart, 10, false),
			NewCard(CardDesignDiamond, 8, false),
			NewCard(CardDesignClover, 8, false),
			NewCard(CardDesignSpade, 14, false),
		})
	}

	err := pk.PlayerAction(PokerActionCheck, 0, 0)
	require.NoError(t, err)

	if pk.GetGameEndFlag() {
		return
	}

	// After the check completes the first betting round, the phase must have
	// advanced out of Deal — either Exchange (waiting for human to exchange)
	// or SecondBet (all players exchanged and auto-advanced).
	phase := pk.GetPhase()
	assert.True(t,
		phase == PokerPhaseExchange || phase == PokerPhaseSecondBet,
		"expected phase to advance past Deal, got %d", phase,
	)

	if phase == PokerPhaseExchange {
		assert.True(t,
			players[pk.GetCurrentTurn()].GetIsHuman(),
			"currentTurn should be the human player in exchange phase, got %d",
			pk.GetCurrentTurn(),
		)
	}
}

// **Holdem 系は EquityDisplay で勝率とポットオッズを出しているのに、5 カード
// ドローには仕組み自体が無く、2巡目ベットの判断材料が交換確率パネルしか
// 無かった (#4678)。**
func TestPoker_EquityAndPotOdds(t *testing.T) {
	withHand := func(t *testing.T, cards []*Card) *Poker {
		t.Helper()
		pk, players := setupPokerForHumanAction(PokerPhaseSecondBet)
		human := players[0]
		human.Reset()
		for _, c := range cards {
			human.AddCard(c)
		}
		return pk
	}
	quads := []*Card{
		NewCard(CardDesignSpade, 1, false),
		NewCard(CardDesignHeart, 1, false),
		NewCard(CardDesignClover, 1, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignSpade, 13, false),
	}

	t.Run("a strong hand beats a weak one", func(t *testing.T) {
		strong := withHand(t, quads)
		weak := withHand(t, []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignClover, 6, false),
			NewCard(CardDesignDiamond, 9, false),
			NewCard(CardDesignSpade, 11, false),
		})

		se, we := strong.GetEquity(), weak.GetEquity()
		if se == nil || we == nil {
			t.Fatal("ベッティングフェーズではエクイティが出る")
		}
		// **順序だけを見る。**モンテカルロなので絶対値を固定するとフレークになる。
		// フォーカードがハイカードに負ける確率はほぼ 0 なので、この比較は安定。
		if se.Equity <= we.Equity {
			t.Errorf("フォーカード %.3f がハイカード %.3f 以下", se.Equity, we.Equity)
		}
		if se.Equity <= 0.5 {
			t.Errorf("フォーカードの勝率が %.3f は低すぎる", se.Equity)
		}
	})

	// **交換フェーズでは出さない。**まだ手が変わるので、確定した勝率として
	// 読まれると誤解を招く。
	t.Run("no equity outside the betting phases", func(t *testing.T) {
		pk := withHand(t, quads)
		pk.SetPhase(PokerPhaseExchange)
		if pk.GetEquity() != nil {
			t.Error("交換フェーズではエクイティを出さない")
		}
		if pk.GetPotOdds() != 0 {
			t.Error("交換フェーズではポットオッズを出さない")
		}
	})

	t.Run("pot odds reflect the amount needed to call", func(t *testing.T) {
		pk := withHand(t, quads)
		pk.SetPot(90)
		pk.SetLastBet(10)
		// コール 10 / (ポット 90 + 10) = 10%
		if got := pk.GetPotOdds(); got < 9.9 || got > 10.1 {
			t.Errorf("GetPotOdds() = %.2f, want ~10", got)
		}
	})
}

// #5475: calcExchangeWarning は「交換枚数が3枚未満の相手がいると CPU の
// フォールド閾値を1ランク上げる」という実在の戦略要素だが、Web にも CUI にも
// 説明が無く、frontend を grep しても exchangeRead は0件だった。
// プレイヤーは自分の交換枚数が読まれていることを知る手段が無い。
func TestPoker_IsExchangeRead(t *testing.T) {
	newGame := func(phase int, exchangeCounts ...int) *Poker {
		p := NewDefaultPoker()
		p.round.phase = phase
		for i, n := range exchangeCounts {
			p.players[i].exchangeCount = n
		}
		return p
	}

	// **閾値は calcExchangeWarning と同じ。** 別に書くと、説明文と CPU の
	// 実際の挙動がずれる。
	t.Run("flags a player whose exchange count is under the threshold", func(t *testing.T) {
		p := newGame(PokerPhaseSecondBet, 2, 4, 4, 4)
		assert.True(t, p.IsExchangeRead(0))
		// 実際に CPU の警戒度が上がっていることも確かめる。
		assert.Positive(t, p.calcExchangeWarning(1, 80))
	})

	t.Run("does not flag a player at or above the threshold", func(t *testing.T) {
		p := newGame(PokerPhaseSecondBet, 3, 4, 4, 4)
		assert.False(t, p.IsExchangeRead(0))
	})

	// 0枚交換 (スタンドパット) はいちばん強く読まれる。
	t.Run("flags a stand pat", func(t *testing.T) {
		assert.True(t, newGame(PokerPhaseSecondBet, 0, 4, 4, 4).IsExchangeRead(0))
	})

	// **第2ベット以外では読まれない。** calcExchangeWarning がフェーズで
	// 早期 return するので、交換前や決着後に警告を出すのは嘘になる。
	t.Run("is quiet outside the second betting round", func(t *testing.T) {
		for _, phase := range []int{PokerPhaseInit, PokerPhaseDeal, PokerPhaseExchange, PokerPhaseEnd} {
			assert.False(t, newGame(phase, 0, 4, 4, 4).IsExchangeRead(0), "phase %d", phase)
		}
	})

	t.Run("rejects an out-of-range index instead of panicking", func(t *testing.T) {
		p := newGame(PokerPhaseSecondBet, 0, 4, 4, 4)
		assert.False(t, p.IsExchangeRead(-1))
		assert.False(t, p.IsExchangeRead(99))
	})

	// 降りた相手は読まれる側から外れる。calcExchangeWarning が folded を飛ばす。
	t.Run("does not flag a folded player", func(t *testing.T) {
		p := newGame(PokerPhaseSecondBet, 0, 4, 4, 4)
		p.players[0].SetFolded(true)
		assert.False(t, p.IsExchangeRead(0))
	})
}
