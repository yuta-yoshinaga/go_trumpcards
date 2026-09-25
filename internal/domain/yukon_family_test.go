//go:build test

package domain

import "testing"

func TestYukonFamilyReadHelpers(t *testing.T) {
	cards := [][]*KlondikeTableauCard{{{Card: NewCard(CardDesignSpade, 7, true), FaceUp: true}}, {{Card: NewCard(CardDesignHeart, 8, true), FaceUp: false}}}
	faceUp := func(card *KlondikeTableauCard) bool { return card.FaceUp }
	if !yukonFamilyAllFaceUp([][]*KlondikeTableauCard{{{FaceUp: true}}}, faceUp) {
		t.Fatal("all face-up tableau reported face-down")
	}
	if yukonFamilyAllFaceUp(cards, faceUp) {
		t.Fatal("face-down card reported as all face-up")
	}
	if got := yukonFamilyGetHint(cards, 4, func(tc *KlondikeTableauCard) *Card { return tc.Card }, faceUp, func(*Card, int) bool { return false }, func(*Card, int) bool { return false }); got != nil {
		t.Fatalf("unexpected hint: %+v", got)
	}
	foundation := make([][]*Card, 4)
	cleared := false
	yukonFamilyCheckGameClear(foundation, CardValueMax, func() { cleared = true })
	if cleared {
		t.Fatal("incomplete foundations marked clear")
	}
	for i := range foundation {
		foundation[i] = make([]*Card, CardValueMax)
	}
	yukonFamilyCheckGameClear(foundation, CardValueMax, func() { cleared = true })
	if !cleared {
		t.Fatal("complete foundations not marked clear")
	}
}
