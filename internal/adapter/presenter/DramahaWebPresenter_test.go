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

func TestDramahaWebPresenter_Output(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	setup := func() (*domain.Dramaha, []*domain.DramahaPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.DramahaPlayer{
			domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
			domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
		return h, players
	}

	t.Run("initial state", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.DramahaPhasePreFlop, out.Phase)
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
		h.SetPhase(domain.DramahaPhasePreFlop)
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
		h.SetPhase(domain.DramahaPhasePreFlop)
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
		h.SetPhase(domain.DramahaPhaseShowdown)
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
		h.SetPhase(domain.DramahaPhaseEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.Cards, 1)
	})

	t.Run("folded CPU cards hidden at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhaseShowdown)
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
		h.SetPhase(domain.DramahaPhaseFlop)
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
		h.SetPhase(domain.DramahaPhaseFlop)
		h.SetSidePots([]domain.SidePot{
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
		h.SetPhase(domain.DramahaPhasePreFlop)
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
		h.SetPhase(domain.DramahaPhasePreFlop)
		h.SetCpuActions([]domain.HoldemCpuAction{
			{PlayerIdx: 1, Action: domain.DramahaActionCall, Amount: 10},
			{PlayerIdx: 2, Action: domain.DramahaActionFold, Amount: 0},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuActions, 2)
		assert.Equal(t, 1, out.CpuActions[0].PlayerIdx)
		assert.Equal(t, domain.DramahaActionCall, out.CpuActions[0].Action)
		assert.Equal(t, 10, out.CpuActions[0].Amount)
		assert.Equal(t, 2, out.CpuActions[1].PlayerIdx)
	})

	t.Run("round results with best hand", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
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
		assert.Equal(t, "", out.RoundResults[0].Kickers)
		assert.Equal(t, 200, out.RoundResults[0].WonAmount)
		assert.Len(t, out.RoundResults[0].BestHand, 2)
		assert.Equal(t, "SPADE", out.RoundResults[0].BestHand[0].Design)
	})

	t.Run("round results with kickers", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
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
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Equal(t, "A, K, Q", out.RoundResults[0].Kickers)
	})

	t.Run("best hand cards at showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhaseShowdown)
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
		h.SetPhase(domain.DramahaPhaseFlop)
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
		h.SetPhase(domain.DramahaPhasePreFlop)

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
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "dramaha.result.win", out.MessageCode)
	})

	t.Run("game end message - human loses", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "dramaha.result.lose", out.MessageCode)
	})

	t.Run("game end message - human folded", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetGameEndFlag(true)
		players[0].SetFolded(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 1, WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "dramaha.result.folded", out.MessageCode)
	})

	t.Run("game end message - no results", func(t *testing.T) {
		h, _ := setup()
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "dramaha.result.gameOver", out.MessageCode)
	})

	t.Run("pot and dealer fields", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseFlop)
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
		gameMock := new(interfaces.MockDramahaGame)
		gameMock.On("GetPhase").Return(domain.DramahaPhaseShowdown)
		gameMock.On("GetPot").Return(0)
		gameMock.On("GetDealerIdx").Return(0)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetLastBet").Return(0)
		gameMock.On("GetMinRaise").Return(0)
		gameMock.On("GetRaiseCount").Return(0)
		gameMock.On("GetCommunityCards").Return([]*domain.Card{})
		gameMock.On("GetSidePots").Return([]domain.SidePot{})
		gameMock.On("GetCpuActions").Return([]domain.HoldemCpuAction{})
		gameMock.On("GetRoundResults").Return([]domain.HoldemResult{})
		gameMock.On("GetPlayerCnt").Return(1)
		gameMock.On("GetConfig").Return(domain.DefaultDramahaConfig())
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

		player := domain.NewDramahaPlayer(false, domain.HoldemStyleTAG)
		player.SetHandRank(99) // out of range
		gameMock.On("GetPlayer", 0).Return(player)

		result := p.Output(gameMock, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Unknown", out.Players[0].HandName)
	})

	t.Run("getHandName negative returns Unknown", func(t *testing.T) {
		gameMock := new(interfaces.MockDramahaGame)
		gameMock.On("GetPhase").Return(domain.DramahaPhaseEnd)
		gameMock.On("GetPot").Return(0)
		gameMock.On("GetDealerIdx").Return(0)
		gameMock.On("GetCurrentTurn").Return(0)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetLastBet").Return(0)
		gameMock.On("GetMinRaise").Return(0)
		gameMock.On("GetRaiseCount").Return(0)
		gameMock.On("GetCommunityCards").Return([]*domain.Card{})
		gameMock.On("GetSidePots").Return([]domain.SidePot{})
		gameMock.On("GetCpuActions").Return([]domain.HoldemCpuAction{})
		gameMock.On("GetRoundResults").Return([]domain.HoldemResult{})
		gameMock.On("GetPlayerCnt").Return(1)
		gameMock.On("GetConfig").Return(domain.DefaultDramahaConfig())
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

		player := domain.NewDramahaPlayer(false, domain.HoldemStyleTAG)
		player.SetHandRank(-1) // negative
		gameMock.On("GetPlayer", 0).Return(player)

		result := p.Output(gameMock, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "Unknown", out.Players[0].HandName)
	})

	t.Run("playStyleName included in output", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.NotEmpty(t, out.Players[1].PlayStyleName)
	})

	t.Run("HUD stats included in output", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementTotalHands()
		players[0].IncrementVPIP()
		players[0].IncrementPFR()

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 2, out.Players[0].TotalHands)
		assert.Equal(t, 50, out.Players[0].VPIP)
		assert.Equal(t, 50, out.Players[0].PFR)
		assert.Equal(t, 0, out.Players[0].ThreeBet)
		assert.Equal(t, "-", out.Players[0].AF)
	})

	t.Run("HUD stats zero when no hands played", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.Players[0].TotalHands)
		assert.Equal(t, 0, out.Players[0].VPIP)
		assert.Equal(t, 0, out.Players[0].PFR)
		assert.Equal(t, 0, out.Players[0].ThreeBet)
		assert.Equal(t, "-", out.Players[0].AF)
	})

	t.Run("HUD 3Bet and AF normal values", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementThreeBetOpportunity()
		players[0].IncrementThreeBetOpportunity()
		players[0].IncrementThreeBet()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopBetRaise()
		players[0].IncrementPostFlopCall()

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 50, out.Players[0].ThreeBet) // 1*100/2=50
		assert.Equal(t, "3.0", out.Players[0].AF)    // 3/1=3.0
	})

	t.Run("HUD AF infinity", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].IncrementTotalHands()
		players[0].IncrementPostFlopBetRaise()

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "∞", out.Players[0].AF)
	})

	t.Run("tournament mode fields included in output", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		cfg := domain.DramahaConfig{
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
		var out controller.HoldemWebOutput
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
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.HandCount)
		assert.Equal(t, 5, out.SmallBlind)
		assert.Equal(t, 10, out.BigBlind)
		assert.False(t, out.TournamentMode)
		assert.Equal(t, 10, out.BlindLevelHands)
		assert.Equal(t, 200, out.BlindMultiplier)
	})
}

