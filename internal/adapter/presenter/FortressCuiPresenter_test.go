//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupFortressCuiMockDefaults(bg *interfaces.MockFortressGame) {
	bg.On("GetPhase").Return(domain.FortressPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
	for i := range domain.FortressTableauCnt {
		tableau[i] = make([]*domain.FortressTableauCard, domain.FortressColumnLen)
		for j := range domain.FortressColumnLen {
			tableau[i][j] = &domain.FortressTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.FortressFoundationCnt][]*domain.Card
	for i := range domain.FortressFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func TestFortressCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		p := new(FortressCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Fortress")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(3)
		p := new(FortressCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, i18n.Tf("fortress.undoToEscape", "count", "3"))
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		p := new(FortressCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FortressPhaseGameClear)

		p := new(FortressCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.FortressPhaseGameOver)

		p := new(FortressCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(0)

		p := new(FortressCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.FortressTableauCnt][]*domain.FortressTableauCard
		bg.On("GetTableau").Return(emptyTableau)

		p := new(FortressCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty foundation", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		setupFortressCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var emptyFoundation [domain.FortressFoundationCnt][]*domain.Card
		bg.On("GetFoundation").Return(emptyFoundation)

		p := new(FortressCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestFortressCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetHint").Return(&domain.FortressHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FortressCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetHint").Return(&domain.FortressHint{
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(FortressCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetHint").Return((*domain.FortressHint)(nil))

		p := new(FortressCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestFortressCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetPhase").Return(domain.FortressPhasePlaying)

		p := new(FortressCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockFortressGame)
		bg.On("GetPhase").Return(domain.FortressPhaseGameOver)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(FortressCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
