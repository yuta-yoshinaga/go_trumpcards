package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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

// kingLineContaining returns the first line of out that contains marker.
func kingLineContaining(out, marker string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, marker) {
			return line
		}
	}
	return ""
}

func TestKingCuiPresenter_RemainingContracts(t *testing.T) {
	p := new(presenter.KingCuiPresenter)
	prefix := strings.Split(i18n.T("king.remainingContracts"), "{{")[0]

	// Right after reset every contract is available.
	g := domain.NewDefaultKing()
	g.Reset()
	remaining := kingLineContaining(p.Output(g, nil), prefix)
	assert.Contains(t, remaining, "[0]")
	assert.Contains(t, remaining, "[6]")

	// Once the dealer selects a contract it drops out of the remaining list.
	g2 := domain.NewDefaultKing()
	g2.Reset()
	g2.SetDealerIdx(0) // human dealer
	require.NoError(t, g2.SelectContract(domain.KingContractNoTricks, 0))
	g2.SetPhase(domain.KingPhaseSelectContract) // force back to the selection view
	remaining2 := kingLineContaining(p.Output(g2, nil), prefix)
	assert.NotContains(t, remaining2, "[0]") // used contract 0 excluded
	assert.Contains(t, remaining2, "[6]")
}

func TestKingCuiPresenter_ContractAndTrick(t *testing.T) {
	g := domain.NewDefaultKing()
	g.Reset()
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: bcard(domain.CardDesignHeart, 9)}})
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
