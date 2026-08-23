//go:build test

package domain

import (
	"fmt"
	"testing"
)

// qualifyGame は「ディーラーが qualify しなかったときの精算に特例がある」
// ゲームを 1 ハンド打ち切って、宣言された勝敗と実際のチップ増減を返す。
type qualifyGame struct {
	name string
	// play は 1 ハンド打ち切り、(宣言された勝敗, 収支, 打ち切れたか) を返す。
	play func() (GameResult, int, bool)
}

// TestQualifyResultAgreesWithSettlement は、**画面が言う勝敗と財布の増減が
// 一致する**ことを、qualify 特例を持つ全ゲームについて見る。
//
// # なぜこの形か
//
// これらのゲームには「ディーラーが qualify しなければアンテ 1:1」という特例が
// あり、**手が弱くてもプレイヤーのチップは増える**。勝敗を手の強弱だけで
// 決めると、チップが増えているのに画面が「負け」と言う。
//
// 実測 (#6213, 2026-08-23): 修正前の CaribbeanStud は 4000 ハンド中 402 件
// (10%) が矛盾していた。ThreeCard 191 件、HighCardFlush 178 件。
//
// **個別の局面を組み立てるテストでは足りない。** どの配りで特例が起きるかは
// 事前に分からないので、実際に何千ハンドも打って矛盾が 0 であることを見る。
// 打ち切れたハンド数も検査する —— 0 ハンドしか回っていなければ、この検査は
// 何も見ていない (実際、OasisPoker は交換ステップを飛ばすと 0/0 になった)。
func TestQualifyResultAgreesWithSettlement(t *testing.T) {
	const hands = 2000

	games := []qualifyGame{
		{"CaribbeanStud", func() (GameResult, int, bool) {
			g := NewDefaultCaribbeanStud()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0) != nil || g.Play() != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"CaribbeanDraw", func() (GameResult, int, bool) {
			g := NewDefaultCaribbeanDraw()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0) != nil || g.Draw(nil) != nil || g.Play() != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"OasisPoker", func() (GameResult, int, bool) {
			g := NewDefaultOasisPoker()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0) != nil || g.Exchange(nil) != nil || g.Play() != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"RussianPoker", func() (GameResult, int, bool) {
			g := NewDefaultRussianPoker()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10) != nil || g.Play() != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"ThreeCard", func() (GameResult, int, bool) {
			g := NewDefaultThreeCard()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0) != nil || g.Play() != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"ThreeCardRummy", func() (GameResult, int, bool) {
			g := NewDefaultThreeCardRummy()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0) != nil || g.Play() != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"HighCardFlush", func() (GameResult, int, bool) {
			g := NewDefaultHighCardFlush()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0, 0) != nil || g.Raise(1) != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
		{"UltimateTexasHoldem", func() (GameResult, int, bool) {
			g := NewDefaultUltimateTexasHoldem()
			g.Reset()
			before := g.GetChips()
			if g.Bet(10, 0) != nil {
				return 0, 0, false
			}
			return g.GetResult(), g.GetChips() - before, true
		}},
	}

	for _, g := range games {
		t.Run(g.name, func(t *testing.T) {
			played, mismatch := 0, 0
			var sample string
			for i := 0; i < hands; i++ {
				res, net, ok := g.play()
				if !ok {
					continue
				}
				played++
				if (res == GameResultLose && net > 0) || (res == GameResultWin && net < 0) {
					mismatch++
					if sample == "" {
						sample = fmt.Sprintf("result=%d net=%+d", res, net)
					}
				}
			}

			// **打ち切れていなければ何も見ていない。**
			if played < hands/2 {
				t.Fatalf("only %d/%d hands played to the end — the check is vacuous", played, hands)
			}
			if mismatch > 0 {
				t.Errorf("%d/%d hands announce a result the chips contradict (e.g. %s)",
					mismatch, played, sample)
			}
		})
	}
}
