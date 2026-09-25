//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func familyCard(suit, rank int, faceUp bool) *KlondikeTableauCard {
	return &KlondikeTableauCard{Card: NewCard(suit, rank, false), FaceUp: faceUp}
}

func familyYukon(tableau [YukonTableauCnt][]*KlondikeTableauCard, foundation [YukonFoundationCnt][]*Card) *Yukon {
	y := NewYukon(&TrumpCards{})
	y.phase = YukonPhasePlaying
	y.SetTableau(tableau)
	y.SetFoundation(foundation)
	return y
}

func familyRussianCard(suit, rank int, faceUp bool) *KlondikeTableauCard {
	return &KlondikeTableauCard{Card: NewCard(suit, rank, false), FaceUp: faceUp}
}

func familyRussian(tableau [RussianSolitaireTableauCnt][]*KlondikeTableauCard, foundation [RussianSolitaireFoundationCnt][]*Card) *RussianSolitaire {
	r := NewRussianSolitaire(&TrumpCards{})
	r.phase = RussianSolitairePhasePlaying
	r.SetTableau(tableau)
	r.SetFoundation(foundation)
	return r
}

func TestRussianSolitaireFamilyCharacterization(t *testing.T) {
	cases := []struct {
		name          string
		make          func() *RussianSolitaire
		move          func(*RussianSolitaire) error
		hint          string
		moveError     string
		autoError     string
		movePhase     int
		moveCount     int
		moveStalemate bool
		autoPhase     int
		autoMoveCount int
		autoStalemate bool
		gameEnded     bool
	}{
		{"multi_card_tableau_move", func() *RussianSolitaire {
			var tab [RussianSolitaireTableauCnt][]*KlondikeTableauCard
			tab[0] = []*KlondikeTableauCard{familyRussianCard(CardDesignSpade, 5, true), familyRussianCard(CardDesignHeart, 10, true), familyRussianCard(CardDesignClover, 2, true)}
			tab[1] = []*KlondikeTableauCard{familyRussianCard(CardDesignSpade, 6, true)}
			return familyRussian(tab, [RussianSolitaireFoundationCnt][]*Card{})
		}, func(r *RussianSolitaire) error { return r.MoveTableauToTableau(0, 0, 1) }, `{"FromCol":0,"CardIndex":0,"ToZone":"tableau","ToCol":1}`, "", "", 0, 1, true, 0, 1, true, false},
		{"tableau_to_foundation", func() *RussianSolitaire {
			var tab [RussianSolitaireTableauCnt][]*KlondikeTableauCard
			var fd [RussianSolitaireFoundationCnt][]*Card
			fd[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
			tab[0] = []*KlondikeTableauCard{familyRussianCard(CardDesignSpade, 2, true)}
			return familyRussian(tab, fd)
		}, func(r *RussianSolitaire) error { return r.MoveTableauToFoundation(0) }, `{"FromCol":0,"CardIndex":0,"ToZone":"foundation","ToCol":0}`, "", "", 0, 1, true, 0, 1, true, false},
		{"autocomplete_several", func() *RussianSolitaire {
			var tab [RussianSolitaireTableauCnt][]*KlondikeTableauCard
			var fd [RussianSolitaireFoundationCnt][]*Card
			for suit := 1; suit <= 4; suit++ {
				for rank := 1; rank <= 10; rank++ {
					fd[suit-1] = append(fd[suit-1], NewCard(suit, rank, false))
				}
				tab[suit-1] = []*KlondikeTableauCard{familyRussianCard(suit, 13, true), familyRussianCard(suit, 12, true), familyRussianCard(suit, 11, true)}
			}
			return familyRussian(tab, fd)
		}, func(r *RussianSolitaire) error { return r.AutoComplete() }, `{"FromCol":0,"CardIndex":2,"ToZone":"foundation","ToCol":0}`, "", "game is not in playing phase", 1, 12, false, 1, 12, false, true},
		{"stalemate", func() *RussianSolitaire {
			var tab [RussianSolitaireTableauCnt][]*KlondikeTableauCard
			tab[0] = []*KlondikeTableauCard{familyRussianCard(CardDesignSpade, 5, true)}
			tab[1] = []*KlondikeTableauCard{familyRussianCard(CardDesignSpade, 6, true)}
			return familyRussian(tab, [RussianSolitaireFoundationCnt][]*Card{})
		}, func(r *RussianSolitaire) error { return r.MoveTableauToTableau(0, 0, 1) }, `{"FromCol":0,"CardIndex":0,"ToZone":"tableau","ToCol":1}`, "", "", 0, 1, true, 0, 1, true, false},
		{"one_move_clear", func() *RussianSolitaire {
			var tab [RussianSolitaireTableauCnt][]*KlondikeTableauCard
			var fd [RussianSolitaireFoundationCnt][]*Card
			for suit := 1; suit <= 4; suit++ {
				maxRank := 13
				if suit == CardDesignSpade {
					maxRank = 12
				}
				for rank := 1; rank <= maxRank; rank++ {
					fd[suit-1] = append(fd[suit-1], NewCard(suit, rank, false))
				}
			}
			tab[0] = []*KlondikeTableauCard{familyRussianCard(CardDesignSpade, 13, true)}
			return familyRussian(tab, fd)
		}, func(r *RussianSolitaire) error { return r.MoveTableauToFoundation(0) }, `{"FromCol":0,"CardIndex":0,"ToZone":"foundation","ToCol":0}`, "", "game is not in playing phase", 1, 1, false, 1, 1, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.make()
			hint, hintErr := json.Marshal(r.GetHint())
			if hintErr != nil || string(hint) != tc.hint {
				t.Fatalf("GetHint() = %s, %v; want %s", hint, hintErr, tc.hint)
			}
			moveErr := tc.move(r)
			gotMoveError := ""
			if moveErr != nil {
				gotMoveError = moveErr.Error()
			}
			if gotMoveError != tc.moveError {
				t.Fatalf("move error = %q, want %q", gotMoveError, tc.moveError)
			}
			afterMove, marshalMoveErr := r.MarshalJSON()
			assertRussianSolitaireCharacterizationJSON(t, afterMove, marshalMoveErr, tc.movePhase, tc.moveCount, tc.moveStalemate)
			autoErr := r.AutoComplete()
			gotAutoError := ""
			if autoErr != nil {
				gotAutoError = autoErr.Error()
			}
			if gotAutoError != tc.autoError {
				t.Fatalf("AutoComplete() error = %q, want %q", gotAutoError, tc.autoError)
			}
			afterAuto, marshalAutoErr := r.MarshalJSON()
			assertRussianSolitaireCharacterizationJSON(t, afterAuto, marshalAutoErr, tc.autoPhase, tc.autoMoveCount, tc.autoStalemate)
			if r.IsStalemate() != tc.autoStalemate || r.GetGameEndFlag() != tc.gameEnded {
				t.Fatalf("flags after AutoComplete: stalemate=%v end=%v", r.IsStalemate(), r.GetGameEndFlag())
			}
		})
	}
}

