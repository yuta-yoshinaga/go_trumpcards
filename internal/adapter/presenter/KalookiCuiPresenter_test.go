//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKalookiCuiMock(phase domain.KalookiPhase, gameEnd bool) (*interfaces.MockKalookiGame, []*domain.KalookiPlayer) {
	m := new(interfaces.MockKalookiGame)
	players := []*domain.KalookiPlayer{
		domain.NewKalookiPlayer(true),
		domain.NewKalookiPlayer(false),
		domain.NewKalookiPlayer(false),
		domain.NewKalookiPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetOpeningThreshold").Return(51)
	m.On("GetDrawPileCount").Return(53)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultKalookiConfig())
	m.On("GetRoundWinnerIdx").Return(-1)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestKalookiCuiPresenter_Output(t *testing.T) {
	p := new(presenter.KalookiCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseDraw, false)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("meld phase", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseMeld, false)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseRoundEnd, false)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseGameEnd, true)
		assert.NotEmpty(t, p.Output(m, nil))
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupKalookiCuiMock(domain.KalookiPhaseDraw, false)
		assert.NotEmpty(t, p.Output(m, errors.New("err")))
	})

	t.Run("opened player with melds", func(t *testing.T) {
		m, players := setupKalookiCuiMock(domain.KalookiPhaseMeld, false)
		players[0].SetHasOpened(true)
		players[0].AppendMeld([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		})
		assert.NotEmpty(t, p.Output(m, nil))
	})
}

func TestKalookiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KalookiCuiPresenter)
	m, _ := setupKalookiCuiMock(domain.KalookiPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
