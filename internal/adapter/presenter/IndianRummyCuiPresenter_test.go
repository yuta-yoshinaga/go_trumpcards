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

func setupIndianRummyCuiMock(phase domain.IndianRummyPhase, gameEnd bool) (*interfaces.MockIndianRummyGame, []*domain.IndianRummyPlayer) {
	m := new(interfaces.MockIndianRummyGame)
	players := []*domain.IndianRummyPlayer{
		domain.NewIndianRummyPlayer(true),
		domain.NewIndianRummyPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRoundNumber").Return(1)
	m.On("GetTargetRounds").Return(3)
	m.On("GetDrawPileCount").Return(60)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetWildJoker").Return(domain.NewCard(domain.CardDesignDiamond, 9, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetPlayerCnt").Return(2)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	return m, players
}

func TestIndianRummyCuiPresenter_Output(t *testing.T) {
	p := new(presenter.IndianRummyCuiPresenter)

	t.Run("draw phase", func(t *testing.T) {
		m, _ := setupIndianRummyCuiMock(domain.IndianRummyPhaseDraw, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "インドラミー")
		assert.Contains(t, out, "ラウンド")
		assert.Contains(t, out, "ワイルドジョーカー")
	})

	t.Run("discard phase", func(t *testing.T) {
		m, _ := setupIndianRummyCuiMock(domain.IndianRummyPhaseDiscard, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "ディスカードフェーズ")
	})

	t.Run("round end", func(t *testing.T) {
		m, _ := setupIndianRummyCuiMock(domain.IndianRummyPhaseRoundEnd, false)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupIndianRummyCuiMock(domain.IndianRummyPhaseGameEnd, true)
		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupIndianRummyCuiMock(domain.IndianRummyPhaseDraw, false)
		out := p.Output(m, errors.New("err"))
		assert.NotEmpty(t, out)
	})
}

func TestIndianRummyCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.IndianRummyCuiPresenter)
	m, _ := setupIndianRummyCuiMock(domain.IndianRummyPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	out := p.ActionLogOutput(m)
	assert.NotEmpty(t, out)
}
