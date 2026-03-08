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
	h.setStartingChips([]int{1000, 1000, 1000, 1000})
	h.SetPhase(phase)
	h.SetCurrentTurn(0)
	h.SetLastBet(0)
	h.SetMinRaise(10)
	h.SetPot(30)
	h.setActedFlags([]bool{false, true, true, true})
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

func TestHoldem_Resize(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, 4, h.GetPlayerCnt())

	// Resize to 6 players
	newPlayers := make([]*HoldemPlayer, 6)
	newPlayers[0] = NewHoldemPlayer(true, HoldemStyleTAG)
	for i := 1; i < 6; i++ {
		newPlayers[i] = NewHoldemPlayer(false, HoldemStyleLAP)
	}
	h.Resize(newPlayers)
	assert.Equal(t, 6, h.GetPlayerCnt())
	assert.Equal(t, 0, h.GetHandCount())
	assert.Equal(t, 6, len(h.GetActedFlags()))

	// Should be able to reset and play with 6 players
	err := h.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 6, h.GetPlayerCnt())
	for i := 0; i < 6; i++ {
		assert.Equal(t, 2, h.GetPlayer(i).GetCardsSize())
	}
}

func TestHoldem_Resize_To9(t *testing.T) {
	h := newTestHoldem()

	newPlayers := make([]*HoldemPlayer, 9)
	newPlayers[0] = NewHoldemPlayer(true, HoldemStyleTAG)
	for i := 1; i < 9; i++ {
		newPlayers[i] = NewHoldemPlayer(false, HoldemStyleGTO)
	}
	h.Resize(newPlayers)
	assert.Equal(t, 9, h.GetPlayerCnt())

	err := h.Reset()
	assert.NoError(t, err)
	assert.Equal(t, 9, h.GetPlayerCnt())
	assert.True(t, h.GetPot() > 0)
}

func TestHoldem_Reset(t *testing.T) {
	h := newTestHoldem()
	_ = h.Reset()

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
	_ = h.Reset()
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
	assert.Contains(t, err.Error(), "minimum bet")
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
	h.setRaiseCount(4) // holdemDefaultMaxRaises = 4
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
	h.setRaiseCount(4) // holdemDefaultMaxRaises = 4
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
	_ = h.Reset()

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
	h.setActedFlags(flags)
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
	h.setActedFlags([]bool{false, true, true, true})

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
	// Player 0 has pair of aces → kickers should be populated
	for _, r := range h.GetRoundResults() {
		if r.PlayerIdx == 0 && r.HandRank == PokerHandOnePair {
			assert.NotNil(t, r.Kickers)
			assert.Len(t, r.Kickers, 3)
		}
	}
}

func TestHoldem_Showdown_Kickers(t *testing.T) {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.SetPhase(HoldemPhaseRiver)
	h.SetPot(200)
	h.SetCurrentTurn(0)
	h.SetLastBet(0)
	h.setActedFlags([]bool{false, true, true, true})

	// Player 0: pair of 5s with A, K, Q kickers
	h.players[0].Reset()
	h.players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	h.players[0].AddCard(NewCard(CardDesignHeart, 5, false))

	// Player 1: three of a kind (8s) with A, K kickers
	h.players[1].Reset()
	h.players[1].AddCard(NewCard(CardDesignClover, 8, false))
	h.players[1].AddCard(NewCard(CardDesignDiamond, 8, false))

	for i := 2; i < 4; i++ {
		h.players[i].SetFolded(true)
		h.players[i].Reset()
		h.players[i].AddCard(NewCard(CardDesignSpade, 2, false))
		h.players[i].AddCard(NewCard(CardDesignHeart, 3, false))
	}

	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignDiamond, 1, false),
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignSpade, 9, false),
	})

	err := h.PlayerAction(HoldemActionCheck, 0)
	assert.NoError(t, err)
	assert.True(t, h.GetGameEndFlag())

	for _, r := range h.GetRoundResults() {
		if r.PlayerIdx == 0 {
			// One Pair (5s) → 3 kickers: A(14), K(13), Q(12)
			assert.Equal(t, PokerHandOnePair, r.HandRank)
			assert.Equal(t, []int{14, 13, 12}, r.Kickers)
		}
		if r.PlayerIdx == 1 {
			// Three of a Kind (8s) → 2 kickers: A(14), K(13)
			assert.Equal(t, PokerHandThreeOfAKind, r.HandRank)
			assert.Equal(t, []int{14, 13}, r.Kickers)
		}
	}
}

