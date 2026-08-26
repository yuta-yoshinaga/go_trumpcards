//go:build !js || !wasm || extra4

package domain

// cpuBid CPUのビッド処理 (executeBid が cpuActions に行動記録を追加する)
func (d *Doudizhu) cpuBid() {
	value := d.evaluateBid(d.players[d.round.currentTurn])
	_ = d.executeBid(value)
}

// evaluateBid 手札の品質に基づいてビッド値を決定
func (d *Doudizhu) evaluateBid(player *DoudizhuPlayer) int {
	if d.config.CpuDifficulty == DoudizhuDifficultyEasy {
		return 0
	}

	score := d.handQuality(player)

	bid := 0
	if score >= 8 {
		bid = 3
	} else if score >= 5 {
		bid = 2
	} else if score >= 3 {
		bid = 1
	}

	if bid > 0 && bid <= d.round.highestBid {
		return 0
	}
	return bid
}

// handQuality 手札の品質をスコアで評価 (0-10+)
func (d *Doudizhu) handQuality(player *DoudizhuPlayer) int {
	score := 0
	freq := make(map[int]int)
	jokerCount := 0

	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if IsJoker(c) {
			jokerCount++
			continue
		}
		s := DoudizhuCardStrength(c)
		freq[s]++
	}

	switch jokerCount {
	case 2:
		score += 4
	case 1:
		score += 2
	}

	for s, cnt := range freq {
		if cnt == 4 {
			score += 3
		}
		if s >= 14 && cnt >= 2 {
			score++
		}
		if s == 15 {
			score++
		}
	}
	return score
}

// findBestPlay 最善のプレイを見つける (インデックスのスライス)
func (d *Doudizhu) findBestPlay(player *DoudizhuPlayer) []int {
	switch d.config.CpuDifficulty {
	case DoudizhuDifficultyEasy:
		return d.findPlayEasy(player)
	case DoudizhuDifficultyHard:
		return d.findPlayHard(player)
	default:
		return d.findPlayNormal(player)
	}
}

// findPlayEasy 最弱の有効な役を出す
func (d *Doudizhu) findPlayEasy(player *DoudizhuPlayer) []int {
	if d.round.tableCombo == nil {
		return d.findLeadEasy(player)
	}
	return d.findResponseEasy(player)
}

// findLeadEasy リードターン (簡単): 最弱の単張を出す
func (d *Doudizhu) findLeadEasy(player *DoudizhuPlayer) []int {
	return []int{0}
}

