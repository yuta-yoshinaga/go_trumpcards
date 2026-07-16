//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupSpiderCuiMockDefaults(sg *interfaces.MockSpiderGame) {
	sg.On("GetPhase").Return(domain.SpiderPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(50).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetScore").Return(500).Maybe()
	sg.On("GetDifficulty").Return(domain.SpiderDifficulty1Suit).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()

	var tableau [domain.SpiderTableauCnt][]*domain.SpiderTableauCard
	for i := 0; i < domain.SpiderTableauCnt; i++ {
		tableau[i] = make([]*domain.SpiderTableauCard, 0)
		for j := 0; j <= i%3; j++ {
			tableau[i] = append(tableau[i], &domain.SpiderTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: j == i%3,
			})
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestSpiderCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		p := new(SpiderCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Spider Solitaire")
		assert.Contains(t, result, "完成スーツ: 0/8")
		assert.Contains(t, result, "山札: 50枚")
		assert.Contains(t, result, "スコア: 500")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		// Difficulty/deals header is shown; no column is empty so no deal warning.
		assert.Contains(t, result, strings.Split(i18n.T("spider.difficultyLine"), "{{")[0])
		assert.NotContains(t, result, i18n.T("spider.dealBlockedEmpty"))
	})

	t.Run("deal-blocked warning when a column is empty", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
		var tableau [domain.SpiderTableauCnt][]*domain.SpiderTableauCard
		for i := 1; i < domain.SpiderTableauCnt; i++ { // column 0 left empty
			tableau[i] = []*domain.SpiderTableauCard{
				{Card: domain.NewCard(domain.CardDesignSpade, 1, false), FaceUp: true},
			}
		}
		sg.On("GetTableau").Return(tableau)
		p := new(SpiderCuiPresenter)
		result := p.Output(sg, nil)
		// Stock remains (default 50) and column 0 is empty → deal is blocked.
		assert.Contains(t, result, i18n.T("spider.dealBlockedEmpty"))
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		p := new(SpiderCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SpiderPhaseGameClear)

		p := new(SpiderCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア！")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SpiderPhaseGameOver)

		p := new(SpiderCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)

		p := new(SpiderCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まりです")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
		var emptyTab [domain.SpiderTableauCnt][]*domain.SpiderTableauCard
		sg.On("GetTableau").Return(emptyTab)

		p := new(SpiderCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("face down card shows ??", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderCuiMockDefaults(sg)
		p := new(SpiderCuiPresenter)
		result := p.Output(sg, nil)
		// Columns with more than 1 card have face-down cards
		assert.Contains(t, result, "??")
	})
}

func TestSpiderCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetHint").Return((*domain.SpiderHint)(nil))

		p := new(SpiderCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint available", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetHint").Return(&domain.SpiderHint{
			FromCol:   0,
			CardIndex: 2,
			ToCol:     3,
		})

		p := new(SpiderCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "タブロー列0[2]")
		assert.Contains(t, result, "タブロー列3")
	})
}

func TestSpiderCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("during game", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetPhase").Return(domain.SpiderPhasePlaying)

		p := new(SpiderCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("after game clear", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetPhase").Return(domain.SpiderPhaseGameClear)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "move", Detail: "test", Cards: nil},
		})

		p := new(SpiderCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "move")
	})

	t.Run("after game over", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetPhase").Return(domain.SpiderPhaseGameOver)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(SpiderCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})
}
