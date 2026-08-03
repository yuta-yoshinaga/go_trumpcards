//go:build !js || !wasm || extra3

package domain

// NainJauneBox は盤上の 5 区画のどれか。
//
// issue #4380 は区画を「♦7 / ♠10 / ♥Q / ♣J / シーケンス」としているが、正しくは
// **♦10 / ♣J / ♠Q / ♥K / ♦7** である。issue は ♦10 を「♠10」、♠Q を「♥Q」と
// スートを取り違えており、**♥K が抜けている**。そして**「シーケンス」区画は
// 存在しない**。
type NainJauneBox int

// Nain Jaune の 5 区画。積む枚数の少ない順。
const (
	// NainJauneBoxTen ♦10
	NainJauneBoxTen NainJauneBox = iota
	// NainJauneBoxJack ♣J
	NainJauneBoxJack
	// NainJauneBoxQueen ♠Q
	NainJauneBoxQueen
	// NainJauneBoxKing ♥K
	NainJauneBoxKing
	// NainJauneBoxDwarf ♦7 (le nain jaune)。盤の中央で、最も多く積まれる。
	NainJauneBoxDwarf
	// NainJauneBoxCount 区画の総数
	NainJauneBoxCount
)

// NainJauneBoxNames は各区画の識別子。
var NainJauneBoxNames = [NainJauneBoxCount]string{"ten", "jack", "queen", "king", "dwarf"}

// String は区画の識別子を返す。
func (b NainJauneBox) String() string {
	if b < 0 || b >= NainJauneBoxCount {
		return "unknown"
	}
	return NainJauneBoxNames[b]
}

// nainJauneBoxCards は各区画を取る札。**スートまで一致していなければならない。**
var nainJauneBoxCards = [NainJauneBoxCount]struct {
	design int
	value  int
}{
	NainJauneBoxTen:   {CardDesignDiamond, 10},
	NainJauneBoxJack:  {CardDesignClover, 11},
	NainJauneBoxQueen: {CardDesignSpade, 12},
	NainJauneBoxKing:  {CardDesignHeart, 13},
	NainJauneBoxDwarf: {CardDesignDiamond, 7},
}

// nainJauneAnte は各プレイヤーが 1 ディールで区画ごとに置く枚数。
//
// **均等ではない。**♦10 に 1、♣J に 2、♠Q に 3、♥K に 4、♦7 に 5。
var nainJauneAnte = [NainJauneBoxCount]int{
	NainJauneBoxTen:   1,
	NainJauneBoxJack:  2,
	NainJauneBoxQueen: 3,
	NainJauneBoxKing:  4,
	NainJauneBoxDwarf: 5,
}

// NainJauneAnteTotal は 1 人が 1 ディールで置く総額。
const NainJauneAnteTotal = 15

// NainJauneBoxForCard は card が取る区画を返す。対応が無ければ false。
func NainJauneBoxForCard(card *Card) (NainJauneBox, bool) {
	if card == nil {
		return 0, false
	}
	for i, spec := range nainJauneBoxCards {
		if card.GetDesign() == spec.design && card.GetValue() == spec.value {
			return NainJauneBox(i), true
		}
	}
	return 0, false
}

// NainJauneBoard は 5 区画のチップ残高。
//
// **取られなかった区画は持ち越す。**次のディールでも全員がまた置くので、♦7 の
// ように出にくい札の区画は膨らんでいく。
type NainJauneBoard struct {
	Chips [NainJauneBoxCount]int `json:"chips"`
}

// Ante は playerCnt 人ぶんのアンティを積む。積んだ総額を返す。
func (b *NainJauneBoard) Ante(playerCnt int) int {
	total := 0
	for i, n := range nainJauneAnte {
		b.Chips[i] += n * playerCnt
		total += n * playerCnt
	}
	return total
}

// Take は区画のチップを空にして、そこにあった枚数を返す。
func (b *NainJauneBoard) Take(box NainJauneBox) int {
	if box < 0 || box >= NainJauneBoxCount {
		return 0
	}
	n := b.Chips[box]
	b.Chips[box] = 0
	return n
}

// Add は区画にチップを足す。
func (b *NainJauneBoard) Add(box NainJauneBox, n int) {
	if box < 0 || box >= NainJauneBoxCount || n <= 0 {
		return
	}
	b.Chips[box] += n
}

// Get は区画の残高を返す。
//
// **値レシーバ。**GetBoard() が値のコピーを返すので、ポインタメソッドだと
// 呼び出し側から見えなくなる。
func (b NainJauneBoard) Get(box NainJauneBox) int {
	if box < 0 || box >= NainJauneBoxCount {
		return 0
	}
	return b.Chips[box]
}

// NainJaunePoints は手札に残った 1 枚の失点を返す。
//
// **枚数ではなく点数で払う。**A=1、2〜10 は額面、J/Q/K は各 10。
func NainJaunePoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch v := c.GetValue(); {
	case v == 1:
		return 1
	case v >= 11:
		return 10
	default:
		return v
	}
}
