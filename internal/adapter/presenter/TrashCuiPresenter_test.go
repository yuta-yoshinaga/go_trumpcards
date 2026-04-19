//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestTrashCuiPresenter_Output(t *testing.T) {
	t.Run("initial state renders slots and header", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{})
		p := new(TrashCuiPresenter)
		out := p.Output(tg, nil)
		assert.Contains(t, out, "Trash")
		assert.Contains(t, out, "CPU")
		assert.Contains(t, out, "あなた")
		assert.Contains(t, out, "山札")
	})

	t.Run("await wild prompts placement", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:   domain.TrashPhaseAwaitWild,
			pending: domain.NewCard(domain.CardDesignSpade, 13, false),
		})
		p := new(TrashCuiPresenter)
		out := p.Output(tg, nil)
		assert.Contains(t, out, "ワイルドを配置")
		assert.Contains(t, out, "ペンディング")
	})

	t.Run("game over banner — human wins", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhaseGameOver, winner: 0, winnerSet: true})
		p := new(TrashCuiPresenter)
		out := p.Output(tg, nil)
		assert.Contains(t, out, "勝ち")
	})

	t.Run("game over banner — cpu wins", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhaseGameOver, winner: 1, winnerSet: true})
		p := new(TrashCuiPresenter)
		out := p.Output(tg, nil)
		assert.Contains(t, out, "CPUの勝ち")
	})

	t.Run("error is surfaced", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{})
		p := new(TrashCuiPresenter)
		out := p.Output(tg, errors.New("no stock"))
		assert.Contains(t, out, "no stock")
	})

	t.Run("discard top rendered", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			discardSize: 1,
			discardTop:  domain.NewCard(domain.CardDesignSpade, 11, false),
		})
		p := new(TrashCuiPresenter)
		out := p.Output(tg, nil)
		assert.Contains(t, out, "top:")
	})
}

func TestTrashCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("hides during play", func(t *testing.T) {
		tg := new(interfaces.MockTrashGame)
		tg.On("GetPhase").Return(domain.TrashPhasePlayerTurn)

		p := new(TrashCuiPresenter)
		out := p.ActionLogOutput(tg)
		assert.NotNil(t, out)
	})

	t.Run("exposes on game over", func(t *testing.T) {
		tg := new(interfaces.MockTrashGame)
		tg.On("GetPhase").Return(domain.TrashPhaseGameOver)
		tg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "end"}})

		p := new(TrashCuiPresenter)
		out := p.ActionLogOutput(tg)
		assert.NotEmpty(t, out)
	})
}
