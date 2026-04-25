//go:build test

package domain

// This file contains test helper methods for SpiteAndMalice.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhase フェーズ設定（テスト用）
func (s *SpiteAndMalice) SetPhase(phase SpiteAndMalicePhase) { s.phase = phase }

// SetCurrent 現在プレイヤー設定（テスト用）
func (s *SpiteAndMalice) SetCurrent(idx int) { s.current = idx }

// SetWinner 勝者設定（テスト用）
func (s *SpiteAndMalice) SetWinner(idx int) { s.winner = idx }

// SetStock ストック設定（テスト用）
func (s *SpiteAndMalice) SetStock(cards []*Card) {
	s.stock = append(s.stock[:0], cards...)
}

// SetCompleted 完成済み山設定（テスト用）
func (s *SpiteAndMalice) SetCompleted(cards []*Card) {
	s.completed = append(s.completed[:0], cards...)
}

// SetFoundation ファウンデーション設定（テスト用）
func (s *SpiteAndMalice) SetFoundation(idx int, cards []*Card) {
	if idx < 0 || idx >= SpiteAndMaliceFoundationCnt {
		return
	}
	s.foundations[idx] = append(s.foundations[idx][:0], cards...)
}

// SetPlayerHand 指定プレイヤーの手札を設定（テスト用）
func (s *SpiteAndMalice) SetPlayerHand(idx int, cards []*Card) {
	if idx < 0 || idx >= SpiteAndMalicePlayerCnt {
		return
	}
	s.players[idx].hand = append(s.players[idx].hand[:0], cards...)
}

// SetPlayerGoal 指定プレイヤーのゴール (末尾がトップ) を設定（テスト用）
func (s *SpiteAndMalice) SetPlayerGoal(idx int, cards []*Card) {
	if idx < 0 || idx >= SpiteAndMalicePlayerCnt {
		return
	}
	s.players[idx].goal = append(s.players[idx].goal[:0], cards...)
}

// SetPlayerSide 指定プレイヤーのサイドパイルを設定（テスト用）
func (s *SpiteAndMalice) SetPlayerSide(pIdx, sIdx int, cards []*Card) {
	if pIdx < 0 || pIdx >= SpiteAndMalicePlayerCnt {
		return
	}
	if sIdx < 0 || sIdx >= SpiteAndMaliceSideCnt {
		return
	}
	s.players[pIdx].sides[sIdx] = append(s.players[pIdx].sides[sIdx][:0], cards...)
}

// SetMoveCount 操作回数設定（テスト用）
func (s *SpiteAndMalice) SetMoveCount(n int) { s.moveCount = n }
