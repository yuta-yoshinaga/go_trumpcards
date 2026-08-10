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

func setupDiplomatCuiMockDefaults(g *interfaces.MockDiplomatGame) {
	g.On("GetPhase").Return(domain.DiplomatPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(72).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	var tableau [domain.DiplomatTableauCnt][]*domain.Card
	for i := range domain.DiplomatTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.DiplomatFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestDiplomatCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)

		result := new(DiplomatCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Diplomat")
		assert.Contains(t, result, i18n.T("diplomat.foundationHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列7:", "all eight columns are rendered")
		assert.Contains(t, result, "72", "the stock count is rendered")
		assert.Contains(t, result, "手数: 0")
	})

	// An empty pile behaves differently from an empty column elsewhere, so the
	// board spells out where a card may come from.
	t.Run("an empty pile says where it can be filled from", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.On("GetTableau").Return([domain.DiplomatTableauCnt][]*domain.Card{})

		assert.Contains(t, new(DiplomatCuiPresenter).Output(g, nil), i18n.T("diplomat.emptyPile"))
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(DiplomatCuiPresenter).Output(g, nil), i18n.T("diplomat.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(DiplomatCuiPresenter).Output(g, nil),
			i18n.Tf("diplomat.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(DiplomatCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		setupDiplomatCuiMockDefaults(g)

		assert.Contains(t, new(DiplomatCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.DiplomatPhase
		want string
	}{
		{"game clear", domain.DiplomatPhaseGameClear, "ゲームクリア"},
		{"game over", domain.DiplomatPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockDiplomatGame)
			setupDiplomatCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(DiplomatCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestDiplomatCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.DiplomatHint
		contains []string
	}{
		{"tableau to a foundation",
			&domain.DiplomatHint{FromZone: "tableau", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"タブロー列1", "基礎札2"}},
		{"between piles",
			&domain.DiplomatHint{FromZone: "tableau", FromIdx: 0, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー列0", "タブロー列5"}},
		{"waste to a foundation",
			&domain.DiplomatHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0},
			[]string{"捨て札", "基礎札0"}},
		{"draw from the stock",
			&domain.DiplomatHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("diplomat.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockDiplomatGame)
			g.On("GetHint").Return(tc.hint)

			result := new(DiplomatCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		g.On("GetHint").Return((*domain.DiplomatHint)(nil))

		assert.Contains(t, new(DiplomatCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestDiplomatCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		g.On("GetPhase").Return(domain.DiplomatPhasePlaying)

		assert.Contains(t, new(DiplomatCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockDiplomatGame)
		g.On("GetPhase").Return(domain.DiplomatPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(DiplomatCuiPresenter).ActionLogOutput(g), "move")
	})
}
