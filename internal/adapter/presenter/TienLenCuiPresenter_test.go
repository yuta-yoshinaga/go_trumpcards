//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupTienLenCuiMock() (*interfaces.MockTienLenGame, []*domain.TienLenPlayer) {
	m := new(interfaces.MockTienLenGame)
	players := makeTienLenPlayers()
	m.On("GetGameEndFlag").Return(false)
	m.On("GetTableCards").Return(([]*domain.Card)(nil))
	m.On("GetCpuActions").Return(([]*domain.TienLenAction)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestTienLenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TienLenCuiPresenter)

	t.Run("initial empty table", func(t *testing.T) {
		m, players := setupTienLenCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Tien Len")
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "自由に出せます")
		assert.Contains(t, result, "あなたのターン")
	})

	t.Run("table cards and cpu pass action", func(t *testing.T) {
		m, _ := setupTienLenCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTableCards")
		m.On("GetTableCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.TienLenAction{
			{PlayerIdx: 1, PlayedCards: nil},
			{PlayerIdx: 2, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)}},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "場")
		assert.Contains(t, result, "CPU 1")
	})

	t.Run("finished player shown", func(t *testing.T) {
		m, players := setupTienLenCuiMock()
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		assert.Contains(t, p.Output(m, nil), "上がり")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupTienLenCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})

	t.Run("game ended rankings", func(t *testing.T) {
		m, players := setupTienLenCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		players[0].SetRank(1)
		players[1].SetRank(2)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		assert.Contains(t, result, "あなた")
	})
}

func TestTienLenCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.TienLenCuiPresenter)

	m := new(interfaces.MockTienLenGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played 1 card(s)"},
	})
	assert.Contains(t, p.ActionLogOutput(m), "play")
}
