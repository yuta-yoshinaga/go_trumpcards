//go:build !js || !wasm || solo

package domain

import "sort"

// findBestPlay 難易度に応じたCPUプレイ戦略のディスパッチャー
func (z *Zheng) findBestPlay(player *ZhengPlayer) []int {
	switch z.config.CpuDifficulty {
	case ZhengDifficultyEasy:
		return z.findBestPlayEasy(player)
	case ZhengDifficultyHard:
		return z.findBestPlayHard(player)
	default:
		return z.findBestPlayNormal(player)
	}
}

// findBestPlayNormal 通常難易度: 出せる最弱のカードセットを探す (爆弾は強度上、自然に温存される)
func (z *Zheng) findBestPlayNormal(player *ZhengPlayer) []int {
	candidates := z.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}
	z.sortCandidatesByStrength(player, candidates, false)
	return candidates[0]
}

// findBestPlayEasy 簡単難易度: 出せるものを最初に見つけた順で出す
func (z *Zheng) findBestPlayEasy(player *ZhengPlayer) []int {
	candidates := z.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

// findBestPlayHard 難しい難易度: 戦略的な判断を行う
func (z *Zheng) findBestPlayHard(player *ZhengPlayer) []int {
	candidates := z.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}

	// 相手が上がり間際なら強い手 (爆弾を含む) で蓋をする
	if len(z.round.tableCards) > 0 && z.opponentMinCards(player) <= 2 {
		z.sortCandidatesByStrength(player, candidates, true)
		return candidates[0]
	}

	// それ以外は爆弾を温存し、通常役の最弱から崩していく
	nonBombs := make([][]int, 0, len(candidates))
	for _, cand := range candidates {
		cards := z.indicesToCards(player, cand)
		if !zhengIsBombType(zhengClassifyPlay(cards)) {
			nonBombs = append(nonBombs, cand)
		}
	}
	if len(nonBombs) > 0 {
		z.sortCandidatesByStrength(player, nonBombs, false)
		return nonBombs[0]
	}
	z.sortCandidatesByStrength(player, candidates, false)
	return candidates[0]
}

// sortCandidatesByStrength 候補をプレイ強度でソートする (descending=true で強い順)
func (z *Zheng) sortCandidatesByStrength(player *ZhengPlayer, candidates [][]int, descending bool) {
	sort.SliceStable(candidates, func(i, j int) bool {
		si := zhengCandidateStrength(z.indicesToCards(player, candidates[i]))
		sj := zhengCandidateStrength(z.indicesToCards(player, candidates[j]))
		if descending {
			return si > sj
		}
		return si < sj
	})
}

// opponentMinCards 対戦相手の中で最小の手札枚数を返す
func (z *Zheng) opponentMinCards(player *ZhengPlayer) int {
	minCards := ZhengMaxCardsPerPlayer
	for _, p := range z.players {
		if p == player || p.GetIsFinished() {
			continue
		}
		if p.GetCardsSize() < minCards {
			minCards = p.GetCardsSize()
		}
	}
	return minCards
}

// indicesToCards インデックス配列からカードスライスを生成
func (z *Zheng) indicesToCards(player *ZhengPlayer, indices []int) []*Card {
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	return cards
}

// findAllPlayableSets 出せる全てのカードセットを探す
func (z *Zheng) findAllPlayableSets(player *ZhengPlayer) [][]int {
	var raw [][]int
	if len(z.round.tableCards) == 0 {
		raw = append(raw, z.findSingles(player)...)
		raw = append(raw, z.findPairs(player)...)
		raw = append(raw, z.findTriples(player)...)
		raw = append(raw, z.findStraights(player, 0)...)
		raw = append(raw, z.findPairRuns(player, 0)...)
	} else {
		switch z.round.tablePlayType {
		case ZhengPlaySingle:
			raw = append(raw, z.findSingles(player)...)
		case ZhengPlayPair:
			raw = append(raw, z.findPairs(player)...)
		case ZhengPlayTriple:
			raw = append(raw, z.findTriples(player)...)
		case ZhengPlayStraight:
			raw = append(raw, z.findStraights(player, len(z.round.tableCards))...)
		case ZhengPlayPairRun:
			raw = append(raw, z.findPairRuns(player, len(z.round.tableCards))...)
		}
	}
	// 爆弾・ジョーカーボムは常に候補化する (リード可、非爆弾役も切れる)。
	// 有効性は zhengIsPlayable のフィルタに委ねる。
	raw = append(raw, z.findBombs(player)...)
	raw = append(raw, z.findJokerBomb(player)...)

	results := make([][]int, 0, len(raw))
	for _, cand := range raw {
		cards := z.indicesToCards(player, cand)
		if zhengIsPlayable(cards, z.round.tableCards, z.round.tablePlayType) {
			results = append(results, cand)
		}
	}
	return results
}

