//go:build !js || !wasm || extra3

package domain

// PochPool は盤上の 9 つのプールのどれか。
//
// issue #4415 は「各ランク + ペア枠」としか書いていないが、原典 (pagat) の盤は
// **9 区画で固定**であり、名前も決まっている。
type PochPool int

// Poch の 9 プール。盤の刻印どおりの並び。
const (
	// PochPoolAce pay suit の A
	PochPoolAce PochPool = iota
	// PochPoolKing pay suit の K
	PochPoolKing
	// PochPoolQueen pay suit の Q
	PochPoolQueen
	// PochPoolJack pay suit の J
	PochPoolJack
	// PochPoolTen pay suit の 10
	PochPoolTen
	// PochPoolMarriage pay suit の K と Q の**両方**
	PochPoolMarriage
	// PochPoolSequence pay suit の 7-8-9 の**三枚すべて**
	PochPoolSequence
	// PochPoolPocher 同ランクの組の勝者が取る
	PochPoolPocher
	// PochPoolCentre 先に手札を出し切った人が取る
	PochPoolCentre
	// PochPoolCount プールの総数
	PochPoolCount
)

// PochPoolNames は各プールの識別子。盤の刻印に合わせてある。
var PochPoolNames = [PochPoolCount]string{
	"ace", "king", "queen", "jack", "ten", "marriage", "sequence", "pocher", "centre",
}

// String はプールの識別子を返す。
func (p PochPool) String() string {
	if p < 0 || p >= PochPoolCount {
		return "unknown"
	}
	return PochPoolNames[p]
}

// pochRankPools は「pay suit の 1 枚を持っていれば取れる」プールと、その
// ランクの対応。
var pochRankPools = []struct {
	pool PochPool
	rank int
}{
	{PochPoolAce, 1},
	{PochPoolKing, 13},
	{PochPoolQueen, 12},
	{PochPoolJack, 11},
	{PochPoolTen, 10},
}

// PochBoard は 9 プールのチップ残高。
//
// **取られなかったプールは持ち越す。**次のディールでも全員が 1 枚ずつ足すので、
// Marriage や Sequence のように条件の厳しいプールは溜まっていく。これが
// issue #4415 に書かれていない、この game の中心的な動機づけである。
type PochBoard struct {
	Chips [PochPoolCount]int `json:"chips"`
}

// Ante は全員が 9 プールすべてに 1 枚ずつ置く。
func (b *PochBoard) Ante(playerCnt int) {
	for i := range b.Chips {
		b.Chips[i] += playerCnt
	}
}

// Take は pool のチップを空にして、そこにあった枚数を返す。
func (b *PochBoard) Take(pool PochPool) int {
	if pool < 0 || pool >= PochPoolCount {
		return 0
	}
	n := b.Chips[pool]
	b.Chips[pool] = 0
	return n
}

// Add は pool にチップを足す。
func (b *PochBoard) Add(pool PochPool, n int) {
	if pool < 0 || pool >= PochPoolCount || n <= 0 {
		return
	}
	b.Chips[pool] += n
}

// Get は pool の残高を返す。
//
// **値レシーバ。**GetBoard() が値のコピーを返すので、ポインタメソッドだと
// 呼び出し側から見えなくなる。読むだけなのでこれで足りる。
func (b PochBoard) Get(pool PochPool) int {
	if pool < 0 || pool >= PochPoolCount {
		return 0
	}
	return b.Chips[pool]
}