func assertRussianSolitaireCharacterizationJSON(t *testing.T, data []byte, marshalErr error, phase, moveCount int, stalemate bool) {
	t.Helper()
	if marshalErr != nil {
		t.Fatalf("MarshalJSON() error = %v", marshalErr)
	}
	var state struct {
		Phase       int  `json:"ps"`
		MoveCount   int  `json:"mc"`
		IsStalemate bool `json:"sl"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("MarshalJSON() returned invalid JSON: %v", err)
	}
	if state.Phase != phase || state.MoveCount != moveCount || state.IsStalemate != stalemate {
		t.Fatalf("MarshalJSON state = {phase:%d moveCount:%d stalemate:%v}, want {%d %d %v}", state.Phase, state.MoveCount, state.IsStalemate, phase, moveCount, stalemate)
	}
}

func TestYukonFamilyCharacterization(t *testing.T) {
	cases := []struct {
		name string
		make func() *Yukon
		move func(*Yukon) error
	}{
		{"multi_card_tableau_move", func() *Yukon {
			var tab [YukonTableauCnt][]*KlondikeTableauCard
			tab[0] = []*KlondikeTableauCard{familyCard(CardDesignHeart, 5, true), familyCard(CardDesignDiamond, 11, true), familyCard(CardDesignSpade, 2, true)}
			tab[1] = []*KlondikeTableauCard{familyCard(CardDesignSpade, 6, true)}
			return familyYukon(tab, [YukonFoundationCnt][]*Card{})
		}, func(y *Yukon) error { return y.MoveTableauToTableau(0, 0, 1) }},
		{"tableau_to_foundation", func() *Yukon {
			var tab [YukonTableauCnt][]*KlondikeTableauCard
			var fd [YukonFoundationCnt][]*Card
			fd[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
			tab[0] = []*KlondikeTableauCard{familyCard(CardDesignSpade, 2, true)}
			return familyYukon(tab, fd)
		}, func(y *Yukon) error { return y.MoveTableauToFoundation(0) }},
		{"autocomplete_several", func() *Yukon {
			var tab [YukonTableauCnt][]*KlondikeTableauCard
			var fd [YukonFoundationCnt][]*Card
			for suit := 1; suit <= 4; suit++ {
				for rank := 1; rank <= 10; rank++ {
					fd[suit-1] = append(fd[suit-1], NewCard(suit, rank, false))
				}
				tab[suit-1] = []*KlondikeTableauCard{familyCard(suit, 11, true), familyCard(suit, 12, true)}
			}
			return familyYukon(tab, fd)
		}, func(y *Yukon) error { return y.AutoComplete() }},
		{"stalemate", func() *Yukon {
			var tab [YukonTableauCnt][]*KlondikeTableauCard
			tab[0] = []*KlondikeTableauCard{familyCard(CardDesignSpade, 5, true)}
			tab[1] = []*KlondikeTableauCard{familyCard(CardDesignClover, 5, true)}
			return familyYukon(tab, [YukonFoundationCnt][]*Card{})
		}, func(y *Yukon) error { return y.MoveTableauToTableau(0, 0, 1) }},
		{"one_move_clear", func() *Yukon {
			var tab [YukonTableauCnt][]*KlondikeTableauCard
			var fd [YukonFoundationCnt][]*Card
			for suit := 1; suit <= 4; suit++ {
				for rank := 1; rank <= 12; rank++ {
					fd[suit-1] = append(fd[suit-1], NewCard(suit, rank, false))
				}
			}
			tab[0] = []*KlondikeTableauCard{familyCard(CardDesignSpade, 13, true)}
			return familyYukon(tab, fd)
		}, func(y *Yukon) error { return y.MoveTableauToFoundation(0) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			y := tc.make()
			hint, _ := json.Marshal(y.GetHint())
			moveErr := tc.move(y)
			afterMove, _ := y.MarshalJSON()
			autoErr := y.AutoComplete()
			afterAuto, _ := y.MarshalJSON()
			t.Logf("hint=%s moveErr=%v move=%s autoErr=%v auto=%s stalemate=%v end=%v", hint, moveErr, afterMove, autoErr, afterAuto, y.IsStalemate(), y.GetGameEndFlag())
		})
	}
}
