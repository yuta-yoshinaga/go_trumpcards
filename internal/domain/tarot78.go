//go:build !js || !wasm || extra4

package domain

// 78 枚タロットデッキ (仏式スート記号) の共有部品。
//
// **同じデッキで遊ぶゲームが 2 つある。** ピエモンテのタロッコは 3 人版 (Scarto)
// と 4 人版 (PiedmonteseTarot) で配り方も精算も違うが、**デッキの形とトリックの
// 出し方の規則は 1 文字も違わない** —— スートに従い、無ければ切り札を出し、
// 切り札が出ていれば上位で応じ、エクスキューズだけがその義務を免れる。
// ここを 2 か所に書くと、片方だけ直した規則が生まれる。

// Tarot78 デッキの構造定数。
const (
	// Tarot78DeckSize はデッキの総枚数。
	Tarot78DeckSize = 78
	// Tarot78SuitCnt はスートの数。
	Tarot78SuitCnt = 4
	// Tarot78KingValue はスート札の最高位 (Roi)。1..14 の 14。
	Tarot78KingValue = 14
	// Tarot78TrumpDesign は切り札 (atout) の仮想デザイン値。1..4 はスート、5 が切り札。
	Tarot78TrumpDesign = 5
	// Tarot78ExcuseDesign はエクスキューズ (Matto) の仮想デザイン値。
	Tarot78ExcuseDesign = 6
	// Tarot78ExcuseValue はエクスキューズの値。
	Tarot78ExcuseValue = 0
	// Tarot78MaxTrump は最高位の切り札 (Mondo)。
	Tarot78MaxTrump = 21
	// Tarot78PetitValue は最低位の切り札 (Bagatto)。
	Tarot78PetitValue = 1
	// Tarot78CourtMin はコート札の最小値 (Valet)。
	Tarot78CourtMin = 11
)

// buildTarot78Deck は 78 枚タロットデッキを構築する。スート札 (design 1..4,
// value 1..14) 56 枚 + 切り札 (design 5, value 1..21) 21 枚 + エクスキューズ
// (design 6, value 0) 1 枚。
func buildTarot78Deck() []*Card {
	deck := make([]*Card, 0, Tarot78DeckSize)
	for suit := 1; suit <= Tarot78SuitCnt; suit++ {
		for val := 1; val <= Tarot78KingValue; val++ {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= Tarot78MaxTrump; val++ {
		deck = append(deck, NewCard(Tarot78TrumpDesign, val, false))
	}
	deck = append(deck, NewCard(Tarot78ExcuseDesign, Tarot78ExcuseValue, false))
	return deck
}

// tarot78IsTrump は切り札かを返す。
func tarot78IsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == Tarot78TrumpDesign
}

// tarot78IsExcuse はエクスキューズ (Matto) かを返す。
func tarot78IsExcuse(c *Card) bool {
	return c != nil && c.GetDesign() == Tarot78ExcuseDesign
}

// tarot78IsBout は 3 大オヌール (Bagatto = 切り札 1 / Mondo = 切り札 21 /
// Matto = エクスキューズ) かを返す。
func tarot78IsBout(c *Card) bool {
	if c == nil {
		return false
	}
	if tarot78IsExcuse(c) {
		return true
	}
	return tarot78IsTrump(c) && (c.GetValue() == Tarot78PetitValue || c.GetValue() == Tarot78MaxTrump)
}

// tarot78Hand はトリックの合法手を決めるのに要る最小限の手札。
//
// **持ち主の型には触らない。** 3 人版と 4 人版でプレイヤー型は別なので、
// 規則の側は「何枚あって i 枚目が何か」だけを知っていればよい。
type tarot78Hand interface {
	GetCardsSize() int
	GetCard(idx int) *Card
}

// tarot78LedSuit はトリックのリードスートを返す (エクスキューズは飛ばす)。
// まだ誰も出していない、またはエクスキューズしか出ていなければ -1。
func tarot78LedSuit(trick []*TrickCard) int {
	for _, tc := range trick {
		if tc == nil || tc.Card == nil {
			continue
		}
		if !tarot78IsExcuse(tc.Card) {
			return tc.Card.GetDesign()
		}
	}
	return -1
}

// tarot78HighestTrumpInTrick はトリック中の最強切り札の値を返す (0 = 切り札なし)。
func tarot78HighestTrumpInTrick(trick []*TrickCard) int {
	best := 0
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		if tarot78IsTrump(tc.Card) && tc.Card.GetValue() > best {
			best = tc.Card.GetValue()
		}
	}
	return best
}

