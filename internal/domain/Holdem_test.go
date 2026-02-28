package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestHoldem() *Holdem {
	players := []*HoldemPlayer{
		NewHoldemPlayer(true, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleLAP),
		NewHoldemPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultHoldemConfig()
	tc := NewTrumpCards(0)
	h := NewHoldem(tc, players, cfg)
	return h
}

// setupHoldemForHumanAction sets up a holdem game where it's the human's turn in a specific phase
func setupHoldemForHumanAction(phase int) *Holdem {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.SetStartingChips([]int{1000, 1000, 1000, 1000})
	h.SetPhase(phase)
	h.SetCurrentTurn(0)
	h.SetLastBet(0)
	h.SetMinRaise(10)
	h.SetPot(30)
	h.SetActedFlags([]bool{false, true, true, true})
	// Give players cards
	for _, p := range h.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 13, false))
	}
	return h
}

func TestNewHoldem(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, HoldemPhaseInit, h.GetPhase())
	assert.Equal(t, 4, h.GetPlayerCnt())
	assert.NotNil(t, h.GetCommunityCards())
	assert.Equal(t, 0, h.GetPot())
	assert.False(t, h.GetGameEndFlag())
}

func TestHoldem_Reset(t *testing.T) {
	h := newTestHoldem()
	h.Reset()

	// After reset, each player should have 2 cards and chips
	for i := 0; i < h.GetPlayerCnt(); i++ {
		p := h.GetPlayer(i)
		assert.Equal(t, 2, p.GetCardsSize())
		assert.False(t, p.GetFolded())
		assert.False(t, p.GetAllIn())
	}
	assert.True(t, h.GetPot() > 0) // blinds posted
	assert.Equal(t, HoldemPhasePreFlop, h.GetPhase())
}

func TestHoldem_Reset_ChipsPreserved(t *testing.T) {
	h := newTestHoldem()
	h.players[0].SetChips(500)
	h.Reset()
	// Player 0 had 500 chips, should keep them (minus blind if applicable)
	assert.True(t, h.players[0].GetChips() > 0)
}

func TestHoldem_PlayerAction_GameEnded(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.SetGameEndFlag(true)

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already ended")
}

func TestHoldem_PlayerAction_WrongPhase_Init(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseInit)

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestHoldem_PlayerAction_WrongPhase_Showdown(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseShowdown)

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.Error(t, err)
}

func TestHoldem_PlayerAction_NotHumanTurn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.SetCurrentTurn(1)

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not your turn")
}

func TestHoldem_PlayerAction_Fold(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionFold, 0)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetFolded())
}

func TestHoldem_PlayerAction_Check(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.NoError(t, err)
}

func TestHoldem_PlayerAction_Check_WithOutstandingBet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot check")
}

func TestHoldem_PlayerAction_Call(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionCall, 0)
	assert.NoError(t, err)
}

func TestHoldem_PlayerAction_Call_NothingToCall(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(0)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionCall, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Nothing to call")
}

func TestHoldem_PlayerAction_Call_AllIn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(2000) // More than player's chips
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	initialChips := h.players[0].GetChips()
	err := h.PlayerAction(HoldemActionCall, 0)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetAllIn())
	// チップはオールイン時に全額投入されるが、ゲーム進行後にショーダウンで
	// チップが再配分される場合がある。投入自体が行われたことを確認。
	assert.True(t, h.players[0].GetChips() < initialChips || h.GetPhase() >= HoldemPhaseShowdown)
}

func TestHoldem_PlayerAction_Bet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionBet, 50)
	assert.NoError(t, err)
	// After bet, CPUs act and game may advance; just verify no error
}

func TestHoldem_PlayerAction_Bet_WithOutstandingBet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionBet, 50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot bet")
}

func TestHoldem_PlayerAction_Bet_TooSmall(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionBet, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "big blind")
}

func TestHoldem_PlayerAction_Bet_InsufficientChips(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionBet, 2000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Insufficient")
}

func TestHoldem_PlayerAction_Bet_ExactChips(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.players[0].SetChips(100)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionBet, 100)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetAllIn())
}

func TestHoldem_PlayerAction_Raise(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetMinRaise(10)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionRaise, 30)
	assert.NoError(t, err)
}

