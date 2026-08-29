//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupSirTommyCuiMockDefaults(g *interfaces.MockSirTommyGame) {
	g.On("GetPhase").Return(domain.SirTommyPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(5).Maybe()
	g.On("GetStockTop").Return(domain.NewCard(domain.CardDesignSpade, 7, false)).Maybe()

	var foundations [domain.SirTommyFoundationCnt][]*domain.Card
	for i := range domain.SirTommyFoundationCnt {
		foundations[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	g.On("GetFoundations").Return(foundations).Maybe()

	var wastes [domain.SirTommyWasteCnt][]*domain.Card
	wastes[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 11, false)}
	g.On("GetWastes").Return(wastes).Maybe()
}

func TestSirTommyCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		setupSirTommyCuiMockDefaults(g)
		p := new(SirTommyCuiPresenter)

		result := p.Output(g, nil)
		assert.Contains(t, result, "Sir Tommy")
		assert.Contains(t, result, "[F0")
		assert.Contains(t, result, "[W0]")
		assert.Contains(t, result, "ストック")
	})

	// **次に必要なランクと先読みを常時出す。**Web はバッジで最大6手先まで見せている (#6348)。
	t.Run("next required rank per foundation", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)

		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		// F0: 空 (0枚) -> 次: A -> 2 -> 3 -> 4 -> 5 -> 6
		// F1: 5枚 (一番上が 5) -> 次: 6 -> 7 -> 8 -> 9 -> 10 -> J
		for v := 1; v <= 5; v++ {
			foundations[1] = append(foundations[1], domain.NewCard(domain.CardDesignHeart, v, false))
		}
		// F2: 12枚 (一番上が Q) -> 次: K（13枚打ち止めで先が無いこと）
		for v := 1; v <= 12; v++ {
			foundations[2] = append(foundations[2], domain.NewCard(domain.CardDesignSpade, v, false))
		}
		// F3: 13枚 (完成) -> 完成（先読みが出ないこと）
		for v := 1; v <= domain.CardValueMax; v++ {
			foundations[3] = append(foundations[3], domain.NewCard(domain.CardDesignClover, v, false))
		}
		// **defaults より先に登録する。**testify は最初に一致した期待値を使うので、
		// 先に setup を呼ぶと `.Maybe()` 側の山が勝ってしまう。
		g.On("GetFoundations").Return(foundations)
		setupSirTommyCuiMockDefaults(g)

		result := new(SirTommyCuiPresenter).Output(g, nil)

		var fLines []string
		for _, l := range strings.Split(result, "\n") {
			if strings.HasPrefix(l, "[F") {
				fLines = append(fLines, l)
			}
		}
		require.Len(t, fLines, domain.SirTommyFoundationCnt)

		// F0 (0枚): 次が A で、その後 2,3,4,5,6 と続く
		assert.Contains(t, fLines[0], "(空) → 次: A")
		assert.Contains(t, fLines[0], "A → 2 → 3 → 4 → 5 → 6")

		// F1 (5枚): 次が 6 で、6->7->8->9->10->J と続く（J が cuiRankLabel を通ること）
		assert.Contains(t, fLines[1], "→ 次: 6")
		assert.Contains(t, fLines[1], "6 → 7 → 8 → 9 → 10 → J")

		// F2 (12枚): 次は K だけで、その先が無いこと（空の追記や → が末尾に残らない）
		assert.Contains(t, fLines[2], "→ 次: K")
		assert.NotContains(t, fLines[2], "K →")
		assert.NotContains(t, fLines[2], "14")
		assert.True(t, strings.HasSuffix(strings.TrimSpace(fLines[2]), "→ 次: K"))

		// F3 (13枚): 完成済みは完成だけで先読みが出ない
		assert.Contains(t, fLines[3], "→ 完成")
		assert.NotContains(t, fLines[3], "(13/13) → 次")
	})

	t.Run("upcoming ranks look ahead in en locale", func(t *testing.T) {
		i18n.SetLang("en")
		t.Cleanup(func() { i18n.SetLang("ja") })

		g := new(interfaces.MockSirTommyGame)
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		for v := 1; v <= 5; v++ {
			foundations[1] = append(foundations[1], domain.NewCard(domain.CardDesignHeart, v, false))
		}
		for v := 1; v <= 12; v++ {
			foundations[2] = append(foundations[2], domain.NewCard(domain.CardDesignSpade, v, false))
		}
		for v := 1; v <= domain.CardValueMax; v++ {
			foundations[3] = append(foundations[3], domain.NewCard(domain.CardDesignClover, v, false))
		}
		g.On("GetFoundations").Return(foundations)
		setupSirTommyCuiMockDefaults(g)

		result := new(SirTommyCuiPresenter).Output(g, nil)

		var fLines []string
		for _, l := range strings.Split(result, "\n") {
			if strings.HasPrefix(l, "[F") {
				fLines = append(fLines, l)
			}
		}
		require.Len(t, fLines, domain.SirTommyFoundationCnt)

		assert.Contains(t, fLines[0], "(empty) → next: A → 2 → 3 → 4 → 5 → 6")
		assert.Contains(t, fLines[1], "→ next: 6 → 7 → 8 → 9 → 10 → J")
		assert.Contains(t, fLines[2], "→ next: K")
		assert.NotContains(t, fLines[2], "K →")
		assert.True(t, strings.HasSuffix(strings.TrimSpace(fLines[2]), "→ next: K"))
		assert.Contains(t, fLines[3], "→ complete")
		assert.NotContains(t, fLines[3], "→ next")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		setupSirTommyCuiMockDefaults(g)
		p := new(SirTommyCuiPresenter)
		result := p.Output(g, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhasePlaying).Maybe()
		g.On("GetMoveCount").Return(1).Maybe()
		g.On("IsStalemate").Return(true).Maybe()
		g.On("UndoToEscape").Return(0).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.SirTommyWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := new(SirTommyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhaseGameClear).Maybe()
		g.On("GetMoveCount").Return(100).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.SirTommyWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := new(SirTommyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhaseGameOver).Maybe()
		g.On("GetMoveCount").Return(10).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		var wastes [domain.SirTommyWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := new(SirTommyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestSirTommyCuiPresenter_HintOutput(t *testing.T) {
	t.Run("stock hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetHint").Return(&domain.SirTommyHint{FromZone: "stock", WasteIdx: -1, FoundationIdx: 2, ToZone: "foundation"})
		result := new(SirTommyCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ストック")
		assert.Contains(t, result, "ファンデーション2")
	})

	// #5552: ファンデーションに置けない局面 — このゲームで最も頻繁に起きる —
	// では、どのウェイストに置くかを助言する。
	t.Run("waste placement hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetHint").Return(&domain.SirTommyHint{FromZone: "stock", WasteIdx: 3, FoundationIdx: -1, ToZone: "waste"})
		result := new(SirTommyCuiPresenter).HintOutput(g)
		assert.Contains(t, result, i18n.Tf("sirtommy.hintPlaceWaste", "waste", "3"))
		// **ファンデーションの案内には落とさない。**-1 が漏れる。
		assert.NotContains(t, result, "-1")
	})

	t.Run("waste hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetHint").Return(&domain.SirTommyHint{FromZone: "waste", WasteIdx: 1, FoundationIdx: 0, ToZone: "foundation"})
		result := new(SirTommyCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ウェイスト1")
		assert.Contains(t, result, "ファンデーション0")
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetHint").Return((*domain.SirTommyHint)(nil))
		result := new(SirTommyCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("unknown zone falls through to no-hint", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetHint").Return(&domain.SirTommyHint{FromZone: "???"})
		result := new(SirTommyCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestSirTommyCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhasePlaying)
		assert.NotEmpty(t, new(SirTommyCuiPresenter).ActionLogOutput(g))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)
		g.On("GetPhase").Return(domain.SirTommyPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move", Detail: "test"}})
		assert.NotEmpty(t, new(SirTommyCuiPresenter).ActionLogOutput(g))
	})
}
