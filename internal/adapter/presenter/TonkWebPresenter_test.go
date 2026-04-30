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

func setupTonkWebMock() *interfaces.MockTonkGame {
	m := new(interfaces.MockTonkGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(41)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TonkPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultTonkConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetKnockerIdx").Return(-1)
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil))
	m.On("GetKnockerDeadwood").Return(([]*domain.Card)(nil))
	m.On("GetOpponentMelds").Return(([][]*domain.Card)(nil))
	m.On("GetOpponentDeadwood").Return(([]*domain.Card)(nil))
	m.On("GetIsTonk").Return(false)
	m.On("GetIsUndercut").Return(false)
	return m
}

func setupTonkWebMockWithPlayers() (*interfaces.MockTonkGame, []*domain.TonkPlayer) {
	m := setupTonkWebMock()
	players := makeTonkPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestTonkWebPresenter_Output(t *testing.T) {
	p := new(presenter.TonkWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupTonkWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.TonkWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 41, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, -1, resObj.KnockerIdx)
		assert.False(t, resObj.IsTonk)
		assert.Nil(t, resObj.DiscardTop)
	})

	t.Run("human cards shown, CPU cards hidden in draw phase", func(t *testing.T) {
		m, players := setupTonkWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("CPU cards shown in round end", func(t *testing.T) {
		m, players := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
	})

	t.Run("CPU cards shown in game end", func(t *testing.T) {
		m, players := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.TonkPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
		assert.Equal(t, 0, resObj.WinnerIdx)
		assert.NotEmpty(t, resObj.MessageCode)
	})

	t.Run("error in lastErr", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		testErr := errors.New("oops")
		result := p.Output(m, testErr)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "oops", resObj.Message)
	})

	t.Run("knocker melds and deadwood serialized", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerMelds")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerDeadwood")
		m.On("GetKnockerIdx").Return(0)
		melds := [][]*domain.Card{
			{
				domain.NewCard(domain.CardDesignSpade, 5, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
				domain.NewCard(domain.CardDesignDiamond, 5, false),
			},
		}
		m.On("GetKnockerMelds").Return(melds)
		m.On("GetKnockerDeadwood").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignClover, 7, false),
		})

		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 0, resObj.KnockerIdx)
		assert.Len(t, resObj.KnockerMelds, 1)
		assert.Len(t, resObj.KnockerMelds[0].Cards, 3)
		assert.Len(t, resObj.KnockerDeadwood, 1)
	})

	t.Run("discard top serialized", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 8, false))

		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, 8, resObj.DiscardTop.Value)
	})

	t.Run("draw phase message code", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tonk.drawPhase", resObj.MessageCode)
	})

	t.Run("discard phase message code", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tonk.discardPhase", resObj.MessageCode)
	})

	t.Run("round end with tonk on deal message code", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsTonk")
		m.On("GetPhase").Return(domain.TonkPhaseRoundEnd)
		m.On("GetIsTonk").Return(true)
		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tonk.tonkOnDeal", resObj.MessageCode)
	})

	t.Run("round end normal message code", func(t *testing.T) {
		m, _ := setupTonkWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TonkPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.TonkWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tonk.roundEnd", resObj.MessageCode)
	})
}

func TestTonkWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TonkWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockTonkGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "knock", Detail: "knocks"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "knock")
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockTonkGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
	})
}
