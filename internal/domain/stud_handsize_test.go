//go:build test

package domain

import "testing"

// TestStudRunOutDealsExactlyTheGamesHandSize は、オールインの走り切りで
// **名前どおりの枚数**しか配らないことを見る。
//
// # なぜこの形か
//
// `advancePhase` はフェーズを進めるのと同時にそのストリートの札を配るので、
// `s.phase` のストリートは既に配り終えている。走り切りを `s.phase` から
// 回すと 1 枚多く配る。
//
// **どのストリートで全員オールインになるかは配り次第**なので、局面を
// 組み立てるテストでは踏めない。実際に何千ハンドも配って、1 件も超えない
// ことを見る。実測 (#6214, 修正前): SevenCardStud 318/2000、
// FiveCardStud 162/2000。
//
// Razz と Stud Hi-Lo は SevenCardStud と同じ実装を共有する。
func TestStudRunOutDealsExactlyTheGamesHandSize(t *testing.T) {
	const hands = 2000

	cases := []struct {
		name string
		max  int
		// runOut は 1 ハンド配って走り切り、席 0 の総枚数を返す。
		runOut func() int
	}{
		{"SevenCardStud", 7, func() int {
			s := NewDefaultSevenCardStud()
			_ = s.Reset()
			s.dealRemainingStreets()
			return len(s.GetPlayer(0).GetAllCards())
		}},
		{"FiveCardStud", 5, func() int {
			s := NewDefaultFiveCardStud()
			_ = s.Reset()
			s.dealRemainingStreets()
			return len(s.GetPlayer(0).GetAllCards())
		}},
		{"FollowTheQueen", 7, func() int {
			s := NewDefaultFollowTheQueen()
			_ = s.Reset()
			s.dealRemainingStreets()
			return len(s.GetPlayer(0).GetAllCards())
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			over, under, worst := 0, 0, 0
			for i := 0; i < hands; i++ {
				n := c.runOut()
				switch {
				case n > c.max:
					over++
					if n > worst {
						worst = n
					}
				case n < c.max:
					// **少なすぎるのも同じくらい悪い。** s.phase+1 に直した
					// あとで switch のケースを取りこぼすとこちらに倒れる。
					under++
				}
			}
			if over > 0 {
				t.Errorf("%d/%d hands dealt more than %d cards (worst %d)", over, hands, c.max, worst)
			}
			if under > 0 {
				t.Errorf("%d/%d hands dealt fewer than %d cards", under, hands, c.max)
			}
		})
	}
}
