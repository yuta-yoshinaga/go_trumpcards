//go:build !js || !wasm || extra2

package domain

import "sort"

// FindContinentalRummyGoOut は手札 15 枚を認められた形に分けられるかを調べ、
// 分けられるならその 1 通り (札のインデックスの組) を返す。
//
// **部分メルドは無い。** この形式は途中で場に出すことができず、15 枚を一度に
// 並べて上がるか、何も出さないかのどちらか。だから「上がれるか」は毎回
// 手札全体の分割問題になる。
//
// 探索は次の 2 点で刈る。Worker の CPU 予算は手元では見えないので、
// 総当たりのままでは持ち込めない (#5462 の隣で同じ轍を踏まないこと)。
//
//   - **起点は必ず一番小さい未使用の素札。** どの組から作るかの順序を固定
//     すると、同じ分割を数え直さずに済む。
//   - **候補は起点と同じスートの札とジョーカーだけ。** 別スートは 1 本の
//     シーケンスに入らないので、そもそも組に含めない。
//
// そのうえで (使用済みビット, 何組目) を記憶して同じ盤面を二度解かない。
func FindContinentalRummyGoOut(cards []*Card) ([][]int, bool) {
	if len(cards) != ContinentalRummyHandSize {
		return nil, false
	}
	for _, layout := range continentalRummyLayouts {
		sizes := append([]int(nil), layout...)
		sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
		s := &continentalSolver{cards: cards, sizes: sizes, seen: map[uint64]bool{}}
		if groups, ok := s.solve(0, 0); ok {
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
type continentalSolver struct {
	cards []*Card
	sizes []int
	seen  map[uint64]bool
}

func (s *continentalSolver) solve(used uint32, sizeIdx int) ([][]int, bool) {
	if sizeIdx == len(s.sizes) {
		return [][]int{}, true
	}
	key := uint64(used) | uint64(sizeIdx)<<32
	if s.seen[key] {
		return nil, false
	}

	anchor := s.lowestFreeNonJoker(used)
	if anchor < 0 {
		// 残りがジョーカーだけ。何のスートの何の並びか決まらないので上がれない。
		s.seen[key] = true
		return nil, false
	}

	need := s.sizes[sizeIdx] - 1
	pool := s.candidatesFor(anchor, used)
	combo := make([]int, 0, need)
	group := make([]*Card, 0, need+1)

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
			rest, ok := s.solve(next, sizeIdx+1)
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

	if got, ok := walk(0); ok {
		return got, true
	}
	s.seen[key] = true
	return nil, false
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