// cardsByRunStrength は手札を (2とジョーカーを除いた) ランク強度ごとのインデックス昇順リストに分類する。
func (z *Zheng) cardsByRunStrength(player *ZhengPlayer) map[int][]int {
	byStr := make(map[int][]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil || c.GetDesign() == CardDesignJoker || c.GetValue() == 2 {
			continue
		}
		s := zhengRankStrength(c)
		byStr[s] = append(byStr[s], i)
	}
	return byStr
}

// cardsByValue は手札を (ジョーカーを除いた) ランクごとのインデックス昇順リストに分類する。
func (z *Zheng) cardsByValue(player *ZhengPlayer) map[int][]int {
	byVal := make(map[int][]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c == nil || c.GetDesign() == CardDesignJoker {
			continue
		}
		byVal[c.GetValue()] = append(byVal[c.GetValue()], i)
	}
	return byVal
}

// findSingles 全シングルカードのインデックスを探す
func (z *Zheng) findSingles(player *ZhengPlayer) [][]int {
	var results [][]int
	for i := 0; i < player.GetCardsSize(); i++ {
		results = append(results, []int{i})
	}
	return results
}

// findPairs 同ランクのペアを探す (ジョーカーはペア不可)
func (z *Zheng) findPairs(player *ZhengPlayer) [][]int {
	var results [][]int
	for _, idxs := range z.cardsByValue(player) {
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				results = append(results, []int{idxs[i], idxs[j]})
			}
		}
	}
	return results
}

// findTriples 同ランクのトリプルを探す
func (z *Zheng) findTriples(player *ZhengPlayer) [][]int {
	var results [][]int
	for _, idxs := range z.cardsByValue(player) {
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				for k := j + 1; k < len(idxs); k++ {
					results = append(results, []int{idxs[i], idxs[j], idxs[k]})
				}
			}
		}
	}
	return results
}

// findBombs 同ランク4枚 (爆弾) を探す
func (z *Zheng) findBombs(player *ZhengPlayer) [][]int {
	var results [][]int
	for _, idxs := range z.cardsByValue(player) {
		if len(idxs) == 4 {
			results = append(results, []int{idxs[0], idxs[1], idxs[2], idxs[3]})
		}
	}
	return results
}

// findJokerBomb ジョーカー2枚 (ジョーカーボム) を探す
func (z *Zheng) findJokerBomb(player *ZhengPlayer) [][]int {
	jokers := make([]int, 0, 2)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c != nil && c.GetDesign() == CardDesignJoker {
			jokers = append(jokers, i)
		}
	}
	if len(jokers) == 2 {
		return [][]int{jokers}
	}
	return nil
}

// findStraights 連続役(ストレート)を探す。exactLen>0 でその長さのみ、0 で長さ3以上すべて。
func (z *Zheng) findStraights(player *ZhengPlayer, exactLen int) [][]int {
	byStr := z.cardsByRunStrength(player)
	var lengths []int
	if exactLen > 0 {
		lengths = []int{exactLen}
	} else {
		for l := 3; l <= player.GetCardsSize(); l++ {
			lengths = append(lengths, l)
		}
	}

	var results [][]int
	for _, l := range lengths {
		if l < 3 {
			continue
		}
		// ランク強度 0..11 (3..A) 内の連続ウィンドウを探す
		for start := 0; start+l <= 12; start++ {
			ok := true
			for s := start; s < start+l; s++ {
				if len(byStr[s]) == 0 {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			cand := make([]int, 0, l)
			for s := start; s < start+l; s++ {
				cand = append(cand, byStr[s][0])
			}
			results = append(results, cand)
		}
	}
	return results
}

// findPairRuns 連続ペア役を探す。exactCards>0 でその枚数のみ、0 で3組(6枚)以上すべて。
func (z *Zheng) findPairRuns(player *ZhengPlayer, exactCards int) [][]int {
	if exactCards > 0 && (exactCards < 6 || exactCards%2 != 0) {
		return nil
	}
	byStr := z.cardsByRunStrength(player)
	var pairLens []int
	if exactCards > 0 {
		pairLens = []int{exactCards / 2}
	} else {
		for l := 3; l*2 <= player.GetCardsSize(); l++ {
			pairLens = append(pairLens, l)
		}
	}

	var results [][]int
	for _, l := range pairLens {
		// ランク強度 0..11 (3..A) 内の連続ウィンドウを探す
		for start := 0; start+l <= 12; start++ {
			ok := true
			for s := start; s < start+l; s++ {
				if len(byStr[s]) < 2 {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			cand := make([]int, 0, l*2)
			for s := start; s < start+l; s++ {
				cand = append(cand, byStr[s][0], byStr[s][1])
			}
			results = append(results, cand)
		}
	}
	return results
}
