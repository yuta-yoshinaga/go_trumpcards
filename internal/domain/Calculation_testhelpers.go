//go:build test

package domain

// This file contains test helper methods for Calculation.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (c *Calculation) SetPhase(p CalculationPhase) { c.phase = p }

// SetFoundations ファンデーション設定（テスト用）
func (c *Calculation) SetFoundations(f [CalculationFoundationCnt][]*Card) { c.foundations = f }

// SetWastes ウェイスト設定（テスト用）
func (c *Calculation) SetWastes(w [CalculationWasteCnt][]*Card) { c.wastes = w }

// SetStock ストック設定（テスト用）
func (c *Calculation) SetStock(s []*Card) { c.stock = s }

// SetIsStalemate 手詰まり状態設定（テスト用）
func (c *Calculation) SetIsStalemate(v bool) { c.isStalemate = v }

// SetMoveCount 移動回数設定（テスト用）
func (c *Calculation) SetMoveCount(n int) { c.moveCount = n }
