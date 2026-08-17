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

func setupWindmillCuiMockDefaults(g *interfaces.MockWindmillGame) {
	g.On("GetPhase").Return(domain.WindmillPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("IsTransferBlocked").Return(false).Maybe()
	g.On("GetStockCount").Return(95).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetCenter").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, true)}).Maybe()

	var sails [domain.WindmillSailCnt]*domain.Card
	for i := range domain.WindmillSailCnt {
		sails[i] = domain.NewCard(domain.CardDesignClover, i+2, true)
	}
	g.On("GetSails").Return(sails).Maybe()

	var corners [domain.WindmillCornerCnt][]*domain.Card
	g.On("GetCorners").Return(corners).Maybe()
}

func TestWindmillCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)

		result := new(WindmillCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Windmill")
		assert.Contains(t, result, i18n.T("windmill.centerHeader"))
		assert.Contains(t, result, i18n.T("windmill.cornerHeader"))
		assert.Contains(t, result, "帆0:")
		assert.Contains(t, result, "帆7:", "all eight sails are rendered")
		assert.Contains(t, result, "95")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("empty waste, empty corners and an empty sail", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetSails")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetCenter")
		g.On("GetWaste").Return([]*domain.Card(nil))
		g.On("GetSails").Return([domain.WindmillSailCnt]*domain.Card{})
		g.On("GetCenter").Return([]*domain.Card(nil))

		result := new(WindmillCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.T("windmill.wasteEmpty"))
		assert.Contains(t, result, "[空]")
	})

	// The block is invisible in the layout, so the board has to state it.
	t.Run("the transfer block is called out", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsTransferBlocked")
		g.On("IsTransferBlocked").Return(true)

		assert.Contains(t, new(WindmillCuiPresenter).Output(g, nil), i18n.T("windmill.transferBlocked"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(WindmillCuiPresenter).Output(g, nil),
			i18n.Tf("windmill.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(WindmillCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillCuiMockDefaults(g)

		assert.Contains(t, new(WindmillCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.WindmillPhase
		want string
	}{
		{"game clear", domain.WindmillPhaseGameClear, "ゲームクリア"},
		{"game over", domain.WindmillPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockWindmillGame)
			setupWindmillCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(WindmillCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestWindmillCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.WindmillHint
		contains []string
	}{
		{"sail to the centre",
			&domain.WindmillHint{FromZone: "sail", FromIdx: 4, ToZone: "center", ToIdx: -1},
			[]string{"帆4", "中央基礎札"}},
		{"sail to a corner",
			&domain.WindmillHint{FromZone: "sail", FromIdx: 0, ToZone: "corner", ToIdx: 2},
			[]string{"帆0", "四隅基礎札2"}},
		{"waste to the centre",
			&domain.WindmillHint{FromZone: "waste", FromIdx: -1, ToZone: "center", ToIdx: -1},
			[]string{"捨て札", "中央基礎札"}},
		{"the corner pull-back",
			&domain.WindmillHint{FromZone: "corner", FromIdx: 1, ToZone: "center", ToIdx: -1},
			[]string{"四隅基礎札1", "中央基礎札"}},
		{"draw from the stock",
			&domain.WindmillHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("windmill.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockWindmillGame)
			g.On("GetHint").Return(tc.hint)

			result := new(WindmillCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		g.On("GetHint").Return((*domain.WindmillHint)(nil))

		assert.Contains(t, new(WindmillCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestWindmillCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		g.On("GetPhase").Return(domain.WindmillPhasePlaying)

		assert.Contains(t, new(WindmillCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		g.On("GetPhase").Return(domain.WindmillPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(WindmillCuiPresenter).ActionLogOutput(g), "move")
	})
}

// #5558: 四隅は K しか受け取らないというウィンドミル固有のルールが、CUI では
// 出力にもヘルプにも出ておらず、試行錯誤でしか学べなかった。
func TestWindmillCuiPresenter_Output_EmptyCornerSaysKingsOnly(t *testing.T) {
	g := new(interfaces.MockWindmillGame)
	setupWindmillCuiMockDefaults(g)

	out := new(WindmillCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.T("windmill.cornerKingsOnly"))

	// **埋まっている四隅の表示は変えない。**カードが乗っていれば規則は自明。
	filled := new(interfaces.MockWindmillGame)
	var corners [domain.WindmillCornerCnt][]*domain.Card
	for i := range domain.WindmillCornerCnt {
		corners[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, true)}
	}
	filled.On("GetCorners").Return(corners)
	setupWindmillCuiMockDefaults(filled)
	assert.NotContains(t, new(WindmillCuiPresenter).Output(filled, nil), i18n.T("windmill.cornerKingsOnly"))
}