func TestDramahaWebPresenter_Output_RebuyAddonFields(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	setup := func() (*domain.Dramaha, []*domain.DramahaPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.DramahaPlayer{
			domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
			domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
		return h, players
	}

	t.Run("default values when rebuy/addon disabled", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
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
		h.SetPhase(domain.DramahaPhaseRebuy)
		cfg := domain.DramahaConfig{
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
		var out controller.HoldemWebOutput
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

func TestDramahaWebPresenter_Output_BettingLimitFields(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	setup := func() (*domain.Dramaha, []*domain.DramahaPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.DramahaPlayer{
			domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
			domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
		return h, players
	}

	t.Run("default Fixed limit", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 0, out.BettingLimit)
		assert.Equal(t, 0, out.RaiseCount)
		assert.Equal(t, 0, out.MaxBetAmount)
	})

	t.Run("tableSize reflects player count", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 4, out.TableSize)
	})

	t.Run("tableSize 6-max", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players6 := make([]*domain.DramahaPlayer, 6)
		players6[0] = domain.NewDramahaPlayer(true, domain.HoldemStyleTAG)
		for i := 1; i < 6; i++ {
			players6[i] = domain.NewDramahaPlayer(false, domain.HoldemStyleLAP)
		}
		h6 := domain.NewDramaha(tc, players6, domain.DefaultDramahaConfig())
		h6.SetPhase(domain.DramahaPhasePreFlop)
		players6[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players6[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(h6, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, 6, out.TableSize)
		assert.Equal(t, 6, len(out.Players))
	})
}

func TestDramahaWebPresenter_Output_MuckFields(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	setup := func() (*domain.Dramaha, []*domain.DramahaPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.DramahaPlayer{
			domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
			domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
		return h, players
	}

	t.Run("muckAvailable true when showdown and human lost", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.True(t, out.MuckAvailable)
	})

	t.Run("muckAvailable false when not showdown", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.False(t, out.MuckAvailable)
	})

	t.Run("mucked result: handRank=0 handName empty bestHand empty mucked=true", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", Kickers: []int{14, 13}, WonAmount: 0, Mucked: true, BestHand: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 5, false),
			}},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
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
		h.SetPhase(domain.DramahaPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "dramaha.muck.prompt", out.MessageCode)
	})

	t.Run("error takes priority over muck prompt", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseShowdown)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, BestHand: nil},
		})

		result := p.Output(h, errors.New("some error"))
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "some error", out.Message)
		assert.Equal(t, "", out.MessageCode)
	})

	t.Run("You mucked message in buildResultMessage", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true, BestHand: nil},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100, BestHand: nil},
		})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "dramaha.result.mucked", out.MessageCode)
	})
}

func TestDramahaWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockDramahaGame)
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
		mockGame := new(interfaces.MockDramahaGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockDramahaGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})
}

func TestDramahaWebPresenter_Equity(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	setup := func() (*domain.Dramaha, []*domain.DramahaPlayer) {
		tc := domain.NewTrumpCards(0)
		players := []*domain.DramahaPlayer{
			domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
			domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
			domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
		}
		h := domain.NewDramaha(tc, players, domain.DefaultDramahaConfig())
		return h, players
	}

	t.Run("equity populated during active phase", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhasePreFlop)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.NotNil(t, out.Equity)
		assert.Greater(t, out.Equity.WinProbability, 0.0)
		assert.NotEmpty(t, out.Equity.HandOdds)
		assert.NotNil(t, out.PotOdds)
	})

	t.Run("equity nil during showdown", func(t *testing.T) {
		h, players := setup()
		h.SetPhase(domain.DramahaPhaseShowdown)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Nil(t, out.Equity)
		assert.Nil(t, out.PotOdds)
	})

	t.Run("equity nil during end phase", func(t *testing.T) {
		h, _ := setup()
		h.SetPhase(domain.DramahaPhaseEnd)
		h.SetGameEndFlag(true)
		h.SetRoundResults([]domain.HoldemResult{})

		result := p.Output(h, nil)
		var out controller.HoldemWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Nil(t, out.Equity)
		assert.Nil(t, out.PotOdds)
	})
}

func TestDramahaWebPresenter_Output_WithMetaAIProfile(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)
	tc := domain.NewTrumpCards(0)
	players := []*domain.DramahaPlayer{
		domain.NewDramahaPlayer(true, domain.HoldemStyleTAG),
		domain.NewDramahaPlayer(false, domain.HoldemStyleLAP),
		domain.NewDramahaPlayer(false, domain.HoldemStyleTAP),
		domain.NewDramahaPlayer(false, domain.HoldemStyleGTO),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := domain.DefaultDramahaConfig()
	cfg.CpuMetaAI = true
	o := domain.NewDramaha(tc, players, cfg)
	_ = o.Reset()

	result := p.Output(o, nil)
	var out controller.HoldemWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.MetaAI)
	assert.True(t, out.MetaAI.Enabled)
	assert.NotNil(t, out.Profile)
}

// --- the split, which Dramaha applies to every pot ---

