//go:build !js || !wasm || extra3

package domain

import "fmt"

// CirullaEval.go は捕獲判定と得点計算を担う。
//
// **Cirulla は「合計 15」だけのゲームではない。** #5457 は捕獲基準が 15 に
// 「固定される」と書いているが、実際は Scopa の捕獲 (同値の 1 枚、または
// 合計が出した札に一致する組) がそのまま生きていて、**そこに 15 が足される**。
// 15 だけにすると、7 で 7 を取るような当たり前の手が打てなくなる。

// CirullaTargetSum は追加の捕獲条件になる合計。
const CirullaTargetSum = 15

// CirullaMaxSubsetCards は部分集合の全探索を許す場札の上限。
//
// **場札は理屈のうえでは積み上がる。** 2^n の全探索なので、上限が無いと
// Worker の CPU 予算 (無料枠 10ms) を静かに超える。
const CirullaMaxSubsetCards = 14

// CirullaCardValue は捕獲に使う札の値を返す (A=1, 2-7=そのまま, J=8, Q=9, K=10)。
func CirullaCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 11:
		return 8
	case 12:
		return 9
	case 13:
		return 10
	default:
		return c.GetValue()
	}
}

// CirullaIsAce はアッソ (A) かを返す。
func CirullaIsAce(c *Card) bool { return c != nil && c.GetValue() == 1 }

// CirullaIsDenari はデナリ (ダイヤ) かを返す。
func CirullaIsDenari(c *Card) bool { return c != nil && c.GetDesign() == CardDesignDiamond }

// CirullaIsSetteBello は 7♦ (セッテベッロ) かを返す。
func CirullaIsSetteBello(c *Card) bool { return CirullaIsDenari(c) && c.GetValue() == 7 }

// CirullaIsWildForBonus は配札ボーナスの判定でワイルドになる札かを返す。
//
// **7♥ だけがワイルドで、しかもボーナスの判定にしか効かない。** 捕獲では
// ただの 7 として扱う。
func CirullaIsWildForBonus(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignHeart && c.GetValue() == 7
}

// cirullaTableHasAce は場にアッソがあるかを返す。
func cirullaTableHasAce(tableCards []*Card) bool {
	for _, c := range tableCards {
		if CirullaIsAce(c) {
			return true
		}
	}
	return false
}

// CirullaAceTakesAll はアッソを出したら場を総取りできるかを返す。
//
// **アッソ・ピリアトゥット。** 場にアッソが無いときだけ成立する ── 場に
// アッソがあると、ただの 1 として同値捕獲になる。
func CirullaAceTakesAll(playedCard *Card, tableCards []*Card) bool {
	return CirullaIsAce(playedCard) && len(tableCards) > 0 && !cirullaTableHasAce(tableCards)
}

// cirullaSubsetsSummingTo は合計が target になる部分集合を列挙する。
func cirullaSubsetsSummingTo(cards []*Card, target int) [][]int {
	if target <= 0 {
		return nil
	}
	n := len(cards)
	if n == 0 || n > CirullaMaxSubsetCards {
		return nil
	}
	result := make([][]int, 0)
	for mask := 1; mask < (1 << n); mask++ {
		sum := 0
		over := false
		for i := 0; i < n; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			sum += CirullaCardValue(cards[i])
			if sum > target {
				over = true
				break
			}
		}
		if over || sum != target {
			continue
		}
		idxs := make([]int, 0)
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				idxs = append(idxs, i)
			}
		}
		result = append(result, idxs)
	}
	return result
}

// cirullaAllIndices は 0..n-1 を返す。
func cirullaAllIndices(n int) []int {
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, i)
	}
	return out
}

