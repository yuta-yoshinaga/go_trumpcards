//go:build test

package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCitadelCharacterization(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Citadel
		move  func(*Citadel) error
	}{
		{"tableau move", func() *Citadel { return characterCitadel([][]int{{4}, {5}}) }, func(c *Citadel) error { return c.MoveTableauToTableau(0, 0, 1) }},
		{"foundation move", func() *Citadel { return characterCitadel([][]int{{2}}) }, func(c *Citadel) error { return c.MoveTableauToFoundation(0) }},
		{"autocomplete several", func() *Citadel {
			c := characterCitadel([][]int{{2}, {3}})
			c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
			return c
		}, func(c *Citadel) error { return c.MoveTableauToFoundation(0) }},
		{"stalemate", func() *Citadel { return characterCitadel([][]int{{13}, {13}}) }, func(c *Citadel) error { return c.MoveTableauToTableau(0, 0, 1) }},
		{"one move clear", func() *Citadel {
			c := characterCitadel([][]int{{13}})
			var full [CitadelFoundationCnt][]*Card
			for suit := range CitadelFoundationCnt {
				maxRank := CardValueMax
				if suit == 0 {
					maxRank = CardValueMax - 1
				}
				for rank := 1; rank <= maxRank; rank++ {
					full[suit] = append(full[suit], NewCard(suit+1, rank, false))
				}
			}
			c.SetFoundation(full)
			c.SetTableau([CitadelTableauCnt][]*CitadelTableauCard{{{Card: NewCard(CardDesignSpade, 13, false), FaceUp: true}}})
			return c
		}, func(c *Citadel) error { return c.MoveTableauToFoundation(0) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			hint := c.GetHint()
			err := tt.move(c)
			postMove, _ := c.MarshalJSON()
			_ = c.AutoComplete()
			postAuto, _ := c.MarshalJSON()
			t.Logf("hint=%+v err=%v move=%s auto=%s stale=%v end=%v", hint, err, postMove, postAuto, c.IsStalemate(), c.GetGameEndFlag())
			assertCitadelCharacterization(t, tt.name, hint, err, string(postMove), string(postAuto), c.IsStalemate(), c.GetGameEndFlag())
		})
	}
}

func characterCitadel(columns [][]int) *Citadel {
	c := NewCitadel(nil)
	c.phase = CitadelPhasePlaying
	var tableau [CitadelTableauCnt][]*CitadelTableauCard
	for col, ranks := range columns {
		for _, rank := range ranks {
			tableau[col] = append(tableau[col], &CitadelTableauCard{Card: NewCard(CardDesignSpade, rank, false), FaceUp: true})
		}
	}
	c.SetTableau(tableau)
	var foundation [CitadelFoundationCnt][]*Card
	for suit := range CitadelFoundationCnt {
		foundation[suit] = []*Card{NewCard(suit+1, 1, false)}
	}
	c.SetFoundation(foundation)
	return c
}

func assertCitadelCharacterization(t *testing.T, name string, hint *CitadelHint, err error, moveJSON, autoJSON string, stale, ended bool) {
	t.Helper()
	switch name {
	case "tableau move":
		require.NoError(t, err)
		assert.Equal(t, &CitadelHint{FromCol: 0, CardIndex: 0, ToZone: "tableau", ToCol: 1}, hint)
		assert.True(t, strings.Contains(moveJSON, `"tb":[[],[{"c":{"d":1,"v":5,"w":false},"f":true},{"c":{"d":1,"v":4,"w":false},"f":true}]`))
		assert.True(t, strings.Contains(autoJSON, `"a":"autocomplete"`))
		assert.False(t, stale)
		assert.False(t, ended)
	case "foundation move":
		require.NoError(t, err)
		assert.Equal(t, &CitadelHint{FromCol: 0, CardIndex: 0, ToZone: "foundation", ToCol: 0}, hint)
		assert.True(t, strings.Contains(moveJSON, `"mc":1`))
		assert.True(t, strings.Contains(autoJSON, `"a":"autocomplete"`))
		assert.True(t, stale)
		assert.False(t, ended)
	case "autocomplete several":
		require.NoError(t, err)
		assert.Equal(t, &CitadelHint{FromCol: 0, CardIndex: 0, ToZone: "foundation", ToCol: 0}, hint)
		assert.True(t, strings.Contains(moveJSON, `"mc":1`))
		assert.True(t, strings.Contains(autoJSON, `"mc":2`))
		assert.True(t, stale)
		assert.False(t, ended)
	case "stalemate":
		require.Error(t, err)
		assert.EqualError(t, err, "cannot place card on tableau")
		assert.NotNil(t, hint)
		assert.True(t, strings.Contains(moveJSON, `"mc":0`))
		assert.True(t, strings.Contains(autoJSON, `"a":"autocomplete"`))
		assert.False(t, stale)
		assert.False(t, ended)
	case "one move clear":
		require.NoError(t, err)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.True(t, strings.Contains(moveJSON, `"mc":1`))
		assert.False(t, strings.Contains(autoJSON, `"a":"autocomplete"`))
		assert.False(t, stale)
		assert.True(t, ended)
	}
}

