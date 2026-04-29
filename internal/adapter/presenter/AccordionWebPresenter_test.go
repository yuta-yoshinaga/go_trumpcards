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

func setupAccordionWebMockDefaults(ag *interfaces.MockAccordionGame) {
	ag.On("GetPhase").Return(domain.AccordionPhasePlaying).Maybe()
	ag.On("GetMoveCount").Return(0).Maybe()
	ag.On("CanUndo").Return(false).Maybe()
	ag.On("IsStalemate").Return(false).Maybe()
	ag.On("UndoToEscape").Return(0).Maybe()
	ag.On("GetPileCount").Return(52).Maybe()
	piles := [][]*domain.Card{{domain.NewCard(domain.CardDesignSpade, 1, false)}}
	ag.On("GetPiles").Return(piles).Maybe()
}

func parseAccordionOutput(t *testing.T, jsonStr string) *controller.AccordionWebOutput {
	t.Helper()
	var out controller.AccordionWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestAccordionWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		setupAccordionWebMockDefaults(ag)
		p := new(AccordionWebPresenter)

		result := parseAccordionOutput(t, p.Output(ag, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, "accordion.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhasePlaying).Maybe()
		ag.On("GetMoveCount").Return(5).Maybe()
		ag.On("CanUndo").Return(true).Maybe()
		ag.On("IsStalemate").Return(true).Maybe()
		ag.On("UndoToEscape").Return(3).Maybe()
		ag.On("GetPileCount").Return(40).Maybe()
		ag.On("GetPiles").Return([][]*domain.Card{}).Maybe()

		p := new(AccordionWebPresenter)
		result := parseAccordionOutput(t, p.Output(ag, nil))
		assert.Equal(t, "accordion.stalemate", result.MessageCode)
		assert.True(t, result.IsStalemate)
	})

	t.Run("game clear", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhaseGameClear).Maybe()
		ag.On("GetMoveCount").Return(51).Maybe()
		ag.On("CanUndo").Return(false).Maybe()
		ag.On("IsStalemate").Return(false).Maybe()
		ag.On("UndoToEscape").Return(0).Maybe()
		ag.On("GetPileCount").Return(1).Maybe()
		ag.On("GetPiles").Return([][]*domain.Card{}).Maybe()

		p := new(AccordionWebPresenter)
		result := parseAccordionOutput(t, p.Output(ag, nil))
		assert.Equal(t, "accordion.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhaseGameOver).Maybe()
		ag.On("GetMoveCount").Return(10).Maybe()
		ag.On("CanUndo").Return(false).Maybe()
		ag.On("IsStalemate").Return(false).Maybe()
		ag.On("UndoToEscape").Return(0).Maybe()
		ag.On("GetPileCount").Return(40).Maybe()
		ag.On("GetPiles").Return([][]*domain.Card{}).Maybe()

		p := new(AccordionWebPresenter)
		result := parseAccordionOutput(t, p.Output(ag, nil))
		assert.Equal(t, "accordion.gameOver", result.MessageCode)
	})

	t.Run("with error", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		setupAccordionWebMockDefaults(ag)
		p := new(AccordionWebPresenter)

		result := parseAccordionOutput(t, p.Output(ag, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})
}

func TestAccordionWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		setupAccordionWebMockDefaults(ag)
		ag.On("GetHint").Return(&domain.AccordionHint{FromIdx: 3, ToIdx: 0})

		p := new(AccordionWebPresenter)
		result := parseAccordionOutput(t, p.HintOutput(ag))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, 3, result.Hint.FromIdx)
		assert.Equal(t, 0, result.Hint.ToIdx)
		assert.Equal(t, "accordion.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		setupAccordionWebMockDefaults(ag)
		ag.On("GetHint").Return((*domain.AccordionHint)(nil))

		p := new(AccordionWebPresenter)
		result := parseAccordionOutput(t, p.HintOutput(ag))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "accordion.noHint", result.MessageCode)
	})
}

func TestAccordionWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhasePlaying)

		ag.On("GetGameEndFlag").Return(false)
		p := new(AccordionWebPresenter)
		result := p.ActionLogOutput(ag)
		assert.Contains(t, result, "entries")
	})

	t.Run("game over", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhaseGameOver)
		ag.On("GetGameEndFlag").Return(true)
		ag.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "move"}})

		p := new(AccordionWebPresenter)
		result := p.ActionLogOutput(ag)
		assert.Contains(t, result, "entries")
	})
}
