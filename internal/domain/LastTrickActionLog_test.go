//go:build test

package domain

import (
	"testing"
)

func assertLastTrickLog(t *testing.T, entry *ActionLogEntry, code string, wantLastBonus string) {
	t.Helper()
	if entry == nil {
		t.Fatal("missing trick-win action log")
	}
	if entry.DetailCode != code {
		t.Fatalf("detail code = %q, want %q", entry.DetailCode, code)
	}
	if _, ok := entry.DetailParams["bonus"]; ok {
		t.Fatalf("detail params contain obsolete bonus: %#v", entry.DetailParams)
	}
	if got := entry.DetailParams["lastBonus"]; got != wantLastBonus {
		t.Fatalf("lastBonus = %q, want %q; params = %#v", got, wantLastBonus, entry.DetailParams)
	}
}

func lastTrickLog(t *testing.T, entries []*ActionLogEntry) *ActionLogEntry {
	t.Helper()
	if len(entries) == 0 {
		t.Fatal("missing action log")
	}
	return entries[len(entries)-1]
}

func TestLastTrickActionLogsUseSeparateDetailCodes(t *testing.T) {
	t.Run("calabresella", func(t *testing.T) {
		g := NewDefaultCalabresella()
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}}
		g.SetPhase(CalabresellaPhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "calabresella.log.trickWin", "")
		g.SetPhase(CalabresellaPhaseTrickEnd)
		g.SetTrickNumber(CalabresellaTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "calabresella.log.trickWinLast", "")
	})

	t.Run("madrasso", func(t *testing.T) {
		g := NewDefaultMadrasso()
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}, {PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)}}
		g.SetPhase(MadrassoPhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "madrasso.log.trickWin", "")
		g.SetPhase(MadrassoPhaseTrickEnd)
		g.SetTrickNumber(MadrassoTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "madrasso.log.trickWinLast", "")
	})

	t.Run("trappola", func(t *testing.T) {
		g := NewDefaultTrappola()
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}, {PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)}}
		g.SetPhase(TrappolaPhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "trappola.log.trickWin", "")
		g.SetPhase(TrappolaPhaseTrickEnd)
		g.SetTrickNumber(TrappolaTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "trappola.log.trickWinLast", "")
	})

	t.Run("tressette", func(t *testing.T) {
		g := NewDefaultTressette()
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}, {PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)}}
		g.SetPhase(TressettePhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "tressette.log.trickWin", "")
		g.SetPhase(TressettePhaseTrickEnd)
		g.SetTrickNumber(TressetteTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "tressette.log.trickWinLast", "")
	})

	t.Run("klaverjas", func(t *testing.T) {
		g := NewDefaultKlaverjas()
		g.SetTrumpSuit(CardDesignDiamond)
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}, {PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)}}
		g.SetPhase(KlaverjasPhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "klaverjas.log.trickWin", "")
		g.SetPhase(KlaverjasPhaseTrickEnd)
		g.SetTrickNumber(KlaverjasTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "klaverjas.log.trickWinLast", "10")
	})

	t.Run("tute", func(t *testing.T) {
		g := NewDefaultTute()
		g.SetTrumpSuit(CardDesignDiamond)
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}, {PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)}}
		g.SetPhase(TutePhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "tute.log.trickWin", "")
		g.SetPhase(TutePhaseTrickEnd)
		g.SetTrickNumber(TuteTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "tute.log.trickWinLast", "10")
	})

	t.Run("twentynine", func(t *testing.T) {
		g := NewDefaultTwentyNine()
		g.SetTrumpSuit(CardDesignDiamond)
		trick := []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)}, {PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)}, {PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)}, {PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)}}
		g.SetPhase(TwentyNinePhaseTrickEnd)
		g.SetTrickNumber(1)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "twentynine.log.trickWin", "")
		g.SetPhase(TwentyNinePhaseTrickEnd)
		g.SetTrickNumber(TwentyNineTrickCount)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		assertLastTrickLog(t, lastTrickLog(t, g.GetActionLog()), "twentynine.log.trickWinLast", "1")
	})
}
