package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newPishtiForPresenter() *domain.Pishti {
	players := []*domain.PishtiPlayer{
		domain.NewPishtiPlayer(true),
		domain.NewPishtiPlayer(false),
		domain.NewPishtiPlayer(false),
		domain.NewPishtiPlayer(false),
	}
	return domain.NewPishti(domain.NewTrumpCards(0), players, domain.DefaultPishtiConfig())
}

func TestPishtiCuiPresenter_Output(t *testing.T) {
	p := new(presenter.PishtiCuiPresenter)

	t.Run("initial state includes header", func(t *testing.T) {
		g := newPishtiForPresenter()
		out := p.Output(g, nil)
		assert.Contains(t, out, "Pişti")
	})

	t.Run("pile with top card rendered", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.SetPile([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		out := p.Output(g, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("empty pile rendered", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.SetPile([]*domain.Card{})
		out := p.Output(g, nil)
		assert.Contains(t, out, "場:")
	})

	t.Run("error displayed", func(t *testing.T) {
		g := newPishtiForPresenter()
		out := p.Output(g, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("game end result", func(t *testing.T) {
		g := newPishtiForPresenter()
		g.GetPlayer(0).AddCaptured([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		g.SetGameEndFlag(true)
		out := p.Output(g, nil)
		assert.Contains(t, out, "ゲーム終了")
	})
}

func TestPishtiCuiPresenter_ActionLog(t *testing.T) {
	p := new(presenter.PishtiCuiPresenter)
	g := newPishtiForPresenter()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
