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

func setupThreeThirteenCuiMock(phase domain.ThreeThirteenPhase, gameEnd bool) (*interfaces.MockThreeThirteenGame, []*domain.ThreeThirteenPlayer) {
	m := new(interfaces.MockThreeThirteenGame)
	players := []*domain.ThreeThirteenPlayer{
		domain.NewThreeThirteenPlayer(true),
		domain.NewThreeThirteenPlayer(false),
		domain.NewThreeThirteenPlayer(false),
		domain.NewThreeThirteenPlayer(false),
	}
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	m.On("GetRound").Return(1)
	m.On("WildRank").Return(3)
	m.On("GetDealCount").Return(3)
	m.On("GetDrawPileCount").Return(91)
	m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
	m.On("GetGameEndFlag").Return(gameEnd)
	m.On("GetPhase").Return(phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetKnockerIdx").Return(-1)
	winner := -1
	if gameEnd {
		winner = 0
	}
	m.On("GetWinnerIdx").Return(winner)
	m.On("GetConfig").Return(domain.DefaultThreeThirteenConfig())
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetPlayerDeadwoodValue", i).Return(5)
	}
	return m, players
}

func TestThreeThirteenCuiPresenter_Output(t *testing.T) {
	p := new(presenter.ThreeThirteenCuiPresenter)

	for _, tc := range []struct {
		name  string
		phase domain.ThreeThirteenPhase
		end   bool
	}{
		{"draw phase", domain.ThreeThirteenPhaseDraw, false},
		{"discard phase", domain.ThreeThirteenPhaseDiscard, false},
		{"round end", domain.ThreeThirteenPhaseRoundEnd, false},
		{"game end", domain.ThreeThirteenPhaseGameEnd, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := setupThreeThirteenCuiMock(tc.phase, tc.end)
			assert.NotEmpty(t, p.Output(m, nil))
		})
	}

	t.Run("cpu deadwood masked during play, revealed at round end", func(t *testing.T) {
		m, _ := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseDraw, false)
		out := p.Output(m, nil)
		assert.Contains(t, out, "デッドウッド?") // CPU hands are hidden
		assert.Contains(t, out, "デッドウッド5") // the human's own deadwood

		m2, _ := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseRoundEnd, false)
		out2 := p.Output(m2, nil)
		assert.NotContains(t, out2, "デッドウッド?") // all hands revealed at round end
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseDraw, false)
		assert.NotEmpty(t, p.Output(m, errors.New("err")))
	})
}

func TestThreeThirteenCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ThreeThirteenCuiPresenter)
	m, _ := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseDraw, false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	assert.NotEmpty(t, p.ActionLogOutput(m))
}
