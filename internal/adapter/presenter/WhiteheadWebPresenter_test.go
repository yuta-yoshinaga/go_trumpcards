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

func setupWhiteheadWebMockDefaults(kg *interfaces.MockWhiteheadGame) {
	kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying).Maybe()
	kg.On("GetMoveCount").Return(0).Maybe()
	kg.On("GetStockCount").Return(24).Maybe()
	kg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	kg.On("GetDrawCount").Return(1).Maybe()
	kg.On("CanUndo").Return(false).Maybe()
	kg.On("GetScore").Return(-52).Maybe()
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone).Maybe()
	kg.On("IsStalemate").Return(false).Maybe()
	kg.On("UndoToEscape").Return(0).Maybe()

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

func parseWhiteheadOutput(t *testing.T, jsonStr string) *controller.WhiteheadWebOutput {
	t.Helper()
	var out controller.WhiteheadWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupWhiteheadOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupWhiteheadOutputMock(g *interfaces.MockWhiteheadGame) {
	setupWhiteheadWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestWhiteheadWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		p := new(WhiteheadWebPresenter)

		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, 24, result.StockCount)
		assert.Empty(t, result.Waste)
		assert.Len(t, result.Tableau, domain.WhiteheadTableauCnt)
		assert.Len(t, result.Foundation, domain.WhiteheadFoundationCnt)
		assert.Equal(t, "whitehead.playing", result.MessageCode)
	})

	t.Run("waste with cards", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetWaste")
		kg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "HEART", result.Waste[0].Design)
		assert.Equal(t, 5, result.Waste[0].Value)
	})

	t.Run("face down card hides data", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		p := new(WhiteheadWebPresenter)

		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		// Column 1 has 2 cards: first face-down
		assert.Len(t, result.Tableau[1], 2)
		assert.False(t, result.Tableau[1][0].FaceUp)
		assert.Nil(t, result.Tableau[1][0].Card)
		assert.True(t, result.Tableau[1][1].FaceUp)
		assert.NotNil(t, result.Tableau[1][1].Card)
	})

	t.Run("foundation with cards", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetFoundation")
		var f [domain.WhiteheadFoundationCnt][]*domain.Card
		f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		kg.On("GetFoundation").Return(f)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.Len(t, result.Foundation[0], 1)
		assert.Equal(t, "SPADE", result.Foundation[0][0].Design)
	})

	t.Run("with error", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		p := new(WhiteheadWebPresenter)

		result := parseWhiteheadOutput(t, p.Output(kg, assert.AnError))
		assert.Equal(t, assert.AnError.Error(), result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetPhase")
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameClear)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.Equal(t, "whitehead.gameClear", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームクリア")
		assert.NotEmpty(t, result.MessageParams)
	})

	t.Run("game over", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetPhase")
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameOver)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.Equal(t, "whitehead.gameOver", result.MessageCode)
		assert.Contains(t, result.Message, "ゲームオーバー")
	})
}

func TestWhiteheadWebPresenter_Output_Stalemate(t *testing.T) {
	t.Run("no escape available", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "IsStalemate")
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "UndoToEscape")
		kg.On("IsStalemate").Return(true)
		kg.On("UndoToEscape").Return(-1)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.True(t, result.IsStalemate)
		assert.Equal(t, "whitehead.stalemate", result.MessageCode)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("escape available with positive count", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadOutputMock(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "IsStalemate")
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "UndoToEscape")
		kg.On("IsStalemate").Return(true)
		kg.On("UndoToEscape").Return(3)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.Output(kg, nil))
		assert.True(t, result.IsStalemate)
		assert.Equal(t, "whitehead.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "3", result.MessageParams["count"])
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestWhiteheadWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadWebMockDefaults(kg)
		kg.On("GetHint").Return(&domain.WhiteheadHint{FromZone: "tableau", FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(WhiteheadWebPresenter).Output(kg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		setupWhiteheadWebMockDefaults(kg)
		kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "IsStalemate")
		kg.On("IsStalemate").Return(true)
		kg.On("GetHint").Return(&domain.WhiteheadHint{FromZone: "tableau", FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(WhiteheadWebPresenter).Output(kg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestWhiteheadWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetHint").Return(&domain.WhiteheadHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 2,
			ToZone:    "foundation",
			ToCol:     0,
		})
		kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)
		kg.On("GetMoveCount").Return(5)
		kg.On("GetStockCount").Return(20)
		kg.On("GetDrawCount").Return(1)
		kg.On("CanUndo").Return(false)
		kg.On("IsStalemate").Return(false)
		kg.On("GetScore").Return(-52)
		kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone)
		kg.On("UndoToEscape").Return(0)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.HintOutput(kg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 0, result.Hint.FromCol)
		assert.Equal(t, 2, result.Hint.CardIndex)
		assert.Equal(t, "foundation", result.Hint.ToZone)
		assert.Equal(t, "whitehead.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetHint").Return((*domain.WhiteheadHint)(nil))
		kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)
		kg.On("GetMoveCount").Return(0)
		kg.On("GetStockCount").Return(24)
		kg.On("GetDrawCount").Return(1)
		kg.On("CanUndo").Return(false)
		kg.On("IsStalemate").Return(false)
		kg.On("GetScore").Return(-52)
		kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone)
		kg.On("UndoToEscape").Return(0)

		p := new(WhiteheadWebPresenter)
		result := parseWhiteheadOutput(t, p.HintOutput(kg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "whitehead.noHint", result.MessageCode)
	})
}

