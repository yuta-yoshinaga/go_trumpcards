package domain

import "math/bits"

// marriageTypeModel represents equivalent physical cards by their suit/rank
// type. Wild cards have no type and are counted separately.
type marriageTypeModel struct {
	types   []marriageType
	counts  []uint8
	wild    uint8
	points  []int
	melds   []marriageTypeMeld
	lowMask uint64
	byType  [][]int
}

type marriageType struct{ suit, rank, points int }
type marriageTypeMeld struct {
	types       []uint8
	wild        uint8
	seq, pure   bool
	need, delta uint64
}

func newMarriageTypeModel(cards []*Card, wildRank int) (marriageTypeModel, bool) {
	m := marriageTypeModel{}
	index := [5][14]int{}
	for s := range index {
		for r := range index[s] {
			index[s][r] = -1
		}
	}
	for _, c := range cards {
		if marriageIsWild(c, wildRank) {
			m.wild++
			continue
		}
		s, r := c.GetDesign(), c.GetValue()
		if s < 1 || s > 4 || r < 1 || r > 13 {
			return m, false
		}
		i := index[s][r]
		if i < 0 {
			i = len(m.types)
			index[s][r] = i
			m.types = append(m.types, marriageType{suit: s, rank: r, points: marriageCardPoints(c, wildRank)})
			m.counts = append(m.counts, 0)
		}
		m.counts[i]++
		if m.counts[i] > 3 {
			return m, false
		}
	}
	if len(m.types) > 27 || m.wild > 31 {
		return m, false
	}
	m.points = make([]int, len(m.types))
	copy(m.points, func() []int {
		a := make([]int, len(m.types))
		for i, t := range m.types {
			a[i] = t.points
		}
		return a
	}())
	m.generateMelds(index)
	var low uint64
	for i := range m.types {
		low |= uint64(1) << (2 * i)
	}
	for i := range m.melds {
		x := &m.melds[i]
		for _, id := range x.types {
			x.need |= uint64(1) << (2 * int(id))
			x.delta += uint64(1) << (2 * int(id))
		}
		if x.wild != 0 {
			x.delta += uint64(x.wild) << (2 * len(m.types))
		}
	}
	m.byType = make([][]int, len(m.types))
	for mi, x := range m.melds {
		for _, id := range x.types {
			m.byType[id] = append(m.byType[id], mi)
		}
	}
	m.lowMask = low
	return m, true
}

type marriageMemo struct {
	keys []uint64
	vals []int16
	size int
	bits uint8
}

func newMarriageMemo() marriageMemo {
	return marriageMemo{keys: make([]uint64, 1024), vals: make([]int16, 1024), bits: 10}
}
func (h *marriageMemo) get(k uint64) (int16, bool) {
	i := (k * 0x9e3779b97f4a7c15) >> (64 - h.bits)
	for probes := 0; probes < len(h.keys); probes++ {
		if h.keys[i] == 0 {
			return 0, false
		}
		if h.keys[i] == k+1 {
			return h.vals[i], true
		}
		i = (i + 1) & uint64(len(h.keys)-1)
	}
	return 0, false
}
func (h *marriageMemo) put(k uint64, v int16) {
	if (h.size+1)*2 > len(h.keys) {
		h.grow()
	}
	i := (k * 0x9e3779b97f4a7c15) >> (64 - h.bits)
	for h.keys[i] != 0 {
		if h.keys[i] == k+1 {
			h.vals[i] = v
			return
		}
		i = (i + 1) & uint64(len(h.keys)-1)
	}
	h.keys[i] = k + 1
	h.vals[i] = v
	h.size++
}
func (h *marriageMemo) grow() {
	old := *h
	*h = marriageMemo{keys: make([]uint64, len(old.keys)*2), vals: make([]int16, len(old.vals)*2), bits: old.bits + 1}
	for i, key := range old.keys {
		if key != 0 {
			h.put(key-1, old.vals[i])
		}
	}
}
func (m marriageTypeModel) first(key uint64) int {
	nz := (key | key>>1) & m.lowMask
	if nz == 0 {
		return -1
	}
	return bits.TrailingZeros64(nz) / 2
}

func (m *marriageTypeModel) generateMelds(index [5][14]int) {
	add := func(ids []uint8, w uint8, seq, pure bool) {
		m.melds = append(m.melds, marriageTypeMeld{types: ids, wild: w, seq: seq, pure: pure})
	}
	// Sets: choose distinct suits from a rank, with at most one wild.
	for r := 1; r <= 13; r++ {
		ids := make([]uint8, 0, 4)
		for s := 1; s <= 4; s++ {
			if i := index[s][r]; i >= 0 {
				ids = append(ids, uint8(i))
			}
		}
		for k := 3; k <= 4; k++ {
			chooseMarriageTypes(ids, k, func(c []uint8) { add(c, 0, false, false) })
		}
		if m.wild > 0 {
			for total := 3; total <= 4; total++ {
				chooseMarriageTypes(ids, total-1, func(c []uint8) { add(c, 1, false, false) })
			}
		}
	}
	// Runs: windows from A through K-A, rejecting repeated ranks.
	for s := 1; s <= 4; s++ {
		for start := 1; start <= 13; start++ {
			for n := 3; start+n-1 <= 14; n++ {
				ids := make([]uint8, 0, n)
				missing := 0
				var seen uint16
				for r := start; r < start+n && missing <= 1; r++ {
					v := r
					if v == 14 {
						v = 1
					}
					if seen&(1<<v) != 0 {
						missing = 99
						break
					}
					seen |= 1 << v
					i := index[s][v]
					if i < 0 {
						missing++
					} else {
						ids = append(ids, uint8(i))
					}
				}
				if missing > 1 {
					// A longer window from the same start only misses more ranks.
					break
				}
				if missing == 0 {
					add(ids, 0, true, true)
				} else if missing == 1 && m.wild > 0 {
					add(ids, 1, true, false)
				}
			}
		}
	}
}

func chooseMarriageTypes(src []uint8, n int, emit func([]uint8)) {
	var walk func(int, []uint8)
	walk = func(i int, c []uint8) {
		if len(c) == n {
			emit(append([]uint8(nil), c...))
			return
		}
		if len(src)-i < n-len(c) {
			return
		}
		for j := i; j < len(src); j++ {
			walk(j+1, append(c, src[j]))
		}
	}
	walk(0, nil)
}

func (m marriageTypeModel) pack(c []uint8, w uint8) uint64 {
	var k uint64
	for i, v := range c {
		k |= uint64(v) << (2 * i)
	}
	k |= uint64(w) << (2 * len(c))
	return k
}
func (m marriageTypeModel) initial() uint64 { return m.pack(m.counts, m.wild) }
