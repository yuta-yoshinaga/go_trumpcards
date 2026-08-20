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

// stalledBidGame は「ビッドが一度も進まない」ドメインを演じる。
//
// 実物のゲームでは上限に当たらせられない——どのゲームも数十手で phase が変わるので、
// 1000 回に届く前に別の条件で抜けてしまい、**上限の分岐を一度も踏まないまま緑になる**。
// 上限が守っているのは「ドメインが壊れたとき」であって、壊れたドメインは自分で
// 用意するしかない。
type stalledBidGame struct{ bids int }

func (g *stalledBidGame) GetGameEndFlag() bool { return false }
func (g *stalledBidGame) GetPhase() int        { return 1 }
func (g *stalledBidGame) IsHumanBidTurn() bool { return false }

// CpuBid は数えるだけで phase も手番も動かさない。#5416 のドメインバグの形。
func (g *stalledBidGame) CpuBid() { g.bids++ }

// runCpuBidsLoop も runCpuTurnsLoop と同じく、進まないドメインから戻れること。
// 兄弟の runCpuTurnsLoop には #4607 由来のテストがあったが、こちらには無かった
// (PR #5418 のレビュー指摘)。
func TestRunCpuBidsLoopTerminatesWhenBiddingNeverAdvances(t *testing.T) {
	g := &stalledBidGame{}

	done := make(chan struct{})
	go func() {
		defer close(done)
		runCpuBidsLoop(g, 1)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("ビッドのループが 3 秒で終わらなかった: 進まないビッドから戻れていない")
	}

	// **「終わった」だけでは上限が効いた証拠にならない。** 他の条件で抜けたのでは
	// なく上限で抜けたことを、回った回数で押さえる。
	if g.bids != maxCpuTurnsPerCall {
		t.Fatalf("上限で抜けていない: CpuBid が %d 回 (期待 %d)", g.bids, maxCpuTurnsPerCall)
	}
}