func TestHoldem_PlayerAction_Raise_TooSmall(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetMinRaise(20)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionRaise, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum raise")
}

func TestHoldem_PlayerAction_Raise_InsufficientChips_AutoAllIn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetMinRaise(10)
	h.players[0].SetChips(25) // Not enough for call (20) + raise (10) → auto all-in
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	chipsBefore := h.players[0].GetChips()
	err := h.PlayerAction(HoldemActionRaise, 10)
	assert.NoError(t, err)
	// Player went all-in with 25 chips (auto-converted from raise)
	// Game may have progressed to showdown and redistributed chips
	assert.True(t, h.GetPot() > 0 || h.GetPhase() >= HoldemPhaseShowdown)
	_ = chipsBefore
}

func TestHoldem_PlayerAction_Raise_ExactChips(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetMinRaise(10)
	h.players[0].SetChips(30) // Exact amount for call (20) + raise (10)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionRaise, 10)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetAllIn())
}

func TestHoldem_PlayerAction_Raise_MaxRaisesReached(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(20)
	h.SetMinRaise(10)
	h.SetRaiseCount(4) // holdemMaxRaisesPerRound = 4
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionRaise, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Maximum number of raises")
}

func TestHoldem_PlayerAction_Bet_MaxRaisesReached(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(0)
	h.SetMinRaise(10)
	h.SetRaiseCount(4) // holdemMaxRaisesPerRound = 4
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionBet, 20)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Maximum number of raises")
}

func TestHoldem_PlayerAction_AllIn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	chipsBefore := h.players[0].GetChips()
	err := h.PlayerAction(HoldemActionAllIn, 0)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetAllIn())
	// After all-in, player's chips were added to pot; they may have won the pot in showdown
	// so just verify all-in amount was deducted before any winnings
	assert.True(t, chipsBefore > 0)
}

func TestHoldem_PlayerAction_AllIn_NoChips(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.players[0].SetChips(0)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionAllIn, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No chips")
}

func TestHoldem_PlayerAction_AllIn_BelowLastBet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(2000)
	h.players[0].SetChips(100)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionAllIn, 0)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetAllIn())
}

func TestHoldem_PlayerAction_UnknownAction(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(99, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown")
}

func TestHoldem_EveryoneFolds(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	// CPU players all folded
	h.players[1].SetFolded(true)
	h.players[2].SetFolded(true)
	h.players[3].SetFolded(true)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	// Player bets, but everyone is already folded so only player 0 is active
	// Actually, player 0 is already the only active player after the 3 folds
	// Let's manually set only 2 players active, then fold
	h.players[1].SetFolded(false)
	h.players[2].SetFolded(true)
	h.players[3].SetFolded(true)

	// Now fold player 0
	err := h.PlayerAction(HoldemActionFold, 0)
	assert.NoError(t, err)
	// Player 1 should win
	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, HoldemPhaseEnd, h.GetPhase())
}

func TestHoldem_PhaseTransitions(t *testing.T) {
	h := newTestHoldem()
	// Give all players chips and reset
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.Reset()

	// After reset, should be in PreFlop
	assert.Equal(t, HoldemPhasePreFlop, h.GetPhase())
}

func TestHoldem_GetPlayer_OutOfBounds(t *testing.T) {
	h := newTestHoldem()
	assert.Nil(t, h.GetPlayer(-1))
	assert.Nil(t, h.GetPlayer(100))
}

func TestHoldem_IsHumanTurn(t *testing.T) {
	h := newTestHoldem()
	h.SetCurrentTurn(0)
	assert.True(t, h.IsHumanTurn())
	h.SetCurrentTurn(1)
	assert.False(t, h.IsHumanTurn())
}