// findResponseEasy レスポンスターン (簡単): 最弱の有効な役を出す、なければパス
func (d *Doudizhu) findResponseEasy(player *DoudizhuPlayer) []int {
	table := d.round.tableCombo
	candidates := d.findAllBeatingCombos(player, table)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

// findPlayNormal 通常難易度のプレイ
func (d *Doudizhu) findPlayNormal(player *DoudizhuPlayer) []int {
	if d.round.tableCombo == nil {
		return d.findLeadNormal(player)
	}
	return d.findResponseNormal(player)
}

// findLeadNormal リードターン (通常): 最弱の単張、対子、三条の順で出す
func (d *Doudizhu) findLeadNormal(player *DoudizhuPlayer) []int {
	freq := d.buildFrequency(player)
	for _, rc := range freq {
		if rc.count == 1 && rc.strength < 15 {
			idx := d.findCardByStrength(player, rc.strength)
			if idx >= 0 {
				return []int{idx}
			}
		}
	}
	for _, rc := range freq {
		if rc.count == 2 && rc.strength < 15 {
			indices := d.findCardsByStrength(player, rc.strength, 2)
			if len(indices) == 2 {
				return indices
			}
		}
	}
	return []int{0}
}

// findResponseNormal レスポンスターン (通常): ボムを温存、最弱で返す
func (d *Doudizhu) findResponseNormal(player *DoudizhuPlayer) []int {
	table := d.round.tableCombo
	candidates := d.findAllBeatingCombos(player, table)
	if len(candidates) == 0 {
		return nil
	}

	playerIdx := d.round.currentTurn
	if d.isTeammate(playerIdx, d.round.lastPlayIdx) && d.shouldLetTeammateWin(playerIdx) {
		return nil
	}

	for _, c := range candidates {
		cards := make([]*Card, len(c))
		for i, idx := range c {
			cards[i] = player.GetCard(idx)
		}
		combo := DoudizhuClassifyCombo(cards)
		if combo != nil && combo.Type != DoudizhuComboBomb && combo.Type != DoudizhuComboRocket {
			return c
		}
	}

	if d.isUrgent() {
		return candidates[0]
	}
	return nil
}

// findPlayHard 高難易度のプレイ
func (d *Doudizhu) findPlayHard(player *DoudizhuPlayer) []int {
	if d.round.tableCombo == nil {
		return d.findLeadHard(player)
	}
	return d.findResponseHard(player)
}

// findLeadHard リードターン (高難易度): 弱い連続役を優先
func (d *Doudizhu) findLeadHard(player *DoudizhuPlayer) []int {
	freq := d.buildFrequency(player)

	if straight := d.findWeakStraight(player, freq); len(straight) > 0 {
		return straight
	}
	if consec := d.findWeakConsecutivePair(player, freq); len(consec) > 0 {
		return consec
	}

	return d.findLeadNormal(player)
}

// findResponseHard レスポンスターン (高難易度): チームメイト考慮あり
func (d *Doudizhu) findResponseHard(player *DoudizhuPlayer) []int {
	table := d.round.tableCombo
	playerIdx := d.round.currentTurn

	if d.isTeammate(playerIdx, d.round.lastPlayIdx) {
		if table.Rank >= 14 || d.shouldLetTeammateWin(playerIdx) {
			return nil
		}
	}

	candidates := d.findAllBeatingCombos(player, table)
	if len(candidates) == 0 {
		return nil
	}

	for _, c := range candidates {
		cards := make([]*Card, len(c))
		for i, idx := range c {
			cards[i] = player.GetCard(idx)
		}
		combo := DoudizhuClassifyCombo(cards)
		if combo != nil && combo.Type != DoudizhuComboBomb && combo.Type != DoudizhuComboRocket {
			return c
		}
	}

	if d.isUrgent() {
		return candidates[0]
	}
	return nil
}

// --- helper methods ---

// isTeammate 2人のプレイヤーがチームメイトかどうか (農民同士)
func (d *Doudizhu) isTeammate(a, b int) bool {
	if a < 0 || b < 0 || a >= DoudizhuPlayerCnt || b >= DoudizhuPlayerCnt {
		return false
	}
	return !d.players[a].GetIsLandlord() && !d.players[b].GetIsLandlord()
}

// shouldLetTeammateWin チームメイトが勝ちそうなら譲る
func (d *Doudizhu) shouldLetTeammateWin(playerIdx int) bool {
	for i := 0; i < DoudizhuPlayerCnt; i++ {
		if i != playerIdx && d.isTeammate(playerIdx, i) && d.players[i].GetCardsSize() <= 2 {
			return true
		}
	}
	return false
}

// isUrgent 地主が残り少ない場合は積極的に
func (d *Doudizhu) isUrgent() bool {
	return d.players[d.round.landlordIdx].GetCardsSize() <= 3
}

// buildFrequency プレイヤーの手札の強さ別頻度を構築
func (d *Doudizhu) buildFrequency(player *DoudizhuPlayer) []rankCount {
	freq := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		s := DoudizhuCardStrength(c)
		freq[s]++
	}
	result := make([]rankCount, 0, len(freq))
	for s, cnt := range freq {
		result = append(result, rankCount{strength: s, count: cnt})
	}
	sortRankCounts(result)
	return result
}

func sortRankCounts(rcs []rankCount) {
	for i := 1; i < len(rcs); i++ {
		for j := i; j > 0 && rcs[j].strength < rcs[j-1].strength; j-- {
			rcs[j], rcs[j-1] = rcs[j-1], rcs[j]
		}
	}
}

// findCardByStrength 指定強さのカードのインデックスを1つ返す
func (d *Doudizhu) findCardByStrength(player *DoudizhuPlayer, strength int) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		if DoudizhuCardStrength(player.GetCard(i)) == strength {
			return i
		}
	}
	return -1
}

