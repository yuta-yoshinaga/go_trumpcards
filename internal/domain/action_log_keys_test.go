//go:build test

package domain

import "testing"

func TestActionLogKeys(t *testing.T) {
	for rank := PokerHandHighCard; rank <= PokerHandFiveOfAKind; rank++ {
		if pokerHandLogKey(rank) == "pokerhand.unknown" {
			t.Fatalf("known poker rank %d returned unknown", rank)
		}
	}
	if pokerHandLogKey(-1) != "pokerhand.unknown" {
		t.Fatal("invalid poker rank did not return unknown key")
	}
	for size := 1; size <= 4; size++ {
		if badugiHandLogKey(size) == "badugi.handRankUnknown" {
			t.Fatalf("known Badugi size %d returned unknown", size)
		}
	}
	if badugiHandLogKey(0) != "badugi.handRankUnknown" {
		t.Fatal("invalid Badugi size did not return unknown key")
	}
	for _, name := range []string{"Royal Flush", "Natural Royal Flush", "Wild Royal Flush", "Four Deuces", "Five of a Kind", "Straight Flush", "Four of a Kind", "Full House", "Flush", "Straight", "Three of a Kind", "Two Pair", "Jacks or Better", "Kings or Better"} {
		if videoPokerHandLogKey(name) == "pokerhand.unknown" {
			t.Fatalf("known video poker hand %q returned unknown", name)
		}
	}
	if videoPokerHandLogKey("") != "pokerhand.unknown" {
		t.Fatal("unknown video poker hand did not return unknown key")
	}
	for _, name := range []string{"Monsieur", "Madame", "Borgne", "Vache", "GrandNeuf", "PetitNeuf"} {
		if aluetteLuetteLogKey(name) == "aluette.luette.unknown" {
			t.Fatalf("known Aluette name %q returned unknown", name)
		}
	}
	if aluetteLuetteLogKey("") != "aluette.luette.unknown" || cirullaBonusLogKey("") != "cirulla.bonus.unknown" {
		t.Fatal("unknown named log value did not return an unknown key")
	}
	for color := 1; color <= 4; color++ {
		if rookColorLogKey(color) == "rook.colorUnknown" {
			t.Fatalf("known Rook color %d returned unknown", color)
		}
	}
	if rookColorLogKey(0) != "rook.colorUnknown" {
		t.Fatal("invalid Rook color did not return unknown key")
	}
}