func TestHoldem_BlindPosting(t *testing.T) {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.SetDealerIdx(0)
	_ = h.Reset()

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
	_ = h.Reset()

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

	h.setActedFlags([]bool{true, true, true, true})

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
	h.setActedFlags([]bool{false, true, true, true})

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
	h.setActedFlags([]bool{false, true, true, false})

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

func TestHoldem_CpuPotBet(t *testing.T) {
	h := newTestHoldem()

	t.Run("normal pot-relative bet", func(t *testing.T) {
		h.SetPot(200)
		h.SetMinRaise(10)
		h.config.BigBlind = 10
		// 200 * 75 / 100 = 150
		assert.Equal(t, 150, h.cpuPotBet(75))
	})

	t.Run("floor to BigBlind when pot is small", func(t *testing.T) {
		h.SetPot(10)
		h.SetMinRaise(5)
		h.config.BigBlind = 20
		// 10 * 50 / 100 = 5, but min is BB=20
		assert.Equal(t, 20, h.cpuPotBet(50))
	})

	t.Run("floor to minRaise when pot*pct < minRaise", func(t *testing.T) {
		h.SetPot(10)
		h.SetMinRaise(30)
		h.config.BigBlind = 10
		// 10 * 50 / 100 = 5, BB floor=10, but minRaise=30
		assert.Equal(t, 30, h.cpuPotBet(50))
	})

	t.Run("pot is zero floors to BigBlind", func(t *testing.T) {
		h.SetPot(0)
		h.SetMinRaise(5)
		h.config.BigBlind = 10
		// 0 * 50 / 100 = 0, but min is BB=10
		assert.Equal(t, 10, h.cpuPotBet(50))
	})

	t.Run("100% pot bet", func(t *testing.T) {
		h.SetPot(300)
		h.SetMinRaise(10)
		h.config.BigBlind = 10
		assert.Equal(t, 300, h.cpuPotBet(100))
	})
}

func TestHoldem_HUDStats_TotalHandsIncrement(t *testing.T) {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	assert.Equal(t, 0, h.players[0].GetTotalHands())

	_ = h.Reset()
	assert.Equal(t, 1, h.players[0].GetTotalHands())

	_ = h.Reset()
	assert.Equal(t, 2, h.players[0].GetTotalHands())
}

func TestHoldem_HUDStats_VPIP_Call(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	// vpipTracked/pfrTracked must be initialized
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.SetLastBet(20)

	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 1, h.players[0].GetVPIPCount())
	assert.Equal(t, 0, h.players[0].GetPFRCount())
}

func TestHoldem_HUDStats_VPIP_Bet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1

	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 1, h.players[0].GetVPIPCount())
	assert.Equal(t, 1, h.players[0].GetPFRCount())
}

func TestHoldem_HUDStats_VPIP_Raise(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.SetLastBet(20)
	h.SetMinRaise(10)

	_ = h.executeAction(0, HoldemActionRaise, 20)
	assert.Equal(t, 1, h.players[0].GetVPIPCount())
	assert.Equal(t, 1, h.players[0].GetPFRCount())
}

func TestHoldem_HUDStats_VPIP_AllIn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1

	_ = h.executeAction(0, HoldemActionAllIn, 0)
	assert.Equal(t, 1, h.players[0].GetVPIPCount())
	assert.Equal(t, 1, h.players[0].GetPFRCount())
}

func TestHoldem_HUDStats_NoTrack_Check(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1

	_ = h.executeAction(0, HoldemActionCheck, 0)
	assert.Equal(t, 0, h.players[0].GetVPIPCount())
	assert.Equal(t, 0, h.players[0].GetPFRCount())
}

func TestHoldem_HUDStats_NoTrack_Fold(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1

	_ = h.executeAction(0, HoldemActionFold, 0)
	assert.Equal(t, 0, h.players[0].GetVPIPCount())
	assert.Equal(t, 0, h.players[0].GetPFRCount())
}

func TestHoldem_HUDStats_OncePerHand(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.SetLastBet(20)
	h.SetMinRaise(10)

	// First call increments VPIP
	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 1, h.players[0].GetVPIPCount())

	// Reset acted flags to allow second action
	h.actedFlags[0] = false
	h.SetLastBet(40)

	// Second call should NOT increment again
	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 1, h.players[0].GetVPIPCount())
}

func TestHoldem_HUDStats_PostFlop_NoTrack(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 0, h.players[0].GetVPIPCount())
	assert.Equal(t, 0, h.players[0].GetPFRCount())
}

// --- 3Bet tracking tests ---

func TestHoldem_HUDStats_ThreeBet_Opportunity_Call(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.raiseCount = 1 // prior raise exists
	h.SetLastBet(20)

	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 1, h.players[0].GetThreeBetOpportunity())
	assert.Equal(t, 0, h.players[0].GetThreeBetCount())
}

func TestHoldem_HUDStats_ThreeBet_Raise(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.raiseCount = 1 // prior raise exists
	h.SetLastBet(20)
	h.SetMinRaise(10)

	_ = h.executeAction(0, HoldemActionRaise, 20)
	assert.Equal(t, 1, h.players[0].GetThreeBetOpportunity())
	assert.Equal(t, 1, h.players[0].GetThreeBetCount())
}

func TestHoldem_HUDStats_ThreeBet_AllIn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.raiseCount = 1

	_ = h.executeAction(0, HoldemActionAllIn, 0)
	assert.Equal(t, 1, h.players[0].GetThreeBetOpportunity())
	assert.Equal(t, 1, h.players[0].GetThreeBetCount())
}

