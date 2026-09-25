//go:build test && (!js || !wasm || solo)

package domain

import "testing"

func TestColumnFindFoundation(t *testing.T) {
	card := &Card{value: 1}
	if got := columnFindFoundation([][]*Card{{}, {}}, card); got != 0 {
		t.Fatalf("foundation index = %d, want 0", got)
	}
}

func TestColumnValidateTableauMove(t *testing.T) {
	card := &Card{value: 5}
	movedCard := &ColumnTableauCard{Card: card, FaceUp: true}
	tableau := [][]*ColumnTableauCard{{movedCard}, {}}
	got, index, err := columnValidateTableauMove(tableau, 0, -1, 1, func(_ *Card, toCol int) bool { return toCol == 1 })
	if err != nil || got != movedCard || index != 0 {
		t.Fatalf("validation = (%p, %d, %v), want (%p, 0, nil)", got, index, err, movedCard)
	}
	if len(tableau[0]) != 1 || len(tableau[1]) != 0 {
		t.Fatal("validation mutated tableau")
	}
}

func TestColumnValidateTableauToFoundation(t *testing.T) {
	card := &Card{value: 1}
	tc := &ColumnTableauCard{Card: card, FaceUp: true}
	tableau := [][]*ColumnTableauCard{{tc}}
	got, fIdx, err := columnValidateTableauToFoundation(tableau, 0, func(got *Card) int {
		if got == card {
			return 2
		}
		return -1
	})
	if err != nil || got != tc || fIdx != 2 {
		t.Fatalf("validation = (%p, %d, %v), want (%p, 2, nil)", got, fIdx, err, tc)
	}
}

func TestColumnAutoCompleteMovesUpdatesSliceViews(t *testing.T) {
	first, second := &Card{value: 1}, &Card{value: 2}
	tableau := [][]*ColumnTableauCard{{{Card: first}}, {{Card: second}}}
	foundation := [][]*Card{{}}
	moved := columnAutoCompleteMoves(tableau, foundation, func(card *Card) int {
		if card == first || card == second {
			return 0
		}
		return -1
	})
	if moved != 2 || len(tableau[0]) != 0 || len(tableau[1]) != 0 || len(foundation[0]) != 2 {
		t.Fatalf("moves=%d tableau=%v foundation=%v", moved, tableau, foundation)
	}
}
