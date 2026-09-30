//go:build !js || !wasm || extra7

package domain

// CanastaMinMeld はキャナスタ系の初回メルド最低点を累積得点から返す。
// Canasta / Bolivia / Samba / HandAndFoot で共有する初回メルド最低点の帯。
// Web は応答の minMeld でこの値を受け取る。
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

// CanastaFamilyCardValue は Canasta 系 4 ゲーム共通の札点数を返す (Joker 50 / 2・A 20 / 8〜K 10 / 黒 3 と 4〜7 は 5)。
// Web の canastaFamilyCardValue と golden vectors で一致を検査する。
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
