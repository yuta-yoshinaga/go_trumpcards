package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestCalcMaxBetAmount_PotLimit(t *testing.T) {
	result := calcMaxBetAmount(domain.BettingLimitPotLimit, 100, 20)
	assert.Equal(t, 120, result)
}

func TestCalcMaxBetAmount_Fixed(t *testing.T) {
	result := calcMaxBetAmount(domain.BettingLimitFixed, 100, 20)
	assert.Equal(t, 0, result)
}

func TestCalcMaxBetAmount_NoLimit(t *testing.T) {
	result := calcMaxBetAmount(domain.BettingLimitNoLimit, 100, 20)
	assert.Equal(t, 0, result)
}
