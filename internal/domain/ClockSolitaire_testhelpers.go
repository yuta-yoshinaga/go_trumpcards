//go:build test

package domain

// This file contains test helper methods for ClockSolitaire.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (cs *ClockSolitaire) SetPhase(phase ClockSolitairePhase) { cs.phase = phase }

// SetPiles パイル設定（テスト用）
func (cs *ClockSolitaire) SetPiles(piles [ClockSolitairePileCount][]*ClockSolitaireCard) {
	cs.piles = piles
}

// SetFaceUpCount 表向き枚数設定（テスト用）
func (cs *ClockSolitaire) SetFaceUpCount(counts [ClockSolitairePileCount]int) {
	cs.faceUpCount = counts
}

// SetCurrentCard 現在のカード設定（テスト用）
func (cs *ClockSolitaire) SetCurrentCard(card *Card) { cs.currentCard = card }

// SetStepCount ステップ数設定（テスト用）
func (cs *ClockSolitaire) SetStepCount(count int) { cs.stepCount = count }