// EnumerateCirullaCaptures は playedCard を出したときの合法な捕獲を列挙する。
//
//   - アッソで場にアッソが無ければ、場の総取りだけ。
//   - それ以外は「同値の 1 枚」「合計が出した札の値」「合計 15」を **すべて** 集める。
//
// **チルッラでは同値の 1 枚が組合せを締め出さない。** Scopa なら場に同値の札が
// あるときは必ずその 1 枚を取らされるが、チルッラは
// "unlike Scopa, if several different captures are available with the played
// card, the player is free to choose which one to make" (pagat)。
// Scopa の優先規則をそのまま持ち込むと、♠7 ♥3 ♣5 の場に ♦7 を出したとき、
// 3+5+7=15 という **この派生の看板ルールのほうが黙って消える**。
func EnumerateCirullaCaptures(playedCard *Card, tableCards []*Card) [][]int {
	if playedCard == nil || len(tableCards) == 0 {
		return nil
	}
	if CirullaAceTakesAll(playedCard, tableCards) {
		return [][]int{cirullaAllIndices(len(tableCards))}
	}
	value := CirullaCardValue(playedCard)

	out := make([][]int, 0)
	// 同値の 1 枚。複数あればそれぞれ別の選択肢。
	for i, c := range tableCards {
		if c != nil && CirullaCardValue(c) == value {
			out = append(out, []int{i})
		}
	}
	out = append(out, cirullaSubsetsSummingTo(tableCards, value)...)
	// **15 は「出した札を足して 15」。** 場札だけで 15 ではない。
	if rest := CirullaTargetSum - value; rest > 0 {
		out = append(out, cirullaSubsetsSummingTo(tableCards, rest)...)
	}
	return cirullaDedupe(out)
}

// cirullaDedupe は同じインデックス集合の重複を落とす。
//
// **合計一致と 15 が同じ組を指すことがある。** 7 を出して場に 8 があれば
// どちらの規則でも同じ 1 枚で、二重に並べると選択肢が水増しされる。
func cirullaDedupe(groups [][]int) [][]int {
	seen := make(map[string]bool, len(groups))
	out := make([][]int, 0, len(groups))
	for _, g := range groups {
		// 列挙はどちらの経路も昇順なので、そのまま並べた文字列で一意に決まる。
		key := fmt.Sprint(g)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, g)
	}
	return out
}

// IsValidCirullaCapture は選ばれた場札が playedCard で合法に捕獲できるかを返す。
func IsValidCirullaCapture(playedCard *Card, tableCards []*Card, selectedIdxs []int) bool {
	if playedCard == nil || len(selectedIdxs) == 0 {
		return false
	}
	seen := make(map[int]bool, len(selectedIdxs))
	sum := 0
	for _, idx := range selectedIdxs {
		if idx < 0 || idx >= len(tableCards) || seen[idx] {
			return false
		}
		seen[idx] = true
		sum += CirullaCardValue(tableCards[idx])
	}

	if CirullaAceTakesAll(playedCard, tableCards) {
		// アッソは総取りのみ。半端に選ばせない。
		return len(selectedIdxs) == len(tableCards)
	}
	value := CirullaCardValue(playedCard)
	// **3 つの規則は並列で、どれを使ってもよい。** 同値の 1 枚を選ぶ手を
	// 強制すると、15 の組が打てなくなる (pagat: プレイヤーが自由に選ぶ)。
	if len(selectedIdxs) == 1 && CirullaCardValue(tableCards[selectedIdxs[0]]) == value {
		return true
	}
	return sum == value || sum+value == CirullaTargetSum
}

// 配札ボーナス。
const (
	// CirullaBonusNone ボーナス無し。
	CirullaBonusNone = ""
	// CirullaBonusBarsega バルセガ (3 枚の合計が 9 以下)。
	CirullaBonusBarsega = "barsega"
	// CirullaBonusBarsegon バルセゴン (3 枚が同位)。
	CirullaBonusBarsegon = "barsegon"
)

// 配札ボーナスの点数。
const (
	// CirullaBarsegaPoints はバルセガの点。
	CirullaBarsegaPoints = 3
	// CirullaBarsegonPoints はバルセゴンの点。**バルセガより桁が違う。**
	CirullaBarsegonPoints = 10
	// CirullaBarsegaMaxSum はバルセガが成立する合計の上限。
	CirullaBarsegaMaxSum = 9
)

