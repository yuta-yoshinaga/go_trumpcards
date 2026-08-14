//go:build test

package domain

// This file contains test helper methods for Slobberhannes.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (s *Slobberhannes) SetPhase(phase SlobberhannesPhase) { s.phase = phase }

// SetTrickNumberForTest 現在のトリック番号を設定する（テスト用）
func (s *Slobberhannes) SetTrickNumberForTest(n int) { s.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する（テスト用）
func (s *Slobberhannes) SetCurrentPlayerIdxForTest(i int) { s.currentPlayerIdx = i }

// SetPenaltiesForTest 指定プレイヤーの罰の内訳を設定する（テスト用）
func (s *Slobberhannes) SetPenaltiesForTest(idx int, first, last, queen bool) {
	p := s.players[idx]
	p.tookFirstTrick, p.tookLastTrick, p.tookQueen = first, last, queen
}

// Finish 現在の累計得点で勝敗を確定させる（テスト用）
func (s *Slobberhannes) Finish() { s.finishGame() }

// SetCurrentTrickForTest 場に出ている札を設定する（テスト用）
func (s *Slobberhannes) SetCurrentTrickForTest(t []*TrickCard) { s.currentTrick = t }
