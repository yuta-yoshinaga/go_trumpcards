//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertMusDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func TestMusDomainErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultMus()
	g.Reset()
	g.SetPhase(domain.MusPhaseDiscard)
	g.SetDiscardTurn(0)
	assertMusDomainError(t, g.PlayerDiscard([]int{-1}), domain.ErrInvalidCard, "mus.errCardIndexOutOfRange")
	assertMusDomainError(t, g.PlayerDiscard([]int{0, 0}), domain.ErrInvalidPlay, "mus.errDuplicateCardIndex")

	g.SetPhase(domain.MusPhaseGrande)
	g.SetBetTeam(0)
	g.SetPendingStake(2)
	assertMusDomainError(t, g.PlayerBet(domain.MusActionPaso, 0), domain.ErrInvalidPlay, "mus.errCannotPass")
	err := g.PlayerBet(domain.MusActionEnvido, 2)
	assertMusDomainError(t, err, domain.ErrInvalidPlay, "mus.errRaiseMustGoUp")
	de, ok := err.(*domain.DomainError)
	require.True(t, ok)
	assert.Equal(t, map[string]string{"val": "2"}, de.MessageParams())

	g.SetPendingStake(-1)
	assertMusDomainError(t, g.PlayerBet(domain.MusActionEnvido, 4), domain.ErrInvalidPlay, "mus.errCannotEnvido")
	g.SetPendingStake(0)
	assertMusDomainError(t, g.PlayerBet(domain.MusActionQuiero, 0), domain.ErrInvalidPlay, "mus.errNoBetToAccept")
	assertMusDomainError(t, g.PlayerBet(domain.MusActionNoQuiero, 0), domain.ErrInvalidPlay, "mus.errNoBetToDecline")
	assertMusDomainError(t, g.PlayerBet(99, 0), domain.ErrInvalidPlay, "mus.errInvalidAction")
}
