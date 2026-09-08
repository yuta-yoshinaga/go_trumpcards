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

func setupTongitsWebMock() *interfaces.MockTongitsGame {
	m := new(interfaces.MockTongitsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(41)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.TongitsPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultTongitsConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetIsTongits").Return(false)
	m.On("IsHumanTurn").Return(false).Maybe()

	return m
}

func setupTongitsWebMockWithPlayers() (*interfaces.MockTongitsGame, []*domain.TongitsPlayer) {
	m := setupTongitsWebMock()
	players := makeTongitsPlayers()
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestTongitsWebPresenter_Output(t *testing.T) {
	p := new(presenter.TongitsWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.TongitsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 41, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, -1, resObj.RemainingPoints)
		assert.False(t, resObj.IsTongits)
		assert.Nil(t, resObj.DiscardTop)
	})

	t.Run("human cards shown, CPU cards hidden in draw phase", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Len(t, resObj.Players[0].Cards, 1)
		assert.Len(t, resObj.Players[1].Cards, 0)
	})

	t.Run("CPU cards shown in round end", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
	})

	t.Run("CPU cards shown in game end", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.TongitsPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
		assert.Equal(t, 0, resObj.WinnerIdx)
		assert.NotEmpty(t, resObj.MessageCode)
	})

	t.Run("error in lastErr", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		testErr := errors.New("oops")
		result := p.Output(m, testErr)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "oops", resObj.Message)
	})

	t.Run("player melds serialized", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		players[1].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[1].Melds, 1)
		assert.Len(t, resObj.Players[1].Melds[0].Cards, 3)
	})

	t.Run("discard top serialized", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 8, false))

		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, 8, resObj.DiscardTop.Value)
	})

	t.Run("draw phase message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.drawPhase", resObj.MessageCode)
	})

	t.Run("discard phase message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard)
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.discardPhase", resObj.MessageCode)
	})

	t.Run("round end with tongits on deal message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetIsTongits")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		m.On("GetIsTongits").Return(true)
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.tongitsOnDeal", resObj.MessageCode)
	})

	t.Run("round end normal message code", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseRoundEnd)
		result := p.Output(m, nil)
		var resObj controller.TongitsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "tongits.roundEnd", resObj.MessageCode)
	})
}

func TestTongitsWebPresenter_RemainingPoints(t *testing.T) {
	p := new(presenter.TongitsWebPresenter)

	decode := func(t *testing.T, m *interfaces.MockTongitsGame) controller.TongitsWebOutput {
		t.Helper()
		var out controller.TongitsWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		return out
	}

	t.Run("human discard turn carries remaining points", func(t *testing.T) {
		m, players := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard)
		m.On("IsHumanTurn").Return(true)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 4, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		out := decode(t, m)
		assert.Equal(t, 9, out.RemainingPoints)
	})

	// **-1 は「まだ聞くべき場面でない」印。**0 にすると「デッドウッド0 =
	// 必ずノック可能」と読めてしまい、ドローフェーズで誤った案内が出る。
	t.Run("outside the human discard turn it is -1, not 0", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "IsHumanTurn")
		m.On("IsHumanTurn").Return(true) // フェーズは Draw のまま

		assert.Equal(t, -1, decode(t, m).RemainingPoints)
	})

	t.Run("cpu discard turn is -1 too", func(t *testing.T) {
		m, _ := setupTongitsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.TongitsPhaseDiscard) // IsHumanTurn は既定 false

		assert.Equal(t, -1, decode(t, m).RemainingPoints)
	})
}

func TestTongitsWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TongitsWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockTongitsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "challenge", Detail: "challenges"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "challenge")
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockTongitsGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.NotEmpty(t, result)
	})
}
