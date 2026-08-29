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

func setupGrandfathersClockCuiMockDefaults(g *interfaces.MockGrandfathersClockGame) {
	g.On("GetPhase").Return(domain.GrandfathersClockPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(false).Maybe()

	var tableau [domain.GrandfathersClockTableauCnt][]*domain.GrandfathersClockTableauCard
	for i := range domain.GrandfathersClockTableauCnt {
		tableau[i] = make([]*domain.GrandfathersClockTableauCard, domain.GrandfathersClockColumnLen)
		for j := range domain.GrandfathersClockColumnLen {
			tableau[i][j] = &domain.GrandfathersClockTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+2, false),
				FaceUp: true,
			}
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.GrandfathersClockFoundationCnt][]*domain.Card
	for i := range domain.GrandfathersClockFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, i+1, false)}
	}
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestGrandfathersClockCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Grandfather's Clock")
		assert.Contains(t, result, i18n.T("grandfathersclock.foundationHeader"))
		assert.Contains(t, result, "[0]")
		assert.Contains(t, result, "[11]", "all twelve faces are rendered")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列7:", "all eight columns are rendered")
		assert.Contains(t, result, "手数: 0")
	})

	// The target is what the player has to plan against, so it must be visible
	// per face rather than left implicit in the clock position.
	t.Run("each face shows its target rank", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, i18n.Tf("grandfathersclock.faceTarget", "rank", "1"))
		assert.Contains(t, result, i18n.Tf("grandfathersclock.faceTarget", "rank", "12"))
	})

	t.Run("completed faces are marked", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		g.On("IsFoundationComplete", mock.AnythingOfType("int")).Return(true)

		assert.Contains(t, new(GrandfathersClockCuiPresenter).Output(g, nil),
			i18n.T("grandfathersclock.faceComplete"))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(GrandfathersClockCuiPresenter).Output(g, nil),
			i18n.Tf("grandfathersclock.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)

		assert.Contains(t, new(GrandfathersClockCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.GrandfathersClockPhase
		want string
	}{
		{"game clear", domain.GrandfathersClockPhaseGameClear, "ゲームクリア"},
		{"game over", domain.GrandfathersClockPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockGrandfathersClockGame)
			setupGrandfathersClockCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(GrandfathersClockCuiPresenter).Output(g, nil), tc.want)
		})
	}

	t.Run("playing phase shows completed faces count and moves", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		for i := range domain.GrandfathersClockFoundationCnt {
			g.On("IsFoundationComplete", i).Return(i < 5)
		}

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "完成した文字盤: 5/12")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("playing phase shows completed faces count in English", func(t *testing.T) {
		origLang := i18n.Lang()
		i18n.SetLang("en")
		defer i18n.SetLang(origLang)

		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		for i := range domain.GrandfathersClockFoundationCnt {
			g.On("IsFoundationComplete", i).Return(i < 5)
		}

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Clock faces complete: 5/12")
		assert.Contains(t, result, "Moves: 0")
	})

	t.Run("game clear phase does not show playing facesComplete line", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.GrandfathersClockPhaseGameClear)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		for i := range domain.GrandfathersClockFoundationCnt {
			g.On("IsFoundationComplete", i).Return(i < 5)
		}

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "ゲームクリア")
		// 件数でなく**ラベルそのもの**が出ないことを見る。"5/12" だけ否定すると、
		// 別の件数で行が出ていても通ってしまう。
		assert.NotContains(t, result, "完成した文字盤:")
	})

	t.Run("game over phase shows summary and does not show playing facesComplete line", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
		g.On("GetPhase").Return(domain.GrandfathersClockPhaseGameOver)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsFoundationComplete")
		for i := range domain.GrandfathersClockFoundationCnt {
			g.On("IsFoundationComplete", i).Return(i < 5)
		}

		result := new(GrandfathersClockCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "ゲームオーバー")
		assert.Contains(t, result, "文字盤 5/12 個を完成")
		// 件数でなく**ラベルそのもの**が出ないことを見る。"5/12" だけ否定すると、
		// 別の件数で行が出ていても通ってしまう。
		assert.NotContains(t, result, "完成した文字盤:")
	})

	t.Run("empty column and empty face", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		setupGrandfathersClockCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetFoundation")
		var emptyTableau [domain.GrandfathersClockTableauCnt][]*domain.GrandfathersClockTableauCard
		var emptyFoundation [domain.GrandfathersClockFoundationCnt][]*domain.Card
		g.On("GetTableau").Return(emptyTableau)
		g.On("GetFoundation").Return(emptyFoundation)

		assert.Contains(t, new(GrandfathersClockCuiPresenter).Output(g, nil), "[空]")
	})
}

func TestGrandfathersClockCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.GrandfathersClockHint
		contains []string
	}{
		{"to a clock face",
			&domain.GrandfathersClockHint{FromCol: 3, ToZone: "foundation", ToIdx: 7},
			[]string{"タブロー列3", "文字盤7"}},
		{"to the tableau",
			&domain.GrandfathersClockHint{FromCol: 1, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー列1", "タブロー列5"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockGrandfathersClockGame)
			g.On("GetHint").Return(tc.hint)

			result := new(GrandfathersClockCuiPresenter).HintOutput(g)
			assert.Contains(t, result, "ヒント")
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		g.On("GetHint").Return((*domain.GrandfathersClockHint)(nil))

		assert.Contains(t, new(GrandfathersClockCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestGrandfathersClockCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		g.On("GetPhase").Return(domain.GrandfathersClockPhasePlaying)

		assert.Contains(t, new(GrandfathersClockCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockGrandfathersClockGame)
		g.On("GetPhase").Return(domain.GrandfathersClockPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(GrandfathersClockCuiPresenter).ActionLogOutput(g), "move")
	})
}
