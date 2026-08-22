//go:build test

package domain

// Test helper methods for Ristikontra. They exist solely for cross-package test setup
// (e.g. presenter tests) and are not part of the production game logic.

// SetPile 場の山を設定する (テスト用)。
func (g *Ristikontra) SetPile(cards []*Card) { g.state.pile = cards }

// SetGameEndFlag ゲーム終了フラグを設定する (テスト用)。
func (g *Ristikontra) SetGameEndFlag(flag bool) { g.state.gameEndFlag = flag }

// SetCurrentTurn 現在の手番を設定する (テスト用)。
func (g *Ristikontra) SetCurrentTurn(idx int) { g.state.currentTurn = idx }

// SetLastCaptureIdx 最後の捕獲者を設定する (テスト用)。
func (g *Ristikontra) SetLastCaptureIdx(idx int) { g.state.lastCaptureIdx = idx }

// SetWinners 勝者リストを設定する (テスト用)。
func (g *Ristikontra) SetWinners(w []int) { g.state.winners = w }
