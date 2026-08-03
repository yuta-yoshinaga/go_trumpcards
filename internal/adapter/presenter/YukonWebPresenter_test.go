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

func setupYukonWebMockDefaults(yg *interfaces.MockYukonGame) {
	yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
	yg.On("GetMoveCount").Return(0).Maybe()
	yg.On("CanUndo").Return(false).Maybe()
	yg.On("IsStalemate").Return(false).Maybe()
	yg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.YukonTableauCnt {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
	}
	yg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.YukonFoundationCnt][]*domain.Card
	yg.On("GetFoundation").Return(foundation).Maybe()
}

func parseYukonOutput(t *testing.T, jsonStr string) *controller.YukonWebOutput {
	t.Helper()
	var out controller.YukonWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupYukonOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupYukonOutputMock(g *interfaces.MockYukonGame) {
	setupYukonWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestYukonWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonOutputMock(yg)
		p := new(YukonWebPresenter)

		result := parseYukonOutput(t, p.Output(yg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, "yukon.playing", result.MessageCode)
	})

	t.Run("stalemate with escape available", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonOutputMock(yg)
		yg.ExpectedCalls = nil
		yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
		yg.On("GetMoveCount").Return(5).Maybe()
		yg.On("CanUndo").Return(true).Maybe()
		yg.On("IsStalemate").Return(true).Maybe()
		yg.On("UndoToEscape").Return(3).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonWebPresenter)
		result := parseYukonOutput(t, p.Output(yg, nil))
		assert.Equal(t, "yukon.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "3", result.MessageParams["count"])
		assert.True(t, result.IsStalemate)
	})

	t.Run("stalemate without escape available", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonOutputMock(yg)
		yg.ExpectedCalls = nil
		yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
		yg.On("GetMoveCount").Return(5).Maybe()
		yg.On("CanUndo").Return(false).Maybe()
		yg.On("IsStalemate").Return(true).Maybe()
		yg.On("UndoToEscape").Return(-1).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonWebPresenter)
		result := parseYukonOutput(t, p.Output(yg, nil))
		assert.Equal(t, "yukon.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("game clear", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonOutputMock(yg)
		yg.ExpectedCalls = nil
		yg.On("GetPhase").Return(domain.YukonPhaseGameClear).Maybe()
		yg.On("GetMoveCount").Return(42).Maybe()
		yg.On("CanUndo").Return(false).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		yg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonWebPresenter)
		result := parseYukonOutput(t, p.Output(yg, nil))
		assert.Equal(t, "yukon.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonOutputMock(yg)
		yg.ExpectedCalls = nil
		yg.On("GetPhase").Return(domain.YukonPhaseGameOver).Maybe()
		yg.On("GetMoveCount").Return(10).Maybe()
		yg.On("CanUndo").Return(false).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		yg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonWebPresenter)
		result := parseYukonOutput(t, p.Output(yg, nil))
		assert.Equal(t, "yukon.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonOutputMock(yg)
		p := new(YukonWebPresenter)

		result := parseYukonOutput(t, p.Output(yg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestYukonWebPresenter_OutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonWebMockDefaults(yg)
		yg.On("GetHint").Return(&domain.YukonHint{FromCol: 1, CardIndex: 0, ToZone: "foundation", ToCol: 3}).Maybe()

		result := new(YukonWebPresenter).Output(yg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonWebMockDefaults(yg)
		yg.ExpectedCalls = filterCalls(yg.ExpectedCalls, "IsStalemate")
		yg.On("IsStalemate").Return(true)
		yg.On("GetHint").Return(&domain.YukonHint{FromCol: 1, CardIndex: 0, ToZone: "foundation", ToCol: 3}).Maybe()

		result := new(YukonWebPresenter).Output(yg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestYukonWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
		yg.On("GetMoveCount").Return(0).Maybe()
		yg.On("CanUndo").Return(false).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		yg.On("UndoToEscape").Return(0).Maybe()
		yg.On("GetHint").Return(&domain.YukonHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(YukonWebPresenter)
		result := parseYukonOutput(t, p.HintOutput(yg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "yukon.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
		yg.On("GetMoveCount").Return(0).Maybe()
		yg.On("CanUndo").Return(false).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		yg.On("UndoToEscape").Return(0).Maybe()
		yg.On("GetHint").Return((*domain.YukonHint)(nil))

		p := new(YukonWebPresenter)
		result := parseYukonOutput(t, p.HintOutput(yg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "yukon.noHint", result.MessageCode)
	})
}

func TestYukonWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhasePlaying)

		yg.On("GetGameEndFlag").Return(false)
		p := new(YukonWebPresenter)
		result := p.ActionLogOutput(yg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhaseGameOver)
		yg.On("GetGameEndFlag").Return(true)
		yg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move"},
		})

		p := new(YukonWebPresenter)
		result := p.ActionLogOutput(yg)
		assert.Contains(t, result, "entries")
	})
}