// TestDramahaWebPresenter_AlwaysReportsASplitPot replaces the clone's
// IsHiLo-flag test. There is no flag to propagate any more: Dramaha halves
// every pot, so the Web output must always tell the client to render two
// sides.
func TestDramahaWebPresenter_AlwaysReportsASplitPot(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	for _, phase := range []int{
		domain.DramahaPhasePreFlop, domain.DramahaPhaseFlop, domain.DramahaPhaseDraw,
		domain.DramahaPhaseTurn, domain.DramahaPhaseRiver, domain.DramahaPhaseEnd,
	} {
		o := domain.NewDefaultDramaha()
		o.SetPhase(phase)

		var out controller.HoldemWebOutput
		if err := json.Unmarshal([]byte(p.Output(o, nil)), &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		assert.True(t, out.IsHiLo, "phase %d: the client has to render both sides of the split", phase)
	}
}

// TestDramahaWebPresenter_SplitResultMessages pins the message codes the
// client keys off. The three outcomes have to be distinguishable: taking both
// halves is not the same event as taking one.
//
// The codes below are the ones the presenter emits today. They live in the
// `dramahahilo` namespace inherited from the clone, and the Web locale defines
// no Dramaha entries at all -- both reported separately. What this test
// guarantees is that the three outcomes stay distinct and stay stable.
func TestDramahaWebPresenter_SplitResultMessages(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)

	cases := []struct {
		name     string
		hi       int
		lo       int
		wantCode string
	}{
		{"both halves", 50, 50, "dramahahilo.result.scoop"},
		{"omaha half only", 50, 0, "dramahahilo.result.hiWin"},
		{"draw half only", 0, 50, "dramahahilo.result.lowWin"},
	}
	seen := map[string]string{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := domain.NewDefaultDramaha()
			o.SetPhase(domain.DramahaPhaseEnd)
			o.SetGameEndFlag(true)
			o.SetRoundResults([]domain.HoldemResult{
				{
					PlayerIdx:    0,
					HandRank:     domain.PokerHandStraight,
					HandName:     "Straight",
					BestHand:     []*domain.Card{},
					WonAmount:    tc.hi + tc.lo,
					HiWonAmount:  tc.hi,
					LowWonAmount: tc.lo,
					LowQualifies: true,
				},
			})

			var out controller.HoldemWebOutput
			if err := json.Unmarshal([]byte(p.Output(o, nil)), &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			assert.Equal(t, tc.wantCode, out.MessageCode)
			if prev, dup := seen[out.MessageCode]; dup {
				t.Errorf("%q and %q share the message code %q", prev, tc.name, out.MessageCode)
			}
			seen[out.MessageCode] = tc.name
		})
	}
	assert.Len(t, seen, len(cases), "each split outcome needs its own code")
}

// TestDramahaWebPresenter_LosingHumanGetsNoSplitCode: with nothing won there is
// no half to announce, so the generic result codes take over.
func TestDramahaWebPresenter_LosingHumanGetsNoSplitCode(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)
	o := domain.NewDefaultDramaha()
	o.SetPhase(domain.DramahaPhaseEnd)
	o.SetGameEndFlag(true)
	o.SetRoundResults([]domain.HoldemResult{
		{PlayerIdx: 1, HandRank: domain.PokerHandStraight, HandName: "Straight", WonAmount: 100, HiWonAmount: 50, LowWonAmount: 50},
	})

	var out controller.HoldemWebOutput
	if err := json.Unmarshal([]byte(p.Output(o, nil)), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assert.Equal(t, "dramaha.result.lose", out.MessageCode)
}

// TestDramahaWebPresenter_RoundResultsCarryBothHalves asserts the per-result
// split fields survive into the Web payload: the amounts for each half and the
// five cards the draw side was judged on.
//
// NOTE: the same information is NOT available per seat -- DramahaPlayer no
// longer implements the presenter's low-hand interface, so Players[i].
// LowBestHand is always empty. That is a gap, reported rather than asserted.
func TestDramahaWebPresenter_RoundResultsCarryBothHalves(t *testing.T) {
	p := new(presenter.DramahaWebPresenter)
	o := domain.NewDefaultDramaha()
	o.SetPhase(domain.DramahaPhaseEnd)

	drawHand := []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignHeart, 3, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
	}
	o.SetRoundResults([]domain.HoldemResult{
		{
			PlayerIdx:    0,
			HandRank:     domain.PokerHandHighCard,
			HandName:     "High Card",
			BestHand:     []*domain.Card{},
			WonAmount:    50,
			HiWonAmount:  0,
			LowWonAmount: 50,
			LowQualifies: true,
			LowBestHand:  drawHand,
		},
	})

	var out controller.HoldemWebOutput
	if err := json.Unmarshal([]byte(p.Output(o, nil)), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if assert.Len(t, out.RoundResults, 1) {
		r := out.RoundResults[0]
		assert.Equal(t, 0, r.HiWonAmount, "this seat lost the Omaha half")
		assert.Equal(t, 50, r.LowWonAmount, "and won the draw half")
		assert.True(t, r.LowQualifies, "five cards always rank, so the draw side always counts")
		assert.Len(t, r.LowBestHand, domain.DramahaHoleCards,
			"the draw side is judged on the whole hole, so all five must be shown")
	}
}
