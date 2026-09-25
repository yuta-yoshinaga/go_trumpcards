//go:build test

package domain

import "testing"

func TestSevenTwentySevenSettleMapOrder(t *testing.T) {
	for iteration := 0; iteration < 50; iteration++ {
		players := []*SevenTwentySevenPlayer{NewSevenTwentySevenPlayer(true, 100), NewSevenTwentySevenPlayer(false, 100), NewSevenTwentySevenPlayer(false, 100), NewSevenTwentySevenPlayer(false, 100)}
		for _, p := range players {
			p.AddCard(NewCard(CardDesignSpade, 3, false))
			p.AddCard(NewCard(CardDesignHeart, 4, false))
		}
		g := NewSevenTwentySeven(NewTrumpCards(0), players, DefaultSevenTwentySevenConfig())
		g.state.pot = 8
		g.settle()
		logs := g.GetActionLog()
		if len(logs) != 4 || logs[0].PlayerIdx != 0 || logs[1].PlayerIdx != 1 || logs[2].PlayerIdx != 2 || logs[3].PlayerIdx != 3 {
			t.Fatalf("iteration %d: winner log seats = %v", iteration, logs)
		}
	}
}
