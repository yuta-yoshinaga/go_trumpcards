package entities

import (
	"testing"
)

func TestDaifugo_Reset(t *testing.T) {
	players := []*DaifugoPlayer{
		NewDaifugoPlayer(true),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
	}
	d := NewDaifugo(NewTrumpCards(2), players)
	d.Reset()

	if d.GetGameEndFlag() {
		t.Error("Game should not be ended after reset")
	}

	totalCards := 0
	for _, p := range d.GetPlayers() {
		totalCards += p.GetCardsSize()
	}
	if totalCards != 54 { // 52 + 2 jokers
		t.Errorf("Total cards should be 54, but got %d", totalCards)
	}
}

func TestDaifugo_CanPlay(t *testing.T) {
	players := []*DaifugoPlayer{
		NewDaifugoPlayer(true),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
	}
	d := NewDaifugo(NewTrumpCards(2), players)
	
	// 3 of Spades
	c3s := NewCard(CardDesignSpade, 3, false)
	// 4 of Spades
	c4s := NewCard(CardDesignSpade, 4, false)
	// 2 of Spades
	c2s := NewCard(CardDesignSpade, 2, false)
	// Joker
	cj := NewCard(CardDesignJoker, 0, false)

	if !d.CanPlay([]*Card{c3s}) {
		t.Error("Should be able to play a single card when field is empty")
	}

	d.lastPlay = []*Card{c3s}
	if !d.CanPlay([]*Card{c4s}) {
		t.Error("4 should be stronger than 3")
	}
	if d.CanPlay([]*Card{c3s}) {
		t.Error("3 should not be stronger than 3")
	}
	if !d.CanPlay([]*Card{c2s}) {
		t.Error("2 should be stronger than 3")
	}
	if !d.CanPlay([]*Card{cj}) {
		t.Error("Joker should be stronger than 3")
	}

	d.lastPlay = []*Card{c2s}
	if !d.CanPlay([]*Card{cj}) {
		t.Error("Joker should be stronger than 2")
	}
	if d.CanPlay([]*Card{c4s}) {
		t.Error("4 should not be stronger than 2")
	}

	// Pairs
	d.lastPlay = nil // Ensure field is empty
	c3h := NewCard(CardDesignHeart, 3, false)
	c4h := NewCard(CardDesignHeart, 4, false)
	if !d.CanPlay([]*Card{c3s, c3h}) {
		t.Error("Should be able to play a pair when field is empty")
	}
	
	d.lastPlay = []*Card{c3s, c3h}
	if d.CanPlay([]*Card{c4s}) {
		t.Error("Should not be able to play a single card over a pair")
	}
	if !d.CanPlay([]*Card{c4s, c4h}) {
		t.Error("Should be able to play a stronger pair")
	}
}

func TestDaifugo_Revolution(t *testing.T) {
	players := []*DaifugoPlayer{
		NewDaifugoPlayer(true),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
	}
	d := NewDaifugo(NewTrumpCards(2), players)
	d.isRevolution = true

	c3s := NewCard(CardDesignSpade, 3, false)
	c4s := NewCard(CardDesignSpade, 4, false)

	d.lastPlay = []*Card{c4s}
	if !d.CanPlay([]*Card{c3s}) {
		t.Error("3 should be stronger than 4 during revolution")
	}
	if d.CanPlay([]*Card{NewCard(CardDesignSpade, 5, false)}) {
		t.Error("5 should not be stronger than 4 during revolution")
	}
}
