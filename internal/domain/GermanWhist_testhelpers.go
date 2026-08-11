//go:build test

package domain

// This file contains test helper methods for GermanWhist.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (g *GermanWhist) SetPhase(phase GermanWhistPhase) { g.phase = phase }

// SetUpCard 表向きの札を設定する（テスト用）。nil で「山札が尽きた」状態。
func (g *GermanWhist) SetUpCard(c *Card) { g.upCard = c }

// SetCurrentPlayerIdx 手番を設定する（テスト用）
func (g *GermanWhist) SetCurrentPlayerIdx(i int) { g.currentPlayerIdx = i }

// Finish 現在の後半トリック数で勝敗を確定させる（テスト用）
func (g *GermanWhist) Finish() { g.finish() }
