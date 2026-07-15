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

type trashMockOpts struct {
	phase       domain.TrashPhase
	current     int
	isCpuTurn   bool
	stockSize   int
	discardSize int
	discardTop  *domain.Card
	pending     *domain.Card
	moveCount   int
	winner      int
	winnerSet   bool // distinguishes explicit winner=0 from the default (-1)
	p0Slots     []domain.TrashSlot
	p1Slots     []domain.TrashSlot
}

// buildTrashMock constructs a fresh MockTrashGame with the supplied options.
// Zero-valued fields receive sensible defaults (34-card stock, 10 face-down slots, winner=-1).
func buildTrashMock(o trashMockOpts) *interfaces.MockTrashGame {
	tg := new(interfaces.MockTrashGame)
	if o.stockSize == 0 {
		o.stockSize = 34
	}
	winner := o.winner
	if !o.winnerSet {
		winner = -1
	}
	defaultSlots := func() []domain.TrashSlot {
		slots := make([]domain.TrashSlot, domain.TrashSlotCnt)
		for i := range slots {
			slots[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: false}
		}
		return slots
	}
	if o.p0Slots == nil {
		o.p0Slots = defaultSlots()
	}
	if o.p1Slots == nil {
		o.p1Slots = defaultSlots()
	}
	tg.On("GetPhase").Return(o.phase).Maybe()
	tg.On("GetCurrent").Return(o.current).Maybe()
	tg.On("GetStockSize").Return(o.stockSize).Maybe()
	tg.On("GetDiscardSize").Return(o.discardSize).Maybe()
	tg.On("GetDiscardTop").Return(o.discardTop).Maybe()
	tg.On("GetPending").Return(o.pending).Maybe()
	tg.On("GetMoveCount").Return(o.moveCount).Maybe()
	tg.On("GetWinner").Return(winner).Maybe()
	tg.On("IsCpuTurn").Return(o.isCpuTurn).Maybe()
	tg.On("IsCpuPlayer", 0).Return(false).Maybe()
	tg.On("IsCpuPlayer", 1).Return(true).Maybe()
	tg.On("GetPlayerSlots", 0).Return(o.p0Slots).Maybe()
	tg.On("GetPlayerSlots", 1).Return(o.p1Slots).Maybe()
	return tg
}

func parseTrashOutput(t *testing.T, jsonStr string) *controller.TrashWebOutput {
	t.Helper()
	var out controller.TrashWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestTrashWebPresenter_Output(t *testing.T) {
	t.Run("player turn", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, "trash.playerTurn", result.MessageCode)
		assert.Equal(t, 34, result.StockSize)
	})

	t.Run("cpu turn", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhasePlayerTurn, current: 1, isCpuTurn: true})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.Equal(t, "trash.cpuTurn", result.MessageCode)
	})

	t.Run("await wild (human)", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:   domain.TrashPhaseAwaitWild,
			pending: domain.NewCard(domain.CardDesignSpade, 13, false),
		})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.Equal(t, "trash.awaitWild", result.MessageCode)
		assert.NotNil(t, result.Pending)
	})

	t.Run("await wild (cpu)", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:     domain.TrashPhaseAwaitWild,
			current:   1,
			isCpuTurn: true,
			pending:   domain.NewCard(domain.CardDesignJoker, 1, false),
		})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.Equal(t, "trash.cpuAwaitWild", result.MessageCode)
	})

	t.Run("game over — player wins", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhaseGameOver, winner: 0, winnerSet: true})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.Equal(t, "trash.gameOverWin", result.MessageCode)
		assert.Equal(t, 0, result.Winner)
	})

	t.Run("game over — cpu wins", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhaseGameOver, winner: 1, winnerSet: true})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.Equal(t, "trash.gameOverLose", result.MessageCode)
		assert.Equal(t, 1, result.Winner)
	})

	t.Run("with error", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{})
		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, errors.New("bad move")))
		assert.Equal(t, "bad move", result.Message)
	})

	t.Run("face-up card surfaces", func(t *testing.T) {
		slots := make([]domain.TrashSlot, domain.TrashSlotCnt)
		slots[0] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignHeart, 1, false), FaceUp: true}
		for i := 1; i < len(slots); i++ {
			slots[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: false}
		}
		tg := buildTrashMock(trashMockOpts{
			stockSize:   30,
			discardSize: 1,
			discardTop:  domain.NewCard(domain.CardDesignSpade, 11, false),
			moveCount:   2,
			p0Slots:     slots,
			p1Slots:     slots,
		})

		p := new(TrashWebPresenter)
		result := parseTrashOutput(t, p.Output(tg, nil))
		assert.NotNil(t, result.Players[0].Slots[0].Card)
		assert.True(t, result.Players[0].Slots[0].FaceUp)
		assert.Nil(t, result.Players[0].Slots[1].Card, "face-down slot must not expose its card")
		assert.NotNil(t, result.DiscardTop)
	})
}

func TestTrashWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("hides during play", func(t *testing.T) {
		tg := new(interfaces.MockTrashGame)
		tg.On("GetPhase").Return(domain.TrashPhasePlayerTurn)

		p := new(TrashWebPresenter)
		result := p.ActionLogOutput(tg)
		assert.Contains(t, result, "entries")
	})

	t.Run("exposes on game over", func(t *testing.T) {
		tg := new(interfaces.MockTrashGame)
		tg.On("GetPhase").Return(domain.TrashPhaseGameOver)
		tg.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "end"}})

		p := new(TrashWebPresenter)
		result := p.ActionLogOutput(tg)
		assert.Contains(t, result, "entries")
	})
}

func TestTrashWebPresenter_HintOutput(t *testing.T) {
	// Web hints are computed client-side, so HintOutput mirrors Output.
	tg := buildTrashMock(trashMockOpts{})
	p := new(TrashWebPresenter)
	assert.Equal(t, p.Output(tg, nil), p.HintOutput(tg))
}
