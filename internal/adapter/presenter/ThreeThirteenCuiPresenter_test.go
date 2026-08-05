//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
		m.On("GetDeadwoodAfterDiscard", i, mock.Anything).Return(4).Maybe()
	}
	return m, players
}

// **捨てる前に分かる情報。**Web は 1 枚選ぶたびに予測デッドウッドを出しているのに、
// CUI は今の値しか出さず、どれを捨てると得かは捨てるまで分からなかった (#4840)。
func TestThreeThirteenCuiPresenter_DiscardPreview(t *testing.T) {
	p := new(presenter.ThreeThirteenCuiPresenter)

	t.Run("lists the deadwood for each discard", func(t *testing.T) {
		m, players := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseDiscard, false)
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeadwoodAfterDiscard")
		m.On("GetDeadwoodAfterDiscard", 0, 0).Return(9)
		m.On("GetDeadwoodAfterDiscard", 0, 1).Return(5)
		m.On("GetDeadwoodAfterDiscard", mock.Anything, mock.Anything).Return(0).Maybe()

		out := p.Output(m, nil)
		assert.Contains(t, out, "捨てた後のデッドウッド: [0]9  [1]5")
	})

	t.Run("not shown in the draw phase", func(t *testing.T) {
		m, _ := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseDraw, false)
		assert.NotContains(t, p.Output(m, nil), "捨てた後のデッドウッド")
	})

	t.Run("not shown while another player is on turn", func(t *testing.T) {
		m, _ := setupThreeThirteenCuiMock(domain.ThreeThirteenPhaseDiscard, false)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(2)
		assert.NotContains(t, p.Output(m, nil), "捨てた後のデッドウッド")
	})
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
