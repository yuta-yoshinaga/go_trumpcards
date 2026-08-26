//go:build !js || !wasm || classic

package domain

// ディロティの捕獲・宣言の規則。
//
// **#5458 の規則説明は 3 か所とも誤っている。**
//
//  1. 「場札の合計が 10 になる組合せを取る」── 基準は **出した札のランク** で
//     あって 10 ではない。10 固定にすると、5 で 2+3 を取る当たり前の手が
//     打てなくなる。
//  2. 「絵札の同ランク取りが独自ルール」── 同ランク取りはこの系統の基本で、
//     ディロティ固有なのはむしろ **絵札はちょうど 1 枚しか取れず、同ランクの
//     絵札が場にあるときは置くこともできない** という制限のほう。
//  3. 「Ksera」── 綴りは **Xeri (ξερή)** で、しかも「場を空にした」だけでは
//     足りない。**1 枚で場の全札を取る**必要があり、その局の初手と、
//     山が尽きたあとの取り残し回収は数えない。
//
// そして #5458 は **宣言 (δήλωση) に一言も触れていない** ── ゲーム名の由来
// そのもので、これを省くと残るのはクセリであってディロティではない。
const (
	// DilotiMaxDeclaration は宣言値の上限。
	DilotiMaxDeclaration = 10
	// DilotiMinDeclaration は宣言値の下限。
	DilotiMinDeclaration = 2
	// DilotiMaxSubsetCards は部分集合の全探索を許す場札の枚数。
	//
	// **Worker の CPU 予算は手元では見えない。** 2^n の全探索は 20 枚を超えた
	// あたりから無料枠の 10ms を食い潰すので、場が膨らんだら列挙を諦める。
	DilotiMaxSubsetCards = 14
)

// DilotiCardValue は捕獲の合計に使う値を返す。A は 1、**絵札は 0** (合計に
// 参加しない)。
func DilotiCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	v := c.GetValue()
	if v >= 11 && v <= 13 {
		return 0
	}
	return v
}

// DilotiIsFaceCard は絵札 (J/Q/K) かどうか。
func DilotiIsFaceCard(c *Card) bool {
	if c == nil {
		return false
	}
	v := c.GetValue()
	return v >= 11 && v <= 13
}

// DilotiTake は 1 つの取り手。TableIdxs は場の緩い札、DeclIdxs は宣言の位置。
type DilotiTake struct {
	// TableIdxs は取る場札の位置。
	TableIdxs []int
	// DeclIdxs は取る宣言の位置。
	DeclIdxs []int
}

// EnumerateDilotiTakes は handCard を出したときに合法な取り手を列挙する。
//
// 数札は「合計がランクに等しい互いに素な束」をいくつでも同時に取れるので、
// 個々の束・値の合う宣言・そして **取れるだけ取る手** を挙げる。最後のひとつが
// 無いとクセリ (場を空にする 10 点) に手が届かない。
//
// 絵札は同ランクの場札をちょうど 1 枚。宣言に絵札は入らず値も 10 以下なので、
// 絵札で宣言を取ることはない。
func EnumerateDilotiTakes(handCard *Card, tableCards []*Card, decls []*DilotiDeclaration) []DilotiTake {
	if handCard == nil {
		return nil
	}
	if DilotiIsFaceCard(handCard) {
		out := make([]DilotiTake, 0, 2)
		for i, c := range tableCards {
			if c != nil && c.GetValue() == handCard.GetValue() {
				out = append(out, DilotiTake{TableIdxs: []int{i}})
			}
		}
		return out
	}

	v := DilotiCardValue(handCard)
	groups := dilotiSubsetsSummingTo(tableCards, v)
	matchingDecls := make([]int, 0, len(decls))
	for i, d := range decls {
		if d != nil && d.Value == v {
			matchingDecls = append(matchingDecls, i)
		}
	}

	out := make([]DilotiTake, 0, len(groups)+2)
	for _, g := range groups {
		out = append(out, DilotiTake{TableIdxs: append([]int(nil), g...)})
	}
	if len(matchingDecls) > 0 {
		out = append(out, DilotiTake{DeclIdxs: append([]int(nil), matchingDecls...)})
	}

	// **取れるだけ取る手。** 互いに素な束を貪欲に積み、宣言も全部載せる。
	greedy := dilotiGreedyDisjoint(groups)
	if len(greedy) > 0 || len(matchingDecls) > 0 {
		max := DilotiTake{TableIdxs: greedy}
		if len(matchingDecls) > 0 {
			max.DeclIdxs = append([]int(nil), matchingDecls...)
		}
		if !dilotiTakeSeen(out, max) {
			out = append(out, max)
		}
	}
	return out
}

