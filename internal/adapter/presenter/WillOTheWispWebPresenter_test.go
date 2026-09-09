//go:build test

package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupWillOTheWispWebMockDefaults(sg *interfaces.MockWillOTheWispGame) {
	sg.On("GetPhase").Return(domain.WillOTheWispPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(24).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()
	sg.On("CanUndo").Return(false).Maybe()
	sg.On("GetScore").Return(500).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.WillOTheWispTableauCnt][]*domain.WillOTheWispTableauCard
	for i := 0; i < domain.WillOTheWispTableauCnt; i++ {
		tableau[i] = make([]*domain.WillOTheWispTableauCard, 0)
		for j := 0; j <= i%3; j++ {
			tableau[i] = append(tableau[i], &domain.WillOTheWispTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: j == i%3,
			})
		}
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func parseWillOTheWispOutput(t *testing.T, jsonStr string) *controller.WillOTheWispWebOutput {
	t.Helper()
	var out controller.WillOTheWispWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupWillOTheWispOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupWillOTheWispOutputMock(g *interfaces.MockWillOTheWispGame) {
	setupWillOTheWispWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestWillOTheWispWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		p := new(WillOTheWispWebPresenter)

		result := parseWillOTheWispOutput(t, p.Output(sg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, 24, result.StockCount)
		assert.Equal(t, 0, result.CompletedSuits)
		assert.Equal(t, 500, result.Score)
		assert.Len(t, result.Tableau, domain.WillOTheWispTableauCnt)
		assert.Equal(t, "willothewisp.playing", result.MessageCode)
	})

	t.Run("face down card hides data", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		p := new(WillOTheWispWebPresenter)

		result := parseWillOTheWispOutput(t, p.Output(sg, nil))
		assert.Len(t, result.Tableau[1], 2)
		assert.False(t, result.Tableau[1][0].FaceUp)
		assert.Nil(t, result.Tableau[1][0].Card)
		assert.True(t, result.Tableau[1][1].FaceUp)
		assert.NotNil(t, result.Tableau[1][1].Card)
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		p := new(WillOTheWispWebPresenter)

		result := parseWillOTheWispOutput(t, p.Output(sg, assert.AnError))
		assert.Equal(t, assert.AnError.Error(), result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.WillOTheWispPhaseGameClear)

		p := new(WillOTheWispWebPresenter)
		result := parseWillOTheWispOutput(t, p.Output(sg, nil))
		assert.Equal(t, "willothewisp.gameClear", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームクリア")
		assert.NotEmpty(t, result.MessageParams)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.WillOTheWispPhaseGameOver)

		p := new(WillOTheWispWebPresenter)
		result := parseWillOTheWispOutput(t, p.Output(sg, nil))
		assert.Equal(t, "willothewisp.gameOver", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームオーバー")
	})
}

func TestWillOTheWispWebPresenter_Output_Stalemate(t *testing.T) {
	t.Run("no escape available", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "UndoToEscape")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(-1)

		p := new(WillOTheWispWebPresenter)
		result := parseWillOTheWispOutput(t, p.Output(sg, nil))
		assert.True(t, result.IsStalemate)
		assert.Equal(t, "willothewisp.stalemate", result.MessageCode)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("escape available with positive count", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "UndoToEscape")
		sg.On("IsStalemate").Return(true)
		sg.On("UndoToEscape").Return(2)

		p := new(WillOTheWispWebPresenter)
		result := parseWillOTheWispOutput(t, p.Output(sg, nil))
		assert.True(t, result.IsStalemate)
		assert.Equal(t, "willothewisp.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "2", result.MessageParams["count"])
	})
}

func TestWillOTheWispWebPresenter_Output_CanUndo(t *testing.T) {
	sg := new(interfaces.MockWillOTheWispGame)
	setupWillOTheWispOutputMock(sg)
	sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "CanUndo")
	sg.On("CanUndo").Return(true)

	p := new(WillOTheWispWebPresenter)
	result := parseWillOTheWispOutput(t, p.Output(sg, nil))
	assert.True(t, result.CanUndo)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestWillOTheWispWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.WillOTheWispHint{FromCol: 2, CardIndex: 0, ToCol: 4}).Maybe()

		result := new(WillOTheWispWebPresenter).Output(sg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispWebMockDefaults(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)
		sg.On("GetHint").Return(&domain.WillOTheWispHint{FromCol: 2, CardIndex: 0, ToCol: 4}).Maybe()

		result := new(WillOTheWispWebPresenter).Output(sg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestWillOTheWispWebPresenter_HintOutput(t *testing.T) {
	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispWebMockDefaults(sg)
		sg.On("GetHint").Return((*domain.WillOTheWispHint)(nil))

		p := new(WillOTheWispWebPresenter)
		result := parseWillOTheWispOutput(t, p.HintOutput(sg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "willothewisp.noHint", result.MessageCode)
	})

	t.Run("hint available", func(t *testing.T) {
		sg := new(interfaces.MockWillOTheWispGame)
		setupWillOTheWispWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.WillOTheWispHint{FromCol: 0, CardIndex: 2, ToCol: 3})

		p := new(WillOTheWispWebPresenter)
		result := parseWillOTheWispOutput(t, p.HintOutput(sg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 0, result.Hint.FromCol)
		assert.Equal(t, 2, result.Hint.CardIndex)
		assert.Equal(t, 3, result.Hint.ToCol)
		assert.Equal(t, "willothewisp.hintAvailable", result.MessageCode)
	})
}

func TestWillOTheWispWebPresenter_ActionLogOutput(t *testing.T) {
	sg := new(interfaces.MockWillOTheWispGame)
	sg.On("GetGameEndFlag").Return(true)
	sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "move", Detail: "x"},
	})

	p := new(WillOTheWispWebPresenter)
	result := p.ActionLogOutput(sg)
	assert.Contains(t, result, "move")
}

// #5593: スコアの決まりはドメインの定数から渡すこと。3 つの値を取り違えると、
// 説明だけが実際の計算と食い違う。
func TestWillOTheWispWebPresenter_ShipsTheScoringRule(t *testing.T) {
	g := domain.NewDefaultWillOTheWisp()
	g.Reset()

	var out controller.WillOTheWispWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(WillOTheWispWebPresenter).Output(g, nil)), &out))
	assert.Equal(t, domain.WillOTheWispStartScore, out.Scoring.Start)
	assert.Equal(t, domain.WillOTheWispMovePenalty, out.Scoring.MovePenalty)
	assert.Equal(t, domain.WillOTheWispSuitBonus, out.Scoring.SuitBonus)
	// 3 つが同じ値でないこと (取り違えを見逃さない)。
	assert.NotEqual(t, out.Scoring.Start, out.Scoring.SuitBonus)
	assert.NotEqual(t, out.Scoring.MovePenalty, out.Scoring.SuitBonus)
}
