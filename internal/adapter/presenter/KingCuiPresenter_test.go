package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestKingCuiPresenter_Output(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	p := new(presenter.KingCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "1/7") // deal line
	assert.Contains(t, out, "[0]") // human indexed hand
}

func TestKingCuiPresenter_ContractAndTrick(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	g.SetCurrentTrick([]*domain.KingTrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	p := new(presenter.KingCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestKingCuiPresenter_DealEndAndGameEnd(t *testing.T) {
	p := new(presenter.KingCuiPresenter)

	g := domain.NewDefaultKing()
	g.Reset()
	g.SetPhase(domain.KingPhaseDealEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// Error output.
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))
}

func TestKingCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.KingCuiPresenter)

	t.Run("negative contract avoid", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()
		g.SetCurrentContract(domain.KingContractNoTricks)
		g.SetTrumpSuit(-1)
		g.SetPhase(domain.KingPhasePlay)
		g.SetCurrentTurn(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 13))
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
	})

	t.Run("positive contract win", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()
		g.SetCurrentContract(domain.KingContractKingTrump)
		g.SetTrumpSuit(domain.CardDesignSpade)
		g.SetPhase(domain.KingPhasePlay)
		g.SetCurrentTurn(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 1))
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
	})

	t.Run("no hint outside play phase", func(t *testing.T) {
		g := domain.NewDefaultKing()
		g.Reset()
		g.SetPhase(domain.KingPhaseSelectContract)
		assert.Contains(t, p.HintOutput(g), "")
	})
}

func TestKingCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	p := new(presenter.KingCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
