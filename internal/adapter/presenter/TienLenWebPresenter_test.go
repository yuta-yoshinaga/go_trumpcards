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

func makeTienLenPlayers() []*domain.TienLenPlayer {
	return []*domain.TienLenPlayer{
		domain.NewTienLenPlayer(true),
		domain.NewTienLenPlayer(false),
		domain.NewTienLenPlayer(false),
		domain.NewTienLenPlayer(false),
	}
}

func setupTienLenWebMock() (*interfaces.MockTienLenGame, []*domain.TienLenPlayer) {
	m := new(interfaces.MockTienLenGame)
	players := makeTienLenPlayers()
	m.On("GetCurrentTurn").Return(0)
	m.On("GetLastPlayPlayerIdx").Return(-1)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetTablePlayType").Return(domain.TienLenPlayInvalid)
	m.On("GetConfig").Return(domain.DefaultTienLenConfig())
	m.On("GetTableCards").Return(([]*domain.Card)(nil))
	m.On("GetCpuActions").Return(([]*domain.TienLenAction)(nil))
	m.On("GetHumanAction").Return((*domain.TienLenAction)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestTienLenWebPresenter_Output(t *testing.T) {
	p := new(presenter.TienLenWebPresenter)

	t.Run("initial state hides CPU cards", func(t *testing.T) {
		m, players := setupTienLenWebMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		var out controller.TienLenWebOutput
		require := assert.New(t)
		require.NoError(json.Unmarshal([]byte(p.Output(m, nil)), &out))
		require.Len(out.Players, 4)
		require.True(out.Players[0].IsHuman)
		require.Len(out.Players[0].Cards, 1) // human cards visible
		require.Len(out.Players[1].Cards, 0) // CPU cards hidden
		require.Equal(1, out.Players[1].CardCount)
	})

	t.Run("table cards and cpu actions", func(t *testing.T) {
		m, _ := setupTienLenWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTableCards")
		m.On("GetTableCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.TienLenAction{
			{PlayerIdx: 1, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)}},
			{PlayerIdx: 2, PlayedCards: nil},
		})

		var out controller.TienLenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		assert.Len(t, out.TableCards, 1)
		assert.Len(t, out.CpuActions, 2)
	})

	t.Run("human action included", func(t *testing.T) {
		m, _ := setupTienLenWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHumanAction")
		m.On("GetHumanAction").Return(&domain.TienLenAction{
			PlayerIdx: 0, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)},
		})
		var out controller.TienLenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		assert.NotNil(t, out.HumanAction)
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupTienLenWebMock()
		var out controller.TienLenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, errors.New("bad play"))), &out))
		assert.Equal(t, "bad play", out.Message)
	})

	t.Run("game end builds rankings", func(t *testing.T) {
		m, players := setupTienLenWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		players[0].SetRank(1)
		players[1].SetRank(2)
		players[2].SetRank(3)
		players[3].SetRank(4)
		var out controller.TienLenWebOutput
		assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))
		assert.Contains(t, out.Message, "あなた:1位")
		assert.Equal(t, "tienlen.result.rankings", out.MessageCode)
		// The rankings param must not embed the "ゲーム終了！" prefix — the frontend
		// template adds it, so embedding it here would double up.
		assert.Contains(t, out.MessageParams["rankings"], "あなた:1位")
		assert.NotContains(t, out.MessageParams["rankings"], "ゲーム終了")
	})
}

func TestTienLenWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TienLenWebPresenter)
	m := new(interfaces.MockTienLenGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played 1 card(s)"},
	})
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// Web はヒント専用のレスポンスを持たない (フロントが算出する)。素通しなので
// 通常の Output と同じものが返ることだけ固定しておく (#5624)。
func TestTienLenWebPresenterHintOutputMatchesOutput(t *testing.T) {
	tl := domain.NewDefaultTienLen()
	tl.Reset()
	p := new(presenter.TienLenWebPresenter)

	assert.JSONEq(t, p.Output(tl, nil), p.HintOutput(tl))
}