// findCardsByStrength 指定強さのカードのインデックスをcount個返す
func (d *Doudizhu) findCardsByStrength(player *DoudizhuPlayer, strength, count int) []int {
	indices := make([]int, 0, count)
	for i := 0; i < player.GetCardsSize(); i++ {
		if DoudizhuCardStrength(player.GetCard(i)) == strength {
			indices = append(indices, i)
			if len(indices) == count {
				break
			}
		}
	}
	return indices
}

// findAllBeatingCombos テーブルの役に勝てるすべての役のインデックスリストを弱い順で返す
func (d *Doudizhu) findAllBeatingCombos(player *DoudizhuPlayer, table *DoudizhuCombo) [][]int {
	var results [][]int

	switch table.Type {
	case DoudizhuComboSingle:
		for i := 0; i < player.GetCardsSize(); i++ {
			s := DoudizhuCardStrength(player.GetCard(i))
			if s > table.Rank {
				results = append(results, []int{i})
			}
		}
	case DoudizhuComboPair:
		freq := d.buildFrequency(player)
		for _, rc := range freq {
			if rc.count >= 2 && rc.strength > table.Rank {
				indices := d.findCardsByStrength(player, rc.strength, 2)
				if len(indices) == 2 {
					results = append(results, indices)
				}
			}
		}
	case DoudizhuComboTrio:
		freq := d.buildFrequency(player)
		for _, rc := range freq {
			if rc.count >= 3 && rc.strength > table.Rank {
				indices := d.findCardsByStrength(player, rc.strength, 3)
				if len(indices) == 3 {
					results = append(results, indices)
				}
			}
		}
	case DoudizhuComboTrioSingle:
		freq := d.buildFrequency(player)
		for _, rc := range freq {
			if rc.count >= 3 && rc.strength > table.Rank {
				trioIdx := d.findCardsByStrength(player, rc.strength, 3)
				if len(trioIdx) == 3 {
					kicker := d.findKickerSingle(player, rc.strength)
					if kicker >= 0 {
						results = append(results, append(trioIdx, kicker))
					}
				}
			}
		}
	case DoudizhuComboTrioPair:
		freq := d.buildFrequency(player)
		for _, rc := range freq {
			if rc.count >= 3 && rc.strength > table.Rank {
				trioIdx := d.findCardsByStrength(player, rc.strength, 3)
				if len(trioIdx) == 3 {
					kickerPair := d.findKickerPair(player, rc.strength)
					if len(kickerPair) == 2 {
						results = append(results, append(trioIdx, kickerPair...))
					}
				}
			}
		}
	case DoudizhuComboStraight:
		results = d.findBeatingStraights(player, table)
	case DoudizhuComboConsecutivePair:
		results = d.findBeatingConsecutivePairs(player, table)
	case DoudizhuComboBomb:
		freq := d.buildFrequency(player)
		for _, rc := range freq {
			if rc.count == 4 && rc.strength > table.Rank {
				indices := d.findCardsByStrength(player, rc.strength, 4)
				if len(indices) == 4 {
					results = append(results, indices)
				}
			}
		}
		if rocket := d.findRocket(player); len(rocket) == 2 {
			results = append(results, rocket)
		}
	default:
	}

	if table.Type != DoudizhuComboBomb && table.Type != DoudizhuComboRocket {
		freq := d.buildFrequency(player)
		for _, rc := range freq {
			if rc.count == 4 {
				indices := d.findCardsByStrength(player, rc.strength, 4)
				if len(indices) == 4 {
					results = append(results, indices)
				}
			}
		}
		if rocket := d.findRocket(player); len(rocket) == 2 {
			results = append(results, rocket)
		}
	}

	return results
}

func (d *Doudizhu) findKickerSingle(player *DoudizhuPlayer, excludeStrength int) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		if DoudizhuCardStrength(player.GetCard(i)) != excludeStrength {
			return i
		}
	}
	return -1
}

func (d *Doudizhu) findKickerPair(player *DoudizhuPlayer, excludeStrength int) []int {
	freq := d.buildFrequency(player)
	for _, rc := range freq {
		if rc.strength != excludeStrength && rc.count >= 2 {
			return d.findCardsByStrength(player, rc.strength, 2)
		}
	}
	return nil
}

