package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertContractRummyDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func contractRummyErrorGame() (*ContractRummy, *ContractRummyPlayer, *ContractRummyPlayer) {
	g := NewDefaultContractRummy()
	g.Reset()
	current := g.GetPlayer(0)
	target := g.GetPlayer(1)
	current.Reset()
	target.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(ContractRummyPhasePlay)
	return g, current, target
}

func TestContractRummy_AdditionalDomainErrorCallSitesHaveCodes(t *testing.T) {
	t.Run("contract slot card count", func(t *testing.T) {
		g, p, _ := contractRummyErrorGame()
		p.AddCard(crCard(CardDesignSpade, 5))
		assertContractRummyDomainError(t, g.PlayerMeldContract([][]int{{0, 1}, {}}), ErrInvalidPlay, "contractrummy.errContractSlotCardCount")
	})

	t.Run("contract card index", func(t *testing.T) {
		g, p, _ := contractRummyErrorGame()
		p.AddCard(crCard(CardDesignSpade, 5))
		p.AddCard(crCard(CardDesignHeart, 5))
		p.AddCard(crCard(CardDesignDiamond, 5))
		assertContractRummyDomainError(t, g.PlayerMeldContract([][]int{{-1, 0, 1}, {0, 1, 2}}), ErrInvalidCard, "contractrummy.errCardIndexOutOfRange")
	})

	t.Run("extra meld minimum", func(t *testing.T) {
		g, p, _ := contractRummyErrorGame()
		p.SetContractMet(true)
		p.AddCard(crCard(CardDesignSpade, 5))
		p.AddCard(crCard(CardDesignHeart, 5))
		assertContractRummyDomainError(t, g.PlayerMeldExtra([]int{0, 1}), ErrInvalidPlay, "contractrummy.errMeldMinimumCards")
	})

	t.Run("layoff target player", func(t *testing.T) {
		g, p, _ := contractRummyErrorGame()
		p.SetContractMet(true)
		p.AddCard(crCard(CardDesignSpade, 5))
		assertContractRummyDomainError(t, g.PlayerLayoff(-1, 0, 0), ErrInvalidPlay, "contractrummy.errTargetPlayerInvalid")
	})

	t.Run("layoff target meld", func(t *testing.T) {
		g, p, target := contractRummyErrorGame()
		p.SetContractMet(true)
		target.SetContractMet(true)
		p.AddCard(crCard(CardDesignSpade, 5))
		assertContractRummyDomainError(t, g.PlayerLayoff(1, 0, 0), ErrInvalidPlay, "contractrummy.errTargetMeldInvalid")
	})

	t.Run("layoff card index", func(t *testing.T) {
		g, p, target := contractRummyErrorGame()
		p.SetContractMet(true)
		target.SetContractMet(true)
		target.SetMelds([][]*Card{{crCard(CardDesignSpade, 5), crCard(CardDesignHeart, 5), crCard(CardDesignDiamond, 5)}})
		assertContractRummyDomainError(t, g.PlayerLayoff(1, 0, -1), ErrInvalidCard, "contractrummy.errCardIndexOutOfRange")
	})

	t.Run("layoff card cannot add", func(t *testing.T) {
		g, p, target := contractRummyErrorGame()
		p.SetContractMet(true)
		target.SetContractMet(true)
		target.SetMelds([][]*Card{{crCard(CardDesignSpade, 5), crCard(CardDesignHeart, 5), crCard(CardDesignDiamond, 5)}})
		p.AddCard(crCard(CardDesignSpade, 9))
		assertContractRummyDomainError(t, g.PlayerLayoff(1, 0, 0), ErrInvalidPlay, "contractrummy.errLayoffCardCannotAdd")
	})
}

func TestContractRummy_ContractSlotErrorsUseSeparateCodes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		round    int
		cards    []*Card
		indices  [][]int
		wantCode string
	}{
		{"set", 1, []*Card{crCard(CardDesignSpade, 5), crCard(CardDesignHeart, 6), crCard(CardDesignDiamond, 7), crCard(CardDesignClover, 8), crCard(CardDesignSpade, 9), crCard(CardDesignHeart, 10)}, [][]int{{0, 1, 2}, {3, 4, 5}}, "contractrummy.errContractSlotSet"},
		{"run", 2, []*Card{crCard(CardDesignSpade, 5), crCard(CardDesignHeart, 5), crCard(CardDesignDiamond, 5), crCard(CardDesignSpade, 5), crCard(CardDesignHeart, 7), crCard(CardDesignDiamond, 9), crCard(CardDesignClover, 11)}, [][]int{{0, 1, 2}, {3, 4, 5, 6}}, "contractrummy.errContractSlotRun"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, p, _ := contractRummyErrorGame()
			g.SetRoundNumber(tc.round)
			setHand(p, tc.cards)
			assertContractRummyDomainError(t, g.PlayerMeldContract(tc.indices), ErrInvalidPlay, tc.wantCode)
		})
	}
}
