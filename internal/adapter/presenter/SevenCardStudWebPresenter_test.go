//go:build test

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

func makeSevenCardStudForPresenter() (*domain.SevenCardStud, []*domain.SevenCardStudPlayer) {
	tc := domain.NewTrumpCards(0)
	players := []*domain.SevenCardStudPlayer{
		domain.NewSevenCardStudPlayer(true, domain.HoldemStyleTAG),
		domain.NewSevenCardStudPlayer(false, domain.HoldemStyleLAP),
		domain.NewSevenCardStudPlayer(false, domain.HoldemStyleTAP),
		domain.NewSevenCardStudPlayer(false, domain.HoldemStyleGTO),
	}
	s := domain.NewSevenCardStud(tc, players, domain.DefaultSevenCardStudConfig())
	return s, players
}

func TestSevenCardStudWebPresenter_Output(t *testing.T) {
	p := new(presenter.SevenCardStudWebPresenter)

	setup := func() (*domain.SevenCardStud, []*domain.SevenCardStudPlayer) {
		return makeSevenCardStudForPresenter()
	}

	t.Run("initial state", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 5, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Equal(t, domain.SevenCardStudPhaseThirdStreet, out.Phase)
		assert.Equal(t, 4, len(out.Players))
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, "", out.Message)
		assert.Nil(t, out.CommunityCard)
		assert.Len(t, out.SidePots, 0)
		assert.Len(t, out.CpuActions, 0)
		assert.Len(t, out.RoundResults, 0)
	})

	t.Run("human hole cards visible", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		players[0].AddHoleCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.True(t, human.IsHuman)
		assert.Len(t, human.HoleCards, 2)
		assert.Equal(t, "SPADE", human.HoleCards[0].Design)
		assert.Equal(t, 10, human.HoleCards[0].Value)
		assert.Equal(t, "HEART", human.HoleCards[1].Design)
	})

	t.Run("human door cards visible", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].AddDoorCard(domain.NewCard(domain.CardDesignClover, 5, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		human := out.Players[0]
		assert.Len(t, human.DoorCards, 1)
		assert.Equal(t, "CLOVER", human.DoorCards[0].Design)
	})

	t.Run("CPU hole cards hidden before showdown", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[1].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.False(t, cpu.IsHuman)
		assert.Len(t, cpu.HoleCards, 0)
	})

	t.Run("CPU door cards always visible", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[1].AddDoorCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.DoorCards, 1)
		assert.Equal(t, "HEART", cpu.DoorCards[0].Design)
	})

	t.Run("CPU cards visible at showdown", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		players[1].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetHandRank(domain.PokerHandOnePair)

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.HoleCards, 1)
		assert.Equal(t, domain.PokerHandOnePair, cpu.HandRank)
		assert.Equal(t, "One Pair", cpu.HandName)
	})

	t.Run("CPU cards visible at end phase", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		players[1].AddHoleCard(domain.NewCard(domain.CardDesignHeart, 7, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.HoleCards, 1)
	})

	t.Run("folded CPU cards hidden at showdown", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		players[1].AddHoleCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].SetFolded(true)

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Len(t, cpu.HoleCards, 0)
		assert.Equal(t, 0, cpu.HandRank)
		assert.Equal(t, "", cpu.HandName)
	})

	t.Run("community card present when card shortage", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseSeventhStreet)
		s.SetCommunityCard(domain.NewCard(domain.CardDesignDiamond, 9, false))

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.NotNil(t, out.CommunityCard)
		assert.Equal(t, "DIAMOND", out.CommunityCard.Design)
		assert.Equal(t, 9, out.CommunityCard.Value)
	})

	t.Run("side pots", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetSidePots([]domain.SidePot{
			{Amount: 100, EligiblePlayers: []int{0, 1}},
			{Amount: 50, EligiblePlayers: []int{0}},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.SidePots, 2)
		assert.Equal(t, 100, out.SidePots[0].Amount)
		assert.Equal(t, []int{0, 1}, out.SidePots[0].EligiblePlayers)
	})

	t.Run("player fields", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		players[0].SetChips(500)
		players[0].SetCurrentBet(20)
		players[2].SetFolded(true)
		players[3].SetAllIn(true)

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 0, out.Players[0].ID)
		assert.True(t, out.Players[0].IsHuman)
		assert.Equal(t, 500, out.Players[0].Chips)
		assert.Equal(t, 20, out.Players[0].CurrentBet)
		assert.True(t, out.Players[2].Folded)
		assert.True(t, out.Players[3].AllIn)
	})

	t.Run("CPU actions", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetCpuActions([]domain.SevenCardStudCpuAction{
			{PlayerIdx: 1, Action: domain.SevenCardStudActionCall, Amount: 10},
			{PlayerIdx: 2, Action: domain.SevenCardStudActionFold, Amount: 0},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.CpuActions, 2)
		assert.Equal(t, 1, out.CpuActions[0].PlayerIdx)
		assert.Equal(t, domain.SevenCardStudActionCall, out.CpuActions[0].Action)
	})

	t.Run("round results", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{
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

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Len(t, out.RoundResults, 1)
		assert.Equal(t, domain.PokerHandFlush, out.RoundResults[0].HandRank)
		assert.Equal(t, "Flush", out.RoundResults[0].HandName)
		assert.Equal(t, 200, out.RoundResults[0].WonAmount)
		assert.Len(t, out.RoundResults[0].BestHand, 2)
	})

	t.Run("round results with kickers", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{
				PlayerIdx: 0,
				HandRank:  domain.PokerHandOnePair,
				HandName:  "One Pair",
				Kickers:   []int{14, 13, 12},
				WonAmount: 200,
			},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "A, K, Q", out.RoundResults[0].Kickers)
	})

	t.Run("error message", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)

		result := p.Output(s, errors.New("test error"))
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, "test error", out.Message)
	})

	t.Run("game end message - win", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, WonAmount: 100},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "sevencardstud.result.win", out.MessageCode)
	})

	t.Run("game end message - lose", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, WonAmount: 0},
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "sevencardstud.result.lose", out.MessageCode)
	})

	t.Run("game end message - folded", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		players[0].SetFolded(true)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "sevencardstud.result.folded", out.MessageCode)
	})

	t.Run("game end message - mucked", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, WonAmount: 0, Mucked: true},
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "sevencardstud.result.mucked", out.MessageCode)
	})

	t.Run("game end message - no results", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "sevencardstud.result.gameOver", out.MessageCode)
	})

	t.Run("muck prompt", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, WonAmount: 0},
			{PlayerIdx: 1, WonAmount: 100},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Empty(t, out.Message)
		assert.Equal(t, "sevencardstud.muck.prompt", out.MessageCode)
	})

	t.Run("config fields in output", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 1, out.Ante)
		assert.Equal(t, 2, out.BringIn)
		assert.Equal(t, 5, out.SmallBet)
		assert.Equal(t, 10, out.BigBet)
		assert.Equal(t, 4, out.TableSize)
	})

	t.Run("bring-in player index", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
		s.SetBringInPlayerIdx(2)

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.Equal(t, 2, out.BringInPlayerIdx)
	})

	t.Run("mucked result hides hand info", func(t *testing.T) {
		s, _ := setup()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{
			{PlayerIdx: 0, HandRank: domain.PokerHandOnePair, HandName: "One Pair", WonAmount: 0, Mucked: true},
			{PlayerIdx: 1, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100},
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		assert.True(t, out.RoundResults[0].Mucked)
		assert.Equal(t, 0, out.RoundResults[0].HandRank)
		assert.Equal(t, "", out.RoundResults[0].HandName)
	})

	t.Run("best hand at showdown", func(t *testing.T) {
		s, players := setup()
		s.SetPhase(domain.SevenCardStudPhaseShowdown)
		players[1].SetHandRank(domain.PokerHandFullHouse)
		players[1].SetBestHand([]*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 3, false),
			domain.NewCard(domain.CardDesignDiamond, 3, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		})

		result := p.Output(s, nil)
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(result), &out)

		cpu := out.Players[1]
		assert.Equal(t, "Full House", cpu.HandName)
		assert.Len(t, cpu.BestHand, 3)
	})
}

func TestSevenCardStudWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SevenCardStudWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockSevenCardStudGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "raise", Detail: "raised to 100"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "raise")
		mockGame.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		mockGame := new(interfaces.MockSevenCardStudGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)
		assert.Contains(t, result, "entries")
		mockGame.AssertExpectations(t)
	})
}

func TestSevenCardStudWebPresenter_HintOutput(t *testing.T) {
	s, _ := makeSevenCardStudForPresenter()
	s.SetPhase(domain.SevenCardStudPhaseThirdStreet)
	p := new(presenter.SevenCardStudWebPresenter)
	// The web presenter computes hints client-side, so HintOutput mirrors Output.
	assert.Equal(t, p.Output(s, nil), p.HintOutput(s))
}

// **#5435: ページはルート名からシカゴだと推測してはいけない。** 3 つのモードが
// 同じページを共有しているので、割れた内訳を出すかどうかはサーバーが立てる
// `isChicago` だけで決まる。スペードの札と取り分がここで落ちると、Web には
// 「ポットが半分だけ入った」理由がどこにも出ない。
func TestSevenCardStudWebPresenter_Output_Chicago(t *testing.T) {
	p := new(presenter.SevenCardStudWebPresenter)

	newChicago := func() *domain.SevenCardStud {
		cfg := domain.DefaultSevenCardStudConfig()
		return domain.NewSevenCardStudChicago(domain.NewTrumpCards(0),
			domain.NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
	}

	outputOf := func(s *domain.SevenCardStud) controller.SevenCardStudWebOutput {
		var out controller.SevenCardStudWebOutput
		_ = json.Unmarshal([]byte(p.Output(s, nil)), &out)
		return out
	}

	t.Run("flags the mode so the shared page can branch", func(t *testing.T) {
		assert.True(t, outputOf(newChicago()).IsChicago)
		plain, _ := makeSevenCardStudForPresenter()
		assert.False(t, outputOf(plain).IsChicago)
	})

	t.Run("carries the spade share and the deciding card", func(t *testing.T) {
		s := newChicago()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{{
			PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
			WonAmount: 100, WonSpade: 50,
			SpadeCard: domain.NewCard(domain.CardDesignSpade, 1, false),
		}})

		r := outputOf(s).RoundResults[0]
		assert.Equal(t, 50, r.WonSpade)
		if assert.NotNil(t, r.SpadeCard) {
			assert.Equal(t, "SPADE", r.SpadeCard.Design)
			assert.Equal(t, 1, r.SpadeCard.Value)
		}
	})

	// 伏せ札にスペードが無かった席は札を持たない。省略されるので `omitempty`
	// のまま nil で届く。
	t.Run("leaves the card unset when nobody held a spade", func(t *testing.T) {
		s := newChicago()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{{
			PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush", WonAmount: 100,
		}})

		r := outputOf(s).RoundResults[0]
		assert.Nil(t, r.SpadeCard)
		assert.Equal(t, 0, r.WonSpade)
	})

	// マックした席は手の情報を一切出さない。スペードもその 1 枚なので同じ扱い。
	t.Run("hides the spade of a mucked hand", func(t *testing.T) {
		s := newChicago()
		s.SetPhase(domain.SevenCardStudPhaseEnd)
		s.SetGameEndFlag(true)
		s.SetRoundResults([]domain.SevenCardStudResult{{
			PlayerIdx: 0, HandRank: domain.PokerHandFlush, HandName: "Flush",
			WonAmount: 0, WonSpade: 0, Mucked: true,
			SpadeCard: domain.NewCard(domain.CardDesignSpade, 1, false),
		}})

		r := outputOf(s).RoundResults[0]
		assert.True(t, r.Mucked)
		assert.Nil(t, r.SpadeCard)
	})
}
