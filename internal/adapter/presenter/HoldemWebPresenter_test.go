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

func TestHoldemWebPresenter_Output(t *testing.T) {
	p := presenter.NewHoldemWebPresenter()

	setup := func() (*domain.Holdem, []*domain.HoldemPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.HoldemPlayer{
			domain.NewHoldemPlayer(true, domain.HoldemStyleTAG),
			domain.NewHoldemPlayer(false, domain.HoldemStyleLAP),
			domain.NewHoldemPlayer(false, domain.HoldemStyleTAP),
			domain.NewHoldemPlayer(false, domain.HoldemStyleLAG),
		}
		h := domain.NewHoldem(tc, players, domain.DefaultHoldemConfig())
		return h, players
	}

	t.Run("initial state", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.HoldemPhasePreFlop, out.Phase)
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
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
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
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.False(t, cpu.IsHuman)
		assert.Len(t, cpu.Cards, 0)
	})

	t.Run("CPU cards visible at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhaseShowdown)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetHandRank(domain.PokerHandOnePair)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 1)
		assert.Equal(t, "SPADE", cpu.Cards[0].Design)
		assert.Equal(t, domain.PokerHandOnePair, cpu.HandRank)
		assert.Equal(t, "One Pair", cpu.HandName)
	})

	t.Run("CPU cards visible at end phase", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 1)
	})

	t.Run("folded CPU cards hidden at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhaseShowdown)
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetFolded(true)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 0)
		assert.Equal(t, 0, cpu.HandRank)
		assert.Equal(t, "", cpu.HandName)
	})

	t.Run("community cards", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetCommunityCards([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 9, false),
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CommunityCards, 3)
		assert.Equal(t, "SPADE", out.CommunityCards[0].Design)
		assert.Equal(t, "CLOVER", out.CommunityCards[1].Design)
		assert.Equal(t, "DIAMOND", out.CommunityCards[2].Design)
	})

	t.Run("side pots", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetSidePots([]domain.HoldemSidePot{
			{Amount: 100, EligiblePlayers: []int{0, 1}},
			{Amount: 50, EligiblePlayers: []int{0}},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.SidePots, 2)
		assert.Equal(t, 100, out.SidePots[0].Amount)
		assert.Equal(t, []int{0, 1}, out.SidePots[0].EligiblePlayers)
		assert.Equal(t, 50, out.SidePots[1].Amount)
	})

	t.Run("player fields", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhasePreFlop)
		players[0].SetChips(500)
		players[0].SetCurrentBet(20)
		players[0].SetFolded(false)
		players[0].SetAllIn(false)
		players[2].SetFolded(true)
		players[3].SetAllIn(true)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
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
		h.SetPhase(domain.HoldemPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.HoldemActionCall, Amount: 10},
			{PlayerIdx: 2, Action: domain.HoldemActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuActions, 2)
		assert.Equal(t, 1, out.CpuActions[0].PlayerIdx)
		assert.Equal(t, domain.HoldemActionCall, out.CpuActions[0].Action)
		assert.Equal(t, 10, out.CpuActions[0].Amount)
		assert.Equal(t, 2, out.CpuActions[1].PlayerIdx)
	})

	t.Run("round results with best hand", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{
				PlayerIdx: 0,
				HandRank:  domain.PokerHandFlush,
				HandName:  "Flush",
				WonAmount: 200,
				BestHand: []*domain.Card{
					domain.NewCard(domain.CardDesignSpade, 1, false),
					domain.NewCard(domain.CardDesignSpade, 5, false),
				},
			},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Equal(t, 0, out.RoundResults[0].PlayerIdx)
		assert.Equal(t, domain.PokerHandFlush, out.RoundResults[0].HandRank)
		assert.Equal(t, "Flush", out.RoundResults[0].HandName)
		assert.Equal(t, 200, out.RoundResults[0].WonAmount)
		assert.Len(t, out.RoundResults[0].BestHand, 2)
		assert.Equal(t, "SPADE", out.RoundResults[0].BestHand[0].Design)
	})

	t.Run("best hand cards at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhaseShowdown)
		players[1].SetHandRank(domain.PokerHandFullHouse)
		players[1].SetBestHand([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Equal(t, "Full House", cpu.HandName)
		assert.Len(t, cpu.BestHand, 3)
	})

	t.Run("best hand empty when not showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhaseFlop)
		players[1].SetHandRank(domain.PokerHandFlush)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.BestHand, 0)
		assert.Equal(t, 0, cpu.HandRank) // not set when not showdown
	})

	t.Run("error message", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, errors.New("invalid action"))
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "invalid action", out.Message)
	})

	t.Run("error message takes priority over game end", func(t *testing.T) {
		h, _ := setup()
		h.SetGameEndFlag(true)

		result := p.Output(h, errors.New("some error"))
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "some error", out.Message)
	})

	t.Run("game end message - human wins", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You are the winner.")
	})

	t.Run("game end message - human loses", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You lose.")
	})

	t.Run("game end message - human folded", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.HoldemPhaseEnd)
		h.SetGameEndFlag(true)
		players[0].SetFolded(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Contains(t, out.Message, "You folded.")
	})

	t.Run("game end message - no results", func(t *testing.T) {
		h, _ := setup()
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Game over.", out.Message)
	})

	t.Run("pot and dealer fields", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhaseFlop)
		h.SetPot(300)
		h.SetDealerIdx(2)
		h.SetCurrentTurn(1)
		h.SetLastBet(50)
		h.SetMinRaise(100)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 300, out.Pot)
		assert.Equal(t, 2, out.DealerIdx)
		assert.Equal(t, 1, out.CurrentTurn)
		assert.Equal(t, 50, out.LastBet)
		assert.Equal(t, 100, out.MinRaise)
	})

	t.Run("getHandName out of range returns Unknown", func(t *testing.T) {
		gameMock := new(interfaces.MockHoldemGame)
		gameMock.On("GetPhase").Return(domain.HoldemPhaseShowdown)
		gameMock.On("GetPot").Return(0)
		gameMock.On("GetDealerIdx").Return(0)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetLastBet").Return(0)
		gameMock.On("GetMinRaise").Return(0)
		gameMock.On("GetCommunityCards").Return([]*domain.Card{})
		gameMock.On("GetSidePots").Return([]domain.HoldemSidePot{})
		gameMock.On("GetCpuActions").Return([]domain.HoldemCpuAction{})
		gameMock.On("GetRoundResults").Return([]domain.HoldemResult{})
		gameMock.On("GetPlayerCnt").Return(1)

		player := domain.NewHoldemPlayer(false, domain.HoldemStyleTAG)
		player.SetHandRank(99) // out of range
		gameMock.On("GetPlayer", 0).Return(player)

		result := p.Output(gameMock, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Unknown", out.Players[0].HandName)
	})

	t.Run("getHandName negative returns Unknown", func(t *testing.T) {
		gameMock := new(interfaces.MockHoldemGame)
		gameMock.On("GetPhase").Return(domain.HoldemPhaseEnd)
		gameMock.On("GetPot").Return(0)
		gameMock.On("GetDealerIdx").Return(0)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetLastBet").Return(0)
		gameMock.On("GetMinRaise").Return(0)
		gameMock.On("GetCommunityCards").Return([]*domain.Card{})
		gameMock.On("GetSidePots").Return([]domain.HoldemSidePot{})
		gameMock.On("GetCpuActions").Return([]domain.HoldemCpuAction{})
		gameMock.On("GetRoundResults").Return([]domain.HoldemResult{})
		gameMock.On("GetPlayerCnt").Return(1)

		player := domain.NewHoldemPlayer(false, domain.HoldemStyleTAG)
		player.SetHandRank(-1) // negative
		gameMock.On("GetPlayer", 0).Return(player)

		result := p.Output(gameMock, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Unknown", out.Players[0].HandName)
	})

	t.Run("playStyleName included in output", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.HoldemPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.NotEmpty(t, out.Players[1].PlayStyleName)
	})
}
