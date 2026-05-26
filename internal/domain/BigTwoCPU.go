package domain

import "sort"

// findBestPlay 難易度に応じたCPUプレイ戦略のディスパッチャー
func (bt *BigTwo) findBestPlay(player *BigTwoPlayer) []int {
	switch bt.config.CpuDifficulty {
	case BigTwoDifficultyEasy:
		return bt.findBestPlayEasy(player)
	case BigTwoDifficultyHard:
		return bt.findBestPlayHard(player)
	default:
		return bt.findBestPlayNormal(player)
	}
}

// findBestPlayNormal 通常難易度: 出せる最弱のカードセットを探す
func (bt *BigTwo) findBestPlayNormal(player *BigTwoPlayer) []int {
	candidates := bt.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		ti := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[i]))
		tj := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[j]))
		si := bigTwoPlayStrength(bt.indicesToCards(player, candidates[i]), ti)
		sj := bigTwoPlayStrength(bt.indicesToCards(player, candidates[j]), tj)
		return si < sj
	})
	return candidates[0]
}

// findBestPlayEasy 簡単難易度: ランダムに出せるものを出す
func (bt *BigTwo) findBestPlayEasy(player *BigTwoPlayer) []int {
	candidates := bt.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

// findBestPlayHard 難しい難易度: 戦略的な判断を行う
func (bt *BigTwo) findBestPlayHard(player *BigTwoPlayer) []int {
	candidates := bt.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}

	if bt.round.tableCards == nil {
		sort.Slice(candidates, func(i, j int) bool {
			ti := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[i]))
			tj := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[j]))
			si := bigTwoPlayStrength(bt.indicesToCards(player, candidates[i]), ti)
			sj := bigTwoPlayStrength(bt.indicesToCards(player, candidates[j]), tj)
			return si < sj
		})
		return candidates[0]
	}

	opMinCards := bt.opponentMinCards(player)
	if opMinCards <= 2 {
		sort.Slice(candidates, func(i, j int) bool {
			ti := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[i]))
			tj := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[j]))
			si := bigTwoPlayStrength(bt.indicesToCards(player, candidates[i]), ti)
			sj := bigTwoPlayStrength(bt.indicesToCards(player, candidates[j]), tj)
			return si > sj
		})
		return candidates[0]
	}

	sort.Slice(candidates, func(i, j int) bool {
		ti := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[i]))
		tj := bigTwoClassifyPlay(bt.indicesToCards(player, candidates[j]))
		si := bigTwoPlayStrength(bt.indicesToCards(player, candidates[i]), ti)
		sj := bigTwoPlayStrength(bt.indicesToCards(player, candidates[j]), tj)
		return si < sj
	})
	return candidates[0]
}

// opponentMinCards 対戦相手の中で最小の手札枚数を返す
func (bt *BigTwo) opponentMinCards(player *BigTwoPlayer) int {
	minCards := BigTwoCardsPerPlayer
	for _, p := range bt.players {
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
func (bt *BigTwo) indicesToCards(player *BigTwoPlayer, indices []int) []*Card {
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	return cards
}

// findAllPlayableSets 出せる全てのカードセットを探す
func (bt *BigTwo) findAllPlayableSets(player *BigTwoPlayer) [][]int {
	var results [][]int

	isFirstPlay := !bt.round.firstPlayDone

	if bt.round.tableCards == nil {
		results = append(results, bt.findSingles(player, isFirstPlay)...)
		results = append(results, bt.findPairs(player, isFirstPlay)...)
		results = append(results, bt.findTriples(player, isFirstPlay)...)
		results = append(results, bt.findFiveCardCombos(player, isFirstPlay)...)
	} else {
		switch bt.round.tablePlayType {
		case BigTwoPlaySingle:
			results = bt.findSingles(player, false)
		case BigTwoPlayPair:
			results = bt.findPairs(player, false)
		case BigTwoPlayTriple:
			results = bt.findTriples(player, false)
		default:
			results = bt.findFiveCardCombos(player, false)
		}
	}

	return results
}

// findSingles 出せるシングルカードのインデックスを探す
func (bt *BigTwo) findSingles(player *BigTwoPlayer, mustIncludeDiamond3 bool) [][]int {
	var results [][]int
	for i := 0; i < player.GetCardsSize(); i++ {
		cards := []*Card{player.GetCard(i)}
		if mustIncludeDiamond3 && !bt.containsDiamond3(cards) {
			continue
		}
		if bigTwoIsPlayable(cards, bt.round.tableCards, bt.round.tablePlayType) {
			results = append(results, []int{i})
		}
	}
	return results
}

// findPairs 出せるペアを探す
func (bt *BigTwo) findPairs(player *BigTwoPlayer, mustIncludeDiamond3 bool) [][]int {
	var results [][]int
	n := player.GetCardsSize()
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			cards := []*Card{player.GetCard(i), player.GetCard(j)}
			if mustIncludeDiamond3 && !bt.containsDiamond3(cards) {
				continue
			}
			if bigTwoIsPlayable(cards, bt.round.tableCards, bt.round.tablePlayType) {
				results = append(results, []int{i, j})
			}
		}
	}
	return results
}

// findTriples 出せるトリプルを探す
func (bt *BigTwo) findTriples(player *BigTwoPlayer, mustIncludeDiamond3 bool) [][]int {
	var results [][]int
	n := player.GetCardsSize()
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			for k := j + 1; k < n; k++ {
				cards := []*Card{player.GetCard(i), player.GetCard(j), player.GetCard(k)}
				if mustIncludeDiamond3 && !bt.containsDiamond3(cards) {
					continue
				}
				if bigTwoIsPlayable(cards, bt.round.tableCards, bt.round.tablePlayType) {
					results = append(results, []int{i, j, k})
				}
			}
		}
	}
	return results
}

// findFiveCardCombos 出せる5枚役を探す
func (bt *BigTwo) findFiveCardCombos(player *BigTwoPlayer, mustIncludeDiamond3 bool) [][]int {
	var results [][]int
	n := player.GetCardsSize()
	if n < 5 {
		return results
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			for k := j + 1; k < n; k++ {
				for l := k + 1; l < n; l++ {
					for m := l + 1; m < n; m++ {
						cards := []*Card{
							player.GetCard(i), player.GetCard(j), player.GetCard(k),
							player.GetCard(l), player.GetCard(m),
						}
						if mustIncludeDiamond3 && !bt.containsDiamond3(cards) {
							continue
						}
						if bigTwoIsPlayable(cards, bt.round.tableCards, bt.round.tablePlayType) {
							results = append(results, []int{i, j, k, l, m})
						}
					}
				}
			}
		}
	}
	return results
}