func TestWhiteheadWebPresenter_Output_DrawCount(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	setupWhiteheadOutputMock(kg)
	kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetDrawCount")
	kg.On("GetDrawCount").Return(3)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.Output(kg, nil))
	assert.Equal(t, 3, result.DrawCount)
}

func TestWhiteheadWebPresenter_Output_CanUndo(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	setupWhiteheadOutputMock(kg)
	kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "CanUndo")
	kg.On("CanUndo").Return(true)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.Output(kg, nil))
	assert.True(t, result.CanUndo)
}

func TestWhiteheadWebPresenter_Output_Score(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	setupWhiteheadOutputMock(kg)
	kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetScore")
	kg.On("GetScore").Return(100)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.Output(kg, nil))
	assert.Equal(t, 100, result.Score)
}

func TestWhiteheadWebPresenter_Output_ScoringMode(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	setupWhiteheadOutputMock(kg)
	kg.ExpectedCalls = filterCalls(kg.ExpectedCalls, "GetScoringMode")
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringVegas)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.Output(kg, nil))
	assert.Equal(t, int(domain.WhiteheadScoringVegas), result.ScoringMode)
}

func TestWhiteheadWebPresenter_HintOutput_DrawCount(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	kg.On("GetHint").Return((*domain.WhiteheadHint)(nil))
	kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)
	kg.On("GetMoveCount").Return(0)
	kg.On("GetStockCount").Return(24)
	kg.On("GetDrawCount").Return(3)
	kg.On("CanUndo").Return(false)
	kg.On("IsStalemate").Return(false)
	kg.On("GetScore").Return(-52)
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone)
	kg.On("UndoToEscape").Return(0)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.HintOutput(kg))
	assert.Equal(t, 3, result.DrawCount)
}

func TestWhiteheadWebPresenter_HintOutput_CanUndo(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	kg.On("GetHint").Return((*domain.WhiteheadHint)(nil))
	kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)
	kg.On("GetMoveCount").Return(0)
	kg.On("GetStockCount").Return(24)
	kg.On("GetDrawCount").Return(1)
	kg.On("CanUndo").Return(true)
	kg.On("IsStalemate").Return(false)
	kg.On("GetScore").Return(-52)
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone)
	kg.On("UndoToEscape").Return(0)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.HintOutput(kg))
	assert.True(t, result.CanUndo)
}

func TestWhiteheadWebPresenter_HintOutput_Score(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	kg.On("GetHint").Return((*domain.WhiteheadHint)(nil))
	kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)
	kg.On("GetMoveCount").Return(0)
	kg.On("GetStockCount").Return(24)
	kg.On("GetDrawCount").Return(1)
	kg.On("CanUndo").Return(false)
	kg.On("IsStalemate").Return(false)
	kg.On("GetScore").Return(200)
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringNone)
	kg.On("UndoToEscape").Return(0)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.HintOutput(kg))
	assert.Equal(t, 200, result.Score)
}

func TestWhiteheadWebPresenter_HintOutput_ScoringMode(t *testing.T) {
	kg := new(interfaces.MockWhiteheadGame)
	kg.On("GetHint").Return((*domain.WhiteheadHint)(nil))
	kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)
	kg.On("GetMoveCount").Return(0)
	kg.On("GetStockCount").Return(24)
	kg.On("GetDrawCount").Return(1)
	kg.On("CanUndo").Return(false)
	kg.On("IsStalemate").Return(false)
	kg.On("GetScore").Return(-52)
	kg.On("GetScoringMode").Return(domain.WhiteheadScoringVegas)
	kg.On("UndoToEscape").Return(0)

	p := new(WhiteheadWebPresenter)
	result := parseWhiteheadOutput(t, p.HintOutput(kg))
	assert.Equal(t, int(domain.WhiteheadScoringVegas), result.ScoringMode)
}

func TestWhiteheadWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetPhase").Return(domain.WhiteheadPhasePlaying)

		kg.On("GetGameEndFlag").Return(false)
		p := new(WhiteheadWebPresenter)
		result := p.ActionLogOutput(kg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})

	t.Run("after game clear", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameClear)
		kg.On("GetGameEndFlag").Return(true)
		kg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "draw", Detail: "test", Cards: nil},
		})

		p := new(WhiteheadWebPresenter)
		result := p.ActionLogOutput(kg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Len(t, out.Entries, 1)
	})

	t.Run("after game over", func(t *testing.T) {
		kg := new(interfaces.MockWhiteheadGame)
		kg.On("GetPhase").Return(domain.WhiteheadPhaseGameOver)
		kg.On("GetGameEndFlag").Return(true)
		kg.On("GetActionLog").Return([]*domain.ActionLogEntry{})

		p := new(WhiteheadWebPresenter)
		result := p.ActionLogOutput(kg)
		var out controller.ActionLogWebOutput
		err := json.Unmarshal([]byte(result), &out)
		assert.NoError(t, err)
		assert.Empty(t, out.Entries)
	})
}
