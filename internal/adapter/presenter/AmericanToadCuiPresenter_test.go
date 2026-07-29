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

func setupAmericanToadCuiMockDefaults(g *interfaces.MockAmericanToadGame) {
	g.On("GetPhase").Return(domain.AmericanToadPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("CanRedeal").Return(false).Maybe()
	g.On("GetStockCount").Return(75).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetReserve").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()

	var tableau [domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard
	for i := range domain.AmericanToadTableauCnt {
		tableau[i] = []*domain.AmericanToadTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.AmericanToadFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestAmericanToadCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)

		result := new(AmericanToadCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "American Toad")
		assert.Contains(t, result, i18n.T("americantoad.foundationHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列7:", "all eight columns are rendered")
		assert.Contains(t, result, "75")
		assert.Contains(t, result, i18n.Tf("americantoad.baseRankLine", "rank", "5"))
		assert.Contains(t, result, "手数: 0")
	})

	// Once the reserve is gone the empty-column rule changes, so the board says so.
	t.Run("an empty reserve is called out", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetReserve")
		g.On("GetReserve").Return([]*domain.Card(nil))

		assert.Contains(t, new(AmericanToadCuiPresenter).Output(g, nil), i18n.T("americantoad.reserveEmpty"))
	})

	t.Run("empty waste and empty columns", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.On("GetWaste").Return([]*domain.Card(nil))
		g.On("GetTableau").Return([domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard{})

		result := new(AmericanToadCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.T("americantoad.wasteEmpty"))
		assert.Contains(t, result, "[空]")
	})

	// The single redeal is easy to miss, so it is announced while it lasts.
	t.Run("an available redeal is announced", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRedeal")
		g.On("CanRedeal").Return(true)

		assert.Contains(t, new(AmericanToadCuiPresenter).Output(g, nil), i18n.T("americantoad.redealAvailable"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(AmericanToadCuiPresenter).Output(g, nil),
			i18n.Tf("americantoad.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(AmericanToadCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		setupAmericanToadCuiMockDefaults(g)

		assert.Contains(t, new(AmericanToadCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.AmericanToadPhase
		want string
	}{
		{"game clear", domain.AmericanToadPhaseGameClear, "ゲームクリア"},
		{"game over", domain.AmericanToadPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockAmericanToadGame)
			setupAmericanToadCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(AmericanToadCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestAmericanToadCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.AmericanToadHint
		contains []string
	}{
		{"reserve to a foundation",
			&domain.AmericanToadHint{FromZone: "reserve", FromIdx: -1, CardIndex: -1, ToZone: "foundation", ToIdx: 1},
			[]string{"リザーブ", "基礎札1"}},
		{"reserve to the tableau",
			&domain.AmericanToadHint{FromZone: "reserve", FromIdx: -1, CardIndex: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"リザーブ", "タブロー列3"}},
		{"waste to a foundation",
			&domain.AmericanToadHint{FromZone: "waste", FromIdx: -1, CardIndex: -1, ToZone: "foundation", ToIdx: 2},
			[]string{"捨て札", "基礎札2"}},
		{"tableau to tableau",
			&domain.AmericanToadHint{FromZone: "tableau", FromIdx: 1, CardIndex: 2, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー列1[2]", "タブロー列5"}},
		{"draw from the stock",
			&domain.AmericanToadHint{FromZone: "stock", FromIdx: -1, CardIndex: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("americantoad.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockAmericanToadGame)
			g.On("GetHint").Return(tc.hint)

			result := new(AmericanToadCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetHint").Return((*domain.AmericanToadHint)(nil))

		assert.Contains(t, new(AmericanToadCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestAmericanToadCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetPhase").Return(domain.AmericanToadPhasePlaying)

		assert.Contains(t, new(AmericanToadCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockAmericanToadGame)
		g.On("GetPhase").Return(domain.AmericanToadPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(AmericanToadCuiPresenter).ActionLogOutput(g), "move")
	})
}
