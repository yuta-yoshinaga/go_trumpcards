//go:build test

package domain

import (
	"errors"
	"testing"
)

func assertSheepsheadCodedError(t *testing.T, err error, sentinel error, code string) {
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

func TestSheepsheadDomainErrorsHaveMessageCodes(t *testing.T) {
	t.Run("last player cannot pass", func(t *testing.T) {
		g := newSSGame(true)
		g.phase, g.currentPlayerIdx, g.passCount = SheepsheadPhasePick, 0, SheepsheadPlayerCnt-1
		assertSheepsheadCodedError(t, g.PlayerPick(false), ErrInvalidPlay, "sheepshead.errLastPlayerCannotPass")
	})

	t.Run("bury count", func(t *testing.T) {
		g := newSSGame(true)
		g.phase, g.pickerIdx, g.currentPlayerIdx = SheepsheadPhaseBury, 0, 0
		assertSheepsheadCodedError(t, g.PlayerBury([]int{0}), ErrInvalidPlay, "sheepshead.errBuryCount")
	})

	t.Run("bury index and duplicate", func(t *testing.T) {
		for _, tc := range []struct {
			name, code string
			indices    []int
		}{
			{"index", "sheepshead.errCardIndexOutOfRange", []int{-1, 0}},
			{"duplicate", "sheepshead.errDuplicateCard", []int{0, 0}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := newSSGame(true)
				g.phase, g.pickerIdx, g.currentPlayerIdx = SheepsheadPhaseBury, 0, 0
				ssSetHand(g.players[0], ssCard(CardDesignSpade, 7), ssCard(CardDesignHeart, 7))
				assertSheepsheadCodedError(t, g.PlayerBury(tc.indices), map[string]error{
					"sheepshead.errCardIndexOutOfRange": ErrInvalidCard,
					"sheepshead.errDuplicateCard":       ErrInvalidPlay,
				}[tc.code], tc.code)
			})
		}
	})

	t.Run("uncallable suit", func(t *testing.T) {
		g := newSSGame(true)
		g.phase, g.pickerIdx, g.currentPlayerIdx = SheepsheadPhaseCall, 0, 0
		assertSheepsheadCodedError(t, g.PlayerCall(CardDesignDiamond), ErrInvalidPlay, "sheepshead.errSuitNotCallable")
	})

	t.Run("play index and follow suit", func(t *testing.T) {
		g := newSSGame(true)
		g.phase, g.currentPlayerIdx = SheepsheadPhasePlay, 0
		ssSetHand(g.players[0], ssCard(CardDesignClover, 7), ssCard(CardDesignSpade, 7))
		assertSheepsheadCodedError(t, g.PlayerPlay(-1), ErrInvalidCard, "sheepshead.errCardIndexOutOfRange")

		g = newSSGame(true)
		g.phase, g.currentPlayerIdx = SheepsheadPhasePlay, 0
		ssSetHand(g.players[0], ssCard(CardDesignClover, 7), ssCard(CardDesignSpade, 7))
		g.currentTrick = []*TrickCard{{PlayerIdx: 1, Card: ssCard(CardDesignClover, 1)}}
		assertSheepsheadCodedError(t, g.PlayerPlay(1), ErrInvalidPlay, "sheepshead.errFollowLeadSuit")
	})
}