// CirullaDealBonus は配られた 3 枚のボーナスを判定する。
//
// **7♥ はワイルド。** ただしボーナスの判定にだけ効き、捕獲ではただの 7。
// バルセゴン (同位 3 枚) はバルセガ (合計 9 以下) より優先する。
func CirullaDealBonus(cards []*Card) (string, int) {
	if len(cards) != CirullaHandSize {
		return CirullaBonusNone, 0
	}
	if cirullaIsThreeOfAKind(cards) {
		return CirullaBonusBarsegon, CirullaBarsegonPoints
	}
	sum := 0
	for _, c := range cards {
		sum += CirullaCardValue(c)
	}
	if sum <= CirullaBarsegaMaxSum {
		return CirullaBonusBarsega, CirullaBarsegaPoints
	}
	// ワイルドを使えば同位 3 枚にできるか。
	if cirullaWildMakesThreeOfAKind(cards) {
		return CirullaBonusBarsegon, CirullaBarsegonPoints
	}
	return CirullaBonusNone, 0
}

// cirullaIsThreeOfAKind は 3 枚が同位かを返す。
func cirullaIsThreeOfAKind(cards []*Card) bool {
	if cards[0] == nil || cards[1] == nil || cards[2] == nil {
		return false
	}
	return cards[0].GetValue() == cards[1].GetValue() && cards[1].GetValue() == cards[2].GetValue()
}

// cirullaWildMakesThreeOfAKind は 7♥ を任意の位として使えば同位 3 枚になるかを返す。
func cirullaWildMakesThreeOfAKind(cards []*Card) bool {
	wild := -1
	for i, c := range cards {
		if CirullaIsWildForBonus(c) {
			wild = i
			break
		}
	}
	if wild < 0 {
		return false
	}
	rest := make([]*Card, 0, 2)
	for i, c := range cards {
		if i != wild {
			rest = append(rest, c)
		}
	}
	return len(rest) == 2 && rest[0] != nil && rest[1] != nil &&
		rest[0].GetValue() == rest[1].GetValue()
}

// cirullaPrimieraValue はプリミエラの札の価値を返す (7 が最高)。
func cirullaPrimieraValue(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 7:
		return 21
	case 6:
		return 18
	case 1:
		return 16
	case 5:
		return 15
	case 4:
		return 14
	case 3:
		return 13
	case 2:
		return 12
	default:
		return 10 // 絵札
	}
}

// CirullaPrimiera は獲得札からプリミエラの点を返す。
//
// **各スートの最高 1 枚ずつの合計。** スートが 1 つでも欠けていれば争えない。
func CirullaPrimiera(cards []*Card) int {
	best := map[int]int{}
	for _, c := range cards {
		if c == nil {
			continue
		}
		if v := cirullaPrimieraValue(c); v > best[c.GetDesign()] {
			best[c.GetDesign()] = v
		}
	}
	if len(best) < 4 {
		return 0
	}
	total := 0
	for _, v := range best {
		total += v
	}
	return total
}

// CirullaPiccola はピッコラの点を返す。
//
// **A♦ から続いた枚数がそのまま点。** A♦ が無ければ 0、A-2 で 2 点、
// A-2-3 で 3 点 ── 途切れたところで止まる。
func CirullaPiccola(cards []*Card) int {
	have := map[int]bool{}
	for _, c := range cards {
		if CirullaIsDenari(c) {
			have[c.GetValue()] = true
		}
	}
	if !have[1] || !have[2] {
		return 0
	}
	n := 2
	for v := 3; v <= 7; v++ {
		if !have[v] {
			break
		}
		n++
	}
	return n
}

// CirullaGrandePoints はグランデの点。
const CirullaGrandePoints = 5

// CirullaHasGrande は K♦・Q♦・J♦ を揃えたかを返す。
func CirullaHasGrande(cards []*Card) bool {
	have := map[int]bool{}
	for _, c := range cards {
		if CirullaIsDenari(c) {
			have[c.GetValue()] = true
		}
	}
	return have[11] && have[12] && have[13]
}
