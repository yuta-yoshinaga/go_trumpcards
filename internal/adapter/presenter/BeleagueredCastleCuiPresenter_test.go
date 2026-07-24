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

func setupBeleagueredCastleCuiMockDefaults(bg *interfaces.MockBeleagueredCastleGame) {
	bg.On("GetPhase").Return(domain.BeleagueredCastlePhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
	for i := range domain.BeleagueredCastleTableauCnt {
		tableau[i] = make([]*domain.BeleagueredCastleTableauCard, domain.BeleagueredCastleColumnLen)
		for j := range domain.BeleagueredCastleColumnLen {
			tableau[i][j] = &domain.BeleagueredCastleTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
	for i := range domain.BeleagueredCastleFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
	}
	bg.On("GetFoundation").Return(foundation).Maybe()
}

func TestBeleagueredCastleCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		p := new(BeleagueredCastleCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Beleaguered Castle")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(3)
		p := new(BeleagueredCastleCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, i18n.Tf("beleagueredcastle.undoToEscape", "count", "3"))
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		p := new(BeleagueredCastleCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameClear)

		p := new(BeleagueredCastleCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameOver)

		p := new(BeleagueredCastleCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(0)

		p := new(BeleagueredCastleCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("empty column", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard
		bg.On("GetTableau").Return(emptyTableau)

		p := new(BeleagueredCastleCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("empty foundation", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		setupBeleagueredCastleCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetFoundation")
		var emptyFoundation [domain.BeleagueredCastleFoundationCnt][]*domain.Card
		bg.On("GetFoundation").Return(emptyFoundation)

		p := new(BeleagueredCastleCuiPresenter)
		result := p.Output(bg, nil)
		assert.NotEmpty(t, result)
	})
}

func TestBeleagueredCastleCuiPresenter_HintOutput(t *testing.T) {
	t.Run("foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetHint").Return(&domain.BeleagueredCastleHint{
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(BeleagueredCastleCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetHint").Return(&domain.BeleagueredCastleHint{
			FromCol:   1,
			CardIndex: 0,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(BeleagueredCastleCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetHint").Return((*domain.BeleagueredCastleHint)(nil))

		p := new(BeleagueredCastleCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestBeleagueredCastleCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhasePlaying)

		p := new(BeleagueredCastleCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockBeleagueredCastleGame)
		bg.On("GetPhase").Return(domain.BeleagueredCastlePhaseGameOver)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(BeleagueredCastleCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
