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

func makeBigTwoPlayers() []*domain.BigTwoPlayer {
	return []*domain.BigTwoPlayer{
		domain.NewBigTwoPlayer(true),
		domain.NewBigTwoPlayer(false),
		domain.NewBigTwoPlayer(false),
		domain.NewBigTwoPlayer(false),
	}
}

func setupBigTwoCuiMock() (*interfaces.MockBigTwoGame, []*domain.BigTwoPlayer) {
	m := new(interfaces.MockBigTwoGame)
	players := makeBigTwoPlayers()
	m.On("GetGameEndFlag").Return(false)
	m.On("GetTableCards").Return(([]*domain.Card)(nil))
	m.On("GetCpuActions").Return(([]*domain.BigTwoAction)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestBigTwoCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BigTwoCuiPresenter)

	t.Run("initial empty table shows title and human turn", func(t *testing.T) {
		m, players := setupBigTwoCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Big Two")
		assert.Contains(t, result, "自由に出せます")
		assert.Contains(t, result, "あなたのターン")
	})

	t.Run("cpu action line uses localized CPU name, not hardcoded", func(t *testing.T) {
		m, _ := setupBigTwoCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.BigTwoAction{
			{PlayerIdx: 1, PlayedCards: nil},
			{PlayerIdx: 2, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)}},
		})
		result := p.Output(m, nil)
		// cuiPlayerName renders "CPU 1"/"CPU 2" via the shared cuiPlayerCpu key.
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
	})

	t.Run("game ended rankings use localized player names", func(t *testing.T) {
		m, players := setupBigTwoCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		players[0].SetRank(1)
		players[1].SetRank(2)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		// Human -> "あなた", CPU -> "CPU 1" (both via cuiPlayerName).
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "1位")
	})

	t.Run("error is shown", func(t *testing.T) {
		m, _ := setupBigTwoCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})
}
