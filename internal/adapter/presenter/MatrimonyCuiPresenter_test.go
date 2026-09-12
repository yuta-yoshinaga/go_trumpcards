//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

type mockMatrimonyGame struct{ mock.Mock }

func (g *mockMatrimonyGame) Reset()                            { g.Called() }
func (g *mockMatrimonyGame) Draw() error                       { return g.Called().Error(0) }
func (g *mockMatrimonyGame) MoveTableauToFoundation(int) error { return g.Called().Error(0) }
func (g *mockMatrimonyGame) MoveWasteToFoundation() error      { return g.Called().Error(0) }
func (g *mockMatrimonyGame) MoveWasteToTableau(int) error      { return g.Called().Error(0) }
func (g *mockMatrimonyGame) MoveStockToTableau(int) error      { return g.Called().Error(0) }
func (g *mockMatrimonyGame) GiveUp()                           { g.Called() }
func (g *mockMatrimonyGame) AutoComplete() error               { return g.Called().Error(0) }
func (g *mockMatrimonyGame) Undo() error                       { return g.Called().Error(0) }
func (g *mockMatrimonyGame) UndoN(int) error                   { return g.Called().Error(0) }
func (g *mockMatrimonyGame) CanUndo() bool                     { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) UndoToEscape() int                 { return g.Called().Int(0) }
func (g *mockMatrimonyGame) AllFaceUp() bool                   { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) GetGameEndFlag() bool              { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) GetHint() *domain.MatrimonyHint {
	v := g.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.MatrimonyHint)
}
func (g *mockMatrimonyGame) GetPhase() domain.MatrimonyPhase {
	return g.Called().Get(0).(domain.MatrimonyPhase)
}
func (g *mockMatrimonyGame) GetMoveCount() int   { return g.Called().Int(0) }
func (g *mockMatrimonyGame) GetStockCount() int  { return g.Called().Int(0) }
func (g *mockMatrimonyGame) GetRedealCount() int { return g.Called().Int(0) }
func (g *mockMatrimonyGame) GetWaste() []*domain.Card {
	v := g.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}
func (g *mockMatrimonyGame) GetTableau() [domain.MatrimonyTableauCnt]*domain.Card {
	return g.Called().Get(0).([domain.MatrimonyTableauCnt]*domain.Card)
}
func (g *mockMatrimonyGame) GetFoundation() [domain.MatrimonyFoundationCnt][]*domain.Card {
	return g.Called().Get(0).([domain.MatrimonyFoundationCnt][]*domain.Card)
}
func (g *mockMatrimonyGame) IsStalemate() bool { return g.Called().Bool(0) }
func (g *mockMatrimonyGame) GetActionLog() []*domain.ActionLogEntry {
	v := g.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func setupMatrimonyCuiMockDefaults(g *mockMatrimonyGame) {
	g.On("GetPhase").Return(domain.MatrimonyPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(88).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	// タブローは 1 枠 1 枚。枠 3 は空きにして、空き枠の描画も踏む。
	var tableau [domain.MatrimonyTableauCnt]*domain.Card
	for i := range domain.MatrimonyTableauCnt {
		tableau[i] = domain.NewCard(domain.CardDesignSpade, (i%13)+1, true)
	}
	tableau[3] = nil
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.MatrimonyFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestMatrimonyCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)

		result := new(MatrimonyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Matrimony")
		assert.Contains(t, result, i18n.T("matrimony.foundationHeader"))
		assert.Contains(t, result, "[枠0]")
		assert.Contains(t, result, "[枠15]", "all sixteen slots are rendered")
		assert.Contains(t, result, "[枠3] (空)")
		assert.Contains(t, result, "88", "the stock count is rendered")
		assert.Contains(t, result, "手数: 0")
	})

	// An empty pile behaves differently from an empty column elsewhere, so the
	// board spells out where a card may come from.
	// リザーブは空くと二度と埋まらない。その一点をここで固定する。
	t.Run("every slot filled renders no empty marker", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		var full [domain.MatrimonyTableauCnt]*domain.Card
		for i := range domain.MatrimonyTableauCnt {
			full[i] = domain.NewCard(domain.CardDesignSpade, 7, true)
		}
		g.On("GetTableau").Return(full)

		// 負のコントロール: 上のテストの "(空)" が常に出ているわけではない。
		assert.NotContains(t, new(MatrimonyCuiPresenter).Output(g, nil), i18n.T("matrimony.emptySlot"))
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(MatrimonyCuiPresenter).Output(g, nil), i18n.T("matrimony.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(MatrimonyCuiPresenter).Output(g, nil),
			i18n.Tf("matrimony.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(MatrimonyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)

		assert.Contains(t, new(MatrimonyCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.MatrimonyPhase
		want string
	}{
		{"game clear", domain.MatrimonyPhaseGameClear, "ゲームクリア"},
		{"game over", domain.MatrimonyPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(mockMatrimonyGame)
			setupMatrimonyCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(MatrimonyCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("game over progress uses the foundation goal", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		setupMatrimonyCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetFoundation")
		g.On("GetPhase").Return(domain.MatrimonyPhaseGameOver)
		var full [domain.MatrimonyFoundationCnt][]*domain.Card
		for i := range full {
			for j := 0; j < domain.MatrimonyFoundationTarget; j++ {
				full[i] = append(full[i], domain.NewCard(domain.CardDesignSpade, (j%13)+1, true))
			}
		}
		g.On("GetFoundation").Return(full)

		result := new(MatrimonyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "52/52")
		assert.Contains(t, result, "100%")
	})
}

func TestMatrimonyCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.MatrimonyHint
		contains []string
	}{
		{"tableau to a foundation",
			&domain.MatrimonyHint{FromZone: "tableau", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"タブロー枠1", "基礎札2"}},
		{"between piles",
			&domain.MatrimonyHint{FromZone: "tableau", FromIdx: 0, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー枠0", "タブロー枠5"}},
		{"waste to a foundation",
			&domain.MatrimonyHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0},
			[]string{"捨て札", "基礎札0"}},
		{"stock into a gap",
			&domain.MatrimonyHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"山札", "タブロー枠3"}},
		{"draw from the stock",
			&domain.MatrimonyHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("matrimony.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(mockMatrimonyGame)
			g.On("GetHint").Return(tc.hint)

			result := new(MatrimonyCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		g.On("GetHint").Return((*domain.MatrimonyHint)(nil))

		assert.Contains(t, new(MatrimonyCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestMatrimonyCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		g.On("GetPhase").Return(domain.MatrimonyPhasePlaying)

		assert.Contains(t, new(MatrimonyCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(mockMatrimonyGame)
		g.On("GetPhase").Return(domain.MatrimonyPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(MatrimonyCuiPresenter).ActionLogOutput(g), "move")
	})
}
