//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupStHelenaCuiMockDefaults(cg *interfaces.MockStHelenaGame) {
	cg.On("GetPhase").Return(domain.StHelenaPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("GetRedealsRemaining").Return(domain.StHelenaMaxRedeals).Maybe()
	cg.On("RestrictionsActive").Return(true).Maybe()
	cg.On("IsStalemate").Return(false).Maybe()
	cg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard
	tableau[0] = []*domain.StHelenaTableauCard{
		{Card: domain.NewCard(domain.CardDesignSpade, 1, false), FaceUp: true},
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.StHelenaFoundationCnt][]*domain.Card
	for i := range domain.StHelenaAscendingFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.StHelenaFoundationSuit(i), 1, false)}
	}
	for i := domain.StHelenaAscendingFoundationCnt; i < domain.StHelenaFoundationCnt; i++ {
		foundation[i] = []*domain.Card{domain.NewCard(domain.StHelenaFoundationSuit(i), domain.CardValueMax, false)}
	}
	cg.On("GetFoundation").Return(foundation).Maybe()
}

func TestStHelenaCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaCuiMockDefaults(cg)
		p := new(StHelenaCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.StHelenaPhaseGameClear)
		p := new(StHelenaCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "GetPhase")
		cg.On("GetPhase").Return(domain.StHelenaPhaseGameOver)
		p := new(StHelenaCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})

	t.Run("stalemate", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "IsStalemate")
		cg.On("IsStalemate").Return(true)
		cg.On("UndoToEscape").Return(0).Maybe()
		p := new(StHelenaCuiPresenter)
		out := p.Output(cg, nil)
		assert.NotEmpty(t, out)
	})
}

func TestStHelenaCuiPresenter_HintOutput(t *testing.T) {
	t.Run("tableau to tableau", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: 3, ToZone: "tableau", ToCol: 4})
		p := new(StHelenaCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})

	t.Run("tableau to foundation", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: 3, ToZone: "foundation", ToCol: 0})
		p := new(StHelenaCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})

	t.Run("redeal", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetHint").Return(&domain.StHelenaHint{FromCol: -1, ToCol: -1, Redeal: true})
		p := new(StHelenaCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})

	t.Run("nil", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetHint").Return((*domain.StHelenaHint)(nil))
		p := new(StHelenaCuiPresenter)
		assert.NotEmpty(t, p.HintOutput(cg))
	})
}

func TestStHelenaCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing returns empty log", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetPhase").Return(domain.StHelenaPhasePlaying)
		p := new(StHelenaCuiPresenter)
		out := p.ActionLogOutput(cg)
		assert.NotNil(t, out)
	})

	t.Run("game over returns log", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		cg.On("GetPhase").Return(domain.StHelenaPhaseGameOver)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "redeal", Detail: "test"},
		})
		p := new(StHelenaCuiPresenter)
		out := p.ActionLogOutput(cg)
		assert.Contains(t, out, "redeal")
	})
}

// **制限は盤の見た目からは読めない。**書かないと、なぜ拒まれたのか分からない
// まま手が止まる。解けたら解けたと言うこと。
func TestStHelenaCuiPresenter_SaysWhichColumnsReachWhichRow(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("while the restriction is on", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "RestrictionsActive")
		cg.On("RestrictionsActive").Return(true)

		out := new(StHelenaCuiPresenter).Output(cg, nil)
		assert.Contains(t, out, i18n.T("sthelena.restrictionsLine"))
		assert.NotContains(t, out, i18n.T("sthelena.restrictionsLifted"))
		// 列ごとの位置も添える。上 4 / 下 4 / 横 4 がそれぞれ出ること。
		assert.Equal(t, domain.StHelenaTopColumnCnt, strings.Count(out, i18n.T("sthelena.columnRowTop")))
		assert.Equal(t, len(domain.StHelenaSideColumns), strings.Count(out, i18n.T("sthelena.columnRowSide")))
		assert.Equal(t,
			domain.StHelenaTableauCnt-domain.StHelenaTopColumnCnt-len(domain.StHelenaSideColumns),
			strings.Count(out, i18n.T("sthelena.columnRowBottom")))
	})

	t.Run("once it is lifted", func(t *testing.T) {
		cg := new(interfaces.MockStHelenaGame)
		setupStHelenaCuiMockDefaults(cg)
		cg.ExpectedCalls = filterCalls(cg.ExpectedCalls, "RestrictionsActive")
		cg.On("RestrictionsActive").Return(false)

		out := new(StHelenaCuiPresenter).Output(cg, nil)
		assert.Contains(t, out, i18n.T("sthelena.restrictionsLifted"))
		// 位置の注記も消える。制限が無いなら意味を持たない。
		assert.NotContains(t, out, i18n.T("sthelena.columnRowTop"))
	})
}
