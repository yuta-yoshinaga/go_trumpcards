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
