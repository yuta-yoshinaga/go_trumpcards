//go:build !js || !wasm || extra2

package domain

// SutdaHands.go は 2 枚の手札から役を判定する。
//
// **20 枚は 1〜10 月が 2 枚ずつ。** ただし 1・3・8 月だけは片方が「光」で、
// どの 1 枚かで役が変わる ── 花札の光札は 1 (松に鶴)・3 (桜に幕)・8 (芒に月)
// で、11・12 月は 20 枚デッキに入らないため、光は必ずこの 3 枚になる。
// 光ッタンが 38 / 18 / 13 の 3 通りしかないのはそのためで、資料の散文に
// 「1 と 2」と書かれているものがあるが、2 月に光札は存在しない。

// 役の位。数が大きいほど強い。
const (
	// SutdaRankKkeut は끗 (数の勝負)。下 1 桁がそのまま加算される。
	SutdaRankKkeut = 600
	// SutdaRankSpecial は特殊役の基準値。
	SutdaRankSpecial = 700
	// SutdaRankTtaeng は땡 (同月のペア) の基準値。月を足す。
	SutdaRankTtaeng = 800
	// SutdaRankGwangTtaeng は광땡 (光札のペア) の基準値。
	SutdaRankGwangTtaeng = 900
)

// 特殊役の加算値 (強い順)。
const (
	// SutdaSpecialAli は알리 (1+2)。
	SutdaSpecialAli = 6
	// SutdaSpecialDoksa は독사 (1+4)。
	SutdaSpecialDoksa = 5
	// SutdaSpecialGupping は구삥 (1+9)。
	SutdaSpecialGupping = 4
	// SutdaSpecialJangpping は장삥 (1+10)。
	SutdaSpecialJangpping = 3
	// SutdaSpecialJangsa は장사 (4+10)。
	SutdaSpecialJangsa = 2
	// SutdaSpecialSeryuk は세륙 (4+6)。
	SutdaSpecialSeryuk = 1
)

// SutdaGwangCopy は光札にあたる複製番号。
//
// 1・3・8 月の 1 枚目を光とする。
const SutdaGwangCopy = 1

// SutdaMonthCnt は月の数 (1〜10)。
const SutdaMonthCnt = 10

// SutdaCopiesPerMonth は 1 月あたりの枚数。
const SutdaCopiesPerMonth = 2

// SutdaHandSize は 1 人の手札枚数。
const SutdaHandSize = 2

// sutdaGwangMonths は光札を持つ月。
var sutdaGwangMonths = map[int]bool{1: true, 3: true, 8: true}

// SutdaIsGwang はその札が光札かを返す。
func SutdaIsGwang(c *Card) bool {
	if c == nil {
		return false
	}
	return sutdaGwangMonths[c.GetDesign()] && c.GetValue() == SutdaGwangCopy
}

// SutdaHand は 2 枚の手札の評価結果。
type SutdaHand struct {
	// Rank は強さ。大きいほど強い。
	Rank int
	// Name は役の安定した識別名 (i18n キー用)。
	Name string
	// Kkeut は끗 の値 (0〜9)。끗 以外の役では -1。
	Kkeut int
}

// sutdaSpecialPairs は特殊役の組み合わせ (月の昇順で引く)。
var sutdaSpecialPairs = map[[2]int]struct {
	bonus int
	name  string
}{
	{1, 2}:  {SutdaSpecialAli, "ali"},
	{1, 4}:  {SutdaSpecialDoksa, "doksa"},
	{1, 9}:  {SutdaSpecialGupping, "gupping"},
	{1, 10}: {SutdaSpecialJangpping, "jangpping"},
	{4, 10}: {SutdaSpecialJangsa, "jangsa"},
	{4, 6}:  {SutdaSpecialSeryuk, "seryuk"},
}

// sutdaGwangPairs は光ッタンの組み合わせと強さ (月の昇順で引く)。
var sutdaGwangPairs = map[[2]int]struct {
	bonus int
	name  string
}{
	{3, 8}: {3, "gwang38"},
	{1, 8}: {2, "gwang18"},
	{1, 3}: {1, "gwang13"},
}

// SutdaEvaluate は 2 枚の手札を評価する。
func SutdaEvaluate(a, b *Card) SutdaHand {
	if a == nil || b == nil {
		return SutdaHand{Rank: 0, Name: "none", Kkeut: -1}
	}
	m1, m2 := a.GetDesign(), b.GetDesign()
	if m1 > m2 {
		m1, m2 = m2, m1
	}
	key := [2]int{m1, m2}

	// **光ッタンが最上位。** 光札 2 枚のときだけ成立する。
	if SutdaIsGwang(a) && SutdaIsGwang(b) {
		if g, ok := sutdaGwangPairs[key]; ok {
			return SutdaHand{Rank: SutdaRankGwangTtaeng + g.bonus, Name: g.name, Kkeut: -1}
		}
	}
	// 땡 (同月のペア)。장땡 (10) が最強。
	if m1 == m2 {
		return SutdaHand{Rank: SutdaRankTtaeng + m1, Name: sutdaTtaengName(m1), Kkeut: -1}
	}
	// 特殊役。끗 より強い。
	if s, ok := sutdaSpecialPairs[key]; ok {
		return SutdaHand{Rank: SutdaRankSpecial + s.bonus, Name: s.name, Kkeut: -1}
	}
	// 끗。合計の下 1 桁で、0 は망통 で最弱。
	kkeut := (m1 + m2) % 10
	return SutdaHand{Rank: SutdaRankKkeut + kkeut, Name: sutdaKkeutName(kkeut), Kkeut: kkeut}
}

// sutdaTtaengName は땡 の識別名を返す。
func sutdaTtaengName(month int) string {
	if month == SutdaMonthCnt {
		return "jangttaeng"
	}
	return "ttaeng" + sutdaDigit(month)
}

// sutdaKkeutName は끗 の識別名を返す。
func sutdaKkeutName(kkeut int) string {
	switch kkeut {
	case 0:
		return "mangtong"
	case 9:
		return "gabo"
	default:
		return "kkeut" + sutdaDigit(kkeut)
	}
}

// sutdaDigit は 1 桁の数字を文字にする (i18n キー用)。
func sutdaDigit(n int) string {
	if n < 0 || n > 9 {
		return "?"
	}
	return string(rune('0' + n))
}

// buildSutdaDeck は 20 枚 (1〜10 月 × 2 枚) を design=月 / value=複製番号 で作る。
func buildSutdaDeck() []*Card {
	deck := make([]*Card, 0, SutdaMonthCnt*SutdaCopiesPerMonth)
	for m := 1; m <= SutdaMonthCnt; m++ {
		for i := 1; i <= SutdaCopiesPerMonth; i++ {
			deck = append(deck, NewCard(m, i, false))
		}
	}
	return deck
}
