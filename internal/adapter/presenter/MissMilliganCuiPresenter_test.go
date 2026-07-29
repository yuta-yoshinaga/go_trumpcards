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

func setupMissMilliganCuiMockDefaults(g *interfaces.MockMissMilliganGame) {
	g.On("GetPhase").Return(domain.MissMilliganPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetWaived").Return([]*domain.Card(nil)).Maybe()
	g.On("CanWaive").Return(false).Maybe()

	var tableau [domain.MissMilliganTableauCnt][]*domain.MissMilliganTableauCard
	for i := range domain.MissMilliganTableauCnt {
		tableau[i] = []*domain.MissMilliganTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.MissMilliganFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestMissMilliganCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)

		result := new(MissMilliganCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Miss Milligan")
		assert.Contains(t, result, i18n.T("missmilligan.foundationHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列7:", "all eight columns are rendered")
		assert.Contains(t, result, "96")
		assert.Contains(t, result, "手数: 0")
	})

	// Holding cards blocks dealing and waiving, so the board has to say so.
	t.Run("waived cards are called out", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaived")
		g.On("GetWaived").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, true)})

		result := new(MissMilliganCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "保持中")
	})

	t.Run("waive availability is announced", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanWaive")
		g.On("CanWaive").Return(true)

		assert.Contains(t, new(MissMilliganCuiPresenter).Output(g, nil),
			i18n.T("missmilligan.waiveAvailable"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(MissMilliganCuiPresenter).Output(g, nil),
			i18n.Tf("missmilligan.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(MissMilliganCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)

		assert.Contains(t, new(MissMilliganCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.MissMilliganPhase
		want string
	}{
		{"game clear", domain.MissMilliganPhaseGameClear, "ゲームクリア"},
		{"game over", domain.MissMilliganPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockMissMilliganGame)
			setupMissMilliganCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(MissMilliganCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("empty column and empty foundation", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		setupMissMilliganCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		var emptyTableau [domain.MissMilliganTableauCnt][]*domain.MissMilliganTableauCard
		g.On("GetTableau").Return(emptyTableau)

		assert.Contains(t, new(MissMilliganCuiPresenter).Output(g, nil), "[空]")
	})
}

func TestMissMilliganCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.MissMilliganHint
		contains []string
	}{
		{"waived back to the tableau",
			&domain.MissMilliganHint{FromZone: "waived", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"保持中の札", "タブロー列3"}},
		{"waived to a foundation",
			&domain.MissMilliganHint{FromZone: "waived", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToIdx: 2},
			[]string{"保持中の札", "基礎札2"}},
		{"tableau to tableau",
			&domain.MissMilliganHint{FromZone: "tableau", FromCol: 1, CardIndex: 2, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー列1[2]", "タブロー列5"}},
		{"deal a row",
			&domain.MissMilliganHint{FromZone: "stock", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToIdx: -1},
			[]string{"山札", i18n.T("missmilligan.hintToDeal")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockMissMilliganGame)
			g.On("GetHint").Return(tc.hint)

			result := new(MissMilliganCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		g.On("GetHint").Return((*domain.MissMilliganHint)(nil))

		assert.Contains(t, new(MissMilliganCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestMissMilliganCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		g.On("GetPhase").Return(domain.MissMilliganPhasePlaying)

		assert.Contains(t, new(MissMilliganCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockMissMilliganGame)
		g.On("GetPhase").Return(domain.MissMilliganPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(MissMilliganCuiPresenter).ActionLogOutput(g), "move")
	})
}
