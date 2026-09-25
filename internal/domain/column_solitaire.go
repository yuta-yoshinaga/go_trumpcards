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

func columnGetHint(
	isPlaying bool,
	tableau [][]*ColumnTableauCard,
	canPlaceOnTableau func(*Card, int) bool,
	findFoundation func(*Card) int,
) *ColumnSolitaireHint {
	if !isPlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range len(tableau) {
		if len(tableau[col]) == 0 {
			continue
		}
		tc := tableau[col][len(tableau[col])-1]
		fIdx := findFoundation(tc.Card)
		if fIdx >= 0 {
			return &ColumnSolitaireHint{
				FromCol:   col,
				CardIndex: len(tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := range len(tableau) {
		fromCards := tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range len(tableau) {
			if toCol == fromCol {
				continue
			}
			if canPlaceOnTableau(card, toCol) {
				return &ColumnSolitaireHint{
					FromCol:   fromCol,
					CardIndex: len(fromCards) - 1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	return nil
}

func columnCheckGameClear(foundation [][]*Card) bool {
	for _, pile := range foundation {
		if len(pile) != CardValueMax {
			return false
		}
	}
	return true
}

func columnCheckStalemate(
	isPlaying bool,
	tableau [][]*ColumnTableauCard,
	canPlaceOnTableau func(*Card, int) bool,
	findFoundation func(*Card) int,
) bool {
	if !isPlaying {
		return false
	}
	hint := columnGetHint(isPlaying, tableau, canPlaceOnTableau, findFoundation)
	return hint == nil
}
