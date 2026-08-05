//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupZhengCuiMock() (*interfaces.MockZhengGame, []*domain.ZhengPlayer) {
	m := new(interfaces.MockZhengGame)
	players := makeZhengPresenterPlayers()
	m.On("GetGameEndFlag").Return(false)
	m.On("GetTableCards").Return(([]*domain.Card)(nil))
	m.On("GetCpuActions").Return(([]*domain.ZhengAction)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestZhengCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ZhengCuiPresenter)

	t.Run("initial empty table", func(t *testing.T) {
		m, players := setupZhengCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Zheng Shangyou")
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "自由に出せます")
		assert.Contains(t, result, "あなたのターン")
		// Combo-strength rules are shown on the human's turn.
		assert.Contains(t, result, "スートは無関係")
	})

	t.Run("invalid combo error is accompanied by the combo rules", func(t *testing.T) {
		m, _ := setupZhengCuiMock()
		result := p.Output(m, errors.New("invalid combo"))
		assert.Contains(t, result, "invalid combo")
		assert.Contains(t, result, "スートは無関係")
	})

	t.Run("table cards and cpu pass action", func(t *testing.T) {
		m, _ := setupZhengCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTableCards")
		m.On("GetTableCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.ZhengAction{
			{PlayerIdx: 1, PlayedCards: nil},
			{PlayerIdx: 2, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)}},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "場")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "パス")
	})

	t.Run("finished player shown", func(t *testing.T) {
		m, players := setupZhengCuiMock()
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		assert.Contains(t, p.Output(m, nil), "上がり")
	})

	t.Run("error shown", func(t *testing.T) {
		m, _ := setupZhengCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})

	t.Run("game ended rankings", func(t *testing.T) {
		m, players := setupZhengCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		players[0].SetRank(1)
		players[1].SetRank(2)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		assert.Contains(t, result, "あなた")
	})
}

func TestZhengCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.ZhengCuiPresenter)

	m := new(interfaces.MockZhengGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played 1 card(s)"},
	})
	assert.Contains(t, p.ActionLogOutput(m), "play")
}

// **CPU 名も i18n と太字を通す。**順位表示と CPU 手番ログだけ `fmt.Sprintf("CPU %d")`
// で自前に組んでいて、同じ画面の他の名前と見た目も経路も違っていた (#4807)。
func TestZheng_CpuNamesGoThroughTheSharedHelper(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	bold1 := color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))
	p := new(presenter.ZhengCuiPresenter)

	t.Run("cpu action log", func(t *testing.T) {
		m, _ := setupZhengCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.ZhengAction{{PlayerIdx: 1, PlayedCards: nil}})
		out := p.Output(m, nil)
		assert.Contains(t, out, bold1)
		// 素の "CPU 1" が太字の外に残っていないこと。
		assert.NotContains(t, strings.ReplaceAll(out, bold1, ""), "CPU 1")
	})
}
