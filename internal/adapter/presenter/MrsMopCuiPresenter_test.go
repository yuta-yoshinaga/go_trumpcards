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

func setupMrsMopCuiMockDefaults(sg *interfaces.MockMrsMopGame) {
	sg.On("GetPhase").Return(domain.MrsMopPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(50).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("GetScore").Return(500).Maybe()
	sg.On("GetDifficulty").Return(domain.MrsMopDifficulty1Suit).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.MrsMopTableauCnt][]*domain.MrsMopTableauCard
	for i := 0; i < domain.MrsMopTableauCnt; i++ {
		tableau[i] = make([]*domain.MrsMopTableauCard, 0)
		for j := 0; j <= i%3; j++ {
			tableau[i] = append(tableau[i], &domain.MrsMopTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: j == i%3,
			})
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func TestMrsMopCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		p := new(MrsMopCuiPresenter)

		result := p.Output(sg, nil)
		assert.Contains(t, result, i18n.T("mrsmop.helpTitle"))
		assert.Contains(t, result, "完成スーツ: 0/8")
		assert.Contains(t, result, "スコア: 500")
		// **山札の行は出ない。**104枚を配り切るので山札そのものが無い。
		assert.NotContains(t, result, "山札")
		assert.NotContains(t, result, "残りディール")
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "手数: 0")
		// Difficulty/deals header is shown; no column is empty so no deal warning.
		assert.Contains(t, result, strings.Split(i18n.T("mrsmop.difficultyLine"), "{{")[0])
	})

	// **空列は障害ではなく自由枠。**配る操作が無いので「配れない」警告も無い。
	// クローン元の Spider では空列が配りを塞ぐので警告していた。
	t.Run("an empty column draws no warning", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
		var tableau [domain.MrsMopTableauCnt][]*domain.MrsMopTableauCard
		for i := 1; i < domain.MrsMopTableauCnt; i++ { // column 0 left empty
			tableau[i] = []*domain.MrsMopTableauCard{
				{Card: domain.NewCard(domain.CardDesignSpade, 1, false), FaceUp: true},
			}
		}
		sg.On("GetTableau").Return(tableau)
		result := new(MrsMopCuiPresenter).Output(sg, nil)
		assert.Contains(t, result, i18n.T("cuiEmptyCol"), "the empty column still shows")
		assert.NotContains(t, result, "配れ", "but nothing says it blocks a deal")
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		p := new(MrsMopCuiPresenter)

		result := p.Output(sg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.MrsMopPhaseGameClear)

		p := new(MrsMopCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームクリア！")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.MrsMopPhaseGameOver)

		p := new(MrsMopCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(0).Maybe()

		p := new(MrsMopCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "手詰まりです")
	})

	t.Run("empty tableau column", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetTableau")
		var emptyTab [domain.MrsMopTableauCnt][]*domain.MrsMopTableauCard
		sg.On("GetTableau").Return(emptyTab)

		p := new(MrsMopCuiPresenter)
		result := p.Output(sg, nil)
		assert.Contains(t, result, "[空]")
	})

	t.Run("face down card shows ??", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		setupMrsMopCuiMockDefaults(sg)
		p := new(MrsMopCuiPresenter)
		result := p.Output(sg, nil)
		// Columns with more than 1 card have face-down cards
		assert.Contains(t, result, "??")
	})
}

func TestMrsMopCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		sg.On("GetHint").Return((*domain.MrsMopHint)(nil))

		p := new(MrsMopCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("hint available", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		sg.On("GetHint").Return(&domain.MrsMopHint{
			FromCol:   0,
			CardIndex: 2,
			ToCol:     3,
		})

		p := new(MrsMopCuiPresenter)
		result := p.HintOutput(sg)
		assert.Contains(t, result, "タブロー列0[2]")
		assert.Contains(t, result, "タブロー列3")
	})
}

func TestMrsMopCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("during game", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		sg.On("GetPhase").Return(domain.MrsMopPhasePlaying)

		p := new(MrsMopCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("after game clear", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		sg.On("GetPhase").Return(domain.MrsMopPhaseGameClear)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "move", Detail: "test", Cards: nil},
		})

		p := new(MrsMopCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "move")
	})

	t.Run("after game over", func(t *testing.T) {
		sg := new(interfaces.MockMrsMopGame)
		sg.On("GetPhase").Return(domain.MrsMopPhaseGameOver)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(MrsMopCuiPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "棋譜はありません")
	})
}
