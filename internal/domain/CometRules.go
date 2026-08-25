//go:build !js || !wasm || solo

package domain

// コメットの規則。
//
// **#5459 は連なりの作り方を取り違えている。** 「同スート連続した数字で繋げ」
// 「各スートの A から始め」と書いてあるが、コメットは **スートを無視して
// ランクだけで**繋ぐ。スートを見るのは派生の Commit のほうで、この 2 つは
// 出典が明確に分けている ── そして **スートを見ないことこそがコメットを
// Michigan (Newmarket 系) と分ける唯一の点**なので、そこを取り違えると
// 追加する意味がなくなる。
//
// 開始も A 固定ではない。**親の左隣が好きな札を出し**、そこから K まで
// 昇っていく。
//
// 札は 52 枚から **8♦ を抜いた 51 枚**、**9♦ が「コメット」でワイルド**。
// 抜いた 8♦ の位置と、配り切れずに伏せた「死に手」に入った札は、そこで
// 連なりが必ず止まる ── これが stops 系の名前の由来。
const (
	// CometRemovedValue は抜く札のランク (8)。
	CometRemovedValue = 8
	// CometRemovedDesign は抜く札のスート (♦)。
	CometRemovedDesign = CardDesignDiamond
	// CometWildValue はコメットのランク (9)。
	CometWildValue = 9
	// CometWildDesign はコメットのスート (♦)。
	CometWildDesign = CardDesignDiamond
	// CometMinRank は連なりの下限 (A)。
	CometMinRank = 1
	// CometMaxRank は連なりの上限 (K)。**K は必ず止まる。**
	CometMaxRank = 13
)

// IsCometWild は 9♦ (コメット) かどうかを返す。
func IsCometWild(c *Card) bool {
	return c != nil && c.GetValue() == CometWildValue && c.GetDesign() == CometWildDesign
}

// IsCometRemoved は 8♦ (抜く札) かどうかを返す。
func IsCometRemoved(c *Card) bool {
	return c != nil && c.GetValue() == CometRemovedValue && c.GetDesign() == CometRemovedDesign
}

// NewCometDeck は 8♦ を抜いた 51 枚を返す。
func NewCometDeck() []*Card {
	src := NewTrumpCards(0)
	src.Shuffle()
	out := make([]*Card, 0, 51)
	for {
		c := src.DrawCard()
		if c == nil {
			break
		}
		// **8♦ は抜く。** 抜いた札の位置で連なりが必ず止まる。
		if IsCometRemoved(c) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// CanPlayComet は need のランクが要るときに card を出せるかを返す。
//
// need が 0 なら新しい連なりの先頭なので何でも出せる。**スートは見ない。**
// コメット (9♦) はどのランクの代わりにもなる。
func CanPlayComet(card *Card, need int) bool {
	if card == nil {
		return false
	}
	if need <= 0 {
		return true
	}
	if IsCometWild(card) {
		return true
	}
	return card.GetValue() == need
}

// CometStopsSequence は card を出したあと連なりが切れるかを返す。
//
// **K とコメットは止まる。** K は上限なので次が無く、コメットは代役なので
// 次のランクが決まらない ── どちらも出した本人が新しい連なりを始める。
func CometStopsSequence(card *Card) bool {
	if card == nil {
		return true
	}
	return IsCometWild(card) || card.GetValue() >= CometMaxRank
}

// CometPlayableIdxs は need のランクが要るときに手札のどれを出せるかを返す。
func CometPlayableIdxs(hand []*Card, need int) []int {
	out := make([]int, 0, len(hand))
	for i, c := range hand {
		if CanPlayComet(c, need) {
			out = append(out, i)
		}
	}
	return out
}

// CometCardPoints は手札に残した 1 枚の失点を返す。
//
// **勝者は相手の残り枚数で稼ぐ。** 出典は 1 枚 1 点で揃っているので、額面や
// 絵札 10 点といった別方式は採らない。
func CometCardPoints(_ *Card) int { return 1 }
