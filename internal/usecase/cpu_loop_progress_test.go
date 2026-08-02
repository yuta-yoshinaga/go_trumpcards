//go:build test

package usecase

import (
	"testing"
	"time"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// **ガードは「落ちない」だけでなく「進む」必要がある。**
//
// runCpuTurnsLoop は phase と IsHumanTurn だけを見て CpuPlay を呼び直す。
// 手札の尽きた CPU で CpuPlay が何もせずに戻ると、同じ状態で呼ばれ続けて
// **パニックの代わりにハングする**——というのが #4607 のレビュー指摘だった。
//
// ループを直接回す。インタラクタ越しでは手番ガードや前段のエラーで
// **ループに入らないまま通ってしまい**、検証が空振りする（実際に一度書いて、
// CpuPlay を完全な no-op にしても通ったので気づいた）。
func TestRunCpuTurnsLoopTerminatesWithEmptyHand(t *testing.T) {
	g := domain.NewDefaultGongZhu()
	g.Reset()
	g.SetPhase(domain.GongZhuPhasePlay)
	cpu := g.GetPlayer(1)
	for cpu.GetCardsSize() > 0 {
		cpu.RemoveCard(0)
	}
	g.SetCurrentPlayerIdx(1)

	phases := trickPhases[domain.GongZhuPhase]{
		play:     domain.GongZhuPhasePlay,
		trickEnd: domain.GongZhuPhaseTrickEnd,
		roundEnd: domain.GongZhuPhaseRoundEnd,
		gameEnd:  domain.GongZhuPhaseGameEnd,
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		runCpuTurnsLoop(g, phases)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("CPU ターンのループが 3 秒で終わらなかった: 手札の尽きた席で進行が止まっている")
	}
}
