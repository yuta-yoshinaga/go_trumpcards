//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newCuiAgnes returns a real Agnes with deterministic state for CUI rendering.
func newCuiAgnes() *domain.Agnes {
	a := domain.NewDefaultAgnes()
	a.Reset()
	a.SetBaseRank(7)

	var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
	// Col 0: a face-down card then a face-up card (exercises "??" and card render).
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

func TestAgnesCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("playing state", func(t *testing.T) {
		a := newCuiAgnes()
		p := new(AgnesCuiPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "SPADE 5") // face-up tableau card
		assert.Contains(t, result, "??")      // face-down card rendering
		assert.Contains(t, result, "SPADE 7") // foundation top
	})

	t.Run("error", func(t *testing.T) {
		a := newCuiAgnes()
		p := new(AgnesCuiPresenter)
		result := p.Output(a, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("game clear", func(t *testing.T) {
		a := newCuiAgnes()
		a.SetPhase(domain.AgnesPhaseGameClear)
		p := new(AgnesCuiPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		a := newCuiAgnes()
		a.SetPhase(domain.AgnesPhaseGameOver)
		p := new(AgnesCuiPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("empty tableau column and empty foundation", func(t *testing.T) {
		a := newCuiAgnes()
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		a.SetTableau(tab)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		p := new(AgnesCuiPresenter)
		result := p.Output(a, nil)
		assert.Contains(t, result, "[空]")
	})
}

// **Web は ag-stalemate-banner で毎レンダー知らせている。**CUI は手数しか出さず、
// 詰んでいても分からなかった (#4830)。
func TestAgnesCuiPresenter_StalemateWarning(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(AgnesCuiPresenter)

	t.Run("no warning while the stock still has cards", func(t *testing.T) {
		a := newCuiAgnes()
		assert.False(t, a.IsStalemate(), "ストックが残っていれば手詰まりではない")
		assert.NotContains(t, p.Output(a, nil), i18n.T("cuiSolitaireStalemate"))
	})

	t.Run("warns once nothing can move", func(t *testing.T) {
		a := newCuiAgnes()
		a.SetStock(nil)
		// どこにも動かせない盤面にする: 各列に孤立した札を 1 枚ずつ。
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		for i := range domain.AgnesTableauCnt {
			tab[i] = []*domain.AgnesTableauCard{
				{Card: domain.NewCard(domain.CardDesignSpade, 13, false), FaceUp: true},
			}
		}
		a.SetTableau(tab)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)

		assert.True(t, a.IsStalemate())
		assert.Nil(t, a.GetHint(), "手詰まりならヒントも手を返さない")
		assert.Contains(t, p.Output(a, nil), i18n.T("cuiSolitaireStalemate"))
	})

	t.Run("no warning outside the playing phase", func(t *testing.T) {
		a := newCuiAgnes()
		a.SetStock(nil)
		a.SetPhase(domain.AgnesPhaseGameOver)
		assert.False(t, a.IsStalemate())
	})
}

func TestAgnesCuiPresenter_HintOutput(t *testing.T) {
	t.Run("no hint", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetPhase(domain.AgnesPhaseGameOver) // GetHint returns nil
		p := new(AgnesCuiPresenter)
		assert.Contains(t, p.HintOutput(a), "ヒントはありません")
	})

	t.Run("tableau to foundation hint", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetBaseRank(5)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, 5, false), FaceUp: true}}
		a.SetTableau(tab)
		p := new(AgnesCuiPresenter)
		result := p.HintOutput(a)
		assert.NotEmpty(t, result)
	})

	t.Run("tableau to tableau hint", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.SetBaseRank(13)
		var f [domain.AgnesFoundationCnt][]*domain.Card
		a.SetFoundation(f)
		var tab [domain.AgnesTableauCnt][]*domain.AgnesTableauCard
		tab[0] = []*domain.AgnesTableauCard{{Card: domain.NewCard(domain.CardDesignSpade, 6, false), FaceUp: true}}
		tab[1] = []*domain.AgnesTableauCard{{Card: domain.NewCard(domain.CardDesignClover, 7, false), FaceUp: true}}
		a.SetTableau(tab)
		p := new(AgnesCuiPresenter)
		result := p.HintOutput(a)
		assert.NotEmpty(t, result)
	})
}

func TestAgnesCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("during game", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		p := new(AgnesCuiPresenter)
		assert.Contains(t, p.ActionLogOutput(a), "棋譜はありません")
	})

	t.Run("after game over", func(t *testing.T) {
		a := domain.NewDefaultAgnes()
		a.Reset()
		a.GiveUp()
		p := new(AgnesCuiPresenter)
		assert.NotEmpty(t, p.ActionLogOutput(a))
	})
}
