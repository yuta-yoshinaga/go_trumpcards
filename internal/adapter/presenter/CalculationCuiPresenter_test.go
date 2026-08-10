//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCalculationCuiMockDefaults(g *interfaces.MockCalculationGame) {
	g.On("GetPhase").Return(domain.CalculationPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(5).Maybe()
	g.On("GetStockTop").Return(domain.NewCard(domain.CardDesignSpade, 7, false)).Maybe()

	var foundations [domain.CalculationFoundationCnt][]*domain.Card
	for i := range domain.CalculationFoundationCnt {
		foundations[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)}
	}
	g.On("GetFoundations").Return(foundations).Maybe()
	g.On("GetNextFoundationRank", mock.Anything).Return(0).Maybe()

	var wastes [domain.CalculationWasteCnt][]*domain.Card
	wastes[0] = []*domain.Card{domain.NewCard(domain.CardDesignHeart, 11, false)}
	g.On("GetWastes").Return(wastes).Maybe()
}

func TestCalculationCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		setupCalculationCuiMockDefaults(g)
		p := new(CalculationCuiPresenter)

		result := p.Output(g, nil)
		assert.Contains(t, result, "Calculation")
		assert.Contains(t, result, "[F0")
		assert.Contains(t, result, "[W0]")
		assert.Contains(t, result, "ストック")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		setupCalculationCuiMockDefaults(g)
		p := new(CalculationCuiPresenter)
		result := p.Output(g, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhasePlaying).Maybe()
		g.On("GetMoveCount").Return(1).Maybe()
		g.On("IsStalemate").Return(true).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		g.On("GetNextFoundationRank", mock.Anything).Return(0).Maybe()
		var wastes [domain.CalculationWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := new(CalculationCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhaseGameClear).Maybe()
		g.On("GetMoveCount").Return(100).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		g.On("GetNextFoundationRank", mock.Anything).Return(0).Maybe()
		var wastes [domain.CalculationWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := new(CalculationCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhaseGameOver).Maybe()
		g.On("GetMoveCount").Return(10).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		g.On("GetFoundations").Return(foundations).Maybe()
		g.On("GetNextFoundationRank", mock.Anything).Return(0).Maybe()
		var wastes [domain.CalculationWasteCnt][]*domain.Card
		g.On("GetWastes").Return(wastes).Maybe()

		result := new(CalculationCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestCalculationCuiPresenter_HintOutput(t *testing.T) {
	t.Run("stock hint", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetHint").Return(&domain.CalculationHint{FromZone: "stock", WasteIdx: -1, FoundationIdx: 2})
		result := new(CalculationCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ストック")
		assert.Contains(t, result, "ファンデーション2")
	})

	t.Run("waste hint", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetHint").Return(&domain.CalculationHint{FromZone: "waste", WasteIdx: 1, FoundationIdx: 0})
		result := new(CalculationCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ウェイスト1")
		assert.Contains(t, result, "ファンデーション0")
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetHint").Return((*domain.CalculationHint)(nil))
		result := new(CalculationCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("unknown zone falls through to no-hint", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetHint").Return(&domain.CalculationHint{FromZone: "???"})
		result := new(CalculationCuiPresenter).HintOutput(g)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestCalculationCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhasePlaying)
		assert.NotEmpty(t, new(CalculationCuiPresenter).ActionLogOutput(g))
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockCalculationGame)
		g.On("GetPhase").Return(domain.CalculationPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move", Detail: "test"}})
		assert.NotEmpty(t, new(CalculationCuiPresenter).ActionLogOutput(g))
	})
}

// **各列が +1/+2/+3/+4 ずつ 13 を法として進む (#4794)。**CUI は一番上の札しか
// 出さず、毎手この暗算を強いていた。
func TestCalculationCuiPresenter_NextRank(t *testing.T) {
	p := new(CalculationCuiPresenter)
	card := func(v int) *domain.Card { return domain.NewCard(domain.CardDesignSpade, v, false) }
	withRanks := func(ranks ...int) *interfaces.MockCalculationGame {
		g := new(interfaces.MockCalculationGame)
		var foundations [domain.CalculationFoundationCnt][]*domain.Card
		for i := range foundations {
			foundations[i] = []*domain.Card{card(i + 1)}
		}
		g.On("GetFoundations").Return(foundations).Maybe()
		for i, r := range ranks {
			g.On("GetNextFoundationRank", i).Return(r).Maybe()
		}
		g.On("GetNextFoundationRank", mock.Anything).Return(0).Maybe()
		g.On("GetWastes").Return([domain.CalculationWasteCnt][]*domain.Card{}).Maybe()
		g.On("GetStockCount").Return(0).Maybe()
		g.On("GetStockTop").Return((*domain.Card)(nil)).Maybe()
		g.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
		g.On("GetPhase").Return(domain.CalculationPhasePlaying).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("IsStalemate").Return(false).Maybe()
		return g
	}

	t.Run("names the next rank for each pile", func(t *testing.T) {
		out := p.Output(withRanks(2, 4, 6, 8), nil)
		assert.Equal(t, 4, strings.Count(out, "→ 次:"))
		assert.Contains(t, out, "→ 次: 2")
		assert.Contains(t, out, "→ 次: 8")
	})

	// **もう置けない山には出さない。**13枚そろった山に「次」は無い。
	t.Run("omits the hint for a pile that is finished", func(t *testing.T) {
		out := p.Output(withRanks(2, 0, 6, 0), nil)
		assert.Equal(t, 2, strings.Count(out, "→ 次:"))
	})
}
