package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBarbuCuiPresenter_Output(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuCuiPresenter)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "1/28") // deal line
	assert.Contains(t, out, "[0]")  // human indexed hand
}

func TestBarbuCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BarbuCuiPresenter)

	t.Run("dominoes playable", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 7)})
		assert.Contains(t, p.HintOutput(b), "置けるカード")
	})

	t.Run("dominoes pass when nothing playable", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		// No 7 and an empty table: nothing can be placed.
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
		assert.Contains(t, p.HintOutput(b), "パス")
	})

	t.Run("trick legal follow", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignHeart, 5), bcard(domain.CardDesignSpade, 9)})
		b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("trick leading (empty trick): all cards legal", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignHeart, 5), bcard(domain.CardDesignSpade, 9)})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("trick void in lead suit: all cards legal", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		// Hand has no hearts, lead is a heart -> void -> every card is legal.
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5), bcard(domain.CardDesignClover, 9)})
		b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("trick with a nil lead entry falls back to all legal", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetContract(domain.BarbuContractNoTricks, 0)
		b.BarbuTestSetPhase(domain.BarbuPhasePlay)
		b.BarbuTestSetCurrentPlayer(0)
		b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
		b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: nil}})
		assert.Contains(t, p.HintOutput(b), "合法手")
	})

	t.Run("none outside the play phase", func(t *testing.T) {
		b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
		b.BarbuTestSetPhase(domain.BarbuPhaseSelectContract)
		assert.Contains(t, p.HintOutput(b), "ヒントはありません")
	})
}

func TestBarbuCuiPresenter_ContractAndTrick(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 5)})
	b.BarbuTestSetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	p := new(presenter.BarbuCuiPresenter)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBarbuCuiPresenter_Dominoes(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{bcard(domain.CardDesignSpade, 7)})
	var table [5]uint16
	table[domain.CardDesignSpade] = 1 << 7
	b.BarbuTestSetTablePlaced(table)
	p := new(presenter.BarbuCuiPresenter)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBarbuCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuCuiPresenter)
	assert.NotEmpty(t, p.Output(b, errors.New("boom")))

	b.BarbuTestSetGameEnd(true)
	out := p.Output(b, nil)
	assert.NotEmpty(t, out)
}

func TestBarbuCuiPresenter_ActionLog(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	p := new(presenter.BarbuCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(b))
}