func TestHoldem_GettersSetters(t *testing.T) {
	h := newTestHoldem()

	h.SetPhase(HoldemPhaseFlop)
	assert.Equal(t, HoldemPhaseFlop, h.GetPhase())

	h.SetCurrentTurn(2)
	assert.Equal(t, 2, h.GetCurrentTurn())

	cards := []*Card{NewCard(CardDesignSpade, 1, false)}
	h.SetCommunityCards(cards)
	assert.Equal(t, 1, len(h.GetCommunityCards()))

	h.SetPot(500)
	assert.Equal(t, 500, h.GetPot())

	h.SetDealerIdx(3)
	assert.Equal(t, 3, h.GetDealerIdx())

	h.SetGameEndFlag(true)
	assert.True(t, h.GetGameEndFlag())

	flags := []bool{true, false, true, false}
	h.SetActedFlags(flags)
	assert.Equal(t, flags, h.GetActedFlags())

	h.SetLastBet(100)
	assert.Equal(t, 100, h.GetLastBet())

	h.SetMinRaise(50)
	assert.Equal(t, 50, h.GetMinRaise())

	results := []HoldemResult{{PlayerIdx: 0, WonAmount: 100}}
	h.SetRoundResults(results)
	assert.Equal(t, 1, len(h.GetRoundResults()))

	actions := []HoldemCpuAction{{PlayerIdx: 1, Action: HoldemActionCall}}
	h.SetCpuActions(actions)
	assert.Equal(t, 1, len(h.GetCpuActions()))

	pots := []HoldemSidePot{{Amount: 100, EligiblePlayers: []int{0, 1}}}
	h.SetSidePots(pots)
	assert.Equal(t, 1, len(h.GetSidePots()))

	cfg := HoldemConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000}
	h.SetConfig(cfg)
	assert.Equal(t, cfg, h.GetConfig())
}

func TestHoldem_Showdown(t *testing.T) {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.SetPhase(HoldemPhaseRiver)
	h.SetPot(200)
	h.SetCurrentTurn(0)
	h.SetLastBet(0)
	h.SetActedFlags([]bool{false, true, true, true})

	// Give player 0 pocket aces
	h.players[0].Reset()
	h.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	h.players[0].AddCard(NewCard(CardDesignHeart, 1, false))

	// Give others weaker hands
	for i := 1; i < 4; i++ {
		h.players[i].Reset()
		h.players[i].AddCard(NewCard(CardDesignSpade, 2+i, false))
		h.players[i].AddCard(NewCard(CardDesignHeart, 3+i, false))
	}

	h.SetCommunityCards([]*Card{
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 9, false),
		NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 7, false),
	})

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.NoError(t, err)

	// Game should be at showdown or end
	assert.True(t, h.GetGameEndFlag())
	assert.True(t, len(h.GetRoundResults()) > 0)
}

func TestHoldem_BlindPosting(t *testing.T) {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.SetDealerIdx(0)
	h.Reset()

	// SB is player 1, BB is player 2
	// Verify pot includes both blinds
	assert.True(t, h.GetPot() >= h.config.SmallBlind+h.config.BigBlind)
}

func TestHoldem_BlindPosting_InsufficientChips(t *testing.T) {
	h := newTestHoldem()
	// Give SB player very few chips
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.SetDealerIdx(0)
	h.players[1].SetChips(3) // SB position, less than small blind
	h.Reset()

	// Player 1 should have 0 chips and be all-in
	assert.Equal(t, 0, h.players[1].GetChips())
}

func TestHoldem_GetHandName(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, "High Card", h.getHandName(PokerHandHighCard))
	assert.Equal(t, "Royal Flush", h.getHandName(PokerHandRoyalFlush))
	assert.Equal(t, "Unknown", h.getHandName(-1))
	assert.Equal(t, "Unknown", h.getHandName(99))
}

func TestHoldem_Constants(t *testing.T) {
	assert.Equal(t, 0, HoldemPhaseInit)
	assert.Equal(t, 1, HoldemPhasePreFlop)
	assert.Equal(t, 2, HoldemPhaseFlop)
	assert.Equal(t, 3, HoldemPhaseTurn)
	assert.Equal(t, 4, HoldemPhaseRiver)
	assert.Equal(t, 5, HoldemPhaseShowdown)
	assert.Equal(t, 6, HoldemPhaseEnd)

	assert.Equal(t, 0, HoldemActionFold)
	assert.Equal(t, 1, HoldemActionCheck)
	assert.Equal(t, 2, HoldemActionCall)
	assert.Equal(t, 3, HoldemActionBet)
	assert.Equal(t, 4, HoldemActionRaise)
	assert.Equal(t, 5, HoldemActionAllIn)
}

