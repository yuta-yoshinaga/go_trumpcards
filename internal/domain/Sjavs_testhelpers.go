//go:build test

package domain

// SetTrumpSuitForTest はテスト用に切札を差し替える。
func (s *Sjavs) SetTrumpSuitForTest(suit int) { s.trumpSuit = suit }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (s *Sjavs) SetPhaseForTest(p SjavsPhase) { s.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (s *Sjavs) SetCurrentPlayerForTest(idx int) { s.currentIdx = idx }

// SetBidderForTest はテスト用にビッド者を差し替える。
func (s *Sjavs) SetBidderForTest(idx int) { s.bidderIdx = idx }

// SetTeamPointsForTest はテスト用にチーム得点を差し替える。
func (s *Sjavs) SetTeamPointsForTest(a, b int) { s.points = []int{a, b} }

// SetTricksWonForTest はテスト用にトリック数を差し替える。
func (s *Sjavs) SetTricksWonForTest(won []int) { s.tricksWon = won }

// SetRemainingForTest はテスト用に残り点を差し替える。
func (s *Sjavs) SetRemainingForTest(a, b int) { s.remaining = []int{a, b} }

// SettleHandForTest はテスト用に精算を走らせる。
func (s *Sjavs) SettleHandForTest() { s.settleHand() }
