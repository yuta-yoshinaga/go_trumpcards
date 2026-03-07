//go:build test

package domain

// This file contains test helper methods for BlackJack.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (b *BlackJack) SetPhase(phase int) {
	b.phase = phase
	b.gameEndFlag = phase == BJPhaseEnd
	if phase == BJPhaseInsurance {
		b.insuranceAvailable = true
	}
}

// SetPlayerHands プレイヤーハンド設定（テスト用）
func (b *BlackJack) SetPlayerHands(hands []*BlackJackHand) {
	b.playerHands = hands
}

// SetCurrentHandIdx 現在操作中のハンドインデックス設定（テスト用）
func (b *BlackJack) SetCurrentHandIdx(idx int) {
	b.currentHandIdx = idx
}
