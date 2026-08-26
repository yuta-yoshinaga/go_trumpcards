//go:build test

package domain

// Ramsch のテスト専用ヘルパ。本番のゲーム進行には使わない。

// SetHandForTest は playerIdx の手札を丸ごと入れ替える（テスト用）。
func (s *Ramsch) SetHandForTest(playerIdx int, cards []*Card) {
	p := s.players[playerIdx]
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// SetSkatForTest は伏せ札 2 枚を固定する（テスト用）。
func (s *Ramsch) SetSkatForTest(cards []*Card) { s.round.skat = cards }

// SetTrickNumberForTest はトリック番号を設定する（テスト用）。
func (s *Ramsch) SetTrickNumberForTest(n int) { s.round.trickNumber = n }

// SetCardPointsForTest は集計済みの点を直接設定する（テスト用）。
// ScoreRound の分岐を、10 トリック配り切らずに検査するために使う。
func (s *Ramsch) SetCardPointsForTest(pts [RamschPlayerCnt]int) {
	for i, n := range pts {
		s.players[i].SetCardPoints(n)
	}
}

// SetDurchmarschForTest は Durchmarsch 状態を直接設定する（テスト用）。
func (s *Ramsch) SetDurchmarschForTest(idx int) {
	s.round.durchmarsch = idx >= 0
	s.round.durchmarschIdx = idx
}

// SetCurrentTrickForTest は場のトリックを固定する（テスト用）。
func (s *Ramsch) SetCurrentTrickForTest(trick []*TrickCard) { s.round.currentTrick = trick }

// SetLeadPlayerIdxForTest はリード役を設定する（テスト用）。
func (s *Ramsch) SetLeadPlayerIdxForTest(idx int) { s.round.leadPlayerIdx = idx }

// TrickWinnerForTest は現在のトリックの勝者を返す（テスト用）。
func (s *Ramsch) TrickWinnerForTest() int { return s.trickWinner() }

// CpuPickPlayForTest は CPU の選択を返す（テスト用）。
func (s *Ramsch) CpuPickPlayForTest(playerIdx int) int { return s.cpuPickPlay(playerIdx) }

// ValidatePlayForTest は着手の合法性を返す（テスト用）。
func (s *Ramsch) ValidatePlayForTest(playerIdx int, card *Card) error {
	return s.validatePlay(playerIdx, card)
}

// IsWildTrumpForTest はその札が切り札かを返す（テスト用）。
func (s *Ramsch) IsWildTrumpForTest(c *Card) bool { return s.isTrump(c) }

// CardStrengthForTest は札の強さを返す（テスト用）。
func (s *Ramsch) CardStrengthForTest(c *Card) int { return s.cardStrength(c) }
