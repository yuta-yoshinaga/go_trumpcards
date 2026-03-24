package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestShortDeckWebPresenter_Output(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)

	setup := func() (*domain.ShortDeck, []*domain.ShortDeckPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.ShortDeckPlayer{
			domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
		return h, players
	}

	t.Run("initial state", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.ShortDeckPhasePreFlop, out.Phase)
		assert.Equal(t, 4, len(out.Players))
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, "", out.Message)
		assert.Len(t, out.CommunityCards, 0)
		assert.Len(t, out.SidePots, 0)
		assert.Len(t, out.CpuActions, 0)
		assert.Len(t, out.RoundResults, 0)
	})

	t.Run("human cards visible", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.True(t, human.IsHuman)
		assert.Len(t, human.Cards, 2)
		assert.Equal(t, "SPADE", human.Cards[0].Design)
		assert.Equal(t, 10, human.Cards[0].Value)
		assert.Equal(t, "HEART", human.Cards[1].Design)
		assert.Equal(t, 11, human.Cards[1].Value)
	})

	t.Run("CPU cards hidden before showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.False(t, cpu.IsHuman)
		assert.Len(t, cpu.Cards, 0)
	})

	t.Run("CPU cards visible at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetHandRank(domain.PokerHandOnePair)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 1)
		assert.Equal(t, "SPADE", cpu.Cards[0].Design)
		assert.Equal(t, domain.PokerHandOnePair, cpu.HandRank)
		assert.Equal(t, "One Pair", cpu.HandName)
	})

	t.Run("CPU cards visible at end phase", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 1)
	})

	t.Run("folded CPU cards hidden at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetFolded(true)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 0)
		assert.Equal(t, 0, cpu.HandRank)
		assert.Equal(t, "", cpu.HandName)
	})

	t.Run("community cards", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 6, false),
			domain.NewCard(domain.CardDesignClover, 8, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CommunityCards, 3)
		assert.Equal(t, "SPADE", out.CommunityCards[0].Design)
		assert.Equal(t, "CLOVER", out.CommunityCards[1].Design)
		assert.Equal(t, "DIAMOND", out.CommunityCards[2].Design)
	})

	t.Run("side pots", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		h.SetSidePots([]domain.ShortDeckSidePot{
			{Amount: 100, EligiblePlayers: []int{0, 1}},
			{Amount: 50, EligiblePlayers: []int{0}},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.SidePots, 2)
		assert.Equal(t, 100, out.SidePots[0].Amount)
		assert.Equal(t, []int{0, 1}, out.SidePots[0].EligiblePlayers)
		assert.Equal(t, 50, out.SidePots[1].Amount)
	})

	t.Run("player fields", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].SetChips(500)
		players[0].SetCurrentBet(20)
		players[0].SetFolded(false)
		players[0].SetAllIn(false)
		players[2].SetFolded(true)
		players[3].SetAllIn(true)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.Players[0].ID)
		assert.True(t, out.Players[0].IsHuman)
		assert.Equal(t, 500, out.Players[0].Chips)
		assert.Equal(t, 20, out.Players[0].CurrentBet)
		assert.False(t, out.Players[0].Folded)
		assert.False(t, out.Players[0].AllIn)

		assert.True(t, out.Players[2].Folded)
		assert.True(t, out.Players[3].AllIn)
	})

	t.Run("CPU actions", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		h.SetCpuActions([]domain.ShortDeckCpuAction{
			{PlayerIdx: 1, Action: domain.ShortDeckActionCall, Amount: 10},
			{PlayerIdx: 2, Action: domain.ShortDeckActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuActions, 2)
		assert.Equal(t, 1, out.CpuActions[0].PlayerIdx)
		assert.Equal(t, domain.ShortDeckActionCall, out.CpuActions[0].Action)
		assert.Equal(t, 10, out.CpuActions[0].Amount)
		assert.Equal(t, 2, out.CpuActions[1].PlayerIdx)
	})

	t.Run("round results with best hand", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.ShortDeckResult{
			{
				PlayerIdx: 0,
				HandRank:  domain.PokerHandFlush,
				HandName:  "Flush",
				WonAmount: 200,
				BestHand: []*domain.Card{
					domain.NewCard(domain.CardDesignSpade, 1, false),
					domain.NewCard(domain.CardDesignSpade, 8, false),
				},
			},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Equal(t, 0, out.RoundResults[0].PlayerIdx)
		assert.Equal(t, domain.PokerHandFlush, out.RoundResults[0].HandRank)
		assert.Equal(t, "Flush", out.RoundResults[0].HandName)
		assert.Equal(t, "", out.RoundResults[0].Kickers)
		assert.Equal(t, 200, out.RoundResults[0].WonAmount)
		assert.Len(t, out.RoundResults[0].BestHand, 2)
		assert.Equal(t, "SPADE", out.RoundResults[0].BestHand[0].Design)
	})

	t.Run("round results with kickers", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.ShortDeckResult{
			{
				PlayerIdx: 0,
				HandRank:  domain.PokerHandOnePair,
				HandName:  "One Pair",
				Kickers:   []int{14, 13, 12},
				WonAmount: 200,
				BestHand:  nil,
			},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Equal(t, "A, K, Q", out.RoundResults[0].Kickers)
	})

	t.Run("best hand cards at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		players[1].SetHandRank(domain.PokerHandFullHouse)
		players[1].SetBestHand([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 6, false),
			domain.NewCard(domain.CardDesignDiamond, 6, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Equal(t, "Full House", cpu.HandName)
		assert.Len(t, cpu.BestHand, 3)
	})

	t.Run("best hand empty when not showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		players[1].SetHandRank(domain.PokerHandFlush)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.BestHand, 0)
		assert.Equal(t, 0, cpu.HandRank) // not set when not showdown
	})

	t.Run("error message", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, errors.New("invalid action"))
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "invalid action", out.Message)
	})

	t.Run("error message takes priority over game end", func(t *testing.T) {
		h, _ := setup()
		h.SetGameEndFlag(true)

		result := p.Output(h, errors.New("some error"))
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "some error", out.Message)
	})

	t.Run("game end message - human wins", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You are the winner.")
		assert.Equal(t, "shortdeck.result.win", out.MessageCode)
	})

	t.Run("game end message - human loses", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You lose.")
		assert.Equal(t, "shortdeck.result.lose", out.MessageCode)
	})

	t.Run("game end message - human folded", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetGameEndFlag(true)
		players[0].SetFolded(true)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 1, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You folded.")
		assert.Equal(t, "shortdeck.result.folded", out.MessageCode)
	})

	t.Run("game end message - no results", func(t *testing.T) {
		h, _ := setup()
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.ShortDeckResult{})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Game over.", out.Message)
		assert.Equal(t, "shortdeck.result.gameOver", out.MessageCode)
	})

	t.Run("pot and dealer fields", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseFlop)
		h.SetPot(300)
		h.SetDealerIdx(2)
		h.SetCurrentTurn(1)
		h.SetLastBet(50)
		h.SetMinRaise(100)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 300, out.Pot)
		assert.Equal(t, 2, out.DealerIdx)
		assert.Equal(t, 1, out.CurrentTurn)
		assert.Equal(t, 50, out.LastBet)
		assert.Equal(t, 100, out.MinRaise)
	})

	t.Run("getHandName out of range returns Unknown", func(t *testing.T) {
		gameMock := new(interfaces.MockShortDeckGame)
		gameMock.On("GetPhase").Return(domain.ShortDeckPhaseShowdown)
		gameMock.On("GetPot").Return(0)
		gameMock.On("GetDealerIdx").Return(0)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetLastBet").Return(0)
		gameMock.On("GetMinRaise").Return(0)
		gameMock.On("GetRaiseCount").Return(0)
		gameMock.On("GetCommunityCards").Return([]*domain.Card{})
		gameMock.On("GetSidePots").Return([]domain.ShortDeckSidePot{})
		gameMock.On("GetCpuActions").Return([]domain.ShortDeckCpuAction{})
		gameMock.On("GetRoundResults").Return([]domain.ShortDeckResult{})
		gameMock.On("GetPlayerCnt").Return(1)
		gameMock.On("GetConfig").Return(domain.DefaultShortDeckConfig())
		gameMock.On("GetHandCount").Return(0)
		gameMock.On("IsRebuyAvailable").Return(false)
		gameMock.On("IsAddonAvailable").Return(false)
		gameMock.On("GetRebuyCounts").Return([]int{0})
		gameMock.On("GetAddonUsed").Return([]bool{false})
		gameMock.On("GetRebuyPhaseType").Return(0)
		gameMock.On("IsMuckAvailable").Return(false)
		gameMock.On("GetEquity").Return((*domain.HoldemEquityResult)(nil))
		gameMock.On("GetPotOdds").Return(0.0)
		gameMock.On("GetHumanProfile").Return((*domain.BettingHumanProfile)(nil))

		player := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		player.SetHandRank(99) // out of range
		gameMock.On("GetPlayer", 0).Return(player)

		result := p.Output(gameMock, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Unknown", out.Players[0].HandName)
	})

	t.Run("getHandName negative returns Unknown", func(t *testing.T) {
		gameMock := new(interfaces.MockShortDeckGame)
		gameMock.On("GetPhase").Return(domain.ShortDeckPhaseEnd)
		gameMock.On("GetPot").Return(0)
		gameMock.On("GetDealerIdx").Return(0)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetLastBet").Return(0)
		gameMock.On("GetMinRaise").Return(0)
		gameMock.On("GetRaiseCount").Return(0)
		gameMock.On("GetCommunityCards").Return([]*domain.Card{})
		gameMock.On("GetSidePots").Return([]domain.ShortDeckSidePot{})
		gameMock.On("GetCpuActions").Return([]domain.ShortDeckCpuAction{})
		gameMock.On("GetRoundResults").Return([]domain.ShortDeckResult{})
		gameMock.On("GetPlayerCnt").Return(1)
		gameMock.On("GetConfig").Return(domain.DefaultShortDeckConfig())
		gameMock.On("GetHandCount").Return(0)
		gameMock.On("IsRebuyAvailable").Return(false)
		gameMock.On("IsAddonAvailable").Return(false)
		gameMock.On("GetRebuyCounts").Return([]int{0})
		gameMock.On("GetAddonUsed").Return([]bool{false})
		gameMock.On("GetRebuyPhaseType").Return(0)
		gameMock.On("IsMuckAvailable").Return(false)
		gameMock.On("GetEquity").Return((*domain.HoldemEquityResult)(nil))
		gameMock.On("GetPotOdds").Return(0.0)
		gameMock.On("GetHumanProfile").Return((*domain.BettingHumanProfile)(nil))

		player := domain.NewShortDeckPlayer(false, domain.HoldemStyleTAG)
		player.SetHandRank(-1) // negative
		gameMock.On("GetPlayer", 0).Return(player)

		result := p.Output(gameMock, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Unknown", out.Players[0].HandName)
	})

	t.Run("playStyleName included in output", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.NotEmpty(t, out.Players[1].PlayStyleName)
	})

	t.Run("HUD stats included in output", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementTotalHands()
		players[0].IncrementVPIP()
		players[0].IncrementPFR()

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 2, out.Players[0].TotalHands)
		assert.Equal(t, 50, out.Players[0].VPIP)
		assert.Equal(t, 50, out.Players[0].PFR)
		assert.Equal(t, 0, out.Players[0].ThreeBet)
		assert.Equal(t, "-", out.Players[0].AF)
	})

	t.Run("HUD stats zero when no hands played", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.Players[0].TotalHands)
		assert.Equal(t, 0, out.Players[0].VPIP)
		assert.Equal(t, 0, out.Players[0].PFR)
		assert.Equal(t, 0, out.Players[0].ThreeBet)
		assert.Equal(t, "-", out.Players[0].AF)
	})

	t.Run("HUD 3Bet and AF normal values", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementThreeBetOpportunity()
		players[0].IncrementThreeBetOpportunity()
		players[0].IncrementThreeBet()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopCall()

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 50, out.Players[0].ThreeBet) // 1*100/2=50
		assert.Equal(t, "3.0", out.Players[0].AF)    // 3/1=3.0
	})

	t.Run("HUD AF infinity", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "∞", out.Players[0].AF)
	})

	t.Run("tournament mode fields included in output", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		cfg := domain.ShortDeckConfig{
			SmallBlind:      10,
			BigBlind:        20,
			InitChips:       1000,
			TournamentMode:  true,
			BlindLevelHands: 5,
			BlindMultiplier: 200,
		}
		h.SetConfig(cfg)
		h.SetHandCount(3)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 3, out.HandCount)
		assert.Equal(t, 10, out.SmallBlind)
		assert.Equal(t, 20, out.BigBlind)
		assert.True(t, out.TournamentMode)
		assert.Equal(t, 5, out.BlindLevelHands)
		assert.Equal(t, 200, out.BlindMultiplier)
	})

	t.Run("tournament mode fields default when not enabled", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.HandCount)
		assert.Equal(t, 5, out.SmallBlind)
		assert.Equal(t, 10, out.BigBlind)
		assert.False(t, out.TournamentMode)
		assert.Equal(t, 10, out.BlindLevelHands)
		assert.Equal(t, 200, out.BlindMultiplier)
	})
}

