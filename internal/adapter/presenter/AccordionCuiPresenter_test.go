//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupAccordionCuiMockDefaults(ag *interfaces.MockAccordionGame) {
	ag.On("GetPhase").Return(domain.AccordionPhasePlaying).Maybe()
	ag.On("GetMoveCount").Return(0).Maybe()
	ag.On("CanUndo").Return(false).Maybe()
	ag.On("IsStalemate").Return(false).Maybe()
	ag.On("UndoToEscape").Return(0).Maybe()
	ag.On("GetPileCount").Return(3).Maybe()
	piles := [][]*domain.Card{
		{domain.NewCard(domain.CardDesignSpade, 1, false)},
		{domain.NewCard(domain.CardDesignHeart, 2, false)},
		{domain.NewCard(domain.CardDesignClover, 3, false), domain.NewCard(domain.CardDesignClover, 4, false)},
	}
	ag.On("GetPiles").Return(piles).Maybe()
}

func TestAccordionCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		setupAccordionCuiMockDefaults(ag)
		p := new(AccordionCuiPresenter)

		result := p.Output(ag, nil)
		assert.Contains(t, result, "Accordion")
		assert.Contains(t, result, "残りパイル")
		assert.Contains(t, result, "[0]")
		// 多層パイル表示
		assert.Contains(t, result, "(+1)")
	})

	t.Run("with error", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		setupAccordionCuiMockDefaults(ag)
		p := new(AccordionCuiPresenter)

		result := p.Output(ag, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhasePlaying).Maybe()
		ag.On("GetMoveCount").Return(5).Maybe()
		ag.On("CanUndo").Return(true).Maybe()
		ag.On("IsStalemate").Return(true).Maybe()
		ag.On("UndoToEscape").Return(0).Maybe()
		ag.On("GetPileCount").Return(40).Maybe()
		ag.On("GetPiles").Return([][]*domain.Card{}).Maybe()

		p := new(AccordionCuiPresenter)
		result := p.Output(ag, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhaseGameClear).Maybe()
		ag.On("GetMoveCount").Return(51).Maybe()
		ag.On("CanUndo").Return(false).Maybe()
		ag.On("IsStalemate").Return(false).Maybe()
		ag.On("GetPileCount").Return(1).Maybe()
		ag.On("GetPiles").Return([][]*domain.Card{}).Maybe()

		p := new(AccordionCuiPresenter)
		result := p.Output(ag, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhaseGameOver).Maybe()
		ag.On("GetMoveCount").Return(10).Maybe()
		ag.On("CanUndo").Return(false).Maybe()
		ag.On("IsStalemate").Return(false).Maybe()
		ag.On("GetPileCount").Return(40).Maybe()
		ag.On("GetPiles").Return([][]*domain.Card{}).Maybe()

		p := new(AccordionCuiPresenter)
		result := p.Output(ag, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})
}

func TestAccordionCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetHint").Return(&domain.AccordionHint{FromIdx: 3, ToIdx: 0})

		p := new(AccordionCuiPresenter)
		result := p.HintOutput(ag)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "パイル3")
		assert.Contains(t, result, "パイル0")
	})

	t.Run("no hint", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetHint").Return((*domain.AccordionHint)(nil))

		p := new(AccordionCuiPresenter)
		result := p.HintOutput(ag)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestAccordionCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhasePlaying)

		p := new(AccordionCuiPresenter)
		result := p.ActionLogOutput(ag)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		ag := new(interfaces.MockAccordionGame)
		ag.On("GetPhase").Return(domain.AccordionPhaseGameOver)
		ag.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(AccordionCuiPresenter)
		result := p.ActionLogOutput(ag)
		assert.NotEmpty(t, result)
	})
}
