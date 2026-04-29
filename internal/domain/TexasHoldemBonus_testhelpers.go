//go:build test

package domain

// This file contains test helper methods for TexasHoldemBonus.
// They exist solely for cross-package test setup and are not part of the
// production game logic.

// ForceResolve runs the showdown logic immediately using the current player,
// dealer, and community card slices without dealing additional cards. Use
// only after seeding all hands deterministically.
func (t *TexasHoldemBonus) ForceResolve() { t.resolve() }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (t *TexasHoldemBonus) SetGameEndFlag(flag bool) { t.gameEndFlag = flag }

// SetResult ゲーム結果設定（テスト用）
func (t *TexasHoldemBonus) SetResult(r GameResult) { t.result = r }
