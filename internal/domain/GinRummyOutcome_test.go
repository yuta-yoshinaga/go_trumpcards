//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// The same three branches scoreRound takes, stated where the presenter can read
// them. The undercut case is the one worth naming: the winner is the opponent,
// not the knocker, so a bare score line reads as the wrong player gaining.
func TestGinRummyRoundOutcomeOf(t *testing.T) {
	t.Run("gin pays the opponent's deadwood plus the bonus", func(t *testing.T) {
		o := domain.GinRummyRoundOutcomeOf(0, 0, 18, true)
		assert.Equal(t, domain.GinRummyOutcomeGin, o.Kind)
		assert.Equal(t, 0, o.WinnerIdx)
		assert.Equal(t, 18+o.Bonus, o.Total)
	})

	t.Run("a plain knock pays the difference with no bonus", func(t *testing.T) {
		o := domain.GinRummyRoundOutcomeOf(0, 5, 20, false)
		assert.Equal(t, domain.GinRummyOutcomeKnock, o.Kind)
		assert.Equal(t, 0, o.WinnerIdx)
		assert.Equal(t, 0, o.Bonus)
		assert.Equal(t, 15, o.Total)
	})

	t.Run("an undercut pays the OPPONENT, not the knocker", func(t *testing.T) {
		o := domain.GinRummyRoundOutcomeOf(0, 10, 4, false)
		assert.Equal(t, domain.GinRummyOutcomeUndercut, o.Kind)
		assert.Equal(t, 1, o.WinnerIdx, "the defender scores an undercut")
		assert.Equal(t, 6+o.Bonus, o.Total)
	})

	t.Run("equal deadwood is an undercut, not a knock", func(t *testing.T) {
		// scoreRound's test is `opponentDeadwood <= knockerDeadwood`; a `<` here
		// would silently hand the round to the wrong player on a tie.
		o := domain.GinRummyRoundOutcomeOf(1, 7, 7, false)
		assert.Equal(t, domain.GinRummyOutcomeUndercut, o.Kind)
		assert.Equal(t, 0, o.WinnerIdx)
		assert.Equal(t, o.Bonus, o.Total, "a tie contributes no difference, only the bonus")
	})
}
