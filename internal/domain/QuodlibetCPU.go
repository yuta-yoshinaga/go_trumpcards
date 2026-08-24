//go:build !js || !wasm || solo

package domain

// QuodlibetCPU.go は CPU の手番の方策。
//
// **コントラクトごとに「良い手」が正反対になる。** プラスは取りに行き、
// マイナスや第 3 の輪は取らないように逃げ、アラリックや赤なしは特定の札を
// 引き受けないようにする ── ひとつの評価関数では書けないので、コントラクト
// から「取りたいか」と「札 1 枚の危険度」を引く形にしている。

// quodlibetWantsTricks は取りに行くコントラクトかを返す。
func quodlibetWantsTricks(contract int) bool {
	return contract == QuodlibetPlus
}

// quodlibetCardRisk はそのコントラクトで札 1 枚が背負う罰点を返す。
//
// 罰点が札に紐づかないコントラクト (プラス / マイナスなど) では 0 を返す。
func quodlibetCardRisk(contract int, c *Card) int {
	if c == nil {
		return 0
	}
	switch contract {
	case QuodlibetAlarich:
		if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
			return QuodlibetKingOfHeartsPenalty
		}
		if c.GetDesign() == CardDesignDiamond && c.GetValue() == 12 {
			return QuodlibetRedRuffianPenalty
		}
	case QuodlibetNoReds:
		if c.GetDesign() == CardDesignHeart {
			return quodlibetHeartPenalty(c)
		}
	case QuodlibetOberUnter:
		switch c.GetValue() {
		case 12:
			return QuodlibetQueenPenalty
		case 11:
			return QuodlibetJackPenalty
		}
	}
	return 0
}

// cpuTrickChoice はトリック系コントラクトで CPU が出す札を選ぶ。
func (q *Quodlibet) cpuTrickChoice(playerIdx int) int {
	valid := q.GetPlayableIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if q.config.CpuDifficulty == QuodlibetCpuDifficultyEasy {
		return valid[quodlibetRandIntn(len(valid))]
	}
	return q.quodlibetSmartTrickChoice(playerIdx, valid)
}

// quodlibetSmartTrickChoice は罰点を避ける (あるいは取りに行く) 札を選ぶ。
func (q *Quodlibet) quodlibetSmartTrickChoice(playerIdx int, valid []int) int {
	p := q.players[playerIdx]
	// **危険な札はまず手放す。** ただし今のトリックを取ってしまう札は除く。
	if idx := q.dumpRiskyCard(playerIdx, valid); idx >= 0 {
		return idx
	}

	winning, losing := -1, -1
	for _, i := range valid {
		if q.wouldWinTrick(p.GetCard(i)) {
			if winning < 0 || QuodlibetCardStrength(p.GetCard(i)) > QuodlibetCardStrength(p.GetCard(winning)) {
				winning = i
			}
			continue
		}
		if losing < 0 || QuodlibetCardStrength(p.GetCard(i)) > QuodlibetCardStrength(p.GetCard(losing)) {
			losing = i
		}
	}

	if quodlibetWantsTricks(q.currentContract) {
		if winning >= 0 {
			return winning
		}
		return q.lowestOf(playerIdx, valid)
	}
	// 取りたくないコントラクト: 負ける中で一番高い札を捨て、無ければ最安を出す。
	if losing >= 0 {
		return losing
	}
	return q.lowestOf(playerIdx, valid)
}

// dumpRiskyCard は「取らずに済む危険な札」を返す (無ければ -1)。
func (q *Quodlibet) dumpRiskyCard(playerIdx int, valid []int) int {
	if quodlibetWantsTricks(q.currentContract) {
		return -1
	}
	p := q.players[playerIdx]
	best, bestRisk := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		risk := quodlibetCardRisk(q.currentContract, c)
		if risk <= bestRisk || q.wouldWinTrick(c) {
			continue
		}
		best, bestRisk = i, risk
	}
	return best
}