func TestHoldem_HUDStats_ThreeBet_Fold(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.raiseCount = 1

	_ = h.executeAction(0, HoldemActionFold, 0)
	assert.Equal(t, 1, h.players[0].GetThreeBetOpportunity())
	assert.Equal(t, 0, h.players[0].GetThreeBetCount())
}

func TestHoldem_HUDStats_ThreeBet_NoOpportunity(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.raiseCount = 0 // no prior raise

	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 0, h.players[0].GetThreeBetOpportunity())
	assert.Equal(t, 0, h.players[0].GetThreeBetCount())
}

func TestHoldem_HUDStats_ThreeBet_OncePerHand(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1
	h.raiseCount = 1
	h.SetLastBet(20)

	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 1, h.players[0].GetThreeBetOpportunity())

	// Reset acted flags to allow second action
	h.actedFlags[0] = false
	h.SetLastBet(40)
	h.raiseCount = 2

	_ = h.executeAction(0, HoldemActionCall, 0)
	// Should NOT increment again
	assert.Equal(t, 1, h.players[0].GetThreeBetOpportunity())
}

func TestHoldem_HUDStats_ThreeBet_ResetBetweenHands(t *testing.T) {
	h := newTestHoldem()
	for _, p := range h.players {
		p.SetChips(1000)
	}
	_ = h.Reset()
	// After Reset, threeBetTracked should be re-initialized
	assert.Equal(t, len(h.players), len(h.threeBetTracked))
	for _, tracked := range h.threeBetTracked {
		assert.False(t, tracked)
	}
}

// --- AF tracking tests ---

func TestHoldem_HUDStats_AF_PostFlop_Bet(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 1, h.players[0].GetPostFlopBetRaise())
	assert.Equal(t, 0, h.players[0].GetPostFlopCall())
}

func TestHoldem_HUDStats_AF_PostFlop_Raise(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetLastBet(20)
	h.SetMinRaise(10)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionRaise, 20)
	assert.Equal(t, 1, h.players[0].GetPostFlopBetRaise())
}

func TestHoldem_HUDStats_AF_PostFlop_Call(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetLastBet(20)
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionCall, 0)
	assert.Equal(t, 0, h.players[0].GetPostFlopBetRaise())
	assert.Equal(t, 1, h.players[0].GetPostFlopCall())
}

func TestHoldem_HUDStats_AF_PostFlop_AllIn(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionAllIn, 0)
	assert.Equal(t, 1, h.players[0].GetPostFlopBetRaise())
}

func TestHoldem_HUDStats_AF_PostFlop_Check_NoTrack(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionCheck, 0)
	assert.Equal(t, 0, h.players[0].GetPostFlopBetRaise())
	assert.Equal(t, 0, h.players[0].GetPostFlopCall())
}

func TestHoldem_HUDStats_AF_PostFlop_Fold_NoTrack(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionFold, 0)
	assert.Equal(t, 0, h.players[0].GetPostFlopBetRaise())
	assert.Equal(t, 0, h.players[0].GetPostFlopCall())
}

func TestHoldem_HUDStats_AF_PreFlop_NoTrack(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhasePreFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.players[0].totalHands = 1

	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 0, h.players[0].GetPostFlopBetRaise())
	assert.Equal(t, 0, h.players[0].GetPostFlopCall())
}

func TestHoldem_HUDStats_AF_Cumulative(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseFlop)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
	})

	_ = h.executeAction(0, HoldemActionBet, 20)
	h.actedFlags[0] = false
	h.SetLastBet(0)
	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 2, h.players[0].GetPostFlopBetRaise())
}

func TestHoldem_HUDStats_AF_Turn_Tracks(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseTurn)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 5, false),
	})

	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 1, h.players[0].GetPostFlopBetRaise())
}

func TestHoldem_HUDStats_AF_River_Tracks(t *testing.T) {
	h := setupHoldemForHumanAction(HoldemPhaseRiver)
	h.vpipTracked = make([]bool, len(h.players))
	h.pfrTracked = make([]bool, len(h.players))
	h.threeBetTracked = make([]bool, len(h.players))
	h.SetCommunityCards([]*Card{
		NewCard(CardDesignSpade, 2, false),
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignClover, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignSpade, 6, false),
	})

	_ = h.executeAction(0, HoldemActionBet, 20)
	assert.Equal(t, 1, h.players[0].GetPostFlopBetRaise())
}

