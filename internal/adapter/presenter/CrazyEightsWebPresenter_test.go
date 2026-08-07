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

func setupCrazyEightsWebMock() *interfaces.MockCrazyEightsGame {
	m := new(interfaces.MockCrazyEightsGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(30)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetChosenSuit").Return(-1)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CrazyEightsPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultCrazyEightsConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupCrazyEightsWebMockWithPlayers() (*interfaces.MockCrazyEightsGame, []*domain.CrazyEightsPlayer) {
	m := setupCrazyEightsWebMock()
	players := makeCrazyEightsPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestCrazyEightsWebPresenter_Output(t *testing.T) {
	p := new(presenter.CrazyEightsWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCrazyEightsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		assert.NotEmpty(t, result)

		var resObj controller.CrazyEightsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentPlayerIdx)
		assert.Equal(t, 0, resObj.Phase) // CrazyEightsPhasePlay
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, 30, resObj.DrawPileCount)
		assert.Equal(t, -1, resObj.ChosenSuit)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Nil(t, resObj.DiscardTop)
		assert.Equal(t, "", resObj.Message)
	})

	t.Run("human cards shown, CPU cards hidden", func(t *testing.T) {
		m, players := setupCrazyEightsWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
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

	t.Run("player scores", func(t *testing.T) {
		m, players := setupCrazyEightsWebMockWithPlayers()
		players[1].SetCumulativeScore(200)
		players[1].SetRoundScore(50)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, 200, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 50, resObj.Players[1].RoundScore)
	})

	t.Run("discard top populated", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		top := domain.NewCard(domain.CardDesignHeart, 7, false)
		m.On("GetDiscardTop").Return(top)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, "HEART", resObj.DiscardTop.Design)
		assert.Equal(t, 7, resObj.DiscardTop.Value)
	})

	t.Run("chosen suit populated", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetChosenSuit")
		m.On("GetChosenSuit").Return(domain.CardDesignHeart)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, domain.CardDesignHeart, resObj.ChosenSuit)
	})

	t.Run("config values", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.CrazyEightsConfig{
			CpuDifficulty: domain.CrazyEightsCpuDifficultyHard,
			PointLimit:    300,
		})

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.CrazyEightsCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 300, resObj.Config.PointLimit)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()

		result := p.Output(m, errors.New("test error"))
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "test error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end human wins", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "あなた")
		assert.Equal(t, "crazyeights.result.humanWin", resObj.MessageCode)
		assert.Nil(t, resObj.MessageParams)
	})

	t.Run("game end CPU wins", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 2")
		assert.Equal(t, "crazyeights.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "2"}, resObj.MessageParams)
	})

	t.Run("game end nil player at winnerIdx", func(t *testing.T) {
		m := setupCrazyEightsWebMock()
		m.On("GetPlayerCnt").Return(0)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(99)
		m.On("GetPlayer", 99).Return((*domain.CrazyEightsPlayer)(nil))

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)

		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "CPU 99")
		assert.Equal(t, "crazyeights.result.cpuWin", resObj.MessageCode)
		assert.Equal(t, map[string]string{"cpuId": "99"}, resObj.MessageParams)
	})

	t.Run("play phase messageCode", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "crazyeights.playPhase", resObj.MessageCode)
	})

	t.Run("choose suit phase messageCode", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CrazyEightsPhaseChooseSuit)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "crazyeights.chooseSuitPhase", resObj.MessageCode)
	})

	t.Run("round end messageCode", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CrazyEightsPhaseRoundEnd)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "crazyeights.roundEnd", resObj.MessageCode)
	})

	t.Run("unrecognized phase no messageCode", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CrazyEightsPhaseGameEnd)
		// GetGameEndFlag remains false (default)

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Empty(t, resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("error takes priority over phase message", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()

		result := p.Output(m, errors.New("some error"))
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, "some error", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("default config values", func(t *testing.T) {
		m, _ := setupCrazyEightsWebMockWithPlayers()

		result := p.Output(m, nil)
		var resObj controller.CrazyEightsWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.Equal(t, int(domain.CrazyEightsCpuDifficultyNormal), resObj.Config.CpuDifficulty)
		assert.Equal(t, 200, resObj.Config.PointLimit)
	})
}

func TestCrazyEightsWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CrazyEightsWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockCrazyEightsGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"detail":"played SPADE 5"`)
		assert.Contains(t, result, `"turnNumber":1`)
		assert.Contains(t, result, `"playerIdx":0`)
		m.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		m := new(interfaces.MockCrazyEightsGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockCrazyEightsGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
		m.AssertExpectations(t)
	})
}

// **受動ヒントとは別に、`command: "hint"` 専用のレスポンスも返す (#4737)。**
// Web はクライアント側でも簡易ヒントを出すが、これはサーバー計算のもの。
func TestCrazyEightsWebPresenter_HintOutput(t *testing.T) {
	m, _ := setupCrazyEightsWebMockWithPlayers()

	out := new(presenter.CrazyEightsWebPresenter).HintOutput(m)
	assert.True(t, json.Valid([]byte(out)), "JSON として妥当")
	assert.Contains(t, out, `"players"`, "状態も一緒に返る")
}
