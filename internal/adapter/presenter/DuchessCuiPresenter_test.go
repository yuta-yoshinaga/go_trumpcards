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

func setupDuchessCuiMockDefaults(g *interfaces.MockDuchessGame) {
	g.On("GetPhase").Return(domain.DuchessPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(35).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("IsAwaitingBaseRank").Return(false).Maybe()

	var reserve [domain.DuchessReserveCnt][]*domain.Card
	for i := range domain.DuchessReserveCnt {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignClover, i+2, true)}
	}
	g.On("GetReserve").Return(reserve).Maybe()

	var tableau [domain.DuchessTableauCnt][]*domain.DuchessTableauCard
	for i := range domain.DuchessTableauCnt {
		tableau[i] = []*domain.DuchessTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.DuchessFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestDuchessCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)

		result := new(DuchessCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Duchess")
		assert.Contains(t, result, i18n.T("duchess.foundationHeader"))
		assert.Contains(t, result, i18n.T("duchess.reserveHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列3:", "all four columns are rendered")
		assert.Contains(t, result, "35")
		assert.Contains(t, result, i18n.Tf("duchess.baseRankLine", "rank", "5"))
		assert.Contains(t, result, "手数: 0")
	})

	// Until the base rank is chosen nothing else is legal, so the board leads with it.
	t.Run("awaiting the base rank is called out", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.On("IsAwaitingBaseRank").Return(true)

		result := new(DuchessCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.T("duchess.awaitingBase"))
		assert.NotContains(t, result, i18n.Tf("duchess.baseRankLine", "rank", "5"))
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(DuchessCuiPresenter).Output(g, nil), i18n.T("duchess.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(DuchessCuiPresenter).Output(g, nil),
			i18n.Tf("duchess.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(DuchessCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)

		assert.Contains(t, new(DuchessCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.DuchessPhase
		want string
	}{
		{"game clear", domain.DuchessPhaseGameClear, "ゲームクリア"},
		{"game over", domain.DuchessPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockDuchessGame)
			setupDuchessCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(DuchessCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("empty column, empty fan and empty foundation", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetReserve")
		var emptyTableau [domain.DuchessTableauCnt][]*domain.DuchessTableauCard
		var emptyReserve [domain.DuchessReserveCnt][]*domain.Card
		g.On("GetTableau").Return(emptyTableau)
		g.On("GetReserve").Return(emptyReserve)

		assert.Contains(t, new(DuchessCuiPresenter).Output(g, nil), "[空]")
	})
}

func TestDuchessCuiPresenter_HintOutput(t *testing.T) {
	// Before the base rank is chosen the hint is an instruction, not a move.
	t.Run("choose the base rank", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		g.On("IsAwaitingBaseRank").Return(true)
		g.On("GetHint").Return(&domain.DuchessHint{
			FromZone: "reserve", FromIdx: 2, CardIndex: -1, ToZone: "foundation", ToIdx: -1,
		})

		assert.Contains(t, new(DuchessCuiPresenter).HintOutput(g),
			i18n.Tf("duchess.hintChooseBase", "idx", "2"))
	})

	for _, tc := range []struct {
		name     string
		hint     *domain.DuchessHint
		contains []string
	}{
		{"reserve to a foundation",
			&domain.DuchessHint{FromZone: "reserve", FromIdx: 2, CardIndex: -1, ToZone: "foundation", ToIdx: 1},
			[]string{"リザーブ扇2", "基礎札1"}},
		{"reserve to the tableau",
			&domain.DuchessHint{FromZone: "reserve", FromIdx: 0, CardIndex: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"リザーブ扇0", "タブロー列3"}},
		{"waste to a foundation",
			&domain.DuchessHint{FromZone: "waste", FromIdx: -1, CardIndex: -1, ToZone: "foundation", ToIdx: 2},
			[]string{"ウェイスト", "基礎札2"}},
		{"tableau to tableau",
			&domain.DuchessHint{FromZone: "tableau", FromIdx: 1, CardIndex: 2, ToZone: "tableau", ToIdx: 3},
			[]string{"タブロー列1[2]", "タブロー列3"}},
		{"draw from the stock",
			&domain.DuchessHint{FromZone: "stock", FromIdx: -1, CardIndex: -1, ToZone: "waste", ToIdx: -1},
			[]string{"山札", i18n.T("duchess.hintToWaste")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockDuchessGame)
			g.On("IsAwaitingBaseRank").Return(false)
			g.On("GetHint").Return(tc.hint)

			result := new(DuchessCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		g.On("GetHint").Return((*domain.DuchessHint)(nil))

		assert.Contains(t, new(DuchessCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestDuchessCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		g.On("GetPhase").Return(domain.DuchessPhasePlaying)

		assert.Contains(t, new(DuchessCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		g.On("GetPhase").Return(domain.DuchessPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(DuchessCuiPresenter).ActionLogOutput(g), "move")
	})
}