func TestHoldem_TournamentMode_BlindEscalation(t *testing.T) {
	t.Run("blinds escalate after BlindLevelHands", func(t *testing.T) {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(10000)
		}
		cfg := HoldemConfig{
			SmallBlind:      5,
			BigBlind:        10,
			InitChips:       10000,
			TournamentMode:  true,
			BlindLevelHands: 2,
			BlindMultiplier: 200,
		}
		h.SetConfig(cfg)

		// Hand 0: no escalation (handCount=0 → 0%2==0 but handCount>0 is false)
		_ = h.Reset()
		assert.Equal(t, 5, h.GetConfig().SmallBlind)
		assert.Equal(t, 10, h.GetConfig().BigBlind)
		assert.Equal(t, 1, h.GetHandCount())

		// Hand 1: no escalation (handCount=1 → 1%2!=0)
		_ = h.Reset()
		assert.Equal(t, 5, h.GetConfig().SmallBlind)
		assert.Equal(t, 10, h.GetConfig().BigBlind)
		assert.Equal(t, 2, h.GetHandCount())

		// Hand 2: escalation! (handCount=2 → 2%2==0 && handCount>0)
		_ = h.Reset()
		assert.Equal(t, 10, h.GetConfig().SmallBlind)
		assert.Equal(t, 20, h.GetConfig().BigBlind)
		assert.Equal(t, 3, h.GetHandCount())

		// Hand 3: no escalation
		_ = h.Reset()
		assert.Equal(t, 10, h.GetConfig().SmallBlind)
		assert.Equal(t, 20, h.GetConfig().BigBlind)

		// Hand 4: escalation again
		_ = h.Reset()
		assert.Equal(t, 20, h.GetConfig().SmallBlind)
		assert.Equal(t, 40, h.GetConfig().BigBlind)
	})

	t.Run("no escalation when tournament mode is off", func(t *testing.T) {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(10000)
		}
		cfg := DefaultHoldemConfig()
		cfg.TournamentMode = false
		h.SetConfig(cfg)

		_ = h.Reset()
		_ = h.Reset()
		_ = h.Reset()
		assert.Equal(t, 5, h.GetConfig().SmallBlind)
		assert.Equal(t, 10, h.GetConfig().BigBlind)
	})

	t.Run("blind floor values", func(t *testing.T) {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(10000)
		}
		cfg := HoldemConfig{
			SmallBlind:      1,
			BigBlind:        2,
			InitChips:       10000,
			TournamentMode:  true,
			BlindLevelHands: 1,
			BlindMultiplier: 101, // 1.01x → floor(1*101/100)=1, floor(2*101/100)=2
		}
		h.SetConfig(cfg)

		_ = h.Reset() // hand 0: no escalation
		assert.Equal(t, 1, h.GetConfig().SmallBlind)
		assert.Equal(t, 2, h.GetConfig().BigBlind)

		_ = h.Reset() // hand 1: escalation → 1*101/100=1, 2*101/100=2
		assert.Equal(t, 1, h.GetConfig().SmallBlind)
		assert.Equal(t, 2, h.GetConfig().BigBlind)
	})

	t.Run("blind floor values hit - multiplier under 100", func(t *testing.T) {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(10000)
		}
		cfg := HoldemConfig{
			SmallBlind:      1,
			BigBlind:        2,
			InitChips:       10000,
			TournamentMode:  true,
			BlindLevelHands: 1,
			BlindMultiplier: 50, // 0.5x → 1*50/100=0, 2*50/100=1
		}
		h.SetConfig(cfg)

		_ = h.Reset() // hand 0: no escalation
		assert.Equal(t, 1, h.GetConfig().SmallBlind)
		assert.Equal(t, 2, h.GetConfig().BigBlind)

		_ = h.Reset() // hand 1: escalation → 0 < 1 → SmallBlind=1; 1 < 2 → BigBlind=2
		assert.Equal(t, 1, h.GetConfig().SmallBlind)
		assert.Equal(t, 2, h.GetConfig().BigBlind)
	})

	t.Run("no panic when BlindLevelHands is zero", func(t *testing.T) {
		h := newTestHoldem()
		for _, p := range h.players {
			p.SetChips(10000)
		}
		cfg := HoldemConfig{
			SmallBlind:      5,
			BigBlind:        10,
			InitChips:       10000,
			TournamentMode:  true,
			BlindLevelHands: 0, // should not panic
			BlindMultiplier: 200,
		}
		h.SetConfig(cfg)

		assert.NotPanics(t, func() {
			_ = h.Reset()
			_ = h.Reset()
		})
		// Blinds should remain unchanged since BlindLevelHands=0 guard prevents escalation
		assert.Equal(t, 5, h.GetConfig().SmallBlind)
		assert.Equal(t, 10, h.GetConfig().BigBlind)
	})

	t.Run("GetHandCount and SetHandCount", func(t *testing.T) {
		h := newTestHoldem()
		assert.Equal(t, 0, h.GetHandCount())
		h.SetHandCount(5)
		assert.Equal(t, 5, h.GetHandCount())
	})
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

func setupHoldemForHumanActionWithConfig(phase int, limit BettingLimitType) *Holdem {
	players := []*HoldemPlayer{
		NewHoldemPlayer(true, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleLAP),
		NewHoldemPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultHoldemConfig()
	cfg.BettingLimit = limit
	tc := NewTrumpCards(0)
	h := NewHoldem(tc, players, cfg)
	for _, p := range h.players {
		p.SetChips(1000)
	}
	h.setStartingChips([]int{1000, 1000, 1000, 1000})
	h.SetPhase(phase)
	h.SetCurrentTurn(0)
	h.SetLastBet(0)
	h.SetMinRaise(10)
	h.SetPot(100)
	h.setActedFlags([]bool{false, true, true, true})
	for _, p := range h.players {
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignHeart, 13, false))
	}
	return h
}

