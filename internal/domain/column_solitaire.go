//go:build !js || !wasm || solo

package domain

import "errors"

// ColumnTableauCard is the common representation used by column solitaire games.
type ColumnTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// ColumnSolitaireHint is the common move hint for column solitaire games.
type ColumnSolitaireHint struct {
	FromCol   int
	CardIndex int
	ToZone    string
	ToCol     int
}

type columnSolitaireRules struct {
	canStack func(card *Card, col []*ColumnTableauCard) bool
}

func columnMoveTableauToTableau(tableau [][]*ColumnTableauCard, fromCol, cardIndex, toCol int, rules columnSolitaireRules) (*ColumnTableauCard, error) {
	if fromCol < 0 || fromCol >= len(tableau) {
		return nil, errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= len(tableau) {
		return nil, errors.New("invalid to column")
	}
	if fromCol == toCol {
		return nil, errors.New("from and to columns are the same")
	}
	from := tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(from) - 1
	}
	if cardIndex < 0 || cardIndex >= len(from) {
		return nil, errors.New("invalid card index")
	}
	if cardIndex != len(from)-1 {
		return nil, errors.New("only the top card can be moved")
	}
	tc := from[cardIndex]
	if !rules.canStack(tc.Card, tableau[toCol]) {
		return nil, errors.New("cannot place card on tableau")
	}
	tableau[toCol] = append(tableau[toCol], tc)
	tableau[fromCol] = from[:cardIndex]
	return tc, nil
}

func columnAllFaceUp(tableau [][]*ColumnTableauCard) bool {
	for _, col := range tableau {
		for _, card := range col {
			if !card.FaceUp {
				return false
			}
		}
	}
	return true
}

func columnFindFoundation(foundation [][]*Card, card *Card) int {
	for i, pile := range foundation {
		if canPlaceOnFoundationPile(pile, card) {
			return i
		}
	}
	return -1
}
