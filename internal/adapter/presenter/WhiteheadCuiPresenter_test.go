//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupWhiteheadCuiMockDefaults(kg *interfaces.MockWhiteheadGame) {
	kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying).Maybe()
	kg.On("GetMoveCount").Return(0).Maybe()
	kg.On("GetStockCount").Return(24).Maybe()
	kg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	kg.On("IsStalemate").Return(false).Maybe()
	kg.On("UndoToEscape").Return(0).Maybe()
	kg.On("CanAutoComplete").Return(false).Maybe()
	kg.On("GetDrawCount").Return(1).Maybe()
	kg.On("GetScore").Return(0).Maybe()
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone).Maybe()

	var tableau [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
	for i := 0; i < domain.WhiteheadTableauCnt; i++ {
		tableau[i] = make([]*domain.WhiteheadTableauCard, 0)
		for j := 0; j <= i; j++ {
			tableau[i] = append(tableau[i], &domain.WhiteheadTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: j == i,
			})
		}
	}
	kg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.WhiteheadFoundationCnt][]*domain.Card
	kg.On("GetFoundation").Return(foundation).Maybe()
}

func TestWhiteheadCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	t.Run("initial state", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		p := new(WhiteheadCuiPresenter)

		result := p.Output(kg, nil)
		assert.Contains(t, result, "Whitehead")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "Stock: 24枚")
		assert.Contains(t, result, "Waste: [空]")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		// Header surfaces draw mode, scoring mode, and the running score.
		assert.Contains(t, result, "ドロー: 1枚 / スコアリング: なし / スコア: 0")
	})

	t.Run("three-draw shows a waste fan with only the top playable", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetDrawCount")
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetWaste")
		kg.On("GetDrawCount").Return(3)
		kg.On("GetWaste").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignClover, 9, false),
		})

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "ドロー: 3枚")
		// All three shown, the top (last) card marked playable.
		assert.Contains(t, result, "SPADE 2")
		assert.Contains(t, result, "CLOVER 9*")
		assert.Contains(t, result, "末尾*のみ操作可")
	})

	t.Run("vegas game clear shows final score", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetPhase")
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetScoringMode")
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetScore")
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameClear)
		kg.On("GetScoringMode").Return(domain.WhiteheadScoringVegas)
		kg.On("GetScore").Return(320)

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "ゲームクリア！")
		assert.Contains(t, result, "スコア: 320")
	})

	t.Run("with waste card", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = nil
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetWaste")
		kg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "Waste: HEART 5")
	})

	t.Run("with error", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		p := new(WhiteheadCuiPresenter)

		result := p.Output(kg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetPhase")
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameClear)

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "ゲームクリア！")
	})

	t.Run("game over", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetPhase")
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameOver)

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "IsStalemate")
		kg.On("CanAutoComplete").Return(false).Maybe()
		kg.On("IsStalemate").Return(true)
		kg.On("UndoToEscape").Return(0).Maybe()

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "手詰まりです")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetTableau")
		var emptyTab [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard
		kg.On("GetTableau").Return(emptyTab)

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		kg.ExpectedCalls = filterCallsWH(kg.ExpectedCalls, "GetFoundation")
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		kg.On("GetFoundation").Return(f)

		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		assert.Contains(t, result, "SPADE 1")
	})

	t.Run("face down card shows ??", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadCuiMockDefaults(kg)
		p := new(WhiteheadCuiPresenter)
		result := p.Output(kg, nil)
		// Column 1 has 2 cards: first face-down
		assert.Contains(t, result, "??")
	})
}

func TestWhiteheadCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	t.Run("no hint", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetHint").Return((*domain.WhiteheadHint)(nil))

		p := new(WhiteheadCuiPresenter)
		result := p.HintOutput(kg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("tableau to foundation hint", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetHint").Return(&domain.WhiteheadHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 2,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(WhiteheadCuiPresenter)
		result := p.HintOutput(kg)
		assert.Contains(t, result, "タブロー列0[2]")
		assert.Contains(t, result, "ファンデーション")
	})

	t.Run("waste to tableau hint", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetHint").Return(&domain.WhiteheadHint{
			FromZone:  "waste",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(WhiteheadCuiPresenter)
		result := p.HintOutput(kg)
		assert.Contains(t, result, "ウェイスト")
		assert.Contains(t, result, "タブロー列3")
	})
}

func TestWhiteheadCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	t.Run("during game", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)

		p := new(WhiteheadCuiPresenter)
		result := p.ActionLogOutput(kg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("after game clear", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameClear)
		kg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw", Detail: "test", Cards: nil},
		})

		p := new(WhiteheadCuiPresenter)
		result := p.ActionLogOutput(kg)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "draw")
	})

	t.Run("after game over", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameOver)
		kg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(WhiteheadCuiPresenter)
		result := p.ActionLogOutput(kg)
		assert.Contains(t, result, "棋譜はありません")
	})
}

// filterCallsWH removes mock expectations for a given method name.
func filterCallsWH(calls []*mock.Call, methodName string) []*mock.Call {
	result := make([]*mock.Call, 0, len(calls))
	for _, c := range calls {
		if c.Method != methodName {
			result = append(result, c)
		}
	}
	return result
}

// **オートコンプリートが使える状態かを CUI は出していなかった (#4776)。**
// ac コマンド自体はあるのに、その存在も、いま使えるかも分からない。
func TestWhiteheadCuiPresenter_AutoCompleteReady(t *testing.T) {
	p := new(WhiteheadCuiPresenter)
	game := func(ready bool, phase domain.WhiteheadPhase) *interfaces.MockWhiteheadGame {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("CanAutoComplete").Return(ready)
		kg.On("GetPhase").Return(phase)
		setupWhiteheadCuiMockDefaults(kg)
		return kg
	}

	t.Run("announces the ready state with the command to use", func(t *testing.T) {
		out := p.Output(game(true, domain.WhiteheadPhasePlaying), nil)
		assert.Contains(t, out, "オートコンプリート可能")
		// **コマンド名まで出す。**使えると分かってもコマンドが分からなければ同じ。
		assert.Contains(t, out, "ac")
	})

	t.Run("says nothing while it is not available", func(t *testing.T) {
		out := p.Output(game(false, domain.WhiteheadPhasePlaying), nil)
		assert.NotContains(t, out, "オートコンプリート可能")
	})

	t.Run("says nothing once the game is cleared", func(t *testing.T) {
		out := p.Output(game(false, domain.WhiteheadPhaseGameClear), nil)
		assert.NotContains(t, out, "オートコンプリート可能")
	})
}
