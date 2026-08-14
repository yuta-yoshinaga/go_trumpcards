//go:build test

package domain

// SetCurrentIdxForTest は現在の手番を設定する
func (b *Bhabhi) SetCurrentIdxForTest(i int) { b.currentIdx = i }

// SetLeadIdxForTest はリードプレイヤーを設定する
func (b *Bhabhi) SetLeadIdxForTest(i int) { b.leadIdx = i }

// SetLeadSuitForTest はリードスートを設定する
func (b *Bhabhi) SetLeadSuitForTest(suit int) { b.leadSuit = suit }

// SetPileForTest は場札を設定する
func (b *Bhabhi) SetPileForTest(pile []*TrickCard) { b.pile = pile }

// SetTrickNumberForTest はトリック数を設定する
func (b *Bhabhi) SetTrickNumberForTest(n int) { b.trickNumber = n }

// PlayForTest は指定プレイヤーに 1 枚出させる
func (b *Bhabhi) PlayForTest(playerIdx, cardIndex int) error { return b.play(playerIdx, cardIndex) }

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (b *Bhabhi) CpuChoiceForTest(playerIdx int) int { return b.chooseCpuCard(playerIdx) }

// FinishGameForTest は終局処理を直接呼ぶ
func (b *Bhabhi) FinishGameForTest() { b.finishGame() }

// TrickWinnerForTest はトリックの勝者を返す
func (b *Bhabhi) TrickWinnerForTest() int { return b.trickWinner() }

// bhabhiHandOf は playerIdx の手札を cards ちょうどに置き換える。
//
// **配りの上に積んではいけない。** Reset() のあとに足すと山から来た札が
// 残り、狙った形にならないうえに配り依存で落ちるテストになる。
func bhabhiHandOf(b *Bhabhi, playerIdx int, cards ...*Card) {
	p := b.GetPlayer(playerIdx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// FinishStalemateForTest は膠着での打ち切りを直接呼ぶ
func (b *Bhabhi) FinishStalemateForTest() { b.finishStalemate() }
