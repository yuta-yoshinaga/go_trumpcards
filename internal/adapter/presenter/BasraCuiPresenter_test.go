package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBasraCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.BasraCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed table
}

func TestBasraCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.BasraCuiPresenter)

	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.BasraPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空テーブル表示。
	g.SetPhase(domain.BasraPhasePlay)
	g.SetTableCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestBasraCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BasraCuiPresenter)

	t.Run("capture hint", func(t *testing.T) {
		g := domain.NewDefaultBasra()
		g.Reset()
		g.SetCurrentTurn(0)
		g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultBasra()
		g.Reset()
		g.SetCurrentTurn(1)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestBasraCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	p := new(presenter.BasraCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
