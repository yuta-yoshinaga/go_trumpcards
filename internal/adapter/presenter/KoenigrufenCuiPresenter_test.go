//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func koenigrufenCuiGame() *domain.Koenigrufen {
	g := domain.NewDefaultKoenigrufen()
	g.Reset()
	return g
}

func TestKoenigrufenCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KoenigrufenCuiPresenter)

	t.Run("bid phase", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetBidPlayerIdx(0)
		result := p.Output(g, nil)
		assert.Contains(t, result, "ケーニッヒルーフェン") // helpTitle
		assert.NotEmpty(t, result)
	})

	t.Run("call phase", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(domain.KoenigrufenPhaseCall)
		result := p.Output(g, nil)
		assert.Contains(t, result, "王呼び")
	})

	t.Run("talon phase", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(domain.KoenigrufenPhaseTalon)
		result := p.Output(g, nil)
		assert.Contains(t, result, "場札交換")
		// The bury-constraint legend guides the CLI player.
		assert.Contains(t, result, i18n.T("koenigrufen.promptTalonLegend"))
	})

	t.Run("play phase renders trump and skus", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(domain.KoenigrufenPhasePlay)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.KoenigrufenTrumpDesign, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.KoenigrufenSkusDesign, 0, false))
		g.SetCurrentTrick([]*domain.KoenigrufenTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		result := p.Output(g, nil)
		assert.Contains(t, result, "T7")   // trump label
		assert.Contains(t, result, "Sküs") // skus label
	})

	t.Run("trick end", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetPhase(domain.KoenigrufenPhaseTrickEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("round end", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(domain.KoenigrufenPhaseRoundEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end banner", func(t *testing.T) {
		g := koenigrufenForceEnd(0, [domain.KoenigrufenPlayerCnt]int{0, 10, 0, 0})
		assert.True(t, g.GetGameEndFlag())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end draw banner (no winner)", func(t *testing.T) {
		g := koenigrufenForceEnd(0, [domain.KoenigrufenPlayerCnt]int{252, 0, 0, 0})
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, -1, g.GetWinnerPlayer())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("error block", func(t *testing.T) {
		g := koenigrufenCuiGame()
		result := p.Output(g, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestKoenigrufenCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.KoenigrufenCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetBidPlayerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("call hint shows suit", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetPhase(domain.KoenigrufenPhaseCall)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint with card index", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(domain.KoenigrufenPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("discard hint uses declarer hand", func(t *testing.T) {
		g := koenigrufenCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.KoenigrufenBidRufer)
		g.SetPhase(domain.KoenigrufenPhaseTalon)
		assert.NotEmpty(t, p.HintOutput(g))
	})
}

func TestKoenigrufenCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KoenigrufenCuiPresenter)
	g := koenigrufenCuiGame()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
