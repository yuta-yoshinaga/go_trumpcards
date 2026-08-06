//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupGinRummyWebMock() *interfaces.MockGinRummyGame {
	m := new(interfaces.MockGinRummyGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(31)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.GinRummyPhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultGinRummyConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetKnockerIdx").Return(-1)
	m.On("GetKnockerMelds").Return(([][]*domain.Card)(nil))
	m.On("LayoffTargets", mock.Anything).Return(([]int)(nil)).Maybe()
	m.On("GetKnockerDeadwood").Return(([]*domain.Card)(nil))
	m.On("GetIsGin").Return(false)
	return m
}

func setupGinRummyWebMockWithPlayers() (*interfaces.MockGinRummyGame, []*domain.GinRummyPlayer) {
	m := setupGinRummyWebMock()
	players := makeGinRummyPlayers()
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestGinRummyWebPresenter_Output(t *testing.T) {
	p := new(presenter.GinRummyWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupGinRummyWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.GinRummyWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 0, resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 31, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Equal(t, -1, resObj.KnockerIdx)
		assert.False(t, resObj.IsGin)
		assert.Nil(t, resObj.DiscardTop)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("human cards shown, CPU cards hidden in draw phase", func(t *testing.T) {
		m, players := setupGinRummyWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 1, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 1)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 5, humanPlayer.Cards[0].Value)

		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 1, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0)
	})

	t.Run("CPU cards shown in round end phase", func(t *testing.T) {
		m, players := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("CPU cards shown in game end phase", func(t *testing.T) {
		m, players := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.GinRummyPhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("CPU cards shown in layoff phase", func(t *testing.T) {
		m, players := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseLayoff)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		cpu1 := resObj.Players[1]
		assert.Len(t, cpu1.Cards, 1)
	})

	t.Run("player scores", func(t *testing.T) {
		m, players := setupGinRummyWebMockWithPlayers()
		players[1].SetCumulativeScore(200)
		players[1].SetRoundScore(50)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 200, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 50, resObj.Players[1].RoundScore)
	})

	t.Run("discard top populated", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, "HEART", resObj.DiscardTop.Design)
		assert.Equal(t, 7, resObj.DiscardTop.Value)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.GinRummyConfig{
			CpuDifficulty: domain.GinRummyCpuDifficultyHard,
			PointLimit:    200,
		})

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.GinRummyCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 200, resObj.Config.PointLimit)
	})

	t.Run("knocker melds and deadwood", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerMelds")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnockerDeadwood")
		m.On("GetKnockerIdx").Return(0)
		m.On("GetKnockerMelds").Return([][]*domain.Card{
			{domain.NewCard(domain.CardDesignSpade, 1, false), domain.NewCard(domain.CardDesignSpade, 2, false), domain.NewCard(domain.CardDesignSpade, 3, false)},
		})
		m.On("GetKnockerDeadwood").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 5, false)})

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 0, resObj.KnockerIdx)
		assert.Len(t, resObj.KnockerMelds, 1)
		assert.Len(t, resObj.KnockerMelds[0].Cards, 3)
		assert.Len(t, resObj.KnockerDeadwood, 1)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		m.On("GetPhase").Return(domain.GinRummyPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "ginrummy.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(1)
		m.On("GetPhase").Return(domain.GinRummyPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 1")
		assert.Equal(t, "ginrummy.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "1"}, resObj.MessageParams)
	})

	t.Run("game end nil player at winnerIdx", func(t *testing.T) {
		m := setupGinRummyWebMock()
		m.On("GetPlayerCnt").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.GinRummyPlayer)(nil))
		m.On("GetPhase").Return(domain.GinRummyPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 99")
		assert.Equal(t, "ginrummy.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})

	t.Run("draw phase messageCode", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "ginrummy.drawPhase", resObj.MessageCode)
	})

	t.Run("discard phase messageCode", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseDiscard)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "ginrummy.discardPhase", resObj.MessageCode)
	})

	t.Run("layoff phase messageCode", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseLayoff)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "ginrummy.layoffPhase", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "ginrummy.roundEnd", resObj.MessageCode)
	})

	t.Run("unrecognized phase no messageCode", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.GinRummyPhaseGameEnd)

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupGinRummyWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.GinRummyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.GinRummyCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 100, resObj.Config.PointLimit)
	})
}

// **レイオフフェーズの主題そのもの。**ディスカードフェーズは meldedIndices で
// メルド/デッドウッドを見せているのに、レイオフには補助が無かった (#4823)。
func TestGinRummyWebPresenter_LayoffTargets(t *testing.T) {
	m, players := setupGinRummyWebMockWithPlayers()
	for players[0].GetCardsSize() > 0 {
		players[0].RemoveCard(0)
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "LayoffTargets")
	m.On("LayoffTargets", mock.MatchedBy(func(c *domain.Card) bool {
		return c != nil && c.GetValue() == 7
	})).Return([]int{0})
	m.On("LayoffTargets", mock.Anything).Return(([]int)(nil))

	var out controller.GinRummyWebOutput
	assert.NoError(t, json.Unmarshal([]byte(new(presenter.GinRummyWebPresenter).Output(m, nil)), &out))

	// 足せる札には番号が付き、足せない札は空配列 (null ではない)。
	assert.Equal(t, [][]int{{0}, {}}, out.LayoffTargets)
}

func TestGinRummyWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GinRummyWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockGinRummyGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw_stock", Detail: "drew from stock", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"draw_stock"`)
		assert.Contains(t, result, `"detail":"drew from stock"`)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockGinRummyGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockGinRummyGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}