// dilotiTakeSeen は同じ取り手が既に挙がっているかを見る。
func dilotiTakeSeen(list []DilotiTake, want DilotiTake) bool {
	for _, t := range list {
		if len(t.TableIdxs) == len(want.TableIdxs) && len(t.DeclIdxs) == len(want.DeclIdxs) &&
			dilotiSameIdxs(t.TableIdxs, want.TableIdxs) && dilotiSameIdxs(t.DeclIdxs, want.DeclIdxs) {
			return true
		}
	}
	return false
}

func dilotiSameIdxs(a, b []int) bool {
	seen := make(map[int]struct{}, len(a))
	for _, v := range a {
		seen[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := seen[v]; !ok {
			return false
		}
	}
	return true
}

// dilotiGreedyDisjoint は互いに素な束を順に取り込み、平らなインデックス列にする。
func dilotiGreedyDisjoint(groups [][]int) []int {
	used := make(map[int]bool)
	out := make([]int, 0)
	for _, g := range groups {
		clash := false
		for _, i := range g {
			if used[i] {
				clash = true
				break
			}
		}
		if clash {
			continue
		}
		for _, i := range g {
			used[i] = true
			out = append(out, i)
		}
	}
	return out
}

// dilotiSubsetsSummingTo は合計が target になる部分集合を列挙する。絵札は除く。
func dilotiSubsetsSummingTo(cards []*Card, target int) [][]int {
	n := len(cards)
	if target <= 0 || n == 0 || n > DilotiMaxSubsetCards {
		return nil
	}
	out := make([][]int, 0)
	for mask := 1; mask < (1 << n); mask++ {
		sum, skip := 0, false
		idxs := make([]int, 0, n)
		for i := 0; i < n; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			if cards[i] == nil || DilotiIsFaceCard(cards[i]) {
				skip = true
				break
			}
			sum += DilotiCardValue(cards[i])
			if sum > target {
				skip = true
				break
			}
			idxs = append(idxs, i)
		}
		if !skip && sum == target {
			out = append(out, idxs)
		}
	}
	return out
}

// CanTrailDiloti は handCard を場に置けるかを返す。
//
// **同ランクの絵札が場にあるなら絵札は置けない。** 置けてしまうと、取れる絵札を
// 見送って場に積むだけの手が成立し、絵札が場に溜まっていく。数札は取れても
// 置ける ── 捕獲は強制ではない。
func CanTrailDiloti(handCard *Card, tableCards []*Card) bool {
	if handCard == nil {
		return false
	}
	if !DilotiIsFaceCard(handCard) {
		return true
	}
	for _, c := range tableCards {
		if c != nil && c.GetValue() == handCard.GetValue() {
			return false
		}
	}
	return true
}

