//go:build !js || !wasm || solo

package domain

import "sort"

// findBestPlay 難易度に応じたCPUプレイ戦略のディスパッチャー
func (tl *TienLen) findBestPlay(player *TienLenPlayer) []int {
	switch tl.config.CpuDifficulty {
	case TienLenDifficultyEasy:
		return tl.findBestPlayEasy(player)
	case TienLenDifficultyHard:
		return tl.findBestPlayHard(player)
	default:
		return tl.findBestPlayNormal(player)
	}
}

// findBestPlayNormal 通常難易度: 出せる最弱のカードセットを探す
func (tl *TienLen) findBestPlayNormal(player *TienLenPlayer) []int {
	candidates := tl.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}
	tl.sortCandidatesByStrength(player, candidates, false)
	return candidates[0]
}

// findBestPlayEasy 簡単難易度: 出せるものを最初に見つけた順で出す
func (tl *TienLen) findBestPlayEasy(player *TienLenPlayer) []int {
	candidates := tl.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

// findBestPlayHard 難しい難易度: 戦略的な判断を行う
func (tl *TienLen) findBestPlayHard(player *TienLenPlayer) []int {
	candidates := tl.findAllPlayableSets(player)
	if len(candidates) == 0 {
		return nil
	}

	// 場が流れているときは弱い手から崩していく
	if tl.round.tableCards == nil {
		tl.sortCandidatesByStrength(player, candidates, false)
		return candidates[0]
	}

	// 相手が上がり間際なら強い手で蓋をする
	if tl.opponentMinCards(player) <= 2 {
		tl.sortCandidatesByStrength(player, candidates, true)
		return candidates[0]
	}

	tl.sortCandidatesByStrength(player, candidates, false)
	return candidates[0]
}

// sortCandidatesByStrength 候補をプレイ強度でソートする (descending=true で強い順)
func (tl *TienLen) sortCandidatesByStrength(player *TienLenPlayer, candidates [][]int, descending bool) {
	sort.SliceStable(candidates, func(i, j int) bool {
		ci := tl.indicesToCards(player, candidates[i])
		cj := tl.indicesToCards(player, candidates[j])
		si := tienLenPlayStrength(ci, tienLenClassifyPlay(ci))
		sj := tienLenPlayStrength(cj, tienLenClassifyPlay(cj))
		if descending {
			return si > sj
		}
		return si < sj
	})
}

// opponentMinCards 対戦相手の中で最小の手札枚数を返す
func (tl *TienLen) opponentMinCards(player *TienLenPlayer) int {
	minCards := TienLenCardsPerPlayer
	for _, p := range tl.players {
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
func (tl *TienLen) indicesToCards(player *TienLenPlayer, indices []int) []*Card {
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	return cards
}

// findAllPlayableSets 出せる全てのカードセットを探す
func (tl *TienLen) findAllPlayableSets(player *TienLenPlayer) [][]int {
	isFirstPlay := !tl.round.firstPlayDone

	var raw [][]int
	if tl.round.tableCards == nil {
		raw = append(raw, tl.findSingles(player)...)
		raw = append(raw, tl.findPairs(player)...)
		raw = append(raw, tl.findTriples(player)...)
		raw = append(raw, tl.findStraights(player, 0)...)
		raw = append(raw, tl.findThreePairRuns(player)...)
		raw = append(raw, tl.findFourOfAKinds(player)...)
	} else {
		switch tl.round.tablePlayType {
		case TienLenPlaySingle:
			raw = append(raw, tl.findSingles(player)...)
			// チョップ役は単体の「2」を切れる
			if tienLenIsSingleTwo(tl.round.tableCards, tl.round.tablePlayType) {
				raw = append(raw, tl.findThreePairRuns(player)...)
				raw = append(raw, tl.findFourOfAKinds(player)...)
			}
		case TienLenPlayPair:
			raw = append(raw, tl.findPairs(player)...)
		case TienLenPlayTriple:
			raw = append(raw, tl.findTriples(player)...)
		case TienLenPlayStraight:
			raw = append(raw, tl.findStraights(player, len(tl.round.tableCards))...)
		case TienLenPlayThreePairRun:
			raw = append(raw, tl.findThreePairRuns(player)...)
			raw = append(raw, tl.findFourOfAKinds(player)...)
		case TienLenPlayFourOfAKind:
			raw = append(raw, tl.findFourOfAKinds(player)...)
		}
	}

	results := make([][]int, 0, len(raw))
	for _, cand := range raw {
		cards := tl.indicesToCards(player, cand)
		if isFirstPlay && !tl.containsSpade3(cards) {
			continue
		}
		if tienLenIsPlayable(cards, tl.round.tableCards, tl.round.tablePlayType) {
			results = append(results, cand)
		}
	}
	return results
}

// cardsByValueStrength は手札を (2を除いた) value-strength ごとのインデックス昇順リストに分類する。
func (tl *TienLen) cardsByValueStrength(player *TienLenPlayer) map[int][]int {
	byStr := make(map[int][]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetValue() == 2 {
			continue
		}
		s := tienLenValueStrength(c.GetValue())
		byStr[s] = append(byStr[s], i)
	}
	return byStr
}

// cardsByValue は手札をランクごとのインデックス昇順リストに分類する。
func (tl *TienLen) cardsByValue(player *TienLenPlayer) map[int][]int {
	byVal := make(map[int][]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		v := player.GetCard(i).GetValue()
		byVal[v] = append(byVal[v], i)
	}
	return byVal
}

// findSingles 全シングルカードのインデックスを探す
func (tl *TienLen) findSingles(player *TienLenPlayer) [][]int {
	var results [][]int
	for i := 0; i < player.GetCardsSize(); i++ {
		results = append(results, []int{i})
	}
	return results
}

// findPairs 同ランクのペアを探す
func (tl *TienLen) findPairs(player *TienLenPlayer) [][]int {
	var results [][]int
	for _, idxs := range tl.cardsByValue(player) {
		for i := 0; i < len(idxs); i++ {
			for j := i + 1; j < len(idxs); j++ {
				results = append(results, []int{idxs[i], idxs[j]})
			}
		}
	}
	return results
}

// findTriples 同ランクのトリプルを探す
func (tl *TienLen) findTriples(player *TienLenPlayer) [][]int {
	var results [][]int
	for _, idxs := range tl.cardsByValue(player) {
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

// findFourOfAKinds 同ランク4枚 (フォーカード) を探す
func (tl *TienLen) findFourOfAKinds(player *TienLenPlayer) [][]int {
	var results [][]int
	for _, idxs := range tl.cardsByValue(player) {
		if len(idxs) == 4 {
			results = append(results, []int{idxs[0], idxs[1], idxs[2], idxs[3]})
		}
	}
	return results
}

// findThreePairRuns 連続する3つのペア (チョップ役) を探す
func (tl *TienLen) findThreePairRuns(player *TienLenPlayer) [][]int {
	byStr := tl.cardsByValueStrength(player)
	var results [][]int
	// value-strength 0..11 (3..A); start+2 must stay within A (<=11)
	for start := 0; start+2 <= 11; start++ {
		a, b, c := byStr[start], byStr[start+1], byStr[start+2]
		if len(a) >= 2 && len(b) >= 2 && len(c) >= 2 {
			results = append(results, []int{a[0], a[1], b[0], b[1], c[0], c[1]})
		}
	}
	return results
}

// findStraights 連続役(ストレート)を探す。exactLen>0 でその長さのみ、0 で長さ3以上すべて。
func (tl *TienLen) findStraights(player *TienLenPlayer, exactLen int) [][]int {
	byStr := tl.cardsByValueStrength(player)
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
		// window of consecutive value-strengths within 0..11 (excludes 2)
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
			top := start + l - 1
			base := make([]int, 0, l-1)
			for s := start; s < top; s++ {
				base = append(base, byStr[s][0]) // 弱いスートを温存
			}
			// 先頭(最強)カードはスート違いをすべて候補化する
			for _, topIdx := range byStr[top] {
				cand := make([]int, 0, l)
				cand = append(cand, base...)
				cand = append(cand, topIdx)
				results = append(results, cand)
			}
		}
	}
	return results
}

// TienLenHint は人間へ勧める着手。Pass が true のときは Indices は空。
type TienLenHint struct {
	// Indices は手札のインデックス (出すカード)。
	Indices []int
	// Pass は「返せる手が無いのでパス」を意味する。
	Pass bool
}

// GetHint は人間の手番で勧める着手を返す。手番でなければ nil。
//
// **CPU の着手選択をそのまま使う。**同じ盤面の判断を 2 か所に書くと、片方だけ
// 直したときにヒントと CPU が別のことを言い出す (#5624)。
//
// **難易度は見ない。**難易度は CPU の強さの設定であって、人間へのアドバイスを
// 弱める設定ではない。Easy を選んだ人が「適当な手」を勧められても困る。
func (tl *TienLen) GetHint() *TienLenHint {
	if tl.round.gameEndFlag {
		return nil
	}
	player := tl.players[tl.round.currentTurn]
	if !player.GetIsHuman() {
		return nil
	}
	indices := tl.findBestPlayNormal(player)
	if len(indices) == 0 {
		return &TienLenHint{Pass: true}
	}
	return &TienLenHint{Indices: indices}
}