func TestFortressCharacterization(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Fortress
		move  func(*Fortress) error
	}{
		{"tableau move", func() *Fortress { return characterFortress([][]int{{4}, {5}}) }, func(c *Fortress) error { return c.MoveTableauToTableau(0, 0, 1) }},
		{"foundation move", func() *Fortress { return characterFortress([][]int{{2}}) }, func(c *Fortress) error { return c.MoveTableauToFoundation(0) }},
		{"autocomplete several", func() *Fortress {
			c := characterFortress([][]int{{2}, {3}})
			c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
			return c
		}, func(c *Fortress) error { return c.MoveTableauToFoundation(0) }},
		{"stalemate", func() *Fortress { return characterFortress([][]int{{13}, {13}}) }, func(c *Fortress) error { return c.MoveTableauToTableau(0, 0, 1) }},
		{"one move clear", func() *Fortress {
			c := characterFortress([][]int{{13}})
			var full [FortressFoundationCnt][]*Card
			for suit := range FortressFoundationCnt {
				maxRank := CardValueMax
				if suit == 0 {
					maxRank = CardValueMax - 1
				}
				for rank := 1; rank <= maxRank; rank++ {
					full[suit] = append(full[suit], NewCard(suit+1, rank, false))
				}
			}
			c.SetFoundation(full)
			c.SetTableau([FortressTableauCnt][]*FortressTableauCard{{{Card: NewCard(CardDesignSpade, 13, false), FaceUp: true}}})
			return c
		}, func(c *Fortress) error { return c.MoveTableauToFoundation(0) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			hint := c.GetHint()
			err := tt.move(c)
			postMove, _ := c.MarshalJSON()
			_ = c.AutoComplete()
			postAuto, _ := c.MarshalJSON()
			assertFortressCharacterization(t, tt.name, hint, err, string(postMove), string(postAuto), c.IsStalemate(), c.GetGameEndFlag())
		})
	}
}

func characterFortress(columns [][]int) *Fortress {
	c := NewFortress(nil)
	c.phase = FortressPhasePlaying
	var tableau [FortressTableauCnt][]*FortressTableauCard
	for col, ranks := range columns {
		for _, rank := range ranks {
			tableau[col] = append(tableau[col], &FortressTableauCard{Card: NewCard(CardDesignSpade, rank, false), FaceUp: true})
		}
	}
	c.SetTableau(tableau)
	var foundation [FortressFoundationCnt][]*Card
	for suit := range FortressFoundationCnt {
		foundation[suit] = []*Card{NewCard(suit+1, 1, false)}
	}
	c.SetFoundation(foundation)
	return c
}

func TestSomersetCharacterization(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Somerset
		move  func(*Somerset) error
	}{
		{"tableau move", func() *Somerset { return characterSomerset([][]int{{4}, {5}}) }, func(c *Somerset) error { return c.MoveTableauToTableau(0, 0, 1) }},
		{"foundation move", func() *Somerset { return characterSomerset([][]int{{2}}) }, func(c *Somerset) error { return c.MoveTableauToFoundation(0) }},
		{"autocomplete several", func() *Somerset {
			c := characterSomerset([][]int{{2}, {3}})
			c.foundation[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
			return c
		}, func(c *Somerset) error { return c.MoveTableauToFoundation(0) }},
		{"stalemate", func() *Somerset { return characterSomerset([][]int{{13}, {13}}) }, func(c *Somerset) error { return c.MoveTableauToTableau(0, 0, 1) }},
		{"one move clear", func() *Somerset {
			c := characterSomerset([][]int{{13}})
			var full [SomersetFoundationCnt][]*Card
			for suit := range SomersetFoundationCnt {
				maxRank := CardValueMax
				if suit == 0 {
					maxRank = CardValueMax - 1
				}
				for rank := 1; rank <= maxRank; rank++ {
					full[suit] = append(full[suit], NewCard(suit+1, rank, false))
				}
			}
			c.SetFoundation(full)
			c.SetTableau([SomersetTableauCnt][]*SomersetTableauCard{{{Card: NewCard(CardDesignSpade, 13, false), FaceUp: true}}})
			return c
		}, func(c *Somerset) error { return c.MoveTableauToFoundation(0) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			hint := c.GetHint()
			err := tt.move(c)
			postMove, _ := c.MarshalJSON()
			_ = c.AutoComplete()
			postAuto, _ := c.MarshalJSON()
			assertSomersetCharacterization(t, tt.name, hint, err, string(postMove), string(postAuto), c.IsStalemate(), c.GetGameEndFlag())
		})
	}
}

