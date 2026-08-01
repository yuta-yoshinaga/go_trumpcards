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

func setupRussianSolitaireWebMockDefaults(rg *interfaces.MockRussianSolitaireGame) {
	rg.On("GetPhase").Return(domain.RussianSolitairePhasePlaying).Maybe()
	rg.On("GetMoveCount").Return(0).Maybe()
	rg.On("CanUndo").Return(false).Maybe()
	rg.On("IsStalemate").Return(false).Maybe()
	rg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.RussianSolitaireTableauCnt {
		tableau[i] = make([]*domain.KlondikeTableauCard, 0)
	}
	rg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.RussianSolitaireFoundationCnt][]*domain.Card
	rg.On("GetFoundation").Return(foundation).Maybe()
}

func parseRussianSolitaireOutput(t *testing.T, jsonStr string) *controller.RussianSolitaireWebOutput {
	t.Helper()
	var out controller.RussianSolitaireWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupRussianSolitaireOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupRussianSolitaireOutputMock(g *interfaces.MockRussianSolitaireGame) {
	setupRussianSolitaireWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestRussianSolitaireWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireOutputMock(rg)
		p := new(RussianSolitaireWebPresenter)

		result := parseRussianSolitaireOutput(t, p.Output(rg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, "russiansolitaire.playing", result.MessageCode)
	})

	t.Run("stalemate with escape available", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.RussianSolitairePhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(5).Maybe()
		rg.On("CanUndo").Return(true).Maybe()
		rg.On("IsStalemate").Return(true).Maybe()
		rg.On("UndoToEscape").Return(3).Maybe()
		var tableau [domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.RussianSolitaireFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(RussianSolitaireWebPresenter)
		result := parseRussianSolitaireOutput(t, p.Output(rg, nil))
		assert.Equal(t, "russiansolitaire.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "3", result.MessageParams["count"])
		assert.True(t, result.IsStalemate)
	})

	t.Run("stalemate without escape available", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.RussianSolitairePhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(5).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(true).Maybe()
		rg.On("UndoToEscape").Return(-1).Maybe()
		var tableau [domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.RussianSolitaireFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(RussianSolitaireWebPresenter)
		result := parseRussianSolitaireOutput(t, p.Output(rg, nil))
		assert.Equal(t, "russiansolitaire.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("game clear", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.RussianSolitairePhaseGameClear).Maybe()
		rg.On("GetMoveCount").Return(42).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.RussianSolitaireFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(RussianSolitaireWebPresenter)
		result := parseRussianSolitaireOutput(t, p.Output(rg, nil))
		assert.Equal(t, "russiansolitaire.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.RussianSolitairePhaseGameOver).Maybe()
		rg.On("GetMoveCount").Return(10).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.RussianSolitaireFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(RussianSolitaireWebPresenter)
		result := parseRussianSolitaireOutput(t, p.Output(rg, nil))
		assert.Equal(t, "russiansolitaire.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireOutputMock(rg)
		p := new(RussianSolitaireWebPresenter)

		result := parseRussianSolitaireOutput(t, p.Output(rg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestRussianSolitaireWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		rsg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireWebMockDefaults(rsg)
		rsg.On("GetHint").Return(&domain.RussianSolitaireHint{FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(RussianSolitaireWebPresenter).Output(rsg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		rsg := new(interfaces.MockRussianSolitaireGame)
		setupRussianSolitaireWebMockDefaults(rsg)
		rsg.ExpectedCalls = filterCalls(rsg.ExpectedCalls, "IsStalemate")
		rsg.On("IsStalemate").Return(true)
		rsg.On("GetHint").Return(&domain.RussianSolitaireHint{FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(RussianSolitaireWebPresenter).Output(rsg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestRussianSolitaireWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		rg.On("GetPhase").Return(domain.RussianSolitairePhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(0).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		rg.On("GetHint").Return(&domain.RussianSolitaireHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(RussianSolitaireWebPresenter)
		result := parseRussianSolitaireOutput(t, p.HintOutput(rg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "russiansolitaire.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		rg.On("GetPhase").Return(domain.RussianSolitairePhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(0).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		rg.On("GetHint").Return((*domain.RussianSolitaireHint)(nil))

		p := new(RussianSolitaireWebPresenter)
		result := parseRussianSolitaireOutput(t, p.HintOutput(rg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "russiansolitaire.noHint", result.MessageCode)
	})
}

func TestRussianSolitaireWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		rg.On("GetPhase").Return(domain.RussianSolitairePhasePlaying)

		rg.On("GetGameEndFlag").Return(false)
		p := new(RussianSolitaireWebPresenter)
		result := p.ActionLogOutput(rg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		rg := new(interfaces.MockRussianSolitaireGame)
		rg.On("GetPhase").Return(domain.RussianSolitairePhaseGameOver)
		rg.On("GetGameEndFlag").Return(true)
		rg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move"},
		})

		p := new(RussianSolitaireWebPresenter)
		result := p.ActionLogOutput(rg)
		assert.Contains(t, result, "entries")
	})
}
