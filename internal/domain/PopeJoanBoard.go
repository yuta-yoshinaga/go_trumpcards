//go:build !js || !wasm || extra3

package domain

// PopeJoanCompartment は盤上の 8 区画のどれか。
//
// issue #4389 は「8 種の絵札 + ポープ」としているが、原典の盤は **8 区画**で
// あり、絵札の数とは関係がない。Game / Matrimony / Intrigue の 3 つは札 1 枚に
// 対応していない。
type PopeJoanCompartment int

// Pope Joan の 8 区画。盤の刻印どおりの並び。
const (
	// PopeJoanAce トランプの A
	PopeJoanAce PopeJoanCompartment = iota
	// PopeJoanKing トランプの K
	PopeJoanKing
	// PopeJoanQueen トランプの Q
	PopeJoanQueen
	// PopeJoanJack トランプの J
	PopeJoanJack
	// PopeJoanGame 先に手札を出し切った人が取る
	PopeJoanGame
	// PopeJoanPope ♦9。この game の名前になっている札
	PopeJoanPope
	// PopeJoanMatrimony トランプの K と Q を**同じ人が出したら**
	PopeJoanMatrimony
	// PopeJoanIntrigue トランプの Q と J を**同じ人が出したら**
	PopeJoanIntrigue
	// PopeJoanCompartmentCount 区画の総数
	PopeJoanCompartmentCount
)

// PopeJoanCompartmentNames は各区画の識別子。盤の刻印に合わせてある。
var PopeJoanCompartmentNames = [PopeJoanCompartmentCount]string{
	"ace", "king", "queen", "jack", "game", "pope", "matrimony", "intrigue",
}

// String は区画の識別子を返す。
func (c PopeJoanCompartment) String() string {
	if c < 0 || c >= PopeJoanCompartmentCount {
		return "unknown"
	}
	return PopeJoanCompartmentNames[c]
}

// popeJoanDress は «dressing the board» の内訳。**プレイヤーが配分するのでは
// なく、ディーラーがこの通りに置く。**合計 15。
var popeJoanDress = [PopeJoanCompartmentCount]int{
	PopeJoanAce:       1,
	PopeJoanKing:      1,
	PopeJoanQueen:     1,
	PopeJoanJack:      1,
	PopeJoanGame:      1,
	PopeJoanPope:      6,
	PopeJoanMatrimony: 2,
	PopeJoanIntrigue:  2,
}

// PopeJoanDressTotal はディーラーが 1 ディールで置く総額。
const PopeJoanDressTotal = 15

// PopeJoanBoard は 8 区画のチップ残高。
//
// **取られなかった区画は持ち越す。**次のディールでもディーラーが同じ内訳を
// 足すので、Pope や Matrimony のように条件の厳しい区画は溜まっていく。これが
// issue #4389 に書かれていない、この game の中心的な動機づけである。
type PopeJoanBoard struct {
	Chips [PopeJoanCompartmentCount]int `json:"chips"`
}

// Dress はディーラーが固定の内訳で置く。置いた総額を返す。
func (b *PopeJoanBoard) Dress() int {
	total := 0
	for i, n := range popeJoanDress {
		b.Chips[i] += n
		total += n
	}
	return total
}

// Take は区画のチップを空にして、そこにあった枚数を返す。
func (b *PopeJoanBoard) Take(c PopeJoanCompartment) int {
	if c < 0 || c >= PopeJoanCompartmentCount {
		return 0
	}
	n := b.Chips[c]
	b.Chips[c] = 0
	return n
}

// Get は区画の残高を返す。
//
// **値レシーバ。**GetBoard() が値のコピーを返すので、ポインタメソッドだと
// 呼び出し側から見えなくなる。読むだけなのでこれで足りる。
func (b PopeJoanBoard) Get(c PopeJoanCompartment) int {
	if c < 0 || c >= PopeJoanCompartmentCount {
		return 0
	}
	return b.Chips[c]
}

// PopeJoanCompartmentForRank はトランプの rank に対応する単札区画を返す。
// 対応が無ければ false。
func PopeJoanCompartmentForRank(rank int) (PopeJoanCompartment, bool) {
	switch rank {
	case 1:
		return PopeJoanAce, true
	case 13:
		return PopeJoanKing, true
	case 12:
		return PopeJoanQueen, true
	case 11:
		return PopeJoanJack, true
	default:
		return 0, false
	}
}
