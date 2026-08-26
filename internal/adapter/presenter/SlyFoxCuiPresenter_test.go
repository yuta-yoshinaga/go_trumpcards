//go:build test

package presenter

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupSlyFoxCuiMockDefaults(g *interfaces.MockSlyFoxGame) {
	g.On("GetPhase").Return(domain.SlyFoxPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(84).Maybe()
	g.On("DealtThisCycle").Return(domain.SlyFoxDealCycle).Maybe()
	g.On("ReserveIsLocked").Return(false).Maybe()

	var tableau [domain.SlyFoxTableauCnt][]*domain.Card
	for i := range domain.SlyFoxTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.SlyFoxFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()

	// The first half build up from the Ace, the second half down from the King.
	for i := range domain.SlyFoxFoundationCnt {
		g.On("IsAscendingFoundation", i).Return(i < domain.SlyFoxAscendingCnt).Maybe()
	}
}

func TestSlyFoxCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)

		result := new(SlyFoxCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Sly Fox")
		assert.Contains(t, result, i18n.T("slyfox.foundationHeader"))
		assert.Contains(t, result, "枠0:")
		assert.Contains(t, result, "枠19:", "all twenty slots are rendered")
		assert.Contains(t, result, "↑", "ascending foundations are marked")
		assert.Contains(t, result, "↓", "descending foundations are marked")
		assert.Contains(t, result, "84", "the stock count is rendered")
		assert.Contains(t, result, "手数: 0")
	})

	// An empty pile behaves differently from an empty column elsewhere, so the
	// board spells out where a card may come from.
	t.Run("an empty pile says where it can be filled from", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.On("GetTableau").Return([domain.SlyFoxTableauCnt][]*domain.Card{})

		assert.Contains(t, new(SlyFoxCuiPresenter).Output(g, nil), i18n.T("slyfox.emptyPile"))
	})

	// **閉じている理由は盤からは読めない。**あと何枚で開くかを書かないと、
	// 「なぜ送れないのか」が分からないまま手が止まる。
	t.Run("says how many more cards open the reserve", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "DealtThisCycle")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "ReserveIsLocked")
		g.On("DealtThisCycle").Return(7)
		g.On("ReserveIsLocked").Return(true)

		out := new(SlyFoxCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.Tf("slyfox.reserveLocked",
			"dealt", "7", "cycle", strconv.Itoa(domain.SlyFoxDealCycle),
			"left", strconv.Itoa(domain.SlyFoxDealCycle-7)))
		assert.NotContains(t, out, i18n.T("slyfox.reserveOpen"))
	})

	t.Run("says so once the reserve is open", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)

		assert.Contains(t, new(SlyFoxCuiPresenter).Output(g, nil), i18n.T("slyfox.reserveOpen"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(SlyFoxCuiPresenter).Output(g, nil),
			i18n.Tf("slyfox.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(SlyFoxCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxCuiMockDefaults(g)

		assert.Contains(t, new(SlyFoxCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.SlyFoxPhase
		want string
	}{
		{"game clear", domain.SlyFoxPhaseGameClear, "ゲームクリア"},
		{"game over", domain.SlyFoxPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSlyFoxGame)
			setupSlyFoxCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(SlyFoxCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestSlyFoxCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.SlyFoxHint
		contains []string
	}{
		{"tableau to a foundation",
			&domain.SlyFoxHint{FromZone: "tableau", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"リザーブ枠1", "基礎札2"}},
		{"between piles",
			&domain.SlyFoxHint{FromZone: "tableau", FromIdx: 0, ToZone: "tableau", ToIdx: 5},
			[]string{"リザーブ枠0", "リザーブ枠5"}},
		{"deal onto a slot",
			&domain.SlyFoxHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3},
			[]string{"山札", "リザーブ枠3"}},
		{"deal straight to a foundation",
			&domain.SlyFoxHint{FromZone: "stock", FromIdx: -1, ToZone: "foundation", ToIdx: 2},
			[]string{"山札", "基礎札2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSlyFoxGame)
			g.On("GetHint").Return(tc.hint)

			result := new(SlyFoxCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		g.On("GetHint").Return((*domain.SlyFoxHint)(nil))

		assert.Contains(t, new(SlyFoxCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestSlyFoxCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		g.On("GetPhase").Return(domain.SlyFoxPhasePlaying)

		assert.Contains(t, new(SlyFoxCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		g.On("GetPhase").Return(domain.SlyFoxPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(SlyFoxCuiPresenter).ActionLogOutput(g), "move")
	})
}

// **添字は指定できるように見せていただけ** (#5739)。`m t <山>` は山番号しか
// 取らず、動かせるのは常に一番上の 1 枚。埋まった札に [1] などの番号を振ると、
// 指定できない札を指定できるように見せることになる。
func TestSlyFoxCuiPresenter_MarksOnlyTheMovableCard(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	g := new(interfaces.MockSlyFoxGame)
	var tableau [domain.SlyFoxTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 4, true),
		domain.NewCard(domain.CardDesignClover, 7, true),
		domain.NewCard(domain.CardDesignSpade, 11, true), // 一番上 = 唯一動かせる札
	}
	g.On("GetTableau").Return(tableau).Maybe()
	setupSlyFoxCuiMockDefaults(g)

	out := new(SlyFoxCuiPresenter).Output(g, nil)

	// 一番上だけが囲まれ、下の 2 枚は地の文で並ぶ。
	assert.Contains(t, out, "枠0: SPADE 4  CLOVER 7  <SPADE 11>")
	// 添字はどのカードにも付かない。
	assert.NotContains(t, out, "[0]SPADE 4")
	assert.NotContains(t, out, "[1]CLOVER 7")
	assert.NotContains(t, out, "[2]SPADE 11")
	// 読み方の説明も出す。
	assert.Contains(t, out, i18n.T("slyfox.pileTopNote"))
	assert.NotContains(t, out, "{{")
}