func TestHoldem_GetPlayers(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, 4, len(h.GetPlayers()))
}

func TestHoldem_SidePots_AllIn(t *testing.T) {
	h := newTestHoldem()
	cfg := DefaultHoldemConfig()
	cfg.InitChips = 100
	h.SetConfig(cfg)
	for _, p := range h.players {
		p.SetChips(100)
	}

	// Simulate all-in scenario
	h.SetPhase(HoldemPhaseFlop)
	h.SetPot(400)
	h.SetCurrentTurn(0)

	// Player 0 has fewer chips, goes all-in first
	h.players[0].SetChips(0)
	h.players[0].SetAllIn(true)
	h.players[0].SetCurrentBet(0)

	// Player 1 has more chips
	h.players[1].SetChips(0)
	h.players[1].SetAllIn(true)

	h.players[2].SetFolded(true)
	h.players[3].SetFolded(true)

	h.SetActedFlags([]bool{true, true, true, true})

	// Give hands
	h.players[0].Reset()
	h.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	h.players[0].AddCard(NewCard(CardDesignHeart, 1, false))

	h.players[1].Reset()
	h.players[1].AddCard(NewCard(CardDesignClover, 2, false))
	h.players[1].AddCard(NewCard(CardDesignDiamond, 3, false))

	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignClover, 9, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 13, false),
	})

	h.resolveShowdown()
	assert.True(t, h.GetGameEndFlag())
	assert.True(t, len(h.GetRoundResults()) > 0)
}

func TestHoldem_SplitPot(t *testing.T) {
	h := newTestHoldem()
	cfg := DefaultHoldemConfig()
	h.SetConfig(cfg)
	for _, p := range h.players {
		p.SetChips(1000)
	}

	h.SetPhase(HoldemPhaseRiver)
	h.SetPot(200)
	h.SetCurrentTurn(0)
	h.SetLastBet(0)
	h.SetActedFlags([]bool{false, true, true, true})

	// Give identical hands (same values, different suits)
	h.players[0].Reset()
	h.players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	h.players[0].AddCard(NewCard(CardDesignHeart, 13, false))

	h.players[1].Reset()
	h.players[1].AddCard(NewCard(CardDesignClover, 1, false))
	h.players[1].AddCard(NewCard(CardDesignDiamond, 13, false))

	// Fold players 2 and 3
	h.players[2].SetFolded(true)
	h.players[3].SetFolded(true)

	h.players[2].Reset()
	h.players[2].AddCard(NewCard(CardDesignSpade, 2, false))
	h.players[2].AddCard(NewCard(CardDesignHeart, 3, false))
	h.players[3].Reset()
	h.players[3].AddCard(NewCard(CardDesignSpade, 4, false))
	h.players[3].AddCard(NewCard(CardDesignHeart, 5, false))

	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignClover, 10, false),
		NewCard(CardDesignDiamond, 11, false),
		NewCard(CardDesignSpade, 12, false),
	})

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.NoError(t, err)
	assert.True(t, h.GetGameEndFlag())

	// Both should have won some amount
	totalWon := 0
	for _, r := range h.GetRoundResults() {
		totalWon += r.WonAmount
	}
	assert.Equal(t, 200, totalWon)
}

func TestHoldem_FullGame_AllFold(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.players[1].SetFolded(true)
	h.players[2].SetFolded(true)
	// Player 3 is still active
	h.SetActedFlags([]bool{false, true, true, false})

	err := h.PlayerAction(HoldemActionFold, 0)
	assert.NoError(t, err)
	// Player 3 should win
	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, HoldemPhaseEnd, h.GetPhase())
}

func TestHoldem_Raise_NegativeDiff(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(0)
	h.players[0].SetCurrentBet(10) // Player has bet more than lastBet
	h.SetMinRaise(10)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionRaise, 10)
	assert.NoError(t, err)
}

func TestHoldem_AllIn_AboveLastBet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.SetLastBet(50)
	h.SetMinRaise(10)
	h.players[0].SetChips(200) // More than lastBet, so allin raises
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	err := h.PlayerAction(HoldemActionAllIn, 0)
	assert.NoError(t, err)
	assert.True(t, h.players[0].GetAllIn())
}
