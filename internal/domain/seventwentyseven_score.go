//go:build !js || !wasm || extra4

package domain

import "strconv"

// Seven Twenty-Seven の点数計算。
//
// **すべて ×2 の整数で持つ。** 絵札が 0.5 点というのがこのゲーム固有の値で、
// float64 で持つと比較のたびに誤差を気にすることになる。倍にすれば絵札が 1、
// 数字が 2n、目標が 14 / 54 になり、超過判定も同点判定も整数のまま済む
// （Call Break が int×10 でスコアを持っているのと同じ判断）。
// 表示するときだけ 2 で割る。

const (
	// SevenTwentySevenScoreScale は内部表現の倍率。絵札 0.5 点を整数で扱うため。
	SevenTwentySevenScoreScale = 2
	// SevenTwentySevenLowTarget は低い側の目標 7 点（内部 14）。
	SevenTwentySevenLowTarget = 7 * SevenTwentySevenScoreScale
	// SevenTwentySevenHighTarget は高い側の目標 27 点（内部 54）。
	SevenTwentySevenHighTarget = 27 * SevenTwentySevenScoreScale
	// sevenTwentySevenAceLow はエースを 1 点として数えたときの内部値。
	sevenTwentySevenAceLow = 1 * SevenTwentySevenScoreScale
	// sevenTwentySevenAceHigh はエースを 11 点として数えたときの内部値。
	sevenTwentySevenAceHigh = 11 * SevenTwentySevenScoreScale
	// sevenTwentySevenFaceValue は絵札 (J/Q/K) の内部値。実点 0.5。
	sevenTwentySevenFaceValue = 1
)

// sevenTwentySevenCardValue は 1 枚のカードの内部点を返す。
// エースは 0 を返す —— 1 か 11 かは手全体で決めるので、単独では定まらない。
func sevenTwentySevenCardValue(c *Card) int {
	switch v := c.GetValue(); {
	case v == 1:
		return 0 // エースは呼び出し側で数える
	case v >= 11: // J / Q / K
		return sevenTwentySevenFaceValue
	default:
		return v * SevenTwentySevenScoreScale
	}
}

// sevenTwentySevenTotals は手札から取りうる合計（内部値）を**昇順・重複なし**で返す。
//
// エースは 1 枚ごとに 1 か 11 を選べるので、n 枚のエースに対して「11 として
// 数える枚数」を 0..n で回せば足りる。どのエースを高くするかは合計に影響しない。
func sevenTwentySevenTotals(cards []*Card) []int {
	base, aces := 0, 0
	for _, c := range cards {
		if c == nil {
			continue
		}
		if c.GetValue() == 1 {
			aces++
			continue
		}
		base += sevenTwentySevenCardValue(c)
	}
	totals := make([]int, 0, aces+1)
	for high := 0; high <= aces; high++ {
		totals = append(totals, base+high*sevenTwentySevenAceHigh+(aces-high)*sevenTwentySevenAceLow)
	}
	return totals
}

// SevenTwentySevenBestFor は target を超えない範囲での最良の合計（内部値）を返す。
// どの数え方でも超えてしまうなら ok=false —— その側の勝負から失格。
func SevenTwentySevenBestFor(cards []*Card, target int) (int, bool) {
	best, ok := 0, false
	for _, total := range sevenTwentySevenTotals(cards) {
		if total > target {
			continue
		}
		if !ok || total > best {
			best, ok = total, true
		}
	}
	return best, ok
}

// SevenTwentySevenFormat は内部値を表示用の文字列にする（14 → "7", 13 → "6.5"）。
func SevenTwentySevenFormat(scaled int) string {
	whole := scaled / SevenTwentySevenScoreScale
	if scaled%SevenTwentySevenScoreScale == 0 {
		return strconv.Itoa(whole)
	}
	return strconv.Itoa(whole) + ".5"
}
