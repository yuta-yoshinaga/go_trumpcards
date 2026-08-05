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

	// **行き先を書く。**Web は `pendingAnnounce` で置ける場所・死に札・ワイルドを
	// 毎回文章化しているのに、CUI は札をそのまま出すだけだった (#4867)。
	t.Run("pending card names its destination", func(t *testing.T) {
		p := new(TrashCuiPresenter)

		// 空きスロットのあるワイルド: 選べる位置を並べる (推奨は HintOutput の役目)。
		slots := make([]domain.TrashSlot, domain.TrashSlotCnt)
		for i := range slots {
			slots[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true}
		}
		slots[1].FaceUp = false
		slots[4].FaceUp = false
		out := p.Output(buildTrashMock(trashMockOpts{
			phase:   domain.TrashPhaseAwaitWild,
			pending: domain.NewCard(domain.CardDesignSpade, 13, false),
			p0Slots: slots,
		}), nil)
		assert.Contains(t, out, "ワイルド: 空きスロット 2, 5 のどこへでも")

		// 数札で、その位置がまだ伏せられている: 位置番号を出す。
		down := make([]domain.TrashSlot, domain.TrashSlotCnt)
		for i := range down {
			down[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: false}
		}
		out = p.Output(buildTrashMock(trashMockOpts{
			pending: domain.NewCard(domain.CardDesignHeart, 4, false),
			p0Slots: down,
		}), nil)
		assert.Contains(t, out, "→ スロット 4 へ")

		// 同じ札でも、その位置が既に表向きなら死に札。
		filled := make([]domain.TrashSlot, domain.TrashSlotCnt)
		copy(filled, down)
		filled[3].FaceUp = true
		out = p.Output(buildTrashMock(trashMockOpts{
			pending: domain.NewCard(domain.CardDesignHeart, 4, false),
			p0Slots: filled,
		}), nil)
		assert.Contains(t, out, "置き場所なし: 捨て札行き")
		assert.NotContains(t, out, "→ スロット 4 へ")

		// スロットを持たないランク (Q) も死に札。
		out = p.Output(buildTrashMock(trashMockOpts{
			pending: domain.NewCard(domain.CardDesignHeart, 12, false),
			p0Slots: down,
		}), nil)
		assert.Contains(t, out, "置き場所なし: 捨て札行き")
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
		assert.Contains(t, out, "上:")
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

func TestTrashCuiPresenter_HintOutput(t *testing.T) {
	p := new(TrashCuiPresenter)

	t.Run("await wild lists candidate slots and recommends the highest open one", func(t *testing.T) {
		slots := make([]domain.TrashSlot, domain.TrashSlotCnt)
		for i := range slots {
			// Slots 1-3 already filled (face up); 4-10 are open.
			slots[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: i < 3}
		}
		tg := buildTrashMock(trashMockOpts{
			phase:   domain.TrashPhaseAwaitWild,
			pending: domain.NewCard(domain.CardDesignSpade, 13, false),
			p0Slots: slots,
		})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "配置候補: 4, 5, 6, 7, 8, 9, 10")
		assert.Contains(t, out, "推奨: 10")
	})

	t.Run("player turn recommends taking a placeable discard", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:      domain.TrashPhasePlayerTurn,
			discardTop: domain.NewCard(domain.CardDesignHeart, 5, false),
		})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "スロット5")
	})

	t.Run("player turn flags a wild discard", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:      domain.TrashPhasePlayerTurn,
			discardTop: domain.NewCard(domain.CardDesignSpade, 13, false),
		})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "ワイルド")
	})

	t.Run("player turn advises drawing when the discard is useless", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:      domain.TrashPhasePlayerTurn,
			discardTop: domain.NewCard(domain.CardDesignSpade, 11, false), // Jack = end-turn card
		})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "山札から引き")
	})

	t.Run("await wild with no open slots falls back to game over", func(t *testing.T) {
		slots := make([]domain.TrashSlot, domain.TrashSlotCnt)
		for i := range slots {
			slots[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true}
		}
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhaseAwaitWild, p0Slots: slots})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "ゲームは終了")
	})

	t.Run("player turn draws when the discard's slot is already filled", func(t *testing.T) {
		slots := make([]domain.TrashSlot, domain.TrashSlotCnt)
		for i := range slots {
			// Slot 5 (index 4) is already filled, so a drawn 5 has nowhere to go.
			slots[i] = domain.TrashSlot{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: i == 4}
		}
		tg := buildTrashMock(trashMockOpts{
			phase:      domain.TrashPhasePlayerTurn,
			discardTop: domain.NewCard(domain.CardDesignHeart, 5, false),
			p0Slots:    slots,
		})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "山札から引き")
	})

	t.Run("declines advice when it is not the human's turn", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{
			phase:      domain.TrashPhasePlayerTurn,
			current:    domain.TrashCpuIdx,
			discardTop: domain.NewCard(domain.CardDesignHeart, 5, false),
		})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "あなたの番ではありません")
	})

	t.Run("game over reports the game is finished", func(t *testing.T) {
		tg := buildTrashMock(trashMockOpts{phase: domain.TrashPhaseGameOver, winner: 1, winnerSet: true})
		out := p.HintOutput(tg)
		assert.Contains(t, out, "ゲームは終了")
	})
}
