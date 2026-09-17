//go:build test

package domain

import (
	"errors"
	"testing"
)

func assertFiveHundredCodedError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want errors.Is(%v)", err, sentinel)
	}
	de, ok := err.(*DomainError)
	if !ok {
		t.Fatalf("error type = %T, want *DomainError", err)
	}
	if de.MessageCode() != code {
		t.Errorf("message code = %q, want %q", de.MessageCode(), code)
	}
	if de.MessageParams() != nil {
		t.Errorf("message params = %#v, want nil", de.MessageParams())
	}
}

func newFiveHundredErrorTest() *FiveHundred {
	players := []*FiveHundredPlayer{
		NewFiveHundredPlayer(true, 0), NewFiveHundredPlayer(false, 1),
		NewFiveHundredPlayer(false, 0), NewFiveHundredPlayer(false, 1),
	}
	return NewFiveHundred(NewTrumpCardsFiveHundred(), players, DefaultFiveHundredConfig())
}

func TestFiveHundredDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("invalid bid", func(t *testing.T) {
		g := newFiveHundredErrorTest()
		g.phase, g.bidPlayerIdx = FiveHundredPhaseBid, 0
		assertFiveHundredCodedError(t, g.PlayerBid(FiveHundredContractSuit, 5, CardDesignSpade), ErrInvalidPlay, "fivehundred.errInvalidBid")
	})

	t.Run("bid must be higher", func(t *testing.T) {
		g := newFiveHundredErrorTest()
		g.phase, g.bidPlayerIdx = FiveHundredPhaseBid, 0
		g.highestBid = &FiveHundredBid{Kind: FiveHundredContractSuit, Tricks: 7, Suit: CardDesignSpade}
		assertFiveHundredCodedError(t, g.PlayerBid(FiveHundredContractSuit, 6, CardDesignSpade), ErrInvalidPlay, "fivehundred.errBidMustBeHigher")
	})

	t.Run("kitty discard validation", func(t *testing.T) {
		for _, tc := range []struct {
			name, code string
			indices    []int
			sentinel   error
		}{
			{"count", "fivehundred.errDiscardThreeCards", []int{0, 1}, ErrInvalidCard},
			{"index", "fivehundred.errCardIndexOutOfRange", []int{0, 1, 9}, ErrInvalidCard},
			{"duplicate", "fivehundred.errDuplicateCard", []int{0, 0, 1}, ErrInvalidCard},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := newFiveHundredErrorTest()
				g.phase, g.declarerIdx = FiveHundredPhaseKittyExchange, 0
				g.players[0].AddCard(NewCard(CardDesignSpade, 7, false))
				g.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
				g.players[0].AddCard(NewCard(CardDesignDiamond, 7, false))
				assertFiveHundredCodedError(t, g.PlayerExchangeKitty(tc.indices), tc.sentinel, tc.code)
			})
		}
	})

	t.Run("play index and follow suit", func(t *testing.T) {
		g := newFiveHundredErrorTest()
		g.phase, g.currentPlayerIdx = FiveHundredPhasePlay, 0
		g.contract = FiveHundredBid{Kind: FiveHundredContractSuit, Tricks: 7, Suit: CardDesignSpade}
		g.players[0].AddCard(NewCard(CardDesignSpade, 7, false))
		assertFiveHundredCodedError(t, g.PlayerPlay(-1, -1), ErrInvalidCard, "fivehundred.errCardIndexOutOfRange")

		g = newFiveHundredErrorTest()
		g.phase, g.currentPlayerIdx = FiveHundredPhasePlay, 0
		g.contract = FiveHundredBid{Kind: FiveHundredContractSuit, Tricks: 7, Suit: CardDesignSpade}
		g.players[0].AddCard(NewCard(CardDesignSpade, 7, false))
		g.players[0].AddCard(NewCard(CardDesignHeart, 7, false))
		g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)}}
		assertFiveHundredCodedError(t, g.PlayerPlay(0, -1), ErrInvalidPlay, "fivehundred.errFollowLeadSuit")
	})
}
