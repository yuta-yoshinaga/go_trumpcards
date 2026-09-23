//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertMachiavelliDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestMachiavelliDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	assertMachiavelliDomainError(t, g.PlayerLayoff(-1, 0), domain.ErrInvalidPlay, "machiavelli.errInvalidMeldIndex")

	g = newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	g.SetTable([][]*domain.Card{{
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignSpade, 6),
		machiavelliCard(domain.CardDesignSpade, 7),
	}})
	assertMachiavelliDomainError(t, g.PlayerLayoff(0, -1), domain.ErrInvalidCard, "machiavelli.errCardIndexOutOfRange")

	g = newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	assertMachiavelliDomainError(t, g.PlayerPlay(nil, nil), domain.ErrInvalidPlay, "machiavelli.errHandCardRequired")

	g = newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(machiavelliCard(domain.CardDesignSpade, 2))
	assertMachiavelliDomainError(t, g.PlayerPlay(machiavelliRefsOf([]*domain.Card{
		machiavelliCard(domain.CardDesignHeart, 5),
		machiavelliCard(domain.CardDesignHeart, 6),
	}), []int{0}), domain.ErrInvalidPlay, "machiavelli.errInvalidMeld")

	g = newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).AddCard(machiavelliCard(domain.CardDesignSpade, 2))
	assertMachiavelliDomainError(t, g.PlayerPlay(machiavelliRefsOf([]*domain.Card{
		machiavelliCard(domain.CardDesignHeart, 5),
		machiavelliCard(domain.CardDesignSpade, 5),
		machiavelliCard(domain.CardDesignClover, 5),
	}), []int{0}), domain.ErrInvalidPlay, "machiavelli.errTableCardsMismatch")

	g = newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	assertMachiavelliDomainError(t, g.PlayerPlay([][]domain.MachiavelliCardRef{{}}, nil), domain.ErrInvalidPlay, "machiavelli.errEmptyMeld")

	g = newTestMachiavelli(2)
	g.SetPhase(domain.MachiavelliPhaseTurn)
	g.SetCurrentPlayerIdx(0)
	assertMachiavelliDomainError(t, g.PlayerPlay([][]domain.MachiavelliCardRef{{{Design: 99, Value: 1}}}, nil), domain.ErrInvalidCard, "machiavelli.errInvalidCard")
}
