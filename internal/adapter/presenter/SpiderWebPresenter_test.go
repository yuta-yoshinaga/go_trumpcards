//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSpiderWebMockDefaults(sg *interfaces.MockSpiderGame) {
	sg.On("GetPhase").Return(domain.SpiderPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(50).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("CanUndo").Return(false).Maybe()
	sg.On("GetScore").Return(500).Maybe()
	sg.On("GetDifficulty").Return(domain.SpiderDifficulty1Suit).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()

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

func parseSpiderOutput(t *testing.T, jsonStr string) *controller.SpiderWebOutput {
	t.Helper()
	var out controller.SpiderWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupSpiderOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupSpiderOutputMock(g *interfaces.MockSpiderGame) {
	setupSpiderWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestSpiderWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		p := new(SpiderWebPresenter)

		result := parseSpiderOutput(t, p.Output(sg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, 50, result.StockCount)
		assert.Equal(t, 0, result.CompletedSuits)
		assert.Equal(t, 500, result.Score)
		assert.Equal(t, 1, result.Difficulty)
		assert.Len(t, result.Tableau, domain.SpiderTableauCnt)
		assert.Equal(t, "spider.playing", result.MessageCode)
	})

	t.Run("face down card hides data", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		p := new(SpiderWebPresenter)

		result := parseSpiderOutput(t, p.Output(sg, nil))
		// Column 1 has 2 cards: first face-down
		assert.Len(t, result.Tableau[1], 2)
		assert.False(t, result.Tableau[1][0].FaceUp)
		assert.Nil(t, result.Tableau[1][0].Card)
		assert.True(t, result.Tableau[1][1].FaceUp)
		assert.NotNil(t, result.Tableau[1][1].Card)
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		p := new(SpiderWebPresenter)

		result := parseSpiderOutput(t, p.Output(sg, assert.AnError))
		assert.Equal(t, assert.AnError.Error(), result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SpiderPhaseGameClear)

		p := new(SpiderWebPresenter)
		result := parseSpiderOutput(t, p.Output(sg, nil))
		assert.Equal(t, "spider.gameClear", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームクリア")
		assert.NotEmpty(t, result.MessageParams)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SpiderPhaseGameOver)

		p := new(SpiderWebPresenter)
		result := parseSpiderOutput(t, p.Output(sg, nil))
		assert.Equal(t, "spider.gameOver", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームオーバー")
	})
}

func TestSpiderWebPresenter_Output_Stalemate(t *testing.T) {
	t.Run("no escape available", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "UndoToEscape")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(-1)

		p := new(SpiderWebPresenter)
		result := parseSpiderOutput(t, p.Output(sg, nil))
		assert.True(t, result.IsStalemate)
		assert.Equal(t, "spider.stalemate", result.MessageCode)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("escape available with positive count", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "UndoToEscape")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(2)

		p := new(SpiderWebPresenter)
		result := parseSpiderOutput(t, p.Output(sg, nil))
		assert.True(t, result.IsStalemate)
		assert.Equal(t, "spider.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "2", result.MessageParams["count"])
	})
}

func TestSpiderWebPresenter_Output_CanUndo(t *testing.T) {
	sg := new(interfaces.MockSpiderGame)
	setupSpiderOutputMock(sg)
	sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "CanUndo")
	sg.On("CanUndo").Return(true)

	p := new(SpiderWebPresenter)
	result := parseSpiderOutput(t, p.Output(sg, nil))
	assert.True(t, result.CanUndo)
}

