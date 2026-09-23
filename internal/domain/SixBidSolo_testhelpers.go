//go:build test

package domain

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (s *SixBidSolo) SetPhaseForTest(p SixBidSoloPhase) { s.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (s *SixBidSolo) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(s.GetPlayer(idx), cards)
}

// SetWidowForTest はウィドウを差し替える (テスト専用)。
func (s *SixBidSolo) SetWidowForTest(cards []*Card) { s.widow = cards }

// SetContractForTest は契約を差し替える (テスト専用)。
func (s *SixBidSolo) SetContractForTest(declarer int, kind SixBidSoloBidKind, trumpSuit int) {
	s.declarerIdx = declarer
	s.highBid = &SixBidSoloBid{Player: declarer, Kind: kind}
	s.trumpSuit = trumpSuit
	s.declared = true
}

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (s *SixBidSolo) SetCurrentPlayerForTest(idx int) { s.currentIdx = idx }

// SetDealerForTest は親を差し替える (テスト専用)。
func (s *SixBidSolo) SetDealerForTest(idx int) { s.dealerIdx = idx }

// SetBidPlayerForTest は宣言中の手番を差し替える (テスト専用)。
func (s *SixBidSolo) SetBidPlayerForTest(idx int) { s.bidIdx = idx }

// SetTrickLeaderForTest はリード席を差し替える (テスト専用)。
func (s *SixBidSolo) SetTrickLeaderForTest(idx int) { s.trickLeader = idx }

// SetPointsForTest は取得カード点を差し替える (テスト専用)。
func (s *SixBidSolo) SetPointsForTest(idx, n int) {
	if idx >= 0 && idx < SixBidSoloPlayerCnt {
		s.points[idx] = n
	}
}

// SetTricksWonForTest は取得トリック数を差し替える (テスト専用)。
func (s *SixBidSolo) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < SixBidSoloPlayerCnt {
		s.tricksWon[idx] = n
	}
}

// SetScoreForTest は通算得点を差し替える (テスト専用)。
func (s *SixBidSolo) SetScoreForTest(idx, n int) {
	if idx >= 0 && idx < SixBidSoloPlayerCnt {
		s.scores[idx] = n
	}
}

// SetHandNumberForTest は局番号を差し替える (テスト専用)。
func (s *SixBidSolo) SetHandNumberForTest(n int) { s.handNumber = n }

// FinishHandForTest は精算を走らせる (テスト専用)。
func (s *SixBidSolo) FinishHandForTest() { s.finishHand() }
