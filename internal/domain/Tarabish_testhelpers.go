//go:build test

package domain

import "errors"

// errPhase はテスト用ヘルパが誤ったフェーズで呼ばれたときのエラー。
var errPhase = errors.New("not the bidding phase")

// This file contains test helper methods for Tarabish.
// They exist solely for cross-package test setup (e.g. presenter tests)
// and are not part of the production game logic.

// SetPhaseForTest フェーズを設定する（テスト用）
func (t *Tarabish) SetPhaseForTest(phase TarabishPhase) { t.phase = phase }

// SetTrickNumberForTest 現在のトリック番号を設定する（テスト用）
func (t *Tarabish) SetTrickNumberForTest(n int) { t.trickNumber = n }

// SetCurrentPlayerIdxForTest 手番を設定する（テスト用）
func (t *Tarabish) SetCurrentPlayerIdxForTest(i int) { t.currentPlayerIdx = i }

// SetCurrentTrickForTest 場に出ている札を設定する（テスト用）
func (t *Tarabish) SetCurrentTrickForTest(tc []*TrickCard) { t.currentTrick = tc }

// SetTrumpSuitForTest 切り札を設定する（テスト用）
func (t *Tarabish) SetTrumpSuitForTest(suit int) { t.trumpSuit = suit }

// SetRoundPointsForTest チームの現ラウンド点を設定する（テスト用）
func (t *Tarabish) SetRoundPointsForTest(team, n int) {
	if team >= 0 && team < TarabishTeamCnt {
		t.roundPoints[team] = n
	}
}

// CountMeldsForTest 配り札からメルドを判定する（テスト用）
func (t *Tarabish) CountMeldsForTest() { t.countMelds() }

// ResolveTrickForTest トリックを解決する（テスト用）
func (t *Tarabish) ResolveTrickForTest() { t.resolveTrick() }

// FinishRoundForTest ラウンドを集計する（テスト用）
func (t *Tarabish) FinishRoundForTest() { t.finishRound() }

// SetDealerIdxForTest 親を設定する（テスト用）
func (t *Tarabish) SetDealerIdxForTest(i int) { t.dealerIdx = i }

// FinishGameForTest 現在のチーム得点で勝敗を確定させる（テスト用）
func (t *Tarabish) FinishGameForTest() { t.finishGame() }

// PassTrumpForTest 指定席が見送る（テスト用。人間以外の見送りを再現する）
func (t *Tarabish) PassTrumpForTest(idx int) error {
	if t.phase != TarabishPhaseBid {
		return errPhase
	}
	t.currentPlayerIdx = idx
	t.advanceBid()
	return nil
}
