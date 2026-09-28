//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (z *Zwicker) SetPhaseForTest(p ZwickerPhase) { z.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (z *Zwicker) SetCurrentPlayerForTest(idx int) { z.currentIdx = idx }

// SetTableCardsForTest はテスト用に場を差し替える。
func (z *Zwicker) SetTableCardsForTest(cards []*Card) { z.tableCards = cards }

// SetTeamScoreForTest はテスト用に累計得点を差し替える。
func (z *Zwicker) SetTeamScoreForTest(team, score int) { z.scores[team] = score }

// SetDealStageForTest はテスト用に配り段階を差し替える。
func (z *Zwicker) SetDealStageForTest(stage int) { z.dealStage = stage }
