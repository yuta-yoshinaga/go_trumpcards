//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSpideretteCuiMockDefaults(sg *interfaces.MockSpideretteGame) {
	sg.On("GetPhase").Return(domain.SpiderettePhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(24).Maybe()
	sg.On("GetDealsRemaining").Return(0).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetScore").Return(500).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.SpideretteTableauCnt][]*domain.SpideretteTableauCard
	for i := 0; i < domain.SpideretteTableauCnt; i++ {
		tableau[i] = make([]*domain.SpideretteTableauCard, 0)
		for j := 0; j <= i%3; j++ {
			tableau[i] = append(tableau[i], &domain.SpideretteTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: j == i%3,
			})
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestSpideretteCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		p := new(SpideretteCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Spiderette")
		assert.Contains(t, result, "完成スーツ: 0/4")
		assert.Contains(t, result, "山札: 24枚")
		assert.Contains(t, result, "スコア: 500")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		p := new(SpideretteCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SpiderettePhaseGameClear)

		p := new(SpideretteCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア！")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SpiderettePhaseGameOver)

		p := new(SpideretteCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(0).Maybe()

		p := new(SpideretteCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まりです")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
		var emptyTab [domain.SpideretteTableauCnt][]*domain.SpideretteTableauCard
		sg.On("GetTableau").Return(emptyTab)

		p := new(SpideretteCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("face down card shows ??", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		setupSpideretteCuiMockDefaults(sg)
		p := new(SpideretteCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "??")
	})
}

func TestSpideretteCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		sg.On("GetHint").Return((*domain.SpideretteHint)(nil))

		p := new(SpideretteCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint available", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		sg.On("GetHint").Return(&domain.SpideretteHint{FromCol: 0, CardIndex: 2, ToCol: 3})

		p := new(SpideretteCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "タブロー列0[2]")
		assert.Contains(t, result, "タブロー列3")
	})
}

func TestSpideretteCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("during game", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		sg.On("GetGameEndFlag").Return(false)

		p := new(SpideretteCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("after game clear", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "move", Detail: "test", Cards: nil},
		})

		p := new(SpideretteCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "move")
	})

	t.Run("after game over", func(t *testing.T) {
		sg := new(interfaces.MockSpideretteGame)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(SpideretteCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})
}

// **生の残り枚数だけでは「あと何回配れるか」が分からない (#4798)。**Web は
// 切り上げたバッジを出しているのに、CUI は7で割る暗算を強いていた。
func TestSpideretteCuiPresenter_DealsRemaining(t *testing.T) {
	p := new(SpideretteCuiPresenter)
	withDeals := func(stock, deals int) *interfaces.MockSpideretteGame {
		sg := new(interfaces.MockSpideretteGame)
		sg.On("GetStockCount").Return(stock)
		sg.On("GetDealsRemaining").Return(deals)
		setupSpideretteCuiMockDefaults(sg)
		return sg
	}

	t.Run("prints the deal count alongside the raw stock", func(t *testing.T) {
		out := p.Output(withDeals(15, 3), nil)
		assert.Contains(t, out, "15")
		assert.Contains(t, out, "残り配布: 3回")
	})

	// **0回でも出す。**行が消えると「まだ配れるのに表示が無い」のか
	// 「もう配れない」のか区別が付かない。
	t.Run("still says zero once the stock is spent", func(t *testing.T) {
		assert.Contains(t, p.Output(withDeals(0, 0), nil), "残り配布: 0回")
	})
}
