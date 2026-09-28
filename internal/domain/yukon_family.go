//go:build !js || !wasm || solo || extra4

package domain

type yukonFamilyHint struct {
	fromCol, cardIndex int
	toZone             string
	toCol              int
}

func yukonFamilyGetHint[T any](tableau [][]T, foundationCount int, cardOf func(T) *Card, faceUp func(T) bool, canPlaceFoundation func(*Card, int) bool, canPlaceTableau func(*Card, int) bool) *yukonFamilyHint {
	for col := range tableau {
		if len(tableau[col]) == 0 {
			continue
		}
		i := len(tableau[col]) - 1
		card := cardOf(tableau[col][i])
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < foundationCount && canPlaceFoundation(card, fIdx) {
			return &yukonFamilyHint{fromCol: col, cardIndex: i, toZone: "foundation", toCol: fIdx}
		}
	}
	for fromCol, cards := range tableau {
		if len(cards) == 0 {
			continue
		}
		firstFaceUp := -1
		for i, tc := range cards {
			if faceUp(tc) {
				firstFaceUp = i
				break
			}
		}
		if firstFaceUp <= 0 {
			continue
		}
		card := cardOf(cards[firstFaceUp])
		for toCol := range tableau {
			if toCol != fromCol && canPlaceTableau(card, toCol) {
				return &yukonFamilyHint{fromCol: fromCol, cardIndex: firstFaceUp, toZone: "tableau", toCol: toCol}
			}
		}
	}
	for fromCol, cards := range tableau {
		for i, tc := range cards {
			if !faceUp(tc) {
				continue
			}
			for toCol := range tableau {
				if toCol != fromCol && canPlaceTableau(cardOf(tc), toCol) {
					return &yukonFamilyHint{fromCol: fromCol, cardIndex: i, toZone: "tableau", toCol: toCol}
				}
			}
		}
	}
	return nil
}

func yukonFamilyAllFaceUp[T any](tableau [][]T, faceUp func(T) bool) bool {
	for _, column := range tableau {
		for _, card := range column {
			if !faceUp(card) {
				return false
			}
		}
	}
	return true
}

func yukonFamilyCheckGameClear(foundation [][]*Card, cardsPerSuit int, setCleared func()) {
	for _, pile := range foundation {
		if len(pile) != cardsPerSuit {
			return
		}
	}
	setCleared()
}
