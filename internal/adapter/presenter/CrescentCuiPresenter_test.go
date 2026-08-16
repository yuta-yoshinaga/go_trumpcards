//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCrescentCuiMockDefaults(cg *interfaces.MockCrescentGame) {
	cg.On("GetPhase").Return(domain.CrescentPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("GetRedealsRemaining").Return(domain.CrescentMaxRedeals).Maybe()
	cg.On("IsStalemate").Return(false).Maybe()
	cg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.CrescentTableauCnt][]*domain.CrescentTableauCard
	tableau[0] = []*domain.CrescentTableauCard{
		{Card: domain.NewCard(domain.CardDesignSpade, 1, false), FaceUp: true},
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CrescentFoundationCnt][]*domain.Card
	for i := range domain.CrescentAscendingFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CrescentFoundationSuit(i), 1, false)}
	}
	for i := domain.CrescentAscendingFoundationCnt; i < domain.CrescentFoundationCnt; i++ {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CrescentFoundationSuit(i), domain.CardValueMax, false)}
	}
	cg.On("GetFoundation").Return(foundation).Maybe()
}

func TestCrescentCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentCuiMockDefaults(cg)
		p := new(CrescentCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CrescentPhaseGameClear)
		p := new(CrescentCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.CrescentPhaseGameOver)
		p := new(CrescentCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("stalemate", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		setupCrescentCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "IsStalemate")
		cg.On("IsStalemate").Return(true)
		cg.On("UndoToEscape").Return(0).Maybe()
		p := new(CrescentCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})
}

func TestCrescentCuiPresenter_HintOutput(t *testing.T) {
	t.Run("tableau to tableau", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: 3, ToZone: "tableau", ToCol: 4})
		p := new(CrescentCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: 3, ToZone: "foundation", ToCol: 0})
		p := new(CrescentCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})

	t.Run("redeal", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetHint").Return(&domain.CrescentHint{FromCol: -1, ToCol: -1, Redeal: true})
		p := new(CrescentCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})

	t.Run("nil", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetHint").Return((*domain.CrescentHint)(nil))
		p := new(CrescentCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})
}

func TestCrescentCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing returns empty log", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetPhase").Return(domain.CrescentPhasePlaying)
		p := new(CrescentCuiPresenter)
		out := p.ActionLogOutput(cg)
		assert.NotNil(t, out)
	})

	t.Run("game over returns log", func(t *testing.T) {
		cg := new(interfaces.MockCrescentGame)
		cg.On("GetPhase").Return(domain.CrescentPhaseGameOver)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "redeal", Detail: "test"},
		})
		p := new(CrescentCuiPresenter)
		out := p.ActionLogOutput(cg)
		assert.Contains(t, out, "redeal")
	})
}