func TestShortDeckWebPresenter_Output_RebuyAddonFields(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)

	setup := func() (*domain.ShortDeck, []*domain.ShortDeckPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.ShortDeckPlayer{
			domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
		return h, players
	}

	t.Run("default values when rebuy/addon disabled", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.False(t, out.RebuyAvailable)
		assert.False(t, out.AddonAvailable)
		assert.Equal(t, []int{0, 0, 0, 0}, out.RebuyCounts)
		assert.Equal(t, []bool{false, false, false, false}, out.AddonUsed)
		assert.False(t, out.RebuyEnabled)
		assert.False(t, out.AddonEnabled)
		assert.Equal(t, 3, out.RebuyMaxCount)
		assert.Equal(t, 1000, out.RebuyChips)
		assert.Equal(t, 1500, out.AddonChips)
		assert.Equal(t, 20, out.RebuyPeriodHands)
		assert.Equal(t, 20, out.AddonAfterHand)
		assert.Equal(t, 0, out.RebuyPhaseType)
	})

	t.Run("values when rebuy/addon enabled with config", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseRebuy)
		cfg := domain.ShortDeckConfig{
			SmallBlind:       10,
			BigBlind:         20,
			InitChips:        1000,
			TournamentMode:   true,
			BlindLevelHands:  5,
			BlindMultiplier:  200,
			RebuyEnabled:     true,
			RebuyMaxCount:    5,
			RebuyChips:       2000,
			RebuyPeriodHands: 30,
			AddonEnabled:     true,
			AddonChips:       3000,
			AddonAfterHand:   25,
		}
		h.SetConfig(cfg)
		h.SetRebuyCounts([]int{2, 1, 0, 3})
		h.SetAddonUsed([]bool{true, false, false, true})
		h.SetRebuyPhaseType(1)

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.True(t, out.RebuyEnabled)
		assert.True(t, out.AddonEnabled)
		assert.Equal(t, 5, out.RebuyMaxCount)
		assert.Equal(t, 2000, out.RebuyChips)
		assert.Equal(t, 3000, out.AddonChips)
		assert.Equal(t, 30, out.RebuyPeriodHands)
		assert.Equal(t, 25, out.AddonAfterHand)
		assert.Equal(t, []int{2, 1, 0, 3}, out.RebuyCounts)
		assert.Equal(t, []bool{true, false, false, true}, out.AddonUsed)
		assert.Equal(t, 1, out.RebuyPhaseType)
	})
}

