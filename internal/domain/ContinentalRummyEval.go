//go:build !js || !wasm || extra2

package domain

import "sort"

// コンチネンタル・ラミーの卓と札。
const (
	// ContinentalRummyPlayerCnt は席数 (人間 1 + CPU 3)。
	//
	// 原典は 2〜12 人。**デッキ枚数は人数で決まる**が、その刻みは「人数 − 1」
	// ではなく 2〜5 人 = 2 組、6〜8 人 = 3 組、9 人以上 = 4 組。4 人卓は
	// 最初の刻みに入るので 2 組 + ジョーカー 2 枚 = 106 枚になる。
	ContinentalRummyPlayerCnt = 4
	// ContinentalRummyDeckCnt は 2〜5 人卓で使うデッキ数。
	ContinentalRummyDeckCnt = 2
	// ContinentalRummyJokerCnt は 2〜5 人卓で混ぜるジョーカーの枚数。
	ContinentalRummyJokerCnt = 2
	// ContinentalRummyHandSize は 1 人あたりの配布枚数。3 枚ずつ 5 回配る。
	ContinentalRummyHandSize = 15
	// ContinentalRummyDealChunk は 1 度に配る枚数。
	ContinentalRummyDealChunk = 3
	// ContinentalRummyMinRun は最短のシーケンス。
	ContinentalRummyMinRun = 3
	// ContinentalRummyMaxRun は最長のシーケンス。
	ContinentalRummyMaxRun = 5
)

// continentalRummyLayouts は上がりに使える **唯一の 3 通り**。
//
// **「15 枚を 3〜5 枚のランに分けられれば上がり」ではない。** 15 を 3〜5 の
// 和に分ける方法は 5+5+5 / 5+4+3+3 / 4+4+4+3 / 3×5 の 4 通りあるが、
// 原典が認めるのは下の 3 通りだけで、**5 枚 3 組 (5+5+5) は上がりにならない**。
// #5464 はこの制約そのものを落としている。
var continentalRummyLayouts = [][]int{
	{3, 3, 3, 3, 3},
	{4, 4, 4, 3},
	{5, 4, 3, 3},
}

// ContinentalRummyLayouts は上がりに使える組み合わせを複製して返す。
// 表示側が「何と何が要るのか」を毎回自分で書き写さずに済むように公開する。
func ContinentalRummyLayouts() [][]int {
	out := make([][]int, len(continentalRummyLayouts))
	for i, l := range continentalRummyLayouts {
		out[i] = append([]int(nil), l...)
	}
	return out
}

// IsContinentalRummyLayout は枚数の並びが認められた上がりの形かを返す。
// 並び順は問わない (4+3+4+4 も {4,4,4,3} と同じ形)。
func IsContinentalRummyLayout(sizes []int) bool {
	got := append([]int(nil), sizes...)
	sort.Sort(sort.Reverse(sort.IntSlice(got)))
	for _, want := range continentalRummyLayouts {
		if len(want) != len(got) {
			continue
		}
		ok := true
		for i := range want {
			if want[i] != got[i] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// IsContinentalRummyJoker はその札がワイルドかを返す。
//
// **ワイルドはジョーカーだけ。** 2 をワイルドに含める流儀もあるが、出典が
// 割れているうえ、2 を潰すと A-2-3 の下端が使えなくなって形が変わる。
func IsContinentalRummyJoker(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignJoker
}

// IsContinentalRummyRun は札の並びが 1 つのシーケンスとして成立するかを返す。
//
// 条件は 3〜5 枚・**同じスート**・連番。ジョーカーは任意の 1 枚の代わりに
// なる。A は上端でも下端でもよいが **両方は兼ねない** (K-A-2 は繋がらない)。
func IsContinentalRummyRun(cards []*Card) bool {
	if len(cards) < ContinentalRummyMinRun || len(cards) > ContinentalRummyMaxRun {
		return false
	}
	jokers := 0
	vals := make([]int, 0, len(cards))
	suit := -1
	for _, c := range cards {
		if c == nil {
			return false
		}
		if IsContinentalRummyJoker(c) {
			jokers++
			continue
		}
		if suit == -1 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return false
		}
		vals = append(vals, c.GetValue())
	}
	// 全部ジョーカーの「ラン」は認めない。何のスートの何の並びか決まらない。
	if len(vals) == 0 {
		return false
	}
	if continentalRunFits(vals, jokers, false) {
		return true
	}
	// A を 14 として上端に置き直した並びも試す。**下端との併用はしない。**
	return continentalRunFits(vals, jokers, true)
}

// continentalRunFits は素札の値がジョーカー jokers 枚で 1 本に繋がるかを返す。
// aceHigh が真なら A (1) を 14 として扱う。
func continentalRunFits(vals []int, jokers int, aceHigh bool) bool {
	v := append([]int(nil), vals...)
	if aceHigh {
		for i := range v {
			if v[i] == 1 {
				v[i] = 14
			}
		}
	}
	sort.Ints(v)
	for i := 1; i < len(v); i++ {
		if v[i] == v[i-1] {
			return false // 同じ札 2 枚は 1 本の並びにならない
		}
	}
	// 端から端までの長さが札数を超えるぶんだけ穴があり、それをジョーカーで埋める。
	span := v[len(v)-1] - v[0] + 1
	total := len(v) + jokers
	if span > total {
		return false
	}
	return span+jokers >= total // 余ったジョーカーは端に足せる
}

// ContinentalRummyCardValue は残り札を数えるときの 1 枚の点数。
//
// 精算は勝った側が集める方式なので通常は使わないが、CPU が手札の重さを測る
// のに要る。ジョーカー 50 / A 11 / 絵札と 10 は 10 / それ以外は数字どおり。
func ContinentalRummyCardValue(c *Card) int {
	switch {
	case c == nil:
		return 0
	case IsContinentalRummyJoker(c):
		return 50
	case c.GetValue() == 1:
		return 11
	case c.GetValue() >= 10:
		return 10
	default:
		return c.GetValue()
	}
}
