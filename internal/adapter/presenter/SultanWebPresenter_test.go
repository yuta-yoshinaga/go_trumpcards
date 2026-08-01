//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupSultanWebMockDefaults(sg *interfaces.MockSultanGame) {
	sg.On("GetPhase").Return(domain.SultanPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("GetStockCount").Return(88).Maybe()
	sg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	sg.On("CanUndo").Return(false).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()
	sg.On("GetRedealCount").Return(0).Maybe()
	sg.On("CanRedeal").Return(false).Maybe()

	divan := make([]*domain.Card, domain.SultanDivanCnt)
	for i := range divan {
		divan[i] = domain.NewCard(domain.CardDesignSpade, i+1, false)
	}
	sg.On("GetDivan").Return(divan).Maybe()

	var foundation [domain.SultanFoundationCnt][]*domain.Card
	for i := range domain.SultanFoundationCnt {
		foundation[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, false)}
	}
	sg.On("GetFoundation").Return(foundation).Maybe()
}

func parseSultanOutput(t *testing.T, jsonStr string) *controller.SultanWebOutput {
	t.Helper()
	var out controller.SultanWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupSultanOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupSultanOutputMock(g *interfaces.MockSultanGame) {
	setupSultanWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestSultanWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		p := new(SultanWebPresenter)

		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 88, result.StockCount)
		assert.Equal(t, 0, result.RedealCount)
		assert.False(t, result.CanRedeal)
		assert.Empty(t, result.Waste)
		assert.Len(t, result.Divan, domain.SultanDivanCnt)
		assert.Len(t, result.Foundation, domain.SultanFoundationCnt)
		assert.Equal(t, "sultan.playing", result.MessageCode)
	})

	t.Run("waste with cards", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetWaste")
		sg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "HEART", result.Waste[0].Design)
		assert.Equal(t, 5, result.Waste[0].Value)
	})

	t.Run("nil divan slot serialised as null", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetDivan")
		divan := make([]*domain.Card, domain.SultanDivanCnt) // all nil
		sg.On("GetDivan").Return(divan)

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.Len(t, result.Divan, domain.SultanDivanCnt)
		assert.Nil(t, result.Divan[0])
	})

	t.Run("redeal available", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "CanRedeal")
		sg.On("CanRedeal").Return(true)

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.True(t, result.CanRedeal)
	})

	t.Run("error message", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		p := new(SultanWebPresenter)

		result := parseSultanOutput(t, p.Output(sg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SultanPhaseGameClear)

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.Equal(t, "sultan.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "GetPhase")
		sg.On("GetPhase").Return(domain.SultanPhaseGameOver)

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.Equal(t, "sultan.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanOutputMock(sg)
		sg.ExpectedCalls = filterCalls(sg.ExpectedCalls, "IsStalemate")
		sg.On("IsStalemate").Return(true)

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.Output(sg, nil))
		assert.Equal(t, "sultan.stalemate", result.MessageCode)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestSultanWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sug := new(interfaces.MockSultanGame)
		setupSultanWebMockDefaults(sug)
		sug.On("GetHint").Return(&domain.SultanHint{FromZone: "divan", FromIdx: 3, ToFoundation: 1}).Maybe()

		result := new(SultanWebPresenter).Output(sug, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		sug := new(interfaces.MockSultanGame)
		setupSultanWebMockDefaults(sug)
		sug.ExpectedCalls = filterCalls(sug.ExpectedCalls, "IsStalemate")
		sug.On("IsStalemate").Return(true)
		sug.On("GetHint").Return(&domain.SultanHint{FromZone: "divan", FromIdx: 3, ToFoundation: 1}).Maybe()

		result := new(SultanWebPresenter).Output(sug, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestSultanWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.SultanHint{FromZone: "divan", FromIdx: 2, ToFoundation: 0})

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.HintOutput(sg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "divan", result.Hint.FromZone)
		assert.Equal(t, "sultan.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		setupSultanWebMockDefaults(sg)
		sg.On("GetHint").Return((*domain.SultanHint)(nil))

		p := new(SultanWebPresenter)
		result := parseSultanOutput(t, p.HintOutput(sg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "sultan.noHint", result.MessageCode)
	})
}

func TestSultanWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetPhase").Return(domain.SultanPhasePlaying)
		sg.On("GetGameEndFlag").Return(false)

		p := new(SultanWebPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		sg := new(interfaces.MockSultanGame)
		sg.On("GetPhase").Return(domain.SultanPhaseGameOver)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(SultanWebPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "draw")
	})
}