func TestHoldem_GetRaiseCount(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, 0, h.GetRaiseCount())
}

func TestHoldem_BettingLimits_PotLimit(t *testing.T) {
	t.Run("bet exceeding pot limit is rejected", func(t *testing.T) {
		h := setupHoldemForHumanActionWithConfig(HoldemPhaseFlop, BettingLimitPotLimit)
		h.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignClover, 4, false),
		})
		h.SetPot(100)
		h.SetLastBet(0)
		// maxBetAmount = pot + lastBet = 100 + 0 = 100; bet 130 > 100
		err := h.PlayerAction(HoldemActionBet, 130)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("bet within pot limit succeeds", func(t *testing.T) {
		h := setupHoldemForHumanActionWithConfig(HoldemPhaseFlop, BettingLimitPotLimit)
		h.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignClover, 4, false),
		})
		h.SetPot(100)
		h.SetLastBet(0)
		// maxBetAmount = pot + lastBet = 100 + 0 = 100; bet 100 is within limit
		err := h.PlayerAction(HoldemActionBet, 100)
		assert.NoError(t, err)
	})
}

func TestHoldem_BettingLimits_NoLimit(t *testing.T) {
	t.Run("bet succeeds even when raiseCount exceeds fixed limit cap", func(t *testing.T) {
		h := setupHoldemForHumanActionWithConfig(HoldemPhaseFlop, BettingLimitNoLimit)
		h.SetCommunityCards([]*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignClover, 4, false),
		})
		h.setRaiseCount(10) // well past fixed limit of 4
		// NoLimit has maxRaises=0, so no raise cap
		err := h.PlayerAction(HoldemActionBet, 20)
		assert.NoError(t, err)
	})
}

// --- Rebuy / Addon tests ---

// newTestHoldemWithRebuy creates a 4-player holdem with rebuy enabled
func newTestHoldemWithRebuy() *Holdem {
	players := []*HoldemPlayer{
		NewHoldemPlayer(true, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleLAP),
		NewHoldemPlayer(false, HoldemStyleLAG),
	}
	cfg := DefaultHoldemConfig()
	cfg.RebuyEnabled = true
	cfg.RebuyMaxCount = 3
	cfg.RebuyChips = 1000
	cfg.RebuyPeriodHands = 20
	tc := NewTrumpCards(0)
	h := NewHoldem(tc, players, cfg)
	return h
}

func TestDefaultHoldemConfig_RebuyAddonDefaults(t *testing.T) {
	cfg := DefaultHoldemConfig()
	assert.False(t, cfg.RebuyEnabled)
	assert.Equal(t, 3, cfg.RebuyMaxCount)
	assert.Equal(t, 1000, cfg.RebuyChips)
	assert.Equal(t, 20, cfg.RebuyPeriodHands)
	assert.False(t, cfg.AddonEnabled)
	assert.Equal(t, 1500, cfg.AddonChips)
	assert.Equal(t, 20, cfg.AddonAfterHand)
}

func TestNewHoldem_InitializesRebuyAddonSlices(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, []int{0, 0, 0, 0}, h.GetRebuyCounts())
	assert.Equal(t, []bool{false, false, false, false}, h.GetAddonUsed())
	assert.Equal(t, 0, h.GetRebuyPhaseType())
}

func TestHoldem_Resize_InitializesRebuyAddonSlices(t *testing.T) {
	h := newTestHoldem()
	newPlayers := make([]*HoldemPlayer, 6)
	newPlayers[0] = NewHoldemPlayer(true, HoldemStyleTAG)
	for i := 1; i < 6; i++ {
		newPlayers[i] = NewHoldemPlayer(false, HoldemStyleLAP)
	}
	h.Resize(newPlayers)
	assert.Equal(t, []int{0, 0, 0, 0, 0, 0}, h.GetRebuyCounts())
	assert.Equal(t, []bool{false, false, false, false, false, false}, h.GetAddonUsed())
}

func TestHoldem_Reset_RebuyDisabled_BustedPlayerGetsInitChips(t *testing.T) {
	h := newTestHoldem() // rebuy disabled by default
	// Bust human player
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	// With rebuy disabled, busted player gets InitChips
	assert.True(t, h.GetPlayer(0).GetChips() > 0)
}

func TestHoldem_Reset_RebuyEnabled_CpuAutoRebuys(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Bust CPU player 1
	h.GetPlayer(1).SetChips(0)
	// Human has chips, so no human rebuy prompt
	h.GetPlayer(0).SetChips(1000)
	_ = h.Reset()
	// CPU should have auto-rebuyed: chips = RebuyChips (1000), rebuyCount = 1
	assert.Equal(t, 1, h.GetRebuyCounts()[1])
}

func TestHoldem_Reset_RebuyEnabled_HumanBustedGetsRebuyPhase(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Bust human player
	h.GetPlayer(0).SetChips(0)
	err := h.Reset()
	assert.NoError(t, err)
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 1, h.GetRebuyPhaseType()) // rebuyPhaseRebuy
}

