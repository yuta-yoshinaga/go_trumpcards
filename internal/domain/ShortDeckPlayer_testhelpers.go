//go:build test

package domain

// This file contains test helper methods for ShortDeckPlayer.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetBestHand ベストハンド設定（テスト用）
func (sp *ShortDeckPlayer) SetBestHand(hand []*Card) { sp.bestHand = hand }