func (d *Doudizhu) findRocket(player *DoudizhuPlayer) []int {
	smallIdx, bigIdx := -1, -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if IsSmallJoker(c) {
			smallIdx = i
		}
		if IsBigJoker(c) {
			bigIdx = i
		}
	}
	if smallIdx >= 0 && bigIdx >= 0 {
		return []int{smallIdx, bigIdx}
	}
	return nil
}

func (d *Doudizhu) findBeatingStraights(player *DoudizhuPlayer, table *DoudizhuCombo) [][]int {
	var results [][]int
	freq := d.buildFrequency(player)

	chainable := make([]int, 0)
	for _, rc := range freq {
		if rc.count >= 1 && isChainable(rc.strength) {
			chainable = append(chainable, rc.strength)
		}
	}

	for start := 0; start <= len(chainable)-table.Length; start++ {
		valid := true
		for i := 1; i < table.Length; i++ {
			if chainable[start+i]-chainable[start+i-1] != 1 {
				valid = false
				break
			}
		}
		if valid && chainable[start] > table.Rank {
			indices := make([]int, 0, table.Length)
			for i := 0; i < table.Length; i++ {
				idx := d.findCardByStrength(player, chainable[start+i])
				if idx >= 0 {
					indices = append(indices, idx)
				}
			}
			if len(indices) == table.Length {
				results = append(results, indices)
			}
		}
	}
	return results
}

func (d *Doudizhu) findBeatingConsecutivePairs(player *DoudizhuPlayer, table *DoudizhuCombo) [][]int {
	var results [][]int
	freq := d.buildFrequency(player)

	pairStrengths := make([]int, 0)
	for _, rc := range freq {
		if rc.count >= 2 && isChainable(rc.strength) {
			pairStrengths = append(pairStrengths, rc.strength)
		}
	}

	for start := 0; start <= len(pairStrengths)-table.Length; start++ {
		valid := true
		for i := 1; i < table.Length; i++ {
			if pairStrengths[start+i]-pairStrengths[start+i-1] != 1 {
				valid = false
				break
			}
		}
		if valid && pairStrengths[start] > table.Rank {
			indices := make([]int, 0, table.Length*2)
			for i := 0; i < table.Length; i++ {
				pair := d.findCardsByStrength(player, pairStrengths[start+i], 2)
				indices = append(indices, pair...)
			}
			if len(indices) == table.Length*2 {
				results = append(results, indices)
			}
		}
	}
	return results
}

// findWeakStraight 弱い順子を探す
func (d *Doudizhu) findWeakStraight(player *DoudizhuPlayer, freq []rankCount) []int {
	chainable := make([]int, 0)
	for _, rc := range freq {
		if rc.count >= 1 && isChainable(rc.strength) {
			chainable = append(chainable, rc.strength)
		}
	}
	for start := 0; start <= len(chainable)-5; start++ {
		valid := true
		for i := 1; i < 5; i++ {
			if chainable[start+i]-chainable[start+i-1] != 1 {
				valid = false
				break
			}
		}
		if valid && chainable[start] <= 10 {
			indices := make([]int, 0, 5)
			for i := 0; i < 5; i++ {
				idx := d.findCardByStrength(player, chainable[start+i])
				if idx >= 0 {
					indices = append(indices, idx)
				}
			}
			if len(indices) == 5 {
				return indices
			}
		}
	}
	return nil
}

// findWeakConsecutivePair 弱い連対を探す
func (d *Doudizhu) findWeakConsecutivePair(player *DoudizhuPlayer, freq []rankCount) []int {
	pairStrengths := make([]int, 0)
	for _, rc := range freq {
		if rc.count >= 2 && isChainable(rc.strength) {
			pairStrengths = append(pairStrengths, rc.strength)
		}
	}
	for start := 0; start <= len(pairStrengths)-3; start++ {
		valid := true
		for i := 1; i < 3; i++ {
			if pairStrengths[start+i]-pairStrengths[start+i-1] != 1 {
				valid = false
				break
			}
		}
		if valid && pairStrengths[start] <= 10 {
			indices := make([]int, 0, 6)
			for i := 0; i < 3; i++ {
				pair := d.findCardsByStrength(player, pairStrengths[start+i], 2)
				indices = append(indices, pair...)
			}
			if len(indices) == 6 {
				return indices
			}
		}
	}
	return nil
}