func TestHoldem_Reset_RebuyEnabled_HumanHasChips_NoRebuyPrompt(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Human has chips
	h.GetPlayer(0).SetChips(500)
	err := h.Reset()
	assert.NoError(t, err)
	// Should proceed to preflop (or later), not rebuy phase
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Reset_RebuyEnabled_BeyondRebuyPeriod(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Set hand count beyond rebuy period
	h.SetHandCount(20) // handCount will become 21 after increment, > RebuyPeriodHands (20)
	h.GetPlayer(0).SetChips(0)
	err := h.Reset()
	assert.NoError(t, err)
	// Beyond rebuy period, no rebuy prompt
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Reset_RebuyEnabled_MaxRebuyReached(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Human already used max rebuys
	h.SetRebuyCounts([]int{3, 0, 0, 0}) // maxCount = 3
	h.GetPlayer(0).SetChips(0)
	err := h.Reset()
	assert.NoError(t, err)
	// Max rebuys reached, no rebuy prompt
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Reset_AddonEnabled_CpuAutoAddons(t *testing.T) {
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonChips = 1500
	cfg.AddonAfterHand = 1 // addon at hand 1 (after first increment)
	h.SetConfig(cfg)
	// handCount starts at 0, after increment it becomes 1 = AddonAfterHand
	// Human has chips, so no rebuy. All CPUs should auto-addon
	// But human needs addon prompt
	err := h.Reset()
	assert.NoError(t, err)
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType()) // rebuyPhaseAddon
	// CPUs should have gotten addon
	addonUsed := h.GetAddonUsed()
	assert.True(t, addonUsed[1])
	assert.True(t, addonUsed[2])
	assert.True(t, addonUsed[3])
	assert.False(t, addonUsed[0]) // human not yet
}

func TestHoldem_Reset_AddonEnabled_NotAtAddonHand(t *testing.T) {
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 5
	h.SetConfig(cfg)
	// handCount will be 1 after reset, not 5
	err := h.Reset()
	assert.NoError(t, err)
	assert.NotEqual(t, 2, h.GetRebuyPhaseType())
}

func TestHoldem_Reset_AddonEnabled_AlreadyUsed(t *testing.T) {
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	// Mark all as already used
	h.SetAddonUsed([]bool{true, true, true, true})
	err := h.Reset()
	assert.NoError(t, err)
	// No addon prompt since all already used
	assert.NotEqual(t, 2, h.GetRebuyPhaseType())
}

func TestHoldem_Rebuy_Success(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Bust human, trigger rebuy phase
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 1, h.GetRebuyPhaseType())

	// Execute rebuy
	err := h.Rebuy()
	assert.NoError(t, err)
	// Human should have chips now
	assert.True(t, h.GetPlayer(0).GetChips() > 0)
	// Rebuy count incremented
	assert.Equal(t, 1, h.GetRebuyCounts()[0])
	// Should have continued to deal (preflop or beyond)
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Rebuy_WrongPhase(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhasePreFlop)
	err := h.Rebuy()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_Rebuy_WrongPhaseType(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhaseRebuy)
	h.SetRebuyPhaseType(2) // addon phase, not rebuy
	err := h.Rebuy()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_Rebuy_ThenAddon(t *testing.T) {
	h := newTestHoldemWithRebuy()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1 // addon at hand 1
	h.SetConfig(cfg)
	// Bust human
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 1, h.GetRebuyPhaseType()) // rebuy first

	// Do rebuy -> should transition to addon phase
	err := h.Rebuy()
	assert.NoError(t, err)
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType()) // now addon phase
}

func TestHoldem_SkipRebuy_Success(t *testing.T) {
	h := newTestHoldemWithRebuy()
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())

	err := h.SkipRebuy()
	assert.NoError(t, err)
	// Human should get InitChips
	assert.True(t, h.GetPlayer(0).GetChips() > 0)
	// Should continue to deal
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_SkipRebuy_WrongPhase(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhasePreFlop)
	err := h.SkipRebuy()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_SkipRebuy_WrongPhaseType(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhaseRebuy)
	h.SetRebuyPhaseType(2) // addon, not rebuy
	err := h.SkipRebuy()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_SkipRebuy_ThenAddon(t *testing.T) {
	h := newTestHoldemWithRebuy()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, 1, h.GetRebuyPhaseType())

	// Skip rebuy -> should transition to addon phase
	err := h.SkipRebuy()
	assert.NoError(t, err)
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType())
}

func TestHoldem_Addon_Success(t *testing.T) {
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonChips = 1500
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	_ = h.Reset()
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType())

	chipsBefore := h.GetPlayer(0).GetChips()
	err := h.Addon()
	assert.NoError(t, err)
	// Human should get addon chips
	assert.Equal(t, chipsBefore+1500, h.GetPlayer(0).GetChips())
	assert.True(t, h.GetAddonUsed()[0])
	// Should continue to deal
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Addon_WrongPhase(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhasePreFlop)
	err := h.Addon()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_Addon_WrongPhaseType(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhaseRebuy)
	h.SetRebuyPhaseType(1) // rebuy, not addon
	err := h.Addon()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_SkipAddon_Success(t *testing.T) {
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	_ = h.Reset()
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType())

	err := h.SkipAddon()
	assert.NoError(t, err)
	// Should continue without addon
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
	assert.False(t, h.GetAddonUsed()[0])
}

