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

func setupEasthavenWebMockDefaults(eg *interfaces.MockEasthavenGame) {
	eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
	eg.On("GetMoveCount").Return(0).Maybe()
	eg.On("GetStockCount").Return(31).Maybe()
	eg.On("CanUndo").Return(false).Maybe()
	eg.On("IsStalemate").Return(false).Maybe()
	eg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.EasthavenTableauCnt {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
	}
	eg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.EasthavenFoundationCnt][]*domain.Card
	eg.On("GetFoundation").Return(foundation).Maybe()
}

func parseEasthavenOutput(t *testing.T, jsonStr string) *controller.EasthavenWebOutput {
	t.Helper()
	var out controller.EasthavenWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupEasthavenOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**
// ようになった (#4483) ので GetHint を呼べるようにする。
func setupEasthavenOutputMock(g *interfaces.MockEasthavenGame) {
	setupEasthavenWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestEasthavenWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenOutputMock(eg)
		p := new(EasthavenWebPresenter)

		result := parseEasthavenOutput(t, p.Output(eg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 31, result.StockCount)
		assert.Equal(t, "easthaven.playing", result.MessageCode)
	})

	t.Run("stalemate with escape", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenOutputMock(eg)
		eg.ExpectedCalls = nil
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(5).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("CanUndo").Return(true).Maybe()
		eg.On("IsStalemate").Return(true).Maybe()
		eg.On("UndoToEscape").Return(3).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenWebPresenter)
		result := parseEasthavenOutput(t, p.Output(eg, nil))
		assert.Equal(t, "easthaven.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "3", result.MessageParams["count"])
	})

	t.Run("stalemate without escape", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenOutputMock(eg)
		eg.ExpectedCalls = nil
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(5).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("CanUndo").Return(false).Maybe()
		eg.On("IsStalemate").Return(true).Maybe()
		eg.On("UndoToEscape").Return(-1).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenWebPresenter)
		result := parseEasthavenOutput(t, p.Output(eg, nil))
		assert.Equal(t, "easthaven.stalemate", result.MessageCode)
	})

	t.Run("game clear", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenOutputMock(eg)
		eg.ExpectedCalls = nil
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameClear).Maybe()
		eg.On("GetMoveCount").Return(42).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("CanUndo").Return(false).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		eg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenWebPresenter)
		result := parseEasthavenOutput(t, p.Output(eg, nil))
		assert.Equal(t, "easthaven.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenOutputMock(eg)
		eg.ExpectedCalls = nil
		eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver).Maybe()
		eg.On("GetMoveCount").Return(10).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("CanUndo").Return(false).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		eg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard
		eg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.EasthavenFoundationCnt][]*domain.Card
		eg.On("GetFoundation").Return(foundation).Maybe()

		p := new(EasthavenWebPresenter)
		result := parseEasthavenOutput(t, p.Output(eg, nil))
		assert.Equal(t, "easthaven.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		setupEasthavenOutputMock(eg)
		p := new(EasthavenWebPresenter)

		result := parseEasthavenOutput(t, p.Output(eg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestEasthavenWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.EasthavenHint{FromCol: 2, CardIndex: 1, ToZone: "foundation", ToCol: 0}

	eg := new(interfaces.MockEasthavenGame)
	setupEasthavenWebMockDefaults(eg)
	eg.On("GetHint").Return(hint).Maybe()

	result := parseEasthavenOutput(t, new(EasthavenWebPresenter).Output(eg, nil))
	if result.Hint == nil {
		t.Fatal("Output must carry the hint -- the frontend reads state.hint")
	}
	assert.Equal(t, 2, result.Hint.FromCol)
}

func TestEasthavenWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(0).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("CanUndo").Return(false).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		eg.On("UndoToEscape").Return(0).Maybe()
		eg.On("GetHint").Return(&domain.EasthavenHint{FromCol: 0, CardIndex: 1, ToZone: "foundation", ToCol: 0})

		p := new(EasthavenWebPresenter)
		result := parseEasthavenOutput(t, p.HintOutput(eg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "easthaven.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		eg := new(interfaces.MockEasthavenGame)
		eg.On("GetPhase").Return(domain.EasthavenPhasePlaying).Maybe()
		eg.On("GetMoveCount").Return(0).Maybe()
		eg.On("GetStockCount").Return(0).Maybe()
		eg.On("CanUndo").Return(false).Maybe()
		eg.On("IsStalemate").Return(false).Maybe()
		eg.On("UndoToEscape").Return(0).Maybe()
		eg.On("GetHint").Return((*domain.EasthavenHint)(nil))

		p := new(EasthavenWebPresenter)
		result := parseEasthavenOutput(t, p.HintOutput(eg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "easthaven.noHint", result.MessageCode)
	})
}

func TestEasthavenWebPresenter_ActionLogOutput(t *testing.T) {
	eg := new(interfaces.MockEasthavenGame)
	eg.On("GetPhase").Return(domain.EasthavenPhaseGameOver)
	eg.On("GetGameEndFlag").Return(true)
	eg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move"}})

	p := new(EasthavenWebPresenter)
	assert.Contains(t, p.ActionLogOutput(eg), "entries")
}
