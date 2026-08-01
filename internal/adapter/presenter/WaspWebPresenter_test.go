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

func setupWaspWebMockDefaults(sg *interfaces.MockWaspGame) {
	sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
	sg.On("GetMoveCount").Return(0).Maybe()
	sg.On("CanUndo").Return(false).Maybe()
	sg.On("IsStalemate").Return(false).Maybe()
	sg.On("UndoToEscape").Return(0).Maybe()
	sg.On("GetStockCount").Return(3).Maybe()
	sg.On("GetCompletedSuits").Return(0).Maybe()

	var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.WaspTableauCnt {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
	}
	sg.On("GetTableau").Return(tableau).Maybe()
}

func parseWaspOutput(t *testing.T, jsonStr string) *controller.WaspWebOutput {
	t.Helper()
	var out controller.WaspWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupWaspOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupWaspOutputMock(g *interfaces.MockWaspGame) {
	setupWaspWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestWaspWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		setupWaspOutputMock(sg)
		p := new(WaspWebPresenter)

		result := parseWaspOutput(t, p.Output(sg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 3, result.StockCount)
		assert.Equal(t, "wasp.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying).Maybe()
		sg.On("GetMoveCount").Return(5).Maybe()
		sg.On("CanUndo").Return(true).Maybe()
		sg.On("IsStalemate").Return(true).Maybe()
		sg.On("UndoToEscape").Return(3).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspWebPresenter)
		result := parseWaspOutput(t, p.Output(sg, nil))
		assert.Equal(t, "wasp.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
	})

	t.Run("game clear", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameClear).Maybe()
		sg.On("GetMoveCount").Return(42).Maybe()
		sg.On("CanUndo").Return(false).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("UndoToEscape").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(4).Maybe()
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspWebPresenter)
		result := parseWaspOutput(t, p.Output(sg, nil))
		assert.Equal(t, "wasp.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameOver).Maybe()
		sg.On("GetMoveCount").Return(10).Maybe()
		sg.On("CanUndo").Return(false).Maybe()
		sg.On("IsStalemate").Return(false).Maybe()
		sg.On("UndoToEscape").Return(0).Maybe()
		sg.On("GetStockCount").Return(0).Maybe()
		sg.On("GetCompletedSuits").Return(0).Maybe()
		var tableau [domain.WaspTableauCnt][]*domain.KlondikeTableauCard
		sg.On("GetTableau").Return(tableau).Maybe()

		p := new(WaspWebPresenter)
		result := parseWaspOutput(t, p.Output(sg, nil))
		assert.Equal(t, "wasp.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		setupWaspOutputMock(sg)
		p := new(WaspWebPresenter)

		result := parseWaspOutput(t, p.Output(sg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestWaspWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		wg := new(interfaces.MockWaspGame)
		setupWaspWebMockDefaults(wg)
		wg.On("GetHint").Return(&domain.WaspHint{FromCol: 1, CardIndex: 0, ToCol: 4}).Maybe()

		result := new(WaspWebPresenter).Output(wg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		wg := new(interfaces.MockWaspGame)
		setupWaspWebMockDefaults(wg)
		wg.ExpectedCalls = filterCalls(wg.ExpectedCalls, "IsStalemate")
		wg.On("IsStalemate").Return(true)
		wg.On("GetHint").Return(&domain.WaspHint{FromCol: 1, CardIndex: 0, ToCol: 4}).Maybe()

		result := new(WaspWebPresenter).Output(wg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestWaspWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		setupWaspWebMockDefaults(sg)
		sg.On("GetHint").Return(&domain.WaspHint{FromCol: 0, CardIndex: 1, ToCol: 3})

		p := new(WaspWebPresenter)
		result := parseWaspOutput(t, p.HintOutput(sg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 3, result.Hint.ToCol)
		assert.Equal(t, "wasp.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		setupWaspWebMockDefaults(sg)
		sg.On("GetHint").Return((*domain.WaspHint)(nil))

		p := new(WaspWebPresenter)
		result := parseWaspOutput(t, p.HintOutput(sg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "wasp.noHint", result.MessageCode)
	})
}

func TestWaspWebPresenter_LegalMovesOutput(t *testing.T) {
	sg := new(interfaces.MockWaspGame)
	setupWaspOutputMock(sg)
	p := new(WaspWebPresenter)

	// Web delegates legal-move previews to the normal state JSON (targets are
	// computed client-side), so the output mirrors Output.
	result := parseWaspOutput(t, p.LegalMovesOutput(sg, 0))
	assert.Equal(t, "wasp.playing", result.MessageCode)
}

func TestWaspWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhasePlaying)

		sg.On("GetGameEndFlag").Return(false)
		p := new(WaspWebPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		sg := new(interfaces.MockWaspGame)
		sg.On("GetPhase").Return(domain.WaspPhaseGameOver)
		sg.On("GetGameEndFlag").Return(true)
		sg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move"}})

		p := new(WaspWebPresenter)
		result := p.ActionLogOutput(sg)
		assert.Contains(t, result, "entries")
	})
}
