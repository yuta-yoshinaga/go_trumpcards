//go:build test

package domain

import "testing"

func TestDaifugoSolverGenerateSequencePlaysMapOrder(t *testing.T) {
	s := &daifugoSolver{}
	hand := []*Card{NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false)}
	for i := 0; i < 50; i++ {
		got := s.generateSequencePlays(hand)
		if len(got) < 2 || got[0].cards[0].GetDesign() != CardDesignSpade {
			t.Fatalf("iteration %d: first suit %v", i, got)
		}
	}
}

func TestDaifugoSolverGenerateSequenceResponsePlaysMapOrder(t *testing.T) {
	s := &daifugoSolver{}
	hand := []*Card{NewCard(CardDesignHeart, 3, false), NewCard(CardDesignHeart, 4, false), NewCard(CardDesignHeart, 5, false), NewCard(CardDesignSpade, 3, false), NewCard(CardDesignSpade, 4, false), NewCard(CardDesignSpade, 5, false)}
	for i := 0; i < 50; i++ {
		got := s.generateSequenceResponsePlays(hand, 3, 0)
		if len(got) < 2 || got[0].cards[0].GetDesign() != CardDesignSpade {
			t.Fatalf("iteration %d: first suit %v", i, got)
		}
	}
}
