//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBigBenCuiMockDefaults(g *interfaces.MockBigBenGame) {
	g.On("GetPhase").Return(domain.BigBenPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(52).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(false).Maybe()

	var tableau [domain.BigBenTableauCnt][]*domain.BigBenTableauCard
	for i := range domain.BigBenTableauCnt {
		tableau[i] = make([]*domain.BigBenTableauCard, domain.BigBenColumnLen)
		for j := range domain.BigBenColumnLen {
			tableau[i][j] = &domain.BigBenTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.BigBenFoundationCnt][]*domain.Card
	for i := range domain.BigBenFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, i+1, false)}
	}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestBigBenCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)

		result := new(BigBenCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Big Ben")
		assert.Contains(t, result, i18n.T("bigben.foundationHeader"))
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "[11]", "all twelve faces are rendered")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列7:", "all eight columns are rendered")
		assert.Contains(t, result, "手数: 0")
	})

	// The target is what the player has to plan against, so it must be visible
	// per face rather than left implicit in the clock position.
	t.Run("each face shows its target rank", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)

		result := new(BigBenCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.Tf("bigben.faceTarget", "rank", "1"))
		assert.Contains(t, result, i18n.Tf("bigben.faceTarget", "rank", "12"))
	})

	t.Run("completed faces are marked", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(true)

		assert.Contains(t, new(BigBenCuiPresenter).Output(g, nil),
			i18n.T("bigben.faceComplete"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(BigBenCuiPresenter).Output(g, nil),
			i18n.Tf("bigben.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(BigBenCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)

		assert.Contains(t, new(BigBenCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.BigBenPhase
		want string
	}{
		{"game clear", domain.BigBenPhaseGameClear, "ゲームクリア"},
		{"game over", domain.BigBenPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockBigBenGame)
			setupBigBenCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(BigBenCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("empty column and empty face", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		setupBigBenCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetFoundation")
		var emptyTableau [domain.BigBenTableauCnt][]*domain.BigBenTableauCard
		var emptyFoundation [domain.BigBenFoundationCnt][]*domain.Card
		g.On("GetTableau").Return(emptyTableau)
		g.On("GetFoundation").Return(emptyFoundation)

		assert.Contains(t, new(BigBenCuiPresenter).Output(g, nil), "[空]")
	})
}

func TestBigBenCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.BigBenHint
		contains []string
	}{
		{"to a clock face",
			&domain.BigBenHint{FromCol: 3, ToZone: "foundation", ToIdx: 7},
			[]string{"タブロー列3", "文字盤7"}},
		{"to the tableau",
			&domain.BigBenHint{FromCol: 1, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー列1", "タブロー列5"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockBigBenGame)
			g.On("GetHint").Return(tc.hint)

			result := new(BigBenCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		g.On("GetHint").Return((*domain.BigBenHint)(nil))

		assert.Contains(t, new(BigBenCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestBigBenCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		g.On("GetPhase").Return(domain.BigBenPhasePlaying)

		assert.Contains(t, new(BigBenCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockBigBenGame)
		g.On("GetPhase").Return(domain.BigBenPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(BigBenCuiPresenter).ActionLogOutput(g), "move")
	})
}

// **山札の残りを画面に出す。**補充がこのゲームの逃げ道なので、あと何枚あるかが
// 読めないと「詰んだのか、配れば動くのか」が分からない。クローン元は山札を
// 持たないので、この行そのものが無かった。
func TestBigBenCuiPresenter_ShowsTheStockCount(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	g := new(interfaces.MockBigBenGame)
	setupBigBenCuiMockDefaults(g)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetStockCount")
	g.On("GetStockCount").Return(37)

	assert.Contains(t, new(BigBenCuiPresenter).Output(g, nil), i18n.Tf("bigben.stockLine", "count", "37"))
}
