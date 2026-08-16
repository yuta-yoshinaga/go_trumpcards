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

func setupCruelWebMockDefaults(cg *interfaces.MockCruelGame) {
	cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
	cg.On("GetMoveCount").Return(0).Maybe()
	cg.On("CanUndo").Return(false).Maybe()
	cg.On("CanAutoComplete").Return(false).Maybe()
	cg.On("IsStalemate").Return(false).Maybe()
	cg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.CruelTableauCnt {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
	}
	cg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.CruelFoundationCnt][]*domain.Card
	cg.On("GetFoundation").Return(foundation).Maybe()
}

func parseCruelOutput(t *testing.T, jsonStr string) *controller.CruelWebOutput {
	t.Helper()
	var out controller.CruelWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupCruelOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupCruelOutputMock(g *interfaces.MockCruelGame) {
	setupCruelWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCruelWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		setupCruelOutputMock(cg)
		p := new(CruelWebPresenter)

		result := parseCruelOutput(t, p.Output(cg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, "cruel.playing", result.MessageCode)
	})

	t.Run("stalemate with escape available", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.ExpectedCalls = nil
		cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
		cg.On("GetMoveCount").Return(5).Maybe()
		cg.On("CanUndo").Return(true).Maybe()
		cg.On("CanAutoComplete").Return(false).Maybe()
		cg.On("IsStalemate").Return(true).Maybe()
		cg.On("UndoToEscape").Return(3).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelWebPresenter)
		result := parseCruelOutput(t, p.Output(cg, nil))
		assert.Equal(t, "cruel.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "3", result.MessageParams["count"])
		assert.True(t, result.IsStalemate)
	})

	t.Run("stalemate without escape available", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
		cg.On("GetMoveCount").Return(5).Maybe()
		cg.On("CanUndo").Return(false).Maybe()
		cg.On("CanAutoComplete").Return(false).Maybe()
		cg.On("IsStalemate").Return(true).Maybe()
		cg.On("UndoToEscape").Return(-1).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelWebPresenter)
		result := parseCruelOutput(t, p.Output(cg, nil))
		assert.Equal(t, "cruel.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("game clear", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhaseGameClear).Maybe()
		cg.On("GetMoveCount").Return(42).Maybe()
		cg.On("CanUndo").Return(false).Maybe()
		cg.On("CanAutoComplete").Return(false).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		cg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelWebPresenter)
		result := parseCruelOutput(t, p.Output(cg, nil))
		assert.Equal(t, "cruel.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhaseGameOver).Maybe()
		cg.On("GetMoveCount").Return(10).Maybe()
		cg.On("CanUndo").Return(false).Maybe()
		cg.On("CanAutoComplete").Return(false).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		cg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
		cg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.CruelFoundationCnt][]*domain.Card
		cg.On("GetFoundation").Return(foundation).Maybe()

		p := new(CruelWebPresenter)
		result := parseCruelOutput(t, p.Output(cg, nil))
		assert.Equal(t, "cruel.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		setupCruelOutputMock(cg)
		p := new(CruelWebPresenter)

		result := parseCruelOutput(t, p.Output(cg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestCruelWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.CruelHint{FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	cg := new(interfaces.MockCruelGame)
	setupCruelWebMockDefaults(cg)
	cg.On("GetHint").Return(hint).Maybe()

	result := parseCruelOutput(t, new(CruelWebPresenter).Output(cg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestCruelWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
		cg.On("GetMoveCount").Return(0).Maybe()
		cg.On("CanUndo").Return(false).Maybe()
		cg.On("CanAutoComplete").Return(false).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		cg.On("UndoToEscape").Return(0).Maybe()
		cg.On("GetHint").Return(&domain.CruelHint{
			FromCol:   0,
			CardIndex: 0,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(CruelWebPresenter)
		result := parseCruelOutput(t, p.HintOutput(cg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "cruel.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying).Maybe()
		cg.On("GetMoveCount").Return(0).Maybe()
		cg.On("CanUndo").Return(false).Maybe()
		cg.On("CanAutoComplete").Return(false).Maybe()
		cg.On("IsStalemate").Return(false).Maybe()
		cg.On("UndoToEscape").Return(0).Maybe()
		cg.On("GetHint").Return((*domain.CruelHint)(nil))

		p := new(CruelWebPresenter)
		result := parseCruelOutput(t, p.HintOutput(cg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "cruel.noHint", result.MessageCode)
	})
}

func TestCruelWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhasePlaying)
		cg.On("GetGameEndFlag").Return(false)

		p := new(CruelWebPresenter)
		result := p.ActionLogOutput(cg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		cg := new(interfaces.MockCruelGame)
		cg.On("GetPhase").Return(domain.CruelPhaseGameOver)
		cg.On("GetGameEndFlag").Return(true)
		cg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move"},
		})

		p := new(CruelWebPresenter)
		result := p.ActionLogOutput(cg)
		assert.Contains(t, result, "entries")
	})
}
