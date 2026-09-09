//go:build test

package presenter

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupWillOTheWispCuiMockDefaults(sg *interfaces.MockWillOTheWispGame) {
	sg.On("GetPhase").Return(domain.WillOTheWispPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(31).Maybe()
	sg.On("GetDealsRemaining").Return(0).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetScore").Return(500).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.WillOTheWispTableauCnt][]*domain.WillOTheWispTableauCard
	for i := 0; i < domain.WillOTheWispTableauCnt; i++ {
		tableau[i] = make([]*domain.WillOTheWispTableauCard, 0)
		for j := 0; j < domain.WillOTheWispInitialPerColumn; j++ {
			tableau[i] = append(tableau[i], &domain.WillOTheWispTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			})
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestWillOTheWispCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		p := new(WillOTheWispCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, "Will o' the Wisp")
		assert.Contains(t, result, "完成スーツ: 0/4")
		assert.Contains(t, result, "山札: 31枚")
		assert.Contains(t, result, "スコア: 500")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		p := new(WillOTheWispCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.WillOTheWispPhaseGameClear)

		p := new(WillOTheWispCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア！")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.WillOTheWispPhaseGameOver)

		p := new(WillOTheWispCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(0).Maybe()

		p := new(WillOTheWispCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まりです")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
		var emptyTab [domain.WillOTheWispTableauCnt][]*domain.WillOTheWispTableauCard
		sg.On("GetTableau").Return(emptyTab)

		p := new(WillOTheWispCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("initial tableau is face up", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispCuiMockDefaults(sg)
		p := new(WillOTheWispCuiPresenter)
		result := p.Output(sg, nil)
		assert.NotContains(t, result, "??")
	})
}

func TestWillOTheWispCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		sg.On("GetHint").Return((*domain.WillOTheWispHint)(nil))

		p := new(WillOTheWispCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint available", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		sg.On("GetHint").Return(&domain.WillOTheWispHint{FromCol: 0, CardIndex: 2, ToCol: 3})

		p := new(WillOTheWispCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "タブロー列0[2]")
		assert.Contains(t, result, "タブロー列3")
	})
}

func TestWillOTheWispCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("during game", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		sg.On("GetGameEndFlag").Return(false)

		p := new(WillOTheWispCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("after game clear", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "move", Detail: "test", Cards: nil},
		})

		p := new(WillOTheWispCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "move")
	})

	t.Run("after game over", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(WillOTheWispCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})
}

// **生の残り枚数だけでは「あと何回配れるか」が分からない (#4798)。**Web は
// 切り上げたバッジを出しているのに、CUI は7で割る暗算を強いていた。
func TestWillOTheWispCuiPresenter_DealsRemaining(t *testing.T) {
	p := new(WillOTheWispCuiPresenter)
	withDeals := func(stock, deals int) *interfaces.MockWillOTheWispGame {
		sg := new(interfaces.MockWillOTheWispGame)
		sg.On("GetStockCount").Return(stock)
		sg.On("GetDealsRemaining").Return(deals)
		setupWillOTheWispCuiMockDefaults(sg)
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

// #5593: スコアは常時出ているのに、手数のペナルティとスート完成のボーナスは
// どこにも書かれておらず、数字が動く理由を推測させていた。
func TestWillOTheWispCuiPresenter_ExplainsTheScoreRule(t *testing.T) {
	i18n.SetLang("ja")
	g := domain.NewDefaultWillOTheWisp()
	g.Reset()

	out := new(WillOTheWispCuiPresenter).Output(g, nil)
	// **数字はドメインの定数から。**訳文に焼き込むと、計算を変えたとき案内だけが古くなる。
	assert.Contains(t, out, i18n.Tf("willothewisp.scoreRule",
		"start", strconv.Itoa(domain.WillOTheWispStartScore),
		"penalty", strconv.Itoa(domain.WillOTheWispMovePenalty),
		"bonus", strconv.Itoa(domain.WillOTheWispSuitBonus)))
	// 3 つが同じ数字ではないこと。1 つの値を 3 箇所に流す実装を弾く。
	assert.NotEqual(t, domain.WillOTheWispStartScore, domain.WillOTheWispSuitBonus)
	assert.NotEqual(t, domain.WillOTheWispMovePenalty, domain.WillOTheWispSuitBonus)
}
