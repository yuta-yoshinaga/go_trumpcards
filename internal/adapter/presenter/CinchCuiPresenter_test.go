package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCinchCuiPresenter_Output_BidPhase(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	p := new(presenter.CinchCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand
}

func TestCinchCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.CinchCuiPresenter)

	g := domain.NewDefaultCinch()
	g.Reset()

	g.SetPhase(domain.CinchPhaseNameTrump)
	g.SetBidWinnerIdx(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.CinchPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.CinchPhaseTrickEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.CinchPhaseRoundEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))
}

func TestCinchCuiPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	g.GetPlayer(0).AddScore(30)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(1)
	g.SetPhase(domain.CinchPhaseRoundEnd)
	g.ScoreRound()

	p := new(presenter.CinchCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
}

func TestCinchCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.CinchCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset() // bid phase, human turn
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("name trump hint", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhaseNameTrump)
		g.SetBidWinnerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhasePlay)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetCurrentTurn(0)
		g.SetLeadPlayerIdx(0)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignHeart, 1))
		g.GetPlayer(0).AddCard(bcard(domain.CardDesignSpade, 2))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultCinch()
		g.Reset()
		g.SetPhase(domain.CinchPhaseTrickEnd)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestCinchCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultCinch()
	g.Reset()
	p := new(presenter.CinchCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
