//go:build test

package domain

// This file contains test helper methods for CasinoHoldem.
// They exist solely for cross-package test setup and are not part of the
// production game logic.

// ForceResolve runs the showdown logic immediately using the current player,
// dealer, and community card slices without dealing additional cards. Use
// only after seeding all hands deterministically.
func (c *CasinoHoldem) ForceResolve() { c.resolve() }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (c *CasinoHoldem) SetGameEndFlag(flag bool) { c.gameEndFlag = flag }

// SetResult ゲーム結果設定（テスト用）
func (c *CasinoHoldem) SetResult(r GameResult) { c.result = r }

// SetDealerQualify ディーラークオリファイ設定（テスト用）
func (c *CasinoHoldem) SetDealerQualify(q bool) { c.dealerQualify = q }
