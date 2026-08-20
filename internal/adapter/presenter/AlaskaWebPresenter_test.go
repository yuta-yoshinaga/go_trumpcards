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

func setupAlaskaWebMockDefaults(rg *interfaces.MockAlaskaGame) {
	rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
	rg.On("GetMoveCount").Return(0).Maybe()
	rg.On("CanUndo").Return(false).Maybe()
	rg.On("IsStalemate").Return(false).Maybe()
	rg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
	for i := range domain.AlaskaTableauCnt {
		tableau[i] = make([]*domain.AlaskaTableauCard, 0)
	}
	rg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.AlaskaFoundationCnt][]*domain.Card
	rg.On("GetFoundation").Return(foundation).Maybe()
}

func parseAlaskaOutput(t *testing.T, jsonStr string) *controller.AlaskaWebOutput {
	t.Helper()
	var out controller.AlaskaWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

// setupAlaskaOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupAlaskaOutputMock(g *interfaces.MockAlaskaGame) {
	setupAlaskaWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestAlaskaWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaOutputMock(rg)
		p := new(AlaskaWebPresenter)

		result := parseAlaskaOutput(t, p.Output(rg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, "alaska.playing", result.MessageCode)
	})

	t.Run("stalemate with escape available", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(5).Maybe()
		rg.On("CanUndo").Return(true).Maybe()
		rg.On("IsStalemate").Return(true).Maybe()
		rg.On("UndoToEscape").Return(3).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaWebPresenter)
		result := parseAlaskaOutput(t, p.Output(rg, nil))
		assert.Equal(t, "alaska.stalemateWithEscape", result.MessageCode)
		assert.Equal(t, "3", result.MessageParams["count"])
		assert.True(t, result.IsStalemate)
	})

	t.Run("stalemate without escape available", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(5).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(true).Maybe()
		rg.On("UndoToEscape").Return(-1).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaWebPresenter)
		result := parseAlaskaOutput(t, p.Output(rg, nil))
		assert.Equal(t, "alaska.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
		assert.Empty(t, result.MessageParams)
	})

	t.Run("game clear", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.AlaskaPhaseGameClear).Maybe()
		rg.On("GetMoveCount").Return(42).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaWebPresenter)
		result := parseAlaskaOutput(t, p.Output(rg, nil))
		assert.Equal(t, "alaska.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaOutputMock(rg)
		rg.ExpectedCalls = nil
		rg.On("GetPhase").Return(domain.AlaskaPhaseGameOver).Maybe()
		rg.On("GetMoveCount").Return(10).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard
		rg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.AlaskaFoundationCnt][]*domain.Card
		rg.On("GetFoundation").Return(foundation).Maybe()

		p := new(AlaskaWebPresenter)
		result := parseAlaskaOutput(t, p.Output(rg, nil))
		assert.Equal(t, "alaska.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		setupAlaskaOutputMock(rg)
		p := new(AlaskaWebPresenter)

		result := parseAlaskaOutput(t, p.Output(rg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestAlaskaWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		rsg := new(interfaces.MockAlaskaGame)
		setupAlaskaWebMockDefaults(rsg)
		rsg.On("GetHint").Return(&domain.AlaskaHint{FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(AlaskaWebPresenter).Output(rsg, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		rsg := new(interfaces.MockAlaskaGame)
		setupAlaskaWebMockDefaults(rsg)
		rsg.ExpectedCalls = filterCalls(rsg.ExpectedCalls, "IsStalemate")
		rsg.On("IsStalemate").Return(true)
		rsg.On("GetHint").Return(&domain.AlaskaHint{FromCol: 2, CardIndex: 0, ToZone: "foundation", ToCol: 1}).Maybe()

		result := new(AlaskaWebPresenter).Output(rsg, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestAlaskaWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(0).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		rg.On("GetHint").Return(&domain.AlaskaHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(AlaskaWebPresenter)
		result := parseAlaskaOutput(t, p.HintOutput(rg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "alaska.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying).Maybe()
		rg.On("GetMoveCount").Return(0).Maybe()
		rg.On("CanUndo").Return(false).Maybe()
		rg.On("IsStalemate").Return(false).Maybe()
		rg.On("UndoToEscape").Return(0).Maybe()
		rg.On("GetHint").Return((*domain.AlaskaHint)(nil))

		p := new(AlaskaWebPresenter)
		result := parseAlaskaOutput(t, p.HintOutput(rg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "alaska.noHint", result.MessageCode)
	})
}

func TestAlaskaWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhasePlaying)

		rg.On("GetGameEndFlag").Return(false)
		p := new(AlaskaWebPresenter)
		result := p.ActionLogOutput(rg)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		rg := new(interfaces.MockAlaskaGame)
		rg.On("GetPhase").Return(domain.AlaskaPhaseGameOver)
		rg.On("GetGameEndFlag").Return(true)
		rg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move"},
		})

		p := new(AlaskaWebPresenter)
		result := p.ActionLogOutput(rg)
		assert.Contains(t, result, "entries")
	})
}
