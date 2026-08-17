//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSevenBridgeWebMock() *interfaces.MockSevenBridgeGame {
	m := new(interfaces.MockSevenBridgeGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(37)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.SevenBridgePhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultSevenBridgeConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetClaimedThisTurn").Return(false).Maybe()
	return m
}

func setupSevenBridgeWebMockWithPlayers() (*interfaces.MockSevenBridgeGame, []*domain.SevenBridgePlayer) {
	m := setupSevenBridgeWebMock()
	players := makeSevenBridgePlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func unmarshalSevenBridge(t *testing.T, s string) controller.SevenBridgeWebOutput {
	t.Helper()
	var out controller.SevenBridgeWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return out
}

func TestSevenBridgeWebPresenter_Output(t *testing.T) {
	p := new(presenter.SevenBridgeWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupSevenBridgeWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.Len(t, out.Players, 2)
		assert.Equal(t, 1, out.RoundNumber)
		assert.Equal(t, 37, out.DrawPileCount)
		assert.Equal(t, -1, out.WinnerIdx)
		assert.Equal(t, -1, out.RoundWinnerIdx)
		assert.Nil(t, out.DiscardTop)
		assert.Equal(t, "sevenbridge.drawPhase", out.MessageCode)
	})

	t.Run("cards hidden in draw phase for CPU", func(t *testing.T) {
		m, players := setupSevenBridgeWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.Len(t, out.Players[0].Cards, 1)
		assert.Equal(t, "SPADE", out.Players[0].Cards[0].Design)
		assert.Equal(t, 1, out.Players[1].CardCount)
		assert.Len(t, out.Players[1].Cards, 0)
	})

	t.Run("cards revealed in round end", func(t *testing.T) {
		m, players := setupSevenBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SevenBridgePhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.Len(t, out.Players[1].Cards, 1)
		assert.Equal(t, "sevenbridge.roundEnd", out.MessageCode)
	})

	t.Run("cards revealed in game end", func(t *testing.T) {
		m, players := setupSevenBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.SevenBridgePhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.Len(t, out.Players[1].Cards, 1)
		assert.Equal(t, 0, out.WinnerIdx)
	})

	t.Run("play phase message code", func(t *testing.T) {
		m, _ := setupSevenBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.SevenBridgePhasePlay)

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.Equal(t, "sevenbridge.playPhase", out.MessageCode)
	})

	t.Run("discard top present", func(t *testing.T) {
		m, _ := setupSevenBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.NotNil(t, out.DiscardTop)
		assert.Equal(t, 7, out.DiscardTop.Value)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupSevenBridgeWebMockWithPlayers()
		out := unmarshalSevenBridge(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("game end message", func(t *testing.T) {
		m, _ := setupSevenBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.NotEmpty(t, out.MessageCode)
	})

	t.Run("meld serialisation", func(t *testing.T) {
		m, players := setupSevenBridgeWebMockWithPlayers()
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignClover, 3, false),
			domain.NewCard(domain.CardDesignHeart, 3, false),
		})

		out := unmarshalSevenBridge(t, p.Output(m, nil))
		assert.Len(t, out.Players[0].Melds, 1)
		assert.Len(t, out.Players[0].Melds[0].Cards, 3)
	})
}

func TestSevenBridgeWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SevenBridgeWebPresenter)

	m := new(interfaces.MockSevenBridgeGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "You draws from stock"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "draw_stock")
}

func TestSevenBridgeWebPresenter_HintOutput(t *testing.T) {
	// Web hints are client-side, so HintOutput mirrors Output.
	p := new(presenter.SevenBridgeWebPresenter)
	m, _ := setupSevenBridgeWebMockWithPlayers()
	assert.Equal(t, p.Output(m, nil), p.HintOutput(m))
}

// #5547: 割り込みで取ったターンかどうかは保存までされているのに、レスポンスに
// 載っていなかった。ページはこの値を読んでバッジを出す。
func TestSevenBridgeWebPresenter_Output_ClaimedThisTurn(t *testing.T) {
	pres := new(presenter.SevenBridgeWebPresenter)

	outputWith := func(claimed bool) controller.SevenBridgeWebOutput {
		m, _ := setupSevenBridgeWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetClaimedThisTurn")
		m.On("GetClaimedThisTurn").Return(claimed)
		var out controller.SevenBridgeWebOutput
		require.NoError(t, json.Unmarshal([]byte(pres.Output(m, nil)), &out))
		return out
	}

	assert.True(t, outputWith(true).ClaimedThisTurn)
	assert.False(t, outputWith(false).ClaimedThisTurn)
}
