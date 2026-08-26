//go:build test

package domain

// This file contains test helper methods for DramahaPlayer.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetBestHand ベストハンド設定（テスト用）
func (op *DramahaPlayer) SetBestHand(hand []*Card) { op.bestHand = hand }

// SetDrawBestHand はドロー側の 5 枚を設定する（テスト用）。役位は渡された
// 手札から計算する —— 手札と役位を別々に渡せると、片方だけ設定したテストが
// 「役位は強いのに札は弱い」矛盾した盤面を作れてしまう。
func (op *DramahaPlayer) SetDrawBestHand(hand []*Card) {
	op.drawBestHand = hand
	if len(hand) == DramahaHoleCards {
		op.drawRank = evalFiveCardHand(hand)
		return
	}
	op.drawRank = PokerHandHighCard
}
