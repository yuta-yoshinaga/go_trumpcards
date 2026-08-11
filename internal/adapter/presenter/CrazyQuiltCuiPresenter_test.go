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

func setupCrazyQuiltCuiMockDefaults(g *interfaces.MockCrazyQuiltGame) {
	g.On("GetPhase").Return(domain.CrazyQuiltPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(32).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	// キルトは 64 マス。マス 5 だけ空けて、空きマスの描画も踏む。
	var quilt [domain.CrazyQuiltCells]*domain.Card
	for i := range domain.CrazyQuiltCells {
		quilt[i] = domain.NewCard(domain.CardDesignSpade, (i%13)+1, true)
	}
	quilt[5] = nil
	g.On("GetQuilt").Return(quilt).Maybe()
	for i := range domain.CrazyQuiltCells {
		g.On("IsAvailable", i).Return(i < 8).Maybe()
	}
	g.On("GetRedealsLeft").Return(domain.CrazyQuiltRedealCnt).Maybe()
	for i := range domain.CrazyQuiltFoundationCnt {
		g.On("IsAscendingFoundation", i).Return(i < domain.CrazyQuiltAscendingCnt).Maybe()
	}

	var foundation [domain.CrazyQuiltFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestCrazyQuiltCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)

		result := new(CrazyQuiltCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Crazy Quilt")
		assert.Contains(t, result, i18n.T("crazyquilt.foundationHeader"))
		// 8 行すべて描く。空きマスは印で示す。
		assert.Contains(t, result, i18n.T("crazyquilt.emptyCell"), "an emptied cell is drawn, not skipped")
		assert.Contains(t, result, "32", "the stock count is rendered")
		assert.Contains(t, result, "組み直し", "the redeal count is rendered")
		assert.Contains(t, result, "↑", "Ace-start foundations are marked")
		assert.Contains(t, result, "↓", "King-start foundations are marked")
		assert.Contains(t, result, "手数: 0")
	})

	// An empty pile behaves differently from an empty column elsewhere, so the
	// board spells out where a card may come from.
	// **取れる札に印が付くこと。**短辺の露出は向きに依存するので、印が無いと
	// 盤面を見ても何が取れるのか分からない。
	t.Run("marks the cards that can be taken", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)

		out := new(CrazyQuiltCuiPresenter).Output(g, nil)
		// 凡例そのものに * が入るので、**札に付いた印**だけを数える。
		assert.Contains(t, out, "*SPADE", "available cards carry a marker")
		assert.Contains(t, out, i18n.T("crazyquilt.availableLegend"))
	})

	// 負のコントロール: 1 枚も取れない盤面では印が出ない。
	t.Run("marks nothing when nothing is available", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAvailable")
		for i := range domain.CrazyQuiltCells {
			g.On("IsAvailable", i).Return(false)
		}

		out := new(CrazyQuiltCuiPresenter).Output(g, nil)
		assert.NotContains(t, out, "*SPADE", "no card is marked")
		// 凡例は常に出るので、それだけは残る。
		assert.Contains(t, out, i18n.T("crazyquilt.availableLegend"))
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(CrazyQuiltCuiPresenter).Output(g, nil), i18n.T("crazyquilt.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(CrazyQuiltCuiPresenter).Output(g, nil),
			i18n.Tf("crazyquilt.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(CrazyQuiltCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltCuiMockDefaults(g)

		assert.Contains(t, new(CrazyQuiltCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.CrazyQuiltPhase
		want string
	}{
		{"game clear", domain.CrazyQuiltPhaseGameClear, "ゲームクリア"},
		{"game over", domain.CrazyQuiltPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockCrazyQuiltGame)
			setupCrazyQuiltCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(CrazyQuiltCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestCrazyQuiltCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.CrazyQuiltHint
		contains []string
	}{
		{"quilt to a foundation",
			&domain.CrazyQuiltHint{FromZone: "quilt", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"キルト1", "組札2"}},
		// キルトを崩す手。行き先はマスではなく捨て札。
		{"quilt onto the waste",
			&domain.CrazyQuiltHint{FromZone: "quilt", FromIdx: 0, ToZone: "waste", ToIdx: -1},
			[]string{"キルト0", "捨て札"}},
		{"waste to a foundation",
			&domain.CrazyQuiltHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0},
			[]string{"捨て札", "組札0"}},
		{"draw from the stock",
			&domain.CrazyQuiltHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", "捨て札"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockCrazyQuiltGame)
			g.On("GetHint").Return(tc.hint)

			result := new(CrazyQuiltCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		g.On("GetHint").Return((*domain.CrazyQuiltHint)(nil))

		assert.Contains(t, new(CrazyQuiltCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestCrazyQuiltCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		g.On("GetPhase").Return(domain.CrazyQuiltPhasePlaying)

		assert.Contains(t, new(CrazyQuiltCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		g.On("GetPhase").Return(domain.CrazyQuiltPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(CrazyQuiltCuiPresenter).ActionLogOutput(g), "move")
	})
}
