package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestTablanetCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.TablanetCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed table
}

func TestTablanetCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.TablanetCuiPresenter)

	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.TablanetPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空テーブル表示。
	g.SetPhase(domain.TablanetPhasePlay)
	g.SetTableCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestTablanetCuiPresenter_Output_CaptureHint(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetPhase(domain.TablanetPhasePlay)
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))  // matches table[0] → capturable
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 9, false)) // captures nothing
	p := new(presenter.TablanetCuiPresenter)
	out := p.Output(g, nil)
	// The ♥5 is annotated with the table index it can capture.
	assert.Contains(t, out, "→ 場[0]")
	// Only the capturing card gets a note; the ♣9 does not.
	assert.Equal(t, 1, strings.Count(out, "→ 場"))
}

func TestTablanetCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TablanetCuiPresenter)

	t.Run("capture hint", func(t *testing.T) {
		g := domain.NewDefaultTablanet()
		g.Reset()
		g.SetCurrentTurn(0)
		g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultTablanet()
		g.Reset()
		g.SetCurrentTurn(1)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestTablanetCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	p := new(presenter.TablanetCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
