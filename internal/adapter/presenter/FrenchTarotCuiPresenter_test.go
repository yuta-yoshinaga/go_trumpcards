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

func frenchTarotCuiGame() *domain.FrenchTarot {
	g := domain.NewDefaultFrenchTarot()
	g.Reset()
	return g
}

func TestFrenchTarotCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.FrenchTarotCuiPresenter)

	t.Run("bid phase", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetBidPlayerIdx(0)
		result := p.Output(g, nil)
		assert.Contains(t, result, "フレンチタロット") // helpTitle
		assert.NotEmpty(t, result)
		// The contract-multiplier legend guides the CLI bidder.
		assert.Contains(t, result, i18n.T("frenchtarot.promptBidLegend"))
	})

	t.Run("chien phase", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidPetite)
		g.SetPhase(domain.FrenchTarotPhaseChien)
		result := p.Output(g, nil)
		assert.Contains(t, result, "シアン")
	})

	t.Run("play phase renders trump and excuse", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidPetite)
		g.SetPhase(domain.FrenchTarotPhasePlay)
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.FrenchTarotTrumpDesign, 7, false))
		g.GetPlayer(0).AddCard(domain.NewCard(domain.FrenchTarotExcuseDesign, 0, false))
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		})
		result := p.Output(g, nil)
		assert.Contains(t, result, "T7")  // trump label
		assert.Contains(t, result, "EXC") // excuse label
	})

	t.Run("trick end", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("round end", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidGarde)
		g.SetPhase(domain.FrenchTarotPhaseRoundEnd)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end banner", func(t *testing.T) {
		g := frenchTarotCuiGame()
		cfg := domain.DefaultFrenchTarotConfig()
		cfg.TargetDeals = 1
		g.SetConfig(cfg)
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidGarde)
		g.SetRoundNumber(1)
		g.SetTrickNumber(domain.FrenchTarotTrickCount)
		g.SetLeadPlayerIdx(0)
		g.SetPhase(domain.FrenchTarotPhaseTrickEnd)
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 5, false)},
			{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 9, false)},
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 3, false)},
			{PlayerIdx: 0, Card: domain.NewCard(domain.FrenchTarotTrumpDesign, 1, false)},
		})
		g.ResolveTrick()
		assert.True(t, g.GetGameEndFlag())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("error block", func(t *testing.T) {
		g := frenchTarotCuiGame()
		result := p.Output(g, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})

	t.Run("game end draw banner (no winner)", func(t *testing.T) {
		// Pre-load scores so a single deal ends the match in a four-way tie:
		// declarer −243, each defender +81 -> all seats level at 81 (winnerPlayer -1).
		g := frenchTarotCuiGame()
		cfg := domain.DefaultFrenchTarotConfig()
		cfg.TargetDeals = 1
		g.SetConfig(cfg)
		g.SetRoundNumber(1)
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidPetite)
		g.SetPlayerScores([domain.FrenchTarotPlayerCnt]int{324, 0, 0, 0})
		g.SetPhase(domain.FrenchTarotPhaseRoundEnd)
		g.ScoreRound()
		assert.True(t, g.GetGameEndFlag())
		assert.Equal(t, -1, g.GetWinnerPlayer())
		assert.NotEmpty(t, p.Output(g, nil))
	})
}

func TestFrenchTarotCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.FrenchTarotCuiPresenter)

	t.Run("bid hint", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetBidPlayerIdx(0)
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("play hint with card index", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidPetite)
		g.SetPhase(domain.FrenchTarotPhasePlay)
		g.SetCurrentPlayerIdx(0)
		result := p.HintOutput(g)
		assert.NotEmpty(t, result)
	})

	t.Run("discard hint uses declarer hand", func(t *testing.T) {
		g := frenchTarotCuiGame()
		g.SetDeclarerIdx(0)
		g.SetContract(domain.FrenchTarotBidPetite)
		g.SetPhase(domain.FrenchTarotPhaseChien)
		result := p.HintOutput(g)
		assert.NotEmpty(t, result)
	})
}

func TestFrenchTarotCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.FrenchTarotCuiPresenter)
	g := frenchTarotCuiGame()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
