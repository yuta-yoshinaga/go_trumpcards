//go:build !js || !wasm || extra2

package domain

// continentalRummyNodeBudget は 1 回の探索で開いてよい節点の上限。
//
// **上限の無い探索を Worker に持ち込まない。** 手元では「少し重い」だけの
// 総当たりが無料枠の 10ms を軽く超えることがある (#5462 の隣で同じ轍を踏まない)。
// 15 枚・3 レイアウトなら実測で 3 桁の節点しか開かないので、4 桁の壁は
// 正しい手を取りこぼさずに暴走だけを止める。
const continentalRummyNodeBudget = 20000

// FindContinentalRummyGoOut は手札 15 枚を認められた形に分けられるかを調べ、
// 分けられるならその 1 通り (札のインデックスの組) を返す。
//
// **部分メルドは無い。** この形式は途中で場に出すことができず、15 枚を一度に
// 並べて上がるか、何も出さないかのどちらか。だから「上がれるか」は毎回
// 手札全体の分割問題になる。
//
// 探索は次の 3 点で刈る。
//
//   - **起点は必ず一番小さい未使用の素札。** どの組から作るかの順序を固定
//     すると、同じ分割を数え直さずに済む。
//   - **候補は起点と同じスートの札とジョーカーだけ。** 別スートは 1 本の
//     シーケンスに入らないので、そもそも組に含めない。
//   - **起点の組の大きさは決め打ちにしない。** 残っている大きさを順に試す。
//     ここを「並べ替えた表の先頭」に固定すると、一番小さい番号の札が最大の
//     組に入る手しか見つからず、**合法な 5+4+3+3 を落とす** ── ♠2-3-4 が
//     先頭に来る手はどう組んでも起点が 3 枚組なので、size-5 を強いた瞬間に
//     詰む。
func FindContinentalRummyGoOut(cards []*Card) ([][]int, bool) {
	if len(cards) != ContinentalRummyHandSize {
		return nil, false
	}
	for _, layout := range continentalRummyLayouts {
		s := &continentalSolver{cards: cards, seen: map[uint64]bool{}, budget: continentalRummyNodeBudget}
		for _, n := range layout {
			s.remaining[n-ContinentalRummyMinRun]++
		}
		if groups, ok := s.solve(0); ok {
			return groups, true
		}
	}
	return nil, false
}

// CanContinentalRummyGoOut は上がれるかどうかだけを返す。
func CanContinentalRummyGoOut(cards []*Card) bool {
	_, ok := FindContinentalRummyGoOut(cards)
	return ok
}

// continentalSolver は 1 つのレイアウトについての分割探索。
//
// remaining は「まだ作っていない組の大きさ」を数えたもの。添字 0 が 3 枚組、
// 1 が 4 枚組、2 が 5 枚組。**順序ではなく個数で持つ**ので、どの大きさから
// 作るかを起点ごとに選び直せる。
type continentalSolver struct {
	cards     []*Card
	remaining [ContinentalRummyMaxRun - ContinentalRummyMinRun + 1]int
	seen      map[uint64]bool
	budget    int
}

// stateKey は (使用済みビット, 残りの組の内訳) を 1 つの鍵にする。
func (s *continentalSolver) stateKey(used uint32) uint64 {
	k := uint64(used)
	for i, n := range s.remaining {
		k |= uint64(n) << (32 + uint(i)*8)
	}
	return k
}

func (s *continentalSolver) solve(used uint32) ([][]int, bool) {
	if s.remaining == [ContinentalRummyMaxRun - ContinentalRummyMinRun + 1]int{} {
		return [][]int{}, true
	}
	if s.budget <= 0 {
		return nil, false
	}
	s.budget--

	key := s.stateKey(used)
	if s.seen[key] {
		return nil, false
	}

	anchor := s.lowestFreeNonJoker(used)
	if anchor < 0 {
		// 残りがジョーカーだけ。何のスートの何の並びか決まらないので上がれない。
		s.seen[key] = true
		return nil, false
	}
	pool := s.candidatesFor(anchor, used)

	// **残っている大きさを順に試す。** 決め打ちにすると合法な手を落とす。
	for size := ContinentalRummyMinRun; size <= ContinentalRummyMaxRun; size++ {
		slot := size - ContinentalRummyMinRun
		if s.remaining[slot] == 0 {
			continue
		}
		s.remaining[slot]--
		if got, ok := s.tryGroup(anchor, size, pool, used); ok {
			s.remaining[slot]++
			return got, true
		}
		s.remaining[slot]++
	}

	s.seen[key] = true
	return nil, false
}

// tryGroup は起点を含む size 枚の組を総当たりし、残りが解けるものを返す。
func (s *continentalSolver) tryGroup(anchor, size int, pool []int, used uint32) ([][]int, bool) {
	need := size - 1
	combo := make([]int, 0, need)
	group := make([]*Card, 0, size)

	var walk func(start int) ([][]int, bool)
	walk = func(start int) ([][]int, bool) {
		if len(combo) == need {
			group = group[:0]
			group = append(group, s.cards[anchor])
			for _, i := range combo {
				group = append(group, s.cards[i])
			}
			if !IsContinentalRummyRun(group) {
				return nil, false
			}
			next := used | 1<<uint(anchor)
			for _, i := range combo {
				next |= 1 << uint(i)
			}
			rest, ok := s.solve(next)
			if !ok {
				return nil, false
			}
			taken := append([]int{anchor}, combo...)
			return append([][]int{taken}, rest...), true
		}
		for p := start; p < len(pool); p++ {
			// 残りの候補が足りないなら打ち切る。
			if len(pool)-p < need-len(combo) {
				return nil, false
			}
			combo = append(combo, pool[p])
			if got, ok := walk(p + 1); ok {
				return got, true
			}
			combo = combo[:len(combo)-1]
		}
		return nil, false
	}
	return walk(0)
}

// lowestFreeNonJoker は未使用の素札のうち一番小さいインデックスを返す。
func (s *continentalSolver) lowestFreeNonJoker(used uint32) int {
	for i, c := range s.cards {
		if used&(1<<uint(i)) == 0 && !IsContinentalRummyJoker(c) {
			return i
		}
	}
	return -1
}

// candidatesFor は起点と 1 本になりうる未使用札のインデックスを返す。
func (s *continentalSolver) candidatesFor(anchor int, used uint32) []int {
	suit := s.cards[anchor].GetDesign()
	out := make([]int, 0, len(s.cards))
	for i, c := range s.cards {
		if i == anchor || used&(1<<uint(i)) != 0 {
			continue
		}
		if IsContinentalRummyJoker(c) || c.GetDesign() == suit {
			out = append(out, i)
		}
	}
	return out
}
