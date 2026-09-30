//go:build test

package domain

import "sort"

// KlaverjasRankRow は強さ早見表の 1 行。
type KlaverjasRankRow struct {
	// Value はカードの値 (1=A, 11=J, 12=Q, 13=K)。
	Value int
	// Points はそのカードの得点。
	Points int
}

// KlaverjasRankTable は強い順に並べた強さ早見表を返す。isTrump で切り札用と
// 非切り札用を切り替える。
//
// **切り札 (J>9>A>10>K>Q>8>7) と非切り札 (A>10>K>Q>J>9>8>7) で順序も配点も
// 別系統**という、このゲームでいちばん覚えにくいところ。Web は #4757 で 2 つの表を
// 出したが値を TS 側に写しており、CUI には説明そのものが無かった (#5645)。
// 3 つ目の写しを作らないよう、判定と同じ関数から表を組み立てて返す。
func KlaverjasRankTable(isTrump bool) []KlaverjasRankRow {
	// 実際の勝敗判定 (trumpStrength / klaverjasPlainStrength) と配点 (cardPoints)
	// をそのまま使うので、片方だけ直しても表がずれない。
	g := &Klaverjas{trumpSuit: CardDesignSpade}
	design := CardDesignSpade
	if !isTrump {
		design = CardDesignHeart
	}
	values := []int{1, 7, 8, 9, 10, 11, 12, 13}
	rows := make([]KlaverjasRankRow, 0, len(values))
	for _, v := range values {
		rows = append(rows, KlaverjasRankRow{Value: v, Points: g.cardPoints(NewCard(design, v, false))})
	}
	strength := func(v int) int {
		if isTrump {
			return g.trumpStrength(v)
		}
		return klaverjasPlainStrength(v)
	}
	sort.SliceStable(rows, func(i, j int) bool { return strength(rows[i].Value) > strength(rows[j].Value) })
	return rows
}
