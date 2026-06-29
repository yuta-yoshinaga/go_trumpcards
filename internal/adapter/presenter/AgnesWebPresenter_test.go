//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newWebAgnes returns a real Agnes with deterministic state for web rendering.
func newWebAgnes() *domain.Agnes {
	a := domain.NewDefaultAgnes()
	a.Reset()
	a.SetBaseRank(7)

	var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	// Col 0: a face-down card (-> card null) and a face-up card.
	tab[0] = []*domain.AgnesTableauCard{
		{Card: domain.NewCard(domain.CardDesignSpade, 10, false), FaceUp: false},
		{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true},
	}
	for i := 1; i < domain.AgnesTableauCnt; i++ {
		tab[i] = []*domain.AgnesTableauCard{
			{Card: domain.NewCard(domain.CardDesignClover, i+1, false), FaceUp: true},
		}
	}
	a.SetTableau(tab)

	var f [domain.AgnesFoundationCnt][]*domain.Card
	f[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 7, false)}
	a.SetFoundation(f)
	a.SetStock([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	return a
}

func TestAgnesWebPresenter_Output(t *testing.T) {
	t.Run("playing state", func(t *testing.T) {
		a := newWebAgnes()
		p := new(AgnesWebPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, `"baseRank":7`)
		assert.Contains(t, result, `"stockCount":1`)
		assert.Contains(t, result, `"tableau"`)
		assert.Contains(t, result, `"foundation"`)
		assert.Contains(t, result, `"agnes.playing"`)
		// Face-down card -> card null with faceUp false.
		assert.Contains(t, result, `"card":null,"faceUp":false`)
	})

	t.Run("error", func(t *testing.T) {
		a := newWebAgnes()
		p := new(AgnesWebPresenter)
		result := p.Output(a, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		a := newWebAgnes()
		a.SetPhase(domain.AgnesPhaseGameClear)
		p := new(AgnesWebPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "agnes.gameClear")
	})

	t.Run("game over", func(t *testing.T) {
		a := newWebAgnes()
		a.SetPhase(domain.AgnesPhaseGameOver)
		p := new(AgnesWebPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "agnes.gameOver")
	})
}

func TestAgnesWebPresenter_HintOutput(t *testing.T) {
	t.Run("hint available", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}}
		a.SetTableau(tab)
		p := new(AgnesWebPresenter)
		result := p.HintOutput(a)
		assert.Contains(t, result, `"agnes.hintAvailable"`)
	})

	t.Run("no hint", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetPhase(domain.AgnesPhaseGameOver)
		p := new(AgnesWebPresenter)
		result := p.HintOutput(a)
		assert.Contains(t, result, `"agnes.noHint"`)
	})
}

func TestAgnesWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		p := new(AgnesWebPresenter)
		_ = p.ActionLogOutput(a)
	})

	t.Run("after game over", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.GiveUp()
		p := new(AgnesWebPresenter)
		_ = p.ActionLogOutput(a)
	})
}
