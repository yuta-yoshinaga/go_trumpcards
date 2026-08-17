//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

	// **次に必要なランクを常時出す。**Web はバッジで見せているのに CUI は
	// 一番上の札しか出しておらず、4 本分の暗算を強いていた (#4868)。
	t.Run("next required rank per foundation", func(t *testing.T) {
		g := new(interfaces.MockSirTommyGame)

		var foundations [domain.SirTommyFoundationCnt][]*domain.Card
		// F0: 空 -> A、F1: 5 の上 -> 6、F2: 13 枚 -> 完成、F3: A の上 -> 2
		foundations[1] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)}
		for v := 1; v <= domain.CardValueMax; v++ {
			foundations[2] = append(foundations[2], domain.NewCard(domain.CardDesignClover, v, false))
		}
		foundations[3] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		// **defaults より先に登録する。**testify は最初に一致した期待値を使うので、
		// 先に setup を呼ぶと `.Maybe()` 側の山が勝ってしまう。
		g.On("GetFoundations").Return(foundations)
		setupSirTommyCuiMockDefaults(g)

		result := new(SirTommyCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "(空) → 次: A")
		assert.Contains(t, result, "→ 次: 6")
		assert.Contains(t, result, "→ 完成")
		assert.Contains(t, result, "→ 次: 2")
		// 完成した山に「次」は出さない。完成判定を外すと `→ 次: 14` が出るので、
		// ランクを指定せず「13/13 の行に次が付かないこと」を見る。
		assert.NotContains(t, result, "(13/13) → 次")
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
