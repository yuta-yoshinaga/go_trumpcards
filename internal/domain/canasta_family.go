//go:build !js || !wasm || extra7

package domain

// CanastaMinMeld はキャナスタ系の初回メルド最低点を累積得点から返す。
// Web の canastaMinMeld と同じ帯 (マイナス 15 / 1500 未満 50 / 3000 未満 90 /
// それ以上 120)。
func CanastaMinMeld(cumulativeScore int) int {
	switch {
	case cumulativeScore < 0:
		return 15
	case cumulativeScore < 1500:
		return 50
	case cumulativeScore < 3000:
		return 90
	default:
		return 120
	}
}

// CanastaFamilyCardValue カードの点数を返す
func CanastaFamilyCardValue(card *Card) int {
	if card.GetDesign() == CardDesignJoker {
		return 50
	}
	v := card.GetValue()
	if v == 2 {
		return 20
	}
	if v == 1 { // Ace
		return 20
	}
	if v == 3 && (card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover) {
		return 5 // 黒3
	}
	if v >= 8 {
		return 10
	}
	return 5
}
