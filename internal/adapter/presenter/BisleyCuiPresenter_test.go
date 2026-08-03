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

// bisleyTestColumnLen は 48 枚を 13 列に配ったときの最短の列の長さ。
const bisleyTestColumnLen = 3

func setupBisleyCuiMockDefaults(bg *interfaces.MockBisleyGame) {
	bg.On("GetPhase").Return(domain.BisleyPhasePlaying).Maybe()
	bg.On("GetMoveCount").Return(0).Maybe()
	bg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.BisleyTableauCnt][]*domain.BisleyTableauCard
	for i := range domain.BisleyTableauCnt {
		tableau[i] = make([]*domain.BisleyTableauCard, bisleyTestColumnLen)
		for j := range bisleyTestColumnLen {
			tableau[i][j] = &domain.BisleyTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	bg.On("GetTableau").Return(tableau).Maybe()

	var ace [domain.BisleyFoundationCnt][]*domain.Card
	var king [domain.BisleyFoundationCnt][]*domain.Card
	for i := range domain.BisleyFoundationCnt {
		ace[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, 1, false)}
		king[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i, domain.CardValueMax, false)}
	}
	bg.On("GetAceFoundations").Return(ace).Maybe()
	bg.On("GetKingFoundations").Return(king).Maybe()
}

func TestBisleyCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		p := new(BisleyCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "Bisley")
		assert.Contains(t, result, i18n.T("bisley.aceFoundationHeader"))
		assert.Contains(t, result, i18n.T("bisley.kingFoundationHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列12:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(3)
		p := new(BisleyCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, i18n.Tf("bisley.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "IsStalemate")
		bg.On("IsStalemate").Return(true)
		bg.On("UndoToEscape").Return(0)
		p := new(BisleyCuiPresenter)

		result := p.Output(bg, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		p := new(BisleyCuiPresenter)

		result := p.Output(bg, errors.New("test error"))
		assert.Contains(t, result, "test error")
	})

	t.Run("game clear", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BisleyPhaseGameClear)

		p := new(BisleyCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetPhase")
		bg.On("GetPhase").Return(domain.BisleyPhaseGameOver)

		p := new(BisleyCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("empty column and empty foundations", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		setupBisleyCuiMockDefaults(bg)
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetTableau")
		bg.ExpectedCalls = filterCalls(bg.ExpectedCalls, "GetKingFoundations")
		var emptyTableau [domain.BisleyTableauCnt][]*domain.BisleyTableauCard
		var emptyKing [domain.BisleyFoundationCnt][]*domain.Card
		bg.On("GetTableau").Return(emptyTableau)
		bg.On("GetKingFoundations").Return(emptyKing)

		p := new(BisleyCuiPresenter)
		result := p.Output(bg, nil)
		assert.Contains(t, result, "[空]")
	})
}

func TestBisleyCuiPresenter_HintOutput(t *testing.T) {
	t.Run("ascending foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetHint").Return(&domain.BisleyHint{FromCol: 0, ToZone: "ace", ToIdx: 2})

		p := new(BisleyCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "タブロー列0")
		assert.Contains(t, result, "昇順基礎札2")
	})

	t.Run("descending foundation hint", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetHint").Return(&domain.BisleyHint{FromCol: 4, ToZone: "king", ToIdx: 1})

		p := new(BisleyCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列4")
		assert.Contains(t, result, "降順基礎札1")
	})

	t.Run("tableau hint", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetHint").Return(&domain.BisleyHint{FromCol: 1, ToZone: "tableau", ToIdx: 3})

		p := new(BisleyCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "タブロー列1")
		assert.Contains(t, result, "タブロー列3")
	})

	t.Run("no hint", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetHint").Return((*domain.BisleyHint)(nil))

		p := new(BisleyCuiPresenter)
		result := p.HintOutput(bg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestBisleyCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetPhase").Return(domain.BisleyPhasePlaying)

		p := new(BisleyCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		bg := new(interfaces.MockBisleyGame)
		bg.On("GetPhase").Return(domain.BisleyPhaseGameOver)
		bg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(BisleyCuiPresenter)
		result := p.ActionLogOutput(bg)
		assert.Contains(t, result, "move")
	})
}