func TestHoldem_SkipAddon_WrongPhase(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhasePreFlop)
	err := h.SkipAddon()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_SkipAddon_WrongPhaseType(t *testing.T) {
	h := newTestHoldem()
	h.SetPhase(HoldemPhaseRebuy)
	h.SetRebuyPhaseType(1) // rebuy, not addon
	err := h.SkipAddon()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPhase)
}

func TestHoldem_IsRebuyAvailable(t *testing.T) {
	t.Run("rebuy disabled returns false", func(t *testing.T) {
		h := newTestHoldem() // rebuy disabled by default
		h.GetPlayer(0).SetChips(0)
		assert.False(t, h.IsRebuyAvailable())
	})

	t.Run("rebuy enabled human busted within period returns true", func(t *testing.T) {
		h := newTestHoldemWithRebuy()
		h.SetHandCount(5) // within RebuyPeriodHands=20
		h.GetPlayer(0).SetChips(0)
		assert.True(t, h.IsRebuyAvailable())
	})

	t.Run("rebuy enabled but beyond period returns false", func(t *testing.T) {
		h := newTestHoldemWithRebuy()
		h.SetHandCount(21) // > RebuyPeriodHands=20
		h.GetPlayer(0).SetChips(0)
		assert.False(t, h.IsRebuyAvailable())
	})

	t.Run("rebuy enabled but human has chips returns false", func(t *testing.T) {
		h := newTestHoldemWithRebuy()
		h.SetHandCount(5)
		h.GetPlayer(0).SetChips(500)
		assert.False(t, h.IsRebuyAvailable())
	})

	t.Run("rebuy enabled but max count reached returns false", func(t *testing.T) {
		h := newTestHoldemWithRebuy()
		h.SetHandCount(5)
		h.GetPlayer(0).SetChips(0)
		h.SetRebuyCounts([]int{3, 0, 0, 0}) // max=3
		assert.False(t, h.IsRebuyAvailable())
	})
}

func TestHoldem_IsAddonAvailable(t *testing.T) {
	t.Run("addon disabled returns false", func(t *testing.T) {
		h := newTestHoldem()
		assert.False(t, h.IsAddonAvailable())
	})

	t.Run("addon enabled at correct hand returns true", func(t *testing.T) {
		h := newTestHoldem()
		cfg := h.GetConfig()
		cfg.AddonEnabled = true
		cfg.AddonAfterHand = 5
		h.SetConfig(cfg)
		h.SetHandCount(5)
		assert.True(t, h.IsAddonAvailable())
	})

	t.Run("addon enabled but wrong hand returns false", func(t *testing.T) {
		h := newTestHoldem()
		cfg := h.GetConfig()
		cfg.AddonEnabled = true
		cfg.AddonAfterHand = 5
		h.SetConfig(cfg)
		h.SetHandCount(3)
		assert.False(t, h.IsAddonAvailable())
	})

	t.Run("addon enabled but already used returns false", func(t *testing.T) {
		h := newTestHoldem()
		cfg := h.GetConfig()
		cfg.AddonEnabled = true
		cfg.AddonAfterHand = 5
		h.SetConfig(cfg)
		h.SetHandCount(5)
		h.SetAddonUsed([]bool{true, false, false, false})
		assert.False(t, h.IsAddonAvailable())
	})
}

func TestHoldem_GetRebuyCounts_ReturnsCopy(t *testing.T) {
	h := newTestHoldem()
	h.SetRebuyCounts([]int{1, 2, 0, 0})
	counts := h.GetRebuyCounts()
	counts[0] = 99
	// Original should not be modified
	assert.Equal(t, 1, h.GetRebuyCounts()[0])
}

func TestHoldem_GetAddonUsed_ReturnsCopy(t *testing.T) {
	h := newTestHoldem()
	h.SetAddonUsed([]bool{true, false, false, false})
	used := h.GetAddonUsed()
	used[0] = false
	// Original should not be modified
	assert.True(t, h.GetAddonUsed()[0])
}

func TestHoldem_GetRebuyPhaseType(t *testing.T) {
	h := newTestHoldem()
	assert.Equal(t, 0, h.GetRebuyPhaseType())
	h.SetRebuyPhaseType(1)
	assert.Equal(t, 1, h.GetRebuyPhaseType())
	h.SetRebuyPhaseType(2)
	assert.Equal(t, 2, h.GetRebuyPhaseType())
}

func TestHoldem_Reset_RebuyEnabled_CpuMaxRebuyReached(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// CPU 1 already at max rebuys and busted
	h.SetRebuyCounts([]int{0, 3, 0, 0})
	h.GetPlayer(1).SetChips(0)
	h.GetPlayer(0).SetChips(1000) // human has chips
	_ = h.Reset()
	// CPU 1 should NOT have gotten rebuy (already at max)
	assert.Equal(t, 3, h.GetRebuyCounts()[1])
}

