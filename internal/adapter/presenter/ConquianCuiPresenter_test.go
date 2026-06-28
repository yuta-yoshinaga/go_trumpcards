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

func makeConquianPlayers() []*domain.ConquianPlayer {
	return []*domain.ConquianPlayer{
		domain.NewConquianPlayer(true),
		domain.NewConquianPlayer(false),
	}
}

func setupConquianCuiMock(phase domain.ConquianPhase, ended bool, winner int) (*interfaces.MockConquianGame, []*domain.ConquianPlayer) {
	m := new(interfaces.MockConquianGame)
	players := makeConquianPlayers()
	m.On("GetRoundNumber").Return(1)
	m.On("GetDrawPileCount").Return(20)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetGameEndFlag").Return(ended)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetTookDiscard").Return(false)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestConquianCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ConquianCuiPresenter)

	t.Run("draw phase shows header and hand", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Conquian (コンキャン)")
		assert.Contains(t, result, "ラウンド: 1")
	})

	t.Run("meld phase shows meld and melded cards", func(t *testing.T) {
		m, players := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		players[0].AddMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// No discard taken → no forced-use warning.
		assert.NotContains(t, result, "必ず")
	})

	t.Run("forced-use warning shown after taking a discard", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseMeld, false, -1)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTookDiscard")
		m.On("GetTookDiscard").Return(true)

		result := p.Output(m, nil)
		assert.Contains(t, result, "必ず")
	})

	t.Run("discard top is displayed", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
		m.ExpectedCalls = nil
		players := makeConquianPlayers()
		m.On("GetRoundNumber").Return(1)
		m.On("GetDrawPileCount").Return(20)
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 7, false))
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPhase").Return(domain.ConquianPhaseDraw)
		m.On("GetCurrentPlayerIdx").Return(0)
		m.On("GetWinnerIdx").Return(-1)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(2)
		m.On("GetPlayer", 0).Return(players[0])
		m.On("GetPlayer", 1).Return(players[1])
		result := p.Output(m, nil)
		assert.Contains(t, result, "捨て札")
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseRoundEnd, false, -1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end with winner", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseGameEnd, true, 0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("game end draw", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseGameEnd, true, -1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "引き分け")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestConquianCuiPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupConquianCuiMock(domain.ConquianPhaseDraw, false, -1)
	p := new(presenter.ConquianCuiPresenter)
	assert.NotPanics(t, func() { p.ActionLogOutput(m) })
}