func characterSomerset(columns [][]int) *Somerset {
	c := NewSomerset(nil)
	c.phase = SomersetPhasePlaying
	var tableau [SomersetTableauCnt][]*SomersetTableauCard
	for col, ranks := range columns {
		for _, rank := range ranks {
			suit := CardDesignSpade
			if rank%2 == 0 {
				suit = CardDesignHeart
			}
			tableau[col] = append(tableau[col], &SomersetTableauCard{Card: NewCard(suit, rank, false), FaceUp: true})
		}
	}
	c.SetTableau(tableau)
	var foundation [SomersetFoundationCnt][]*Card
	for suit := range SomersetFoundationCnt {
		foundation[suit] = []*Card{NewCard(suit+1, 1, false)}
	}
	c.SetFoundation(foundation)
	return c
}

func assertFortressCharacterization(t *testing.T, name string, hint *FortressHint, err error, moveJSON, autoJSON string, stale, ended bool) {
	t.Helper()
	switch name {
	case "tableau move":
		require.NoError(t, err)
		assert.Equal(t, &FortressHint{FromCol: 0, CardIndex: 0, ToZone: "tableau", ToCol: 1}, hint)
		assert.True(t, strings.Contains(moveJSON, "\"tb\":[[],[{\"c\":{\"d\":1,\"v\":5,\"w\":false}"))
		assert.True(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.False(t, stale)
		assert.False(t, ended)
	case "foundation move":
		require.NoError(t, err)
		assert.Equal(t, &FortressHint{FromCol: 0, CardIndex: 0, ToZone: "foundation", ToCol: 0}, hint)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":1"))
		assert.True(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.True(t, stale)
		assert.False(t, ended)
	case "autocomplete several":
		require.NoError(t, err)
		assert.Equal(t, &FortressHint{FromCol: 0, CardIndex: 0, ToZone: "foundation", ToCol: 0}, hint)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":1"))
		assert.True(t, strings.Contains(autoJSON, "\"mc\":1"))
		assert.True(t, stale)
		assert.False(t, ended)
	case "stalemate":
		require.Error(t, err)
		assert.EqualError(t, err, "cannot place card on tableau")
		assert.NotNil(t, hint)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":0"))
		assert.True(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.False(t, stale)
		assert.False(t, ended)
	case "one move clear":
		require.NoError(t, err)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":1"))
		assert.False(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.False(t, stale)
		assert.True(t, ended)
	}
}

func assertSomersetCharacterization(t *testing.T, name string, hint *SomersetHint, err error, moveJSON, autoJSON string, stale, ended bool) {
	t.Helper()
	switch name {
	case "tableau move":
		require.NoError(t, err)
		assert.Equal(t, &SomersetHint{FromCol: 0, CardIndex: 0, ToZone: "tableau", ToCol: 1}, hint)
		assert.True(t, strings.Contains(moveJSON, "\"tb\":[[],[{\"c\":{\"d\":1,\"v\":5,\"w\":false},\"f\":true},{\"c\":{\"d\":3,\"v\":4,\"w\":false},\"f\":true}]"))
		assert.True(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.False(t, stale)
		assert.False(t, ended)
	case "foundation move":
		require.NoError(t, err)
		assert.Equal(t, &SomersetHint{FromCol: 0, CardIndex: 0, ToZone: "foundation", ToCol: 2}, hint)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":1"))
		assert.True(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.True(t, stale)
		assert.False(t, ended)
	case "autocomplete several":
		require.NoError(t, err)
		assert.Equal(t, &SomersetHint{FromCol: 0, CardIndex: 0, ToZone: "foundation", ToCol: 2}, hint)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":1"))
		assert.True(t, strings.Contains(autoJSON, "\"mc\":1"))
		assert.False(t, stale)
		assert.False(t, ended)
	case "stalemate":
		require.Error(t, err)
		assert.EqualError(t, err, "cannot place card on tableau")
		assert.NotNil(t, hint)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":0"))
		assert.True(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.False(t, stale)
		assert.False(t, ended)
	case "one move clear":
		require.NoError(t, err)
		assert.Equal(t, "foundation", hint.ToZone)
		assert.True(t, strings.Contains(moveJSON, "\"mc\":1"))
		assert.False(t, strings.Contains(autoJSON, "\"a\":\"autocomplete\""))
		assert.False(t, stale)
		assert.True(t, ended)
	}
}
