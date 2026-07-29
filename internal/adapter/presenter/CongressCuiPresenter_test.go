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

func setupCongressCuiMockDefaults(g *interfaces.MockCongressGame) {
	g.On("GetPhase").Return(domain.CongressPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	var tableau [domain.CongressTableauCnt][]*domain.Card
	for i := range domain.CongressTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CongressFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestCongressCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)

		result := new(CongressCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Congress")
		assert.Contains(t, result, i18n.T("congress.foundationHeader"))
		assert.Contains(t, result, "山0:")
		assert.Contains(t, result, "山7:", "all eight piles are rendered")
		assert.Contains(t, result, "96")
		assert.Contains(t, result, "手数: 0")
	})

	// An empty pile behaves differently from an empty column elsewhere, so the
	// board spells out where a card may come from.
	t.Run("an empty pile says where it can be filled from", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.On("GetTableau").Return([domain.CongressTableauCnt][]*domain.Card{})

		assert.Contains(t, new(CongressCuiPresenter).Output(g, nil), i18n.T("congress.emptyPile"))
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(CongressCuiPresenter).Output(g, nil), i18n.T("congress.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(CongressCuiPresenter).Output(g, nil),
			i18n.Tf("congress.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(CongressCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		setupCongressCuiMockDefaults(g)

		assert.Contains(t, new(CongressCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.CongressPhase
		want string
	}{
		{"game clear", domain.CongressPhaseGameClear, "ゲームクリア"},
		{"game over", domain.CongressPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockCongressGame)
			setupCongressCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(CongressCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestCongressCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.CongressHint
		contains []string
	}{
		{"tableau to a foundation",
			&domain.CongressHint{FromZone: "tableau", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"タブロー山1", "基礎札2"}},
		{"between piles",
			&domain.CongressHint{FromZone: "tableau", FromIdx: 0, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー山0", "タブロー山5"}},
		{"waste to a foundation",
			&domain.CongressHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0},
			[]string{"捨て札", "基礎札0"}},
		{"stock into a gap",
			&domain.CongressHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"山札", "タブロー山3"}},
		{"draw from the stock",
			&domain.CongressHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("congress.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockCongressGame)
			g.On("GetHint").Return(tc.hint)

			result := new(CongressCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		g.On("GetHint").Return((*domain.CongressHint)(nil))

		assert.Contains(t, new(CongressCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestCongressCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		g.On("GetPhase").Return(domain.CongressPhasePlaying)

		assert.Contains(t, new(CongressCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockCongressGame)
		g.On("GetPhase").Return(domain.CongressPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(CongressCuiPresenter).ActionLogOutput(g), "move")
	})
}
