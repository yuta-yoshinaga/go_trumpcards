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

func setupNapoleonsSquareCuiMockDefaults(g *interfaces.MockNapoleonsSquareGame) {
	g.On("GetPhase").Return(domain.NapoleonsSquarePhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(48).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, true)}).Maybe()

	var tableau [domain.NapoleonsSquareTableauCnt][]*domain.NapoleonsSquareTableauCard
	for i := range domain.NapoleonsSquareTableauCnt {
		tableau[i] = make([]*domain.NapoleonsSquareTableauCard, domain.NapoleonsSquareColumnLen)
		for j := range domain.NapoleonsSquareColumnLen {
			tableau[i][j] = &domain.NapoleonsSquareTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.NapoleonsSquareFoundationCnt][]*domain.Card
	for i := range domain.NapoleonsSquareFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade+i%4, 1, false)}
	}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestNapoleonsSquareCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)

		result := new(NapoleonsSquareCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Napoleon's Square")
		assert.Contains(t, result, i18n.T("napoleonssquare.foundationHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列11:", "all twelve columns are rendered")
		assert.Contains(t, result, "手数: 0")
		assert.Contains(t, result, "48")
	})

	t.Run("empty waste", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetWaste")
		g.On("GetWaste").Return([]*domain.Card(nil))

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).Output(g, nil),
			i18n.T("napoleonssquare.wasteEmpty"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).Output(g, nil),
			i18n.Tf("napoleonssquare.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(NapoleonsSquareCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	phases := []struct {
		name string
		val  domain.NapoleonsSquarePhase
		want string
	}{
		{"game clear", domain.NapoleonsSquarePhaseGameClear, "ゲームクリア"},
		{"game over", domain.NapoleonsSquarePhaseGameOver, "ゲームオーバー"},
	}
	for _, tc := range phases {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockNapoleonsSquareGame)
			setupNapoleonsSquareCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(NapoleonsSquareCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("empty column and empty foundation", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		setupNapoleonsSquareCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetFoundation")
		var emptyTableau [domain.NapoleonsSquareTableauCnt][]*domain.NapoleonsSquareTableauCard
		var emptyFoundation [domain.NapoleonsSquareFoundationCnt][]*domain.Card
		g.On("GetTableau").Return(emptyTableau)
		g.On("GetFoundation").Return(emptyFoundation)

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).Output(g, nil), "[空]")
	})
}

func TestNapoleonsSquareCuiPresenter_HintOutput(t *testing.T) {
	cases := []struct {
		name     string
		hint     *domain.NapoleonsSquareHint
		contains []string
	}{
		{"waste to foundation",
			&domain.NapoleonsSquareHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: 2},
			[]string{"ウェイスト", "基礎札2"}},
		{"tableau to foundation",
			&domain.NapoleonsSquareHint{FromZone: "tableau", FromCol: 3, CardIndex: 0, ToZone: "foundation", ToCol: 5},
			[]string{"タブロー列3", "基礎札5"}},
		{"tableau to tableau",
			&domain.NapoleonsSquareHint{FromZone: "tableau", FromCol: 1, CardIndex: 2, ToZone: "tableau", ToCol: 9},
			[]string{"タブロー列1[2]", "タブロー列9"}},
		{"turn the stock",
			&domain.NapoleonsSquareHint{FromZone: "stock", FromCol: -1, CardIndex: -1, ToZone: "waste", ToCol: -1},
			[]string{"山札", "ウェイスト"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockNapoleonsSquareGame)
			g.On("GetHint").Return(tc.hint)

			result := new(NapoleonsSquareCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		g.On("GetHint").Return((*domain.NapoleonsSquareHint)(nil))

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestNapoleonsSquareCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		g.On("GetPhase").Return(domain.NapoleonsSquarePhasePlaying)

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockNapoleonsSquareGame)
		g.On("GetPhase").Return(domain.NapoleonsSquarePhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(NapoleonsSquareCuiPresenter).ActionLogOutput(g), "move")
	})
}

// #5554: 勝利条件は 8 つの組札を積み切ることなのに、進捗 (収納枚数) は
// ゲームオーバーになるまでどちらの UI にも出ていなかった。
func TestNapoleonsSquareCuiPresenter_Output_FoundationProgress(t *testing.T) {
	g := new(interfaces.MockNapoleonsSquareGame)
	// 8 山に 1 枚ずつ + 1 山に 2 枚目 = 9 枚。先に登録した期待が優先される。
	var f [domain.NapoleonsSquareFoundationCnt][]*domain.Card
	for i := range domain.NapoleonsSquareFoundationCnt {
		f[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	}
	f[0] = append(f[0], domain.NewCard(domain.CardDesignSpade, 2, false))
	g.On("GetFoundation").Return(f)
	setupNapoleonsSquareCuiMockDefaults(g)

	out := new(NapoleonsSquareCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.Tf("napoleonssquare.foundationProgress",
		"count", "9",
		"total", strconv.Itoa(domain.NapoleonsSquareTotalCards),
		"percent", "9"))
}