// IsValidDilotiCapture は選んだ取り手が合法かを検査する。
func IsValidDilotiCapture(handCard *Card, tableCards []*Card, decls []*DilotiDeclaration,
	tableIdxs, declIdxs []int) bool {
	if handCard == nil || (len(tableIdxs) == 0 && len(declIdxs) == 0) {
		return false
	}
	if !dilotiIdxsInRange(tableIdxs, len(tableCards)) || !dilotiIdxsInRange(declIdxs, len(decls)) {
		return false
	}

	if DilotiIsFaceCard(handCard) {
		// 絵札は同ランクをちょうど 1 枚。宣言には触れない。
		if len(declIdxs) != 0 || len(tableIdxs) != 1 {
			return false
		}
		c := tableCards[tableIdxs[0]]
		return c != nil && c.GetValue() == handCard.GetValue()
	}

	v := DilotiCardValue(handCard)
	for _, i := range declIdxs {
		if decls[i] == nil || decls[i].Value != v {
			return false
		}
	}
	if len(tableIdxs) == 0 {
		return true
	}
	picked := make([]int, 0, len(tableIdxs))
	sum := 0
	for _, i := range tableIdxs {
		c := tableCards[i]
		if c == nil || DilotiIsFaceCard(c) {
			return false
		}
		sum += DilotiCardValue(c)
		picked = append(picked, DilotiCardValue(c))
	}
	if v <= 0 || sum%v != 0 {
		return false
	}
	// **合計が割り切れるだけでは足りない。** 1+9 は合計 10 でも 5+5 には
	// 割れないので、5 の捕獲としては不正。
	return dilotiCanPartition(picked, v)
}

// dilotiCanPartition は values を合計 target の束に余さず割れるかを返す。
//
// **合計が割り切れることと束に割れることは別物。** 1+9 は合計 10 だが 5+5 には
// 割れないので、5 の捕獲としては不正になる。
func dilotiCanPartition(values []int, target int) bool {
	if len(values) == 0 {
		return true
	}
	if target <= 0 {
		return false
	}
	used := make([]bool, len(values))
	return dilotiFillBuckets(values, used, target, target, 0, len(values))
}

// dilotiFillBuckets は need を埋めながら束を作る。
// target は束ひとつぶんの目標値、remaining は未使用の枚数。
func dilotiFillBuckets(values []int, used []bool, need, target, start, remaining int) bool {
	if need == 0 {
		if remaining == 0 {
			return true
		}
		// 束がひとつ埋まった。次の束を最初から作り直す。
		return dilotiFillBuckets(values, used, target, target, 0, remaining)
	}
	if remaining == 0 {
		return false
	}
	for i := start; i < len(values); i++ {
		if used[i] || values[i] > need {
			continue
		}
		used[i] = true
		ok := dilotiFillBuckets(values, used, need-values[i], target, i+1, remaining-1)
		used[i] = false
		if ok {
			return true
		}
	}
	return false
}

func dilotiIdxsInRange(idxs []int, n int) bool {
	seen := make(map[int]struct{}, len(idxs))
	for _, i := range idxs {
		if i < 0 || i >= n {
			return false
		}
		if _, dup := seen[i]; dup {
			return false
		}
		seen[i] = struct{}{}
	}
	return true
}

// DilotiDeclCandidate は作れる宣言の候補。
type DilotiDeclCandidate struct {
	// Value は宣言値。
	Value int
	// TableIdxs は巻き込む場札。
	TableIdxs []int
}

// EnumerateDilotiDeclarations は handCard で作れる宣言を列挙する。
//
// **宣言には裏付けが要る。** 宣言値と同じ札を別に手札へ残していなければ
// 宣言できず、宣言した側はそれを捕獲するまで手放せない。値の上限は 10。
func EnumerateDilotiDeclarations(handCard *Card, handIdx int, hand, tableCards []*Card) []DilotiDeclCandidate {
	if handCard == nil || DilotiIsFaceCard(handCard) {
		return nil
	}
	base := DilotiCardValue(handCard)
	out := make([]DilotiDeclCandidate, 0)
	for declared := DilotiMinDeclaration; declared <= DilotiMaxDeclaration; declared++ {
		if declared <= base {
			continue
		}
		if !dilotiHoldsValue(hand, handIdx, declared) {
			continue
		}
		for _, s := range dilotiSubsetsSummingTo(tableCards, declared-base) {
			out = append(out, DilotiDeclCandidate{Value: declared, TableIdxs: append([]int(nil), s...)})
		}
	}
	return out
}

// dilotiHoldsValue は skipIdx 以外の手札に値 v の数札があるかを返す。
func dilotiHoldsValue(hand []*Card, skipIdx, v int) bool {
	for i, c := range hand {
		if i == skipIdx || c == nil || DilotiIsFaceCard(c) {
			continue
		}
		if DilotiCardValue(c) == v {
			return true
		}
	}
	return false
}
