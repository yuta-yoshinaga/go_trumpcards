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

func setupRoyalCotillionCuiMockDefaults(g *interfaces.MockRoyalCotillionGame) {
	g.On("GetPhase").Return(domain.RoyalCotillionPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(76).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	// タブローは 1 枠 1 枚。枠 3 は空きにして、空き枠の描画も踏む。
	var tableau [domain.RoyalCotillionTableauCnt]*domain.Card
	for i := range domain.RoyalCotillionTableauCnt {
		tableau[i] = domain.NewCard(domain.CardDesignSpade, (i%13)+1, true)
	}
	tableau[3] = nil
	g.On("GetTableau").Return(tableau).Maybe()

	var reserve [domain.RoyalCotillionReserveCnt][]*domain.Card
	for i := range domain.RoyalCotillionReserveCnt {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, i+2, true)}
	}
	reserve[2] = nil
	g.On("GetReserve").Return(reserve).Maybe()

	var foundation [domain.RoyalCotillionFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()

	for i := range domain.RoyalCotillionFoundationCnt {
		g.On("IsOddFoundation", i).Return(i < domain.RoyalCotillionOddCnt).Maybe()
	}
}

func TestRoyalCotillionCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)

		result := new(RoyalCotillionCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Royal Cotillion")
		assert.Contains(t, result, i18n.T("royalcotillion.foundationHeader"))
		assert.Contains(t, result, "[枠0]")
		assert.Contains(t, result, "[枠15]", "all sixteen slots are rendered")
		assert.Contains(t, result, "リザーブ0:")
		assert.Contains(t, result, "リザーブ3:", "all four reserve piles are rendered")
		assert.Contains(t, result, "(空)", "an empty slot says so")
		assert.Contains(t, result, "補充されません", "an emptied reserve never comes back")
		assert.Contains(t, result, "A:", "the Ace-start series is marked")
		assert.Contains(t, result, "2:", "the deuce-start series is marked")
		assert.Contains(t, result, "76", "the stock count is rendered")
		assert.Contains(t, result, "手数: 0")
	})

	// An empty pile behaves differently from an empty column elsewhere, so the
	// board spells out where a card may come from.
	// リザーブは空くと二度と埋まらない。その一点をここで固定する。
	t.Run("an emptied reserve says it never comes back", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetReserve")
		g.On("GetReserve").Return([domain.RoyalCotillionReserveCnt][]*domain.Card{})

		out := new(RoyalCotillionCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("royalcotillion.emptyReserve"))
	})

	t.Run("every slot filled renders no empty marker", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		var full [domain.RoyalCotillionTableauCnt]*domain.Card
		for i := range domain.RoyalCotillionTableauCnt {
			full[i] = domain.NewCard(domain.CardDesignSpade, 7, true)
		}
		g.On("GetTableau").Return(full)

		// 負のコントロール: 上のテストの "(空)" が常に出ているわけではない。
		assert.NotContains(t, new(RoyalCotillionCuiPresenter).Output(g, nil), i18n.T("royalcotillion.emptySlot"))
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(RoyalCotillionCuiPresenter).Output(g, nil), i18n.T("royalcotillion.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(RoyalCotillionCuiPresenter).Output(g, nil),
			i18n.Tf("royalcotillion.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(RoyalCotillionCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		setupRoyalCotillionCuiMockDefaults(g)

		assert.Contains(t, new(RoyalCotillionCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.RoyalCotillionPhase
		want string
	}{
		{"game clear", domain.RoyalCotillionPhaseGameClear, "ゲームクリア"},
		{"game over", domain.RoyalCotillionPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockRoyalCotillionGame)
			setupRoyalCotillionCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(RoyalCotillionCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestRoyalCotillionCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.RoyalCotillionHint
		contains []string
	}{
		{"tableau to a foundation",
			&domain.RoyalCotillionHint{FromZone: "tableau", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"タブロー枠1", "基礎札2"}},
		{"between piles",
			&domain.RoyalCotillionHint{FromZone: "tableau", FromIdx: 0, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー枠0", "タブロー枠5"}},
		{"waste to a foundation",
			&domain.RoyalCotillionHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: 0},
			[]string{"捨て札", "基礎札0"}},
		{"stock into a gap",
			&domain.RoyalCotillionHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"山札", "タブロー枠3"}},
		{"draw from the stock",
			&domain.RoyalCotillionHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("royalcotillion.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockRoyalCotillionGame)
			g.On("GetHint").Return(tc.hint)

			result := new(RoyalCotillionCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		g.On("GetHint").Return((*domain.RoyalCotillionHint)(nil))

		assert.Contains(t, new(RoyalCotillionCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestRoyalCotillionCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		g.On("GetPhase").Return(domain.RoyalCotillionPhasePlaying)

		assert.Contains(t, new(RoyalCotillionCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockRoyalCotillionGame)
		g.On("GetPhase").Return(domain.RoyalCotillionPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(RoyalCotillionCuiPresenter).ActionLogOutput(g), "move")
	})
}