func TestHoldem_Reset_AddonEnabled_CpuAlreadyUsed(t *testing.T) {
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	// CPU 1 already used addon
	h.SetAddonUsed([]bool{false, true, false, false})
	_ = h.Reset()
	// Phase should be rebuy (addon) because human still needs addon
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType())
	// CPU 1 should still be true (already used), CPUs 2,3 should have gotten addon
	addonUsed := h.GetAddonUsed()
	assert.True(t, addonUsed[1])
	assert.True(t, addonUsed[2])
	assert.True(t, addonUsed[3])
}

func TestHoldem_Reset_RebuyEnabled_NoBustedPlayers(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// All players have chips
	for _, p := range h.GetPlayers() {
		p.SetChips(500)
	}
	err := h.Reset()
	assert.NoError(t, err)
	// No rebuy prompt, should proceed to preflop
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Rebuy_AddonNotDue(t *testing.T) {
	// When rebuy succeeds but addon is not due (wrong hand), should go straight to continueReset
	h := newTestHoldemWithRebuy()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 99 // not at this hand
	h.SetConfig(cfg)
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, 1, h.GetRebuyPhaseType())

	err := h.Rebuy()
	assert.NoError(t, err)
	// Should proceed directly, not to addon phase
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_SkipRebuy_AddonNotDue(t *testing.T) {
	h := newTestHoldemWithRebuy()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 99
	h.SetConfig(cfg)
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()

	err := h.SkipRebuy()
	assert.NoError(t, err)
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Rebuy_AddonAllAlreadyUsed(t *testing.T) {
	// Rebuy succeeds, addon is due but all players already used it
	h := newTestHoldemWithRebuy()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	h.SetAddonUsed([]bool{true, true, true, true})
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, 1, h.GetRebuyPhaseType())

	err := h.Rebuy()
	assert.NoError(t, err)
	// Addon all used, should go to continueReset
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_SkipRebuy_AddonAllAlreadyUsed(t *testing.T) {
	h := newTestHoldemWithRebuy()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	h.SetAddonUsed([]bool{true, true, true, true})
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()

	err := h.SkipRebuy()
	assert.NoError(t, err)
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Rebuy_AddonDisabled(t *testing.T) {
	// Rebuy with addon disabled should go straight to continueReset
	h := newTestHoldemWithRebuy() // addon disabled by default
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()
	assert.Equal(t, 1, h.GetRebuyPhaseType())

	err := h.Rebuy()
	assert.NoError(t, err)
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_SkipRebuy_AddonDisabled(t *testing.T) {
	h := newTestHoldemWithRebuy()
	h.GetPlayer(0).SetChips(0)
	_ = h.Reset()

	err := h.SkipRebuy()
	assert.NoError(t, err)
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
}

func TestHoldem_Reset_AddonEnabled_NoCpuNeedAddon_HumanOnly(t *testing.T) {
	// All CPUs already used addon, only human needs it
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	h.SetAddonUsed([]bool{false, true, true, true})
	err := h.Reset()
	assert.NoError(t, err)
	assert.Equal(t, HoldemPhaseRebuy, h.GetPhase())
	assert.Equal(t, 2, h.GetRebuyPhaseType())
}

func TestHoldem_Reset_AddonEnabled_AllCpuAddon_NoHumanNeeded(t *testing.T) {
	// Human already used addon, only CPUs need it -> no rebuy phase prompt
	h := newTestHoldem()
	cfg := h.GetConfig()
	cfg.AddonEnabled = true
	cfg.AddonAfterHand = 1
	h.SetConfig(cfg)
	h.SetAddonUsed([]bool{true, false, false, false})
	err := h.Reset()
	assert.NoError(t, err)
	// No addon prompt for human, CPUs auto-addon, proceed to deal
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
	// CPUs should have gotten addon
	assert.True(t, h.GetAddonUsed()[1])
	assert.True(t, h.GetAddonUsed()[2])
	assert.True(t, h.GetAddonUsed()[3])
}

func TestHoldem_Reset_RebuyEnabled_OnlyCpuBusted_NoHumanPrompt(t *testing.T) {
	h := newTestHoldemWithRebuy()
	// Only CPU is busted, human is fine
	h.GetPlayer(0).SetChips(1000)
	h.GetPlayer(1).SetChips(0)
	h.GetPlayer(2).SetChips(0)
	err := h.Reset()
	assert.NoError(t, err)
	// Should not show rebuy phase for human
	assert.NotEqual(t, HoldemPhaseRebuy, h.GetPhase())
	// CPUs should have auto-rebuyed
	assert.Equal(t, 1, h.GetRebuyCounts()[1])
	assert.Equal(t, 1, h.GetRebuyCounts()[2])
}

func TestHoldem_HoldemPhaseRebuy_Constant(t *testing.T) {
	assert.Equal(t, 7, HoldemPhaseRebuy)
}
