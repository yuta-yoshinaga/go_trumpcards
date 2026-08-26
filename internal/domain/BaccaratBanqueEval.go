//go:build !js || !wasm || extra2

package domain

import "math/rand"

// バカラ・バンクの評価。
//
// **#5462 は自分自身と矛盾している。** 冒頭では「バンクは勝ち続ける限り
// 同じ席に固定される」と正しく書きながら、要件 6 で「負けたら次の参加者に
// バンクが移る」と書いている ── 後者はシュマン・ド・フェールの規則で、
// バカラ・バンクではない。
//
// **バンカーはシューを配り切るまで席を降りない。** 降りるのは、自分から
// 退くか、資金が尽きたときだけ。1 回負けただけでは動かない ── そこが
// 「バンク」と呼ばれる所以で、シュマン・ド・フェールとの唯一にして最大の
// 違いでもある。
//
// 3 枚目を引く規則も、issue の冒頭は「合計 5 の子を除き裁量」と裏返しに
// 書いている。正しくは **0-4 は必ず引き、6-7 は必ず止まり、裁量があるのは
// 5 のときだけ**。実装メモのほうにはこちらが書かれている。

// 卓の形。
const (
	// BaccaratBanqueDeckCount はシューに束ねる組数。**3 組が慣例。**
	BaccaratBanqueDeckCount = 3
	// BaccaratBanqueDeckSize はシューの総枚数。
	BaccaratBanqueDeckSize = 52 * BaccaratBanqueDeckCount
	// BaccaratBanqueNatural はナチュラルの下限 (8 か 9 で即勝負)。
	BaccaratBanqueNatural = 8
	// BaccaratBanqueDiscretionTotal は子に裁量が生じる唯一の合計。
	BaccaratBanqueDiscretionTotal = 5
	// BaccaratBanqueMustDrawMax はこれ以下なら子は必ず引く。
	BaccaratBanqueMustDrawMax = 4
	// BaccaratBanqueMustStandMin はこれ以上なら子は必ず止まる (7 まで)。
	BaccaratBanqueMustStandMin = 6
)

// BaccaratBanqueCardValue は札の点を返す。**10 と絵札は 0、A は 1。**
func BaccaratBanqueCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v >= 10 {
		return 0
	}
	return v
}

// BaccaratBanqueTotal は手の合計を返す。**10 で割った余り。**
func BaccaratBanqueTotal(cards []*Card) int {
	sum := 0
	for _, c := range cards {
		sum += BaccaratBanqueCardValue(c)
	}
	return sum % 10
}

// BaccaratBanqueIsNatural は 2 枚で 8 か 9 かを返す。
//
// **ナチュラルが出たら 3 枚目は無い。** 引く引かないの判断ごと飛ばす。
func BaccaratBanqueIsNatural(cards []*Card) bool {
	if len(cards) != 2 {
		return false
	}
	return BaccaratBanqueTotal(cards) >= BaccaratBanqueNatural
}

// 子の 3 枚目の扱い。
const (
	// BaccaratBanqueDrawMust は必ず引く (合計 0-4)。
	BaccaratBanqueDrawMust = "must"
	// BaccaratBanqueDrawStand は必ず止まる (合計 6-7)。
	BaccaratBanqueDrawStand = "stand"
	// BaccaratBanqueDrawFree は裁量 (合計 5 のときだけ)。
	BaccaratBanqueDrawFree = "free"
	// BaccaratBanqueDrawNatural はナチュラルなので 3 枚目が無い。
	BaccaratBanqueDrawNatural = "natural"
)

// BaccaratBanquePunterRule は子が 3 枚目をどう扱うかを返す。
//
// **裁量があるのは合計 5 のときだけ。** 0-4 は必ず引き、6-7 は必ず止まる ──
// #5462 の冒頭はここを裏返しに書いている。
func BaccaratBanquePunterRule(cards []*Card) string {
	if BaccaratBanqueIsNatural(cards) {
		return BaccaratBanqueDrawNatural
	}
	switch total := BaccaratBanqueTotal(cards); {
	case total <= BaccaratBanqueMustDrawMax:
		return BaccaratBanqueDrawMust
	case total >= BaccaratBanqueMustStandMin:
		return BaccaratBanqueDrawStand
	default:
		return BaccaratBanqueDrawFree
	}
}

// BaccaratBanqueBankerRule はバンカーが 3 枚目をどう扱うかを返す。
//
// **バンカーはどの合計でも自由。** プント・バンコのような固定表は無く、
// 両方の子の結果を見てから決められる ── ナチュラルのときだけ引けない。
func BaccaratBanqueBankerRule(cards []*Card) string {
	if BaccaratBanqueIsNatural(cards) {
		return BaccaratBanqueDrawNatural
	}
	return BaccaratBanqueDrawFree
}

// 1 つの子との勝敗。
const (
	// BaccaratBanqueOutcomeBankerWin はバンカーの勝ち。
	BaccaratBanqueOutcomeBankerWin = "bankerWin"
	// BaccaratBanqueOutcomePunterWin は子の勝ち。
	BaccaratBanqueOutcomePunterWin = "punterWin"
	// BaccaratBanqueOutcomeTie は引き分け (賭けは戻る)。
	BaccaratBanqueOutcomeTie = "tie"
)

// BaccaratBanqueCompare は 1 つの子とバンカーの勝敗を返す。
//
// **左右は別勘定。** それぞれ自分の札だけでバンカーと比べ、片方が勝って
// もう片方が負けることがある ── そこがこの形式の要。
func BaccaratBanqueCompare(bankerTotal, punterTotal int) string {
	switch {
	case bankerTotal > punterTotal:
		return BaccaratBanqueOutcomeBankerWin
	case punterTotal > bankerTotal:
		return BaccaratBanqueOutcomePunterWin
	default:
		return BaccaratBanqueOutcomeTie
	}
}

// NewBaccaratBanqueShoe は 3 組を混ぜたシューを返す。
func NewBaccaratBanqueShoe() []*Card {
	out := make([]*Card, 0, BaccaratBanqueDeckSize)
	for i := 0; i < BaccaratBanqueDeckCount; i++ {
		src := NewTrumpCards(0)
		src.Shuffle()
		for {
			c := src.DrawCard()
			if c == nil {
				break
			}
			out = append(out, c)
		}
	}
	// **束ねてからもう一度混ぜる。** 組ごとに並んだままだと、同じ札が
	// 決まった間隔で出てくる ── 3 組を「inter-shuffled」にするのが規則。
	for i := len(out) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) //nolint:gosec // ゲームの配りに暗号強度は要らない
		out[i], out[j] = out[j], out[i]
	}
	return out
}