// tarot78ValidPlayIndices はマストフォロー + 切り札義務 + オーバートランプ義務を
// 満たす手札インデックスを返す。
//
// **エクスキューズはいつでも足せる。** フォロー義務を免れる 1 枚なので、
// 合法手の集合から落とすと「出せるはずの札が出せない」画面になる。
func tarot78ValidPlayIndices(hand tarot78Hand, trick []*TrickCard) []int {
	n := hand.GetCardsSize()
	all := make([]int, 0, n)
	for i := 0; i < n; i++ {
		all = append(all, i)
	}
	if len(trick) == 0 {
		return all
	}
	led := tarot78LedSuit(trick)
	if led == -1 {
		return all
	}
	excuseIdx := -1
	for i := 0; i < n; i++ {
		if tarot78IsExcuse(hand.GetCard(i)) {
			excuseIdx = i
		}
	}
	highestTrump := tarot78HighestTrumpInTrick(trick)
	var base []int
	if led == Tarot78TrumpDesign {
		base = tarot78TrumpFollowIndices(hand, highestTrump)
	} else {
		base = tarot78SuitFollowIndices(hand, led, highestTrump)
	}
	if excuseIdx >= 0 {
		base = appendUniqueIndex(base, excuseIdx)
	}
	if len(base) == 0 {
		return all
	}
	return base
}

// tarot78TrumpFollowIndices は切り札リード時の合法な非エクスキューズ札を返す。
func tarot78TrumpFollowIndices(hand tarot78Hand, highestTrump int) []int {
	trumps := tarot78SuitOf(hand, Tarot78TrumpDesign)
	if len(trumps) == 0 {
		return tarot78NonExcuseIndices(hand)
	}
	higher := filterIndices(trumps, func(idx int) bool {
		c := hand.GetCard(idx)
		return c != nil && c.GetValue() > highestTrump
	})
	if len(higher) > 0 {
		return higher
	}
	return trumps
}

// tarot78SuitFollowIndices はスートリード時の合法な非エクスキューズ札を返す。
func tarot78SuitFollowIndices(hand tarot78Hand, led, highestTrump int) []int {
	ledCards := tarot78SuitOf(hand, led)
	if len(ledCards) > 0 {
		return ledCards
	}
	trumps := tarot78SuitOf(hand, Tarot78TrumpDesign)
	if len(trumps) == 0 {
		return tarot78NonExcuseIndices(hand)
	}
	higher := filterIndices(trumps, func(idx int) bool {
		c := hand.GetCard(idx)
		return c != nil && c.GetValue() > highestTrump
	})
	if highestTrump > 0 && len(higher) > 0 {
		return higher
	}
	return trumps
}

// tarot78SuitOf は指定 design の (エクスキューズを除く) 手札インデックスを返す。
func tarot78SuitOf(hand tarot78Hand, design int) []int {
	var out []int
	for i := 0; i < hand.GetCardsSize(); i++ {
		c := hand.GetCard(i)
		if c == nil || tarot78IsExcuse(c) {
			continue
		}
		if c.GetDesign() == design {
			out = append(out, i)
		}
	}
	return out
}

// tarot78NonExcuseIndices はエクスキューズを除く全手札インデックスを返す。
func tarot78NonExcuseIndices(hand tarot78Hand) []int {
	var out []int
	for i := 0; i < hand.GetCardsSize(); i++ {
		if !tarot78IsExcuse(hand.GetCard(i)) {
			out = append(out, i)
		}
	}
	return out
}

// appendUniqueIndex は重複しないようにインデックスを足す。
func appendUniqueIndex(dst []int, idx int) []int {
	for _, v := range dst {
		if v == idx {
			return dst
		}
	}
	return append(dst, idx)
}
