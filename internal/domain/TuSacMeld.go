//go:build !js || !wasm || solo

package domain

// TuSacMeld は場に出した組み合わせ。
type TuSacMeld struct {
	Kind  TuSacMeldKind
	Cards []*Card
}

// TuSacClassifyMeld は札の並びがどのメルドかを判定する。
//
// **3 つの形しかない。** 同色同種 3 枚 / 異色の車馬砲 3 枚 / 卒 5 枚 ──
// 標準デッキのラミーにある「同スートの連番」は、四色牌に数字の並びが無いので
// 存在しない。数字の大小を持ち込むと別のゲームになる。
func TuSacClassifyMeld(cards []*Card) TuSacMeldKind {
	if tuSacHasNilCard(cards) {
		return TuSacMeldNone
	}
	switch len(cards) {
	case TuSacSetSize:
		if kind := tuSacClassifyTriple(cards); kind != TuSacMeldNone {
			return kind
		}
	case TuSacSoldierSetSize:
		if tuSacAllSoldiers(cards) {
			return TuSacMeldSoldierSet
		}
	}
	return TuSacMeldNone
}

// tuSacClassifyTriple は 3 枚の組を判定する。
func tuSacClassifyTriple(cards []*Card) TuSacMeldKind {
	// 同色・同種 3 枚。
	sameColor := cards[0].GetDesign() == cards[1].GetDesign() &&
		cards[1].GetDesign() == cards[2].GetDesign()
	samePiece := cards[0].GetValue() == cards[1].GetValue() &&
		cards[1].GetValue() == cards[2].GetValue()
	if sameColor && samePiece {
		return TuSacMeldSameColorSet
	}

	// **異色の車・馬・砲。** 3 種がそろい、色が 3 つとも違うことの両方が要る。
	// 片方だけを見ると、同色の車馬砲や、異色の車車馬が通ってしまう。
	if !tuSacIsChariotHorseCannonTrio(cards) {
		return TuSacMeldNone
	}
	if cards[0].GetDesign() == cards[1].GetDesign() ||
		cards[1].GetDesign() == cards[2].GetDesign() ||
		cards[0].GetDesign() == cards[2].GetDesign() {
		return TuSacMeldNone
	}
	return TuSacMeldChariotTrio
}

// tuSacIsChariotHorseCannonTrio は 車・馬・砲 が 1 枚ずつかを返す。
func tuSacIsChariotHorseCannonTrio(cards []*Card) bool {
	seen := map[int]bool{}
	for _, c := range cards {
		if !TuSacIsChariotHorseCannon(c.GetValue()) || seen[c.GetValue()] {
			return false
		}
		seen[c.GetValue()] = true
	}
	return len(seen) == TuSacTrioSize
}

// tuSacAllSoldiers は全部が卒かを返す。
func tuSacAllSoldiers(cards []*Card) bool {
	for _, c := range cards {
		if c.GetValue() != TuSacPieceSoldier {
			return false
		}
	}
	return true
}

// tuSacHasNilCard は nil が混ざっているか (または空か) を返す。
//
// **同名の helper が OichoKabu.go にある。** あちらは extra タグ、こちらは
// solo タグだが、ホストビルドは両方を含むので名前が衝突する ── タグで
// 分かれていても、無印のビルドでは同じパッケージに同居する。
func tuSacHasNilCard(cards []*Card) bool {
	for _, c := range cards {
		if c == nil {
			return true
		}
	}
	return len(cards) == 0
}

// TuSacFindMeld は手札から指定の添字の札を取り出し、メルドかを判定する。
//
// **添字は手札の位置。** 札そのものを渡させると、同じ色・同じ駒が 4 枚ある
// このデッキでは「どの 1 枚か」が決まらない。
func TuSacFindMeld(hand []*Card, indexes []int) ([]*Card, TuSacMeldKind) {
	seen := map[int]bool{}
	picked := make([]*Card, 0, len(indexes))
	for _, i := range indexes {
		if i < 0 || i >= len(hand) || seen[i] {
			return nil, TuSacMeldNone
		}
		seen[i] = true
		picked = append(picked, hand[i])
	}
	return picked, TuSacClassifyMeld(picked)
}
