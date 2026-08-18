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

func makeThirtyOnePlayers() []*domain.ThirtyOnePlayer {
	players := []*domain.ThirtyOnePlayer{
		domain.NewThirtyOnePlayer(true),
		domain.NewThirtyOnePlayer(false),
		domain.NewThirtyOnePlayer(false),
		domain.NewThirtyOnePlayer(false),
	}
	for _, p := range players {
		p.SetLives(3)
	}
	return players
}

func setupThirtyOneWebMock() (*interfaces.MockThirtyOneGame, []*domain.ThirtyOnePlayer) {
	m := new(interfaces.MockThirtyOneGame)
	players := makeThirtyOnePlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(39)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ThirtyOnePhaseDraw)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultThirtyOneConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetKnockerIdx").Return(-1)
	m.On("GetThirtyOneIdx").Return(-1)
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetRoundLosers").Return([]int{})
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestThirtyOneWebPresenter_Output(t *testing.T) {
	p := new(presenter.ThirtyOneWebPresenter)

	t.Run("initial draw phase hides CPU cards", func(t *testing.T) {
		m, players := setupThirtyOneWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		var resObj controller.ThirtyOneWebOutput
		require := json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NoError(t, require)
		assert.Len(t, resObj.Players, 4)
		assert.Equal(t, 39, resObj.DrawPileCount)
		assert.Equal(t, "thirtyone.drawPhase", resObj.MessageCode)
		assert.Len(t, resObj.Players[0].Cards, 1) // human shown
		assert.Len(t, resObj.Players[1].Cards, 0) // CPU hidden
		assert.Equal(t, 3, resObj.Players[0].Lives)
	})

	t.Run("discard phase message", func(t *testing.T) {
		m, _ := setupThirtyOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThirtyOnePhaseDiscard)
		var resObj controller.ThirtyOneWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "thirtyone.discardPhase", resObj.MessageCode)
	})

	t.Run("round end reveals CPU cards and scores", func(t *testing.T) {
		m, players := setupThirtyOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ThirtyOnePhaseRoundEnd)
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))
		var resObj controller.ThirtyOneWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Len(t, resObj.Players[1].Cards, 1)
		assert.Equal(t, 10, resObj.Players[1].Score)
		assert.Equal(t, "thirtyone.roundEnd", resObj.MessageCode)
	})

	t.Run("round end thirty-one message", func(t *testing.T) {
		m, _ := setupThirtyOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetThirtyOneIdx")
		m.On("GetPhase").Return(domain.ThirtyOnePhaseRoundEnd)
		m.On("GetThirtyOneIdx").Return(0)
		var resObj controller.ThirtyOneWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, "thirtyone.thirtyOneHit", resObj.MessageCode)
	})

	t.Run("game end has winner message", func(t *testing.T) {
		m, _ := setupThirtyOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetPhase").Return(domain.ThirtyOnePhaseGameEnd)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		var resObj controller.ThirtyOneWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.Equal(t, 0, resObj.WinnerIdx)
		assert.NotEmpty(t, resObj.MessageCode)
	})

	t.Run("discard top serialized", func(t *testing.T) {
		m, _ := setupThirtyOneWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 8, false))
		var resObj controller.ThirtyOneWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, nil)), &resObj)
		assert.NotNil(t, resObj.DiscardTop)
		assert.Equal(t, 8, resObj.DiscardTop.Value)
	})

	t.Run("error surfaced", func(t *testing.T) {
		m, _ := setupThirtyOneWebMock()
		var resObj controller.ThirtyOneWebOutput
		_ = json.Unmarshal([]byte(p.Output(m, errors.New("oops"))), &resObj)
		assert.Equal(t, "oops", resObj.Message)
	})
}

func TestThirtyOneWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ThirtyOneWebPresenter)
	m := new(interfaces.MockThirtyOneGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "knock", Detail: "knocks"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	assert.Contains(t, p.ActionLogOutput(m), "knock")
}

// #5623: 難易度の違いは CPU のノック閾値 (25/27/29) がすべてなのに、Web には
// Easy/Normal/Hard というラベルしか届いていなかった。数字を画面側に書き写すと
// 定数を変えたときに黙って古くなるので、レスポンスで運ぶ。
func TestThirtyOneWebPresenterCarriesTheKnockThresholds(t *testing.T) {
	g := domain.NewDefaultThirtyOne()
	g.Reset()

	var out controller.ThirtyOneWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.ThirtyOneWebPresenter).Output(g, nil)), &out))

	assert.Equal(t, domain.ThirtyOneKnockThresholdEasy, out.Config.KnockThresholds.Easy)
	assert.Equal(t, domain.ThirtyOneKnockThresholdNormal, out.Config.KnockThresholds.Normal)
	assert.Equal(t, domain.ThirtyOneKnockThresholdHard, out.Config.KnockThresholds.Hard)
	// 難易度を選ぶ意味がある = 3 つが別の数字。
	assert.NotEqual(t, out.Config.KnockThresholds.Easy, out.Config.KnockThresholds.Hard)
}