func TestSpiderWebPresenter_Output_Difficulty(t *testing.T) {
	sg := new(interfaces.MockSpiderGame)
	setupSpiderOutputMock(sg)
	sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetDifficulty")
	sg.On("GetDifficulty").Return(domain.SpiderDifficulty4Suit)

	p := new(SpiderWebPresenter)
	result := parseSpiderOutput(t, p.Output(sg, nil))
	assert.Equal(t, int(domain.SpiderDifficulty4Suit), result.Difficulty)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSpiderWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.SpiderHint{FromCol: 3, CardIndex: 0, ToCol: 5}).Maybe()

		result := new(SpiderWebPresenter).Output(sg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		setupSpiderWebMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("GetHint").Return(&domain.SpiderHint{FromCol: 3, CardIndex: 0, ToCol: 5}).Maybe()

		result := new(SpiderWebPresenter).Output(sg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestSpiderWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetHint").Return(&domain.SpiderHint{
			FromCol:   0,
			CardIndex: 2,
			ToCol:     3,
		})
		sg.On("GetPhase").Return(domain.SpiderPhasePlaying)
		sg.On("GetMoveCount").Return(5)
		sg.On("GetStockCount").Return(40)
		sg.On("GetCompletedSuits").Return(1)
		sg.On("CanUndo").Return(false)
		sg.On("GetScore").Return(450)
		sg.On("GetDifficulty").Return(domain.SpiderDifficulty1Suit)
		sg.On("IsStalemate").Return(false)
		sg.On("UndoToEscape").Return(0)

		p := new(SpiderWebPresenter)
		result := parseSpiderOutput(t, p.HintOutput(sg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 0, result.Hint.FromCol)
		assert.Equal(t, 2, result.Hint.CardIndex)
		assert.Equal(t, 3, result.Hint.ToCol)
		assert.Equal(t, "spider.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetHint").Return((*domain.SpiderHint)(nil))
		sg.On("GetPhase").Return(domain.SpiderPhasePlaying)
		sg.On("GetMoveCount").Return(0)
		sg.On("GetStockCount").Return(50)
		sg.On("GetCompletedSuits").Return(0)
		sg.On("CanUndo").Return(false)
		sg.On("GetScore").Return(500)
		sg.On("GetDifficulty").Return(domain.SpiderDifficulty1Suit)
		sg.On("IsStalemate").Return(false)
		sg.On("UndoToEscape").Return(0)

		p := new(SpiderWebPresenter)
		result := parseSpiderOutput(t, p.HintOutput(sg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "spider.noHint", result.MessageCode)
	})
}

func TestSpiderWebPresenter_HintOutput_CanUndo(t *testing.T) {
	sg := new(interfaces.MockSpiderGame)
	sg.On("GetHint").Return((*domain.SpiderHint)(nil))
	sg.On("GetPhase").Return(domain.SpiderPhasePlaying)
	sg.On("GetMoveCount").Return(0)
	sg.On("GetStockCount").Return(50)
	sg.On("GetCompletedSuits").Return(0)
	sg.On("CanUndo").Return(true)
	sg.On("GetScore").Return(500)
	sg.On("GetDifficulty").Return(domain.SpiderDifficulty1Suit)
	sg.On("IsStalemate").Return(false)
	sg.On("UndoToEscape").Return(0)

	p := new(SpiderWebPresenter)
	result := parseSpiderOutput(t, p.HintOutput(sg))
	assert.True(t, result.CanUndo)
}

func TestSpiderWebPresenter_HintOutput_Score(t *testing.T) {
	sg := new(interfaces.MockSpiderGame)
	sg.On("GetHint").Return((*domain.SpiderHint)(nil))
	sg.On("GetPhase").Return(domain.SpiderPhasePlaying)
	sg.On("GetMoveCount").Return(0)
	sg.On("GetStockCount").Return(50)
	sg.On("GetCompletedSuits").Return(0)
	sg.On("CanUndo").Return(false)
	sg.On("GetScore").Return(200)
	sg.On("GetDifficulty").Return(domain.SpiderDifficulty1Suit)
	sg.On("IsStalemate").Return(false)
	sg.On("UndoToEscape").Return(0)

	p := new(SpiderWebPresenter)
	result := parseSpiderOutput(t, p.HintOutput(sg))
	assert.Equal(t, 200, result.Score)
}

func TestSpiderWebPresenter_HintOutput_Difficulty(t *testing.T) {
	sg := new(interfaces.MockSpiderGame)
	sg.On("GetHint").Return((*domain.SpiderHint)(nil))
	sg.On("GetPhase").Return(domain.SpiderPhasePlaying)
	sg.On("GetMoveCount").Return(0)
	sg.On("GetStockCount").Return(50)
	sg.On("GetCompletedSuits").Return(0)
	sg.On("CanUndo").Return(false)
	sg.On("GetScore").Return(500)
	sg.On("GetDifficulty").Return(domain.SpiderDifficulty4Suit)
	sg.On("IsStalemate").Return(false)
	sg.On("UndoToEscape").Return(0)

	p := new(SpiderWebPresenter)
	result := parseSpiderOutput(t, p.HintOutput(sg))
	assert.Equal(t, int(domain.SpiderDifficulty4Suit), result.Difficulty)
}

func TestSpiderWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetPhase").Return(domain.SpiderPhasePlaying)

		sg.On("GetGameEndFlag").Return(false)
		p := new(SpiderWebPresenter)
		result := p.ActionLogOutput(sg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})

	t.Run("after game clear", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetPhase").Return(domain.SpiderPhaseGameClear)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "move", Detail: "test", Cards: nil},
		})

		p := new(SpiderWebPresenter)
		result := p.ActionLogOutput(sg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Len(t, out.Entries, 1)
	})

	t.Run("after game over", func(t *testing.T) {
		sg := new(interfaces.MockSpiderGame)
		sg.On("GetPhase").Return(domain.SpiderPhaseGameOver)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(SpiderWebPresenter)
		result := p.ActionLogOutput(sg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})
}