func TestShortDeckWebPresenter_Output_BettingLimitFields(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)

	setup := func() (*domain.ShortDeck, []*domain.ShortDeckPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.ShortDeckPlayer{
			domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
		return h, players
	}

	t.Run("default Fixed limit", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 0, out.BettingLimit)
		assert.Equal(t, 0, out.RaiseCount)
		assert.Equal(t, 0, out.MaxBetAmount)
	})

	t.Run("tableSize reflects player count", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 4, out.TableSize)
	})

	t.Run("tableSize 6-max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players6 := make([]*domain.ShortDeckPlayer, 6)
		players6[0] = domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG)
		for i := 1; i < 6; i++ {
			players6[i] = domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP)
		}
		h6 := domain.NewShortDeck(tc, players6, domain.DefaultShortDeckConfig())
		h6.SetPhase(domain.ShortDeckPhasePreFlop)
		players6[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players6[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h6, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 6, out.TableSize)
		assert.Equal(t, 6, len(out.Players))
	})
}

func TestShortDeckWebPresenter_Output_MuckFields(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)

	setup := func() (*domain.ShortDeck, []*domain.ShortDeckPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.ShortDeckPlayer{
			domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
		return h, players
	}

	t.Run("muckAvailable true when showdown and human lost", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.True(t, out.MuckAvailable)
	})

	t.Run("muckAvailable false when not showdown", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.False(t, out.MuckAvailable)
	})

	t.Run("mucked result: handRank=0 handName empty bestHand empty mucked=true", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 13}, WonAmount: 0, Mucked: true, BestHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 8, false),
			}},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.RoundResults[0].HandRank)
		assert.Equal(t, "", out.RoundResults[0].HandName)
		assert.Equal(t, "", out.RoundResults[0].Kickers)
		assert.Len(t, out.RoundResults[0].BestHand, 0)
		assert.True(t, out.RoundResults[0].Mucked)
		// non-mucked result is normal
		assert.Equal(t, domain.PokerHandFlush, out.RoundResults[1].HandRank)
		assert.False(t, out.RoundResults[1].Mucked)
	})

	t.Run("muck prompt message when IsMuckAvailable", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Muck or show your hand.", out.Message)
		assert.Equal(t, "shortdeck.muck.prompt", out.MessageCode)
	})

	t.Run("error takes priority over muck prompt", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, errors.New("some error"))
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "some error", out.Message)
		assert.Equal(t, "", out.MessageCode)
	})

	t.Run("You mucked message in buildResultMessage", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.ShortDeckResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "You mucked.", out.Message)
		assert.Equal(t, "shortdeck.result.mucked", out.MessageCode)
	})
}

func TestShortDeckWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockShortDeckGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raised to 100", Cards: []*domain.Card{domain.NewCard(domain.CardDesignDiamond, 10, true)}},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"actionType":"raise"`)
		assert.Contains(t, result, `"detail":"raised to 100"`)
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockShortDeckGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockShortDeckGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})
}

func TestShortDeckWebPresenter_Equity(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)

	setup := func() (*domain.ShortDeck, []*domain.ShortDeckPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.ShortDeckPlayer{
			domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
			domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewShortDeck(tc, players, domain.DefaultShortDeckConfig())
		return h, players
	}

	t.Run("equity populated during active phase", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.NotNil(t, out.Equity)
		assert.Greater(t, out.Equity.WinProbability, 0.0)
		assert.NotEmpty(t, out.Equity.HandOdds)
		assert.NotNil(t, out.PotOdds)
	})

	t.Run("equity nil during showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.ShortDeckPhaseShowdown)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Nil(t, out.Equity)
		assert.Nil(t, out.PotOdds)
	})

	t.Run("equity nil during end phase", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.ShortDeckPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.ShortDeckResult{})

		result := p.Output(h, nil)
		var out controller.ShortDeckWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Nil(t, out.Equity)
		assert.Nil(t, out.PotOdds)
	})
}

func TestShortDeckWebPresenter_Output_WithMetaAIProfile(t *testing.T) {
	p := new(presenter.ShortDeckWebPresenter)
	tc := domain.NewTrumpCards(0)
	players := []*domain.ShortDeckPlayer{
		domain.NewShortDeckPlayer(true, domain.HoldemStyleTAG),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleLAP),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleTAP),
		domain.NewShortDeckPlayer(false, domain.HoldemStyleGTO),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := domain.DefaultShortDeckConfig()
	cfg.CpuMetaAI = true
	o := domain.NewShortDeck(tc, players, cfg)
	_ = o.Reset()

	result := p.Output(o, nil)
	var out controller.ShortDeckWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.MetaAI)
	assert.True(t, out.MetaAI.Enabled)
	assert.NotNil(t, out.Profile)
}
