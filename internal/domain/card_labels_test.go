//go:build test

package domain

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionLogCardLabels(t *testing.T) {
	assert.Equal(t, "♠A", cuiCardName(NewCard(CardDesignSpade, 1, false)))
	assert.Equal(t, "JK", cuiCardName(NewCard(CardDesignJoker, 1, false)))
	assert.Equal(t, "JK", mightyCardStr(NewCard(CardDesignJoker, 1, false)))
	assert.Equal(t, "JK", napoleonCardStr(NewCard(CardDesignJoker, 1, false)))
	assert.Equal(t, "♠A", cardLogStr(NewCard(CardDesignSpade, 1, false)))

	for _, value := range []int{1, 9} {
		assert.True(t, isUnsunKarutaNumber(value))
	}
	for _, value := range []int{0, UnsunKarutaSota} {
		assert.False(t, isUnsunKarutaNumber(value))
	}
}

func TestRookPlayActionLogUsesStructuredCardParams(t *testing.T) {
	g := NewDefaultRook()
	g.playCard(0, NewCard(1, 14, false))
	entry := requireLastActionLog(t, g.GetActionLog())
	assert.Equal(t, "rook.log.play", entry.DetailCode)
	assert.Equal(t, "rook.colorRed", entry.DetailParams["colorKey"])
	assert.Equal(t, "14", entry.DetailParams["value"])
	assert.NotContains(t, entry.DetailParams, "card")

	g.playCard(0, NewCard(RookBirdDesign, 0, false))
	entry = requireLastActionLog(t, g.GetActionLog())
	assert.Equal(t, "rook.log.playBird", entry.DetailCode)
	assert.Equal(t, "rook.birdName", entry.DetailParams["birdKey"])
	assert.NotContains(t, entry.DetailParams, "card")
}

func TestUnsunKarutaPlayActionLogUsesStructuredCardParams(t *testing.T) {
	g := NewDefaultUnsunKaruta()
	g.playCard(0, NewCard(UnsunKarutaSuitPao, 1, false))
	entry := requireLastActionLog(t, g.GetActionLog())
	assert.Equal(t, "unsunkaruta.log.playNumber", entry.DetailCode)
	assert.Equal(t, "unsunkaruta.suit.pao", entry.DetailParams["suitKey"])
	assert.Equal(t, "1", entry.DetailParams["value"])
	assert.NotContains(t, entry.DetailParams, "card")

	g.playCard(0, NewCard(UnsunKarutaSuitPao, UnsunKarutaSota, false))
	entry = requireLastActionLog(t, g.GetActionLog())
	assert.Equal(t, "unsunkaruta.log.playRank", entry.DetailCode)
	assert.Equal(t, "unsunkaruta.suit.pao", entry.DetailParams["suitKey"])
	assert.Equal(t, "unsunkaruta.rank.sota", entry.DetailParams["rankKey"])
	assert.NotContains(t, entry.DetailParams, "card")
}

func requireLastActionLog(t *testing.T, entries []*ActionLogEntry) *ActionLogEntry {
	t.Helper()
	require.NotEmpty(t, entries)
	return entries[len(entries)-1]
}

func TestUnsunKarutaRankKeysCoverAllBranches(t *testing.T) {
	known := map[int]string{
		UnsunKarutaSota: "sota", UnsunKarutaUma: "uma", UnsunKarutaKiri: "kiri",
		UnsunKarutaUn: "un", UnsunKarutaSun: "sun", UnsunKarutaRobai: "robai",
	}
	for value, want := range known {
		assert.Equal(t, want, UnsunKarutaRankName(value))
	}
	assert.Equal(t, strconv.Itoa(7), UnsunKarutaRankName(7))
}
