//go:build test

package domain

// SetPhaseForTest はフェーズを設定する
func (c *Cucumber) SetPhaseForTest(p CucumberPhase) { c.phase = p }

// SetCurrentPlayerIdxForTest は現在の手番を設定する
func (c *Cucumber) SetCurrentPlayerIdxForTest(i int) { c.currentPlayerIdx = i }

// SetLeadPlayerIdxForTest はリード席を設定する
func (c *Cucumber) SetLeadPlayerIdxForTest(i int) { c.leadPlayerIdx = i }

// SetCurrentTrickForTest は現在のトリックを設定する
func (c *Cucumber) SetCurrentTrickForTest(tc []*TrickCard) { c.currentTrick = tc }

// GiveHandForTest は指定席の手札を cards ちょうどに置き換える
func (c *Cucumber) GiveHandForTest(playerIdx int, cards ...*Card) {
	p := c.players[playerIdx]
	p.Reset()
	for _, card := range cards {
		p.AddCard(card)
	}
}

// PlayForTest は指定席に 1 枚出させる
func (c *Cucumber) PlayForTest(playerIdx, cardIndex int) error { return c.play(playerIdx, cardIndex) }

// CpuChoiceForTest は CPU が選ぶ手札インデックスを返す
func (c *Cucumber) CpuChoiceForTest(playerIdx int) int { return c.chooseCpuCard(playerIdx) }

// CucumberRankForTest はランクの数値を返す
func CucumberRankForTest(card *Card) int { return cucumberRank(card) }
