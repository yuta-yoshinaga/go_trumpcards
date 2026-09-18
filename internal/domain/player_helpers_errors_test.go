//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type playerHelpersErrorTestHand struct {
	cards []*Card
}

func (p *playerHelpersErrorTestHand) GetCardsSize() int   { return len(p.cards) }
func (p *playerHelpersErrorTestHand) GetCard(i int) *Card { return p.cards[i] }

type playerHelpersErrorTestEndgame struct {
	endgame bool
	valid   bool
}

func (g playerHelpersErrorTestEndgame) IsEndgame() bool { return g.endgame }
func (g playerHelpersErrorTestEndgame) cardSatisfiesFollow(int, *Card) bool {
	return g.valid
}

func assertPlayerHelpersDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, errors.Is(err, sentinel))
	de, ok := err.(*DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
	assert.Nil(t, de.MessageParams())
}

func TestPlayerHelpersDomainErrorsHaveMessageCodes(t *testing.T) {
	lead := NewCard(CardDesignSpade, 1, false)
	offSuit := NewCard(CardDesignHeart, 2, false)
	hand := &playerHelpersErrorTestHand{cards: []*Card{lead, offSuit}}
	trick := []*TrickCard{{PlayerIdx: 1, Card: lead}}

	assertPlayerHelpersDomainError(t,
		validateFollowSuit(trick, []*playerHelpersErrorTestHand{hand}, 0, offSuit),
		ErrInvalidPlay, "shared.errFollowLeadSuit")

	assertPlayerHelpersDomainError(t,
		validateCardIsPlayable([]int{0}, hand, offSuit),
		ErrInvalidPlay, "shared.errFollowRuleViolation")

	assertPlayerHelpersDomainError(t,
		validateEndgameFollow(nil, playerHelpersErrorTestEndgame{}, 0, nil),
		ErrInvalidCard, "shared.errCardNil")

	assertPlayerHelpersDomainError(t,
		validateEndgameFollow(trick, playerHelpersErrorTestEndgame{endgame: true}, 0, offSuit),
		ErrInvalidCard, "shared.errMustFollowEndgameRule")
}
