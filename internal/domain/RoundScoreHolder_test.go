//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundScoreHolder_GetSetRoundScore(t *testing.T) {
	h := &RoundScoreHolder{}
	assert.Equal(t, 0, h.GetRoundScore())
	h.SetRoundScore(42)
	assert.Equal(t, 42, h.GetRoundScore())
}

func TestRoundScoreHolder_GetSetCumulativeScore(t *testing.T) {
	h := &RoundScoreHolder{}
	assert.Equal(t, 0, h.GetCumulativeScore())
	h.SetCumulativeScore(100)
	assert.Equal(t, 100, h.GetCumulativeScore())
}

func TestRoundScoreHolder_CommitRoundScore(t *testing.T) {
	h := &RoundScoreHolder{}
	h.SetRoundScore(10)
	h.CommitRoundScore()
	assert.Equal(t, 10, h.GetCumulativeScore())
	assert.Equal(t, 10, h.GetRoundScore())

	h.SetRoundScore(20)
	h.CommitRoundScore()
	assert.Equal(t, 30, h.GetCumulativeScore())
}
