//go:build test

package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **手札の尽きた席に手番が回ってもサーバを落とさない。**
//
// 各ゲームの cpuSelectPlayCard は「出せる札が無い」とき 0 を返し、
// Player.RemoveCard は範囲外で nil を返す (Player.go:89-96)。この 2 つが
// 噛み合うと playCard が nil を触り、**HTTP ハンドラごとパニックする**。
// E2E が GongZhu で実際にサーバを落とした (#4606) のが発端で、同じ形は
// 53 ゲームにあった。
//
// 代表として、発生源 (GongZhu) と、別系統の 4 人・2 人ゲームを見る。
func TestCpuPlayWithEmptyHandDoesNotPanic(t *testing.T) {
	cases := []struct {
		name  string
		drive func()
	}{
		{"GongZhu", func() {
			g := domain.NewDefaultGongZhu()
			g.Reset()
			g.SetPhase(domain.GongZhuPhasePlay)
			emptyHand(g.GetPlayer(1).GetCardsSize(), func(i int) { g.GetPlayer(1).RemoveCard(0) })
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Hearts", func() {
			g := domain.NewDefaultHearts()
			g.Reset()
			g.SetPhase(domain.HeartsPhasePlay)
			emptyHand(g.GetPlayer(1).GetCardsSize(), func(i int) { g.GetPlayer(1).RemoveCard(0) })
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Spades", func() {
			g := domain.NewDefaultSpades()
			g.Reset()
			g.SetPhase(domain.SpadesPhasePlay)
			emptyHand(g.GetPlayer(1).GetCardsSize(), func(i int) { g.GetPlayer(1).RemoveCard(0) })
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
		{"Bezique", func() {
			g := domain.NewDefaultBezique()
			g.Reset()
			g.SetPhase(domain.BeziquePhasePlay)
			emptyHand(g.GetPlayer(1).GetCardsSize(), func(i int) { g.GetPlayer(1).RemoveCard(0) })
			g.SetCurrentPlayerIdx(1)
			g.CpuPlay()
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("手札の無い CPU でパニックした: %v", r)
				}
			}()
			c.drive()
		})
	}
}

// emptyHand n 回 remove を呼んで手札を空にする。
func emptyHand(n int, remove func(int)) {
	for i := 0; i < n; i++ {
		remove(i)
	}
}
