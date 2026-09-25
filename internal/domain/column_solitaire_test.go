//go:build test && (!js || !wasm || solo)

package domain

import "testing"

func TestColumnMoveTableauToTableau(t *testing.T) {
	card := &Card{value: 5}
	tableau := [][]*ColumnTableauCard{{{Card: card, FaceUp: true}}, {}}
	moved, err := columnMoveTableauToTableau(tableau, 0, -1, 1, columnSolitaireRules{canStack: func(_ *Card, col []*ColumnTableauCard) bool { return len(col) == 0 }})
	if err != nil {
		t.Fatalf("move returned error: %v", err)
	}
	if moved.Card != card || len(tableau[0]) != 0 || len(tableau[1]) != 1 || tableau[1][0] != moved {
		t.Fatal("move did not update the tableau view")
	}
}

func TestColumnFindFoundation(t *testing.T) {
	card := &Card{value: 1}
	if got := columnFindFoundation([][]*Card{{}, {}}, card); got != 0 {
		t.Fatalf("foundation index = %d, want 0", got)
	}
}
