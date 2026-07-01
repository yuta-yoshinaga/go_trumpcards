package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestLooCuiPresenter_Output_DecidePhase(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	g.SetDecidePlayerIdx(0)
	p := new(presenter.LooCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand
}

func TestLooCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.LooCuiPresenter)

	g := domain.NewDefaultLoo()
	g.Reset()

	g.SetPhase(domain.LooPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(0)
	g.GetPlayer(0).SetPlaying(true)
	g.SetCurrentTrick([]*domain.LooTrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.LooPhaseTrickEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.LooPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))
}

func TestLooCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.LooCuiPresenter)

	t.Run("decide hint", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetDecidePlayerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetPhase(domain.LooPhasePlay)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTurn(0)
		g.SetLeadPlayerIdx(0)
		g.GetPlayer(0).SetPlaying(true)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 1))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultLoo()
		g.Reset()
		g.SetPhase(domain.LooPhaseTrickEnd)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestLooCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultLoo()
	g.Reset()
	p := new(presenter.LooCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