// wouldWinTrick は今その札を出したら現時点のトリックを取るかを返す。
//
// **後ろにまだ人が残っていても「取る」と数える。** 残りが被せてくれるかは
// 分からないので、危険を引き受けない側に倒す。
func (q *Quodlibet) wouldWinTrick(c *Card) bool {
	if c == nil {
		return false
	}
	if len(q.currentTrick) == 0 {
		return true
	}
	lead := q.currentTrick[0].Card.GetDesign()
	if c.GetDesign() != lead {
		return false
	}
	best := -1
	for _, tc := range q.currentTrick {
		if tc == nil || tc.Card == nil || tc.Card.GetDesign() != lead {
			continue
		}
		if s := QuodlibetCardStrength(tc.Card); s > best {
			best = s
		}
	}
	return QuodlibetCardStrength(c) > best
}

// lowestOf は候補のうち最も弱い札のインデックスを返す。
func (q *Quodlibet) lowestOf(playerIdx int, valid []int) int {
	p := q.players[playerIdx]
	best := valid[0]
	for _, i := range valid[1:] {
		if QuodlibetCardStrength(p.GetCard(i)) < QuodlibetCardStrength(p.GetCard(best)) {
			best = i
		}
	}
	return best
}

// cpuSheddingChoice はシェディング系で CPU が出す札を選ぶ (-1 = パス)。
func (q *Quodlibet) cpuSheddingChoice(playerIdx int) int {
	valid := q.GetSheddingPlayableIndices(playerIdx)
	if len(valid) == 0 {
		return -1
	}
	if q.config.CpuDifficulty == QuodlibetCpuDifficultyEasy {
		return valid[quodlibetRandIntn(len(valid))]
	}
	return q.smartSheddingChoice(playerIdx, valid)
}

// smartSheddingChoice は難易度に関わらず「良い」シェディングの手を返す。
//
// **端の札から出す。** 真ん中を先に出すと自分の連なりが途切れる。
func (q *Quodlibet) smartSheddingChoice(playerIdx int, valid []int) int {
	p := q.players[playerIdx]
	best := valid[0]
	for _, i := range valid[1:] {
		if quodlibetDistanceFromAnchor(p.GetCard(i)) > quodlibetDistanceFromAnchor(p.GetCard(best)) {
			best = i
		}
	}
	return best
}

// quodlibetDistanceFromAnchor は小食いの起点 (J) からの距離を返す。
func quodlibetDistanceFromAnchor(c *Card) int {
	if c == nil {
		return 0
	}
	d := QuodlibetRankIndex(c.GetValue()) - QuodlibetSnackAnchorIndex
	if d < 0 {
		return -d
	}
	return d
}

// cpuPickContract は CPU のディーラーが選ぶコントラクトを返す。
func (q *Quodlibet) cpuPickContract() int {
	return q.cpuPickContractFor(q.dealerIdx)
}

// cpuPickContractFor は席 playerIdx の手札から選ぶコントラクトを返す。
//
// **選ぶ人にとって一番安く済む種目を選ぶ。** 危険札を抱えている輪では、
// その札が効かない種目へ逃げるのが自然な打ち方。
func (q *Quodlibet) cpuPickContractFor(playerIdx int) int {
	avail := q.GetAvailableContracts()
	if len(avail) == 0 {
		return -1
	}
	if len(avail) == 1 || q.config.CpuDifficulty == QuodlibetCpuDifficultyEasy {
		return avail[quodlibetRandIntn(len(avail))]
	}
	return q.smartPickContractFor(playerIdx)
}

// smartPickContractFor は難易度に関わらず、その席にとって一番安い種目を返す。
func (q *Quodlibet) smartPickContractFor(playerIdx int) int {
	avail := q.GetAvailableContracts()
	if len(avail) == 0 {
		return -1
	}
	p := q.players[playerIdx]
	best, bestRisk := avail[0], -1
	for _, c := range avail {
		risk := 0
		for i := 0; i < p.GetCardsSize(); i++ {
			risk += quodlibetCardRisk(c, p.GetCard(i))
		}
		if bestRisk < 0 || risk < bestRisk {
			best, bestRisk = c, risk
		}
	}
	return best
}
