//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func scartoCuiGame() *domain.Scarto {
	g := domain.NewDefaultScarto()
	g.Reset()
	return g
}

func TestScartoCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ScartoCuiPresenter)

	t.Run("scarto phase", func(t *testing.T) {
		g := scartoCuiGame() // human dealer in scarto phase
		result := p.Output(g, nil)
		assert.Contains(t, result, "スカルト") // helpTitle
		assert.NotEmpty(t, result)
		// The discardable list and the excluded-kinds legend both appear.
		assert.Contains(t, result, i18n.T("scarto.discardableLegend"))
		discardablePrefix := strings.SplitN(i18n.T("scarto.discardableList"), "{{", 2)[0]
		assert.Contains(t, result, discardablePrefix)
	})

	t.Run("scarto phase with no discardable cards shows none", func(t *testing.T) {
		g := scartoCuiGame()
		dealer := g.GetPlayer(g.GetDealerIdx())
		dealer.Reset()
		// A hand of only trumps has nothing legally discardable.
		dealer.AddCard(domain.NewCard(domain.ScartoTrumpDesign, 3, false))
		dealer.AddCard(domain.NewCard(domain.ScartoTrumpDesign, 8, false))
		result := p.Output(g, nil)
		assert.Contains(t, result, i18n.Tf("scarto.discardableList", "cards", i18n.T("scarto.discardableNone")))
	})

	t.Run("play phase renders trump and excuse", func(t *testing.T) {
		g := scartoCuiGame()
		g.SetPhase(domain.ScartoPhasePlay)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.ScartoTrumpDesign, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.ScartoExcuseDesign, 0, false))
		g.SetCurrentTrick([]*domain.ScartoTrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		result := p.Output(g, nil)
		assert.Contains(t, result, "T7")  // trump label
		assert.Contains(t, result, "EXC") // excuse label
	})

	t.Run("trick end", func(t *testing.T) {
		g := scartoCuiGame()
		g.SetPhase(domain.ScartoPhaseTrickEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("round end", func(t *testing.T) {
		g := scartoCuiGame()
		g.SetPhase(domain.ScartoPhaseRoundEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end banner", func(t *testing.T) {
		g := scartoCuiGame()
		cfg := domain.DefaultScartoConfig()
		cfg.TargetDeals = 1
		g.SetConfig(cfg)
		g.SetRoundNumber(1)
		g.SetPlayerScores([domain.ScartoPlayerCnt]int{300, 0, 0})
		g.SetPhase(domain.ScartoPhaseRoundEnd)
		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end draw banner (no winner)", func(t *testing.T) {
		g := scartoCuiGame()
		cfg := domain.DefaultScartoConfig()
		cfg.TargetDeals = 1
		g.SetConfig(cfg)
		g.SetRoundNumber(1)
		g.SetPlayerScores([domain.ScartoPlayerCnt]int{5, 5, 5})
		g.SetPhase(domain.ScartoPhaseRoundEnd)
		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, -1, g.GetWinnerPlayer())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("error block", func(t *testing.T) {
		g := scartoCuiGame()
		result := p.Output(g, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestScartoCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ScartoCuiPresenter)

	t.Run("scarto hint uses dealer hand", func(t *testing.T) {
		g := scartoCuiGame() // human dealer in scarto phase
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint with card index", func(t *testing.T) {
		g := scartoCuiGame()
		g.SetPhase(domain.ScartoPhasePlay)
		g.SetCurrentPlayerIdx(0)
		result := p.HintOutput(g)
		assert.NotEmpty(t, result)
	})
}

func TestScartoCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ScartoCuiPresenter)
	g := scartoCuiGame()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
