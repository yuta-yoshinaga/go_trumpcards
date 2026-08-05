package domain

import "encoding/json"

// CribbageScoreDetail クリベッジのスコア内訳
type CribbageScoreDetail struct {
	Fifteens int
	Pairs    int
	Runs     int
	Flush    int
	Nobs     int
	Total    int
	Cards    []*Card // スコア対象カード (hand + starter)
}

// cribbageScoreDetailJSON is the JSON wire format for CribbageScoreDetail.
type cribbageScoreDetailJSON struct {
	Fifteens int     `json:"f"`
	Pairs    int     `json:"p"`
	Runs     int     `json:"r"`
	Flush    int     `json:"fl"`
	Nobs     int     `json:"n"`
	Total    int     `json:"t"`
	Cards    []*Card `json:"cs"`
}

// MarshalJSON implements json.Marshaler.
func (d CribbageScoreDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(cribbageScoreDetailJSON(d))
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *CribbageScoreDetail) UnmarshalJSON(data []byte) error {
	var j cribbageScoreDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Fifteens = j.Fifteens
	d.Pairs = j.Pairs
	d.Runs = j.Runs
	d.Flush = j.Flush
	d.Nobs = j.Nobs
	d.Total = j.Total
	d.Cards = j.Cards
	return nil
}

// cribbageCardValue クリベッジ用カード値 (A=1, 2-9=face, 10/J/Q/K=10)
func cribbageCardValue(c *Card) int {
	v := c.GetValue()
	if v >= 10 {
		return 10
	}
	return v
}

// CribbageScoreFifteens 15になる組み合わせを数える (各2点)
func CribbageScoreFifteens(cards []*Card) int {
	count := 0
	n := len(cards)
	// 全部分集合を列挙 (2^n - 1)
	for mask := 1; mask < (1 << n); mask++ {
		sum := 0
		for i := range n {
			if mask&(1<<i) != 0 {
				sum += cribbageCardValue(cards[i])
			}
		}
		if sum == 15 {
			count++
		}
	}
	return count * 2
}

// CribbageScorePairs 同じランクのペアを数える (各2点)
func CribbageScorePairs(cards []*Card) int {
	count := 0
	for i := range len(cards) {
		for j := i + 1; j < len(cards); j++ {
			if cards[i].GetValue() == cards[j].GetValue() {
				count++
			}
		}
	}
	return count * 2
}

// CribbageScoreRuns ランを検出してスコアを返す
// ランクの連続3枚以上がランとなる。重複ランクがある場合は掛け算 (ダブルラン等)
func CribbageScoreRuns(cards []*Card) int {
	// ランクごとのカード数をカウント
	rankCount := make(map[int]int)
	for _, c := range cards {
		rankCount[c.GetValue()]++
	}

	// 存在するランクをソート
	ranks := make([]int, 0, len(rankCount))
	for r := range rankCount {
		ranks = append(ranks, r)
	}
	sortInts(ranks)

	bestScore := 0

	// 連続するランクの区間を見つける
	for i := 0; i < len(ranks); i++ {
		// 連続区間の開始
		j := i
		for j < len(ranks)-1 && ranks[j+1] == ranks[j]+1 {
			j++
		}
		runLen := j - i + 1
		if runLen >= 3 {
			// ランの長さ × 各ランクの出現回数の積
			multiplier := 1
			for k := i; k <= j; k++ {
				multiplier *= rankCount[ranks[k]]
			}
			score := runLen * multiplier
			if score > bestScore {
				bestScore = score
			}
		}
		// 次の区間へ (ただし重複チェックのため i は j まで進めない)
	}

	return bestScore
}

// CribbageScoreFlush フラッシュのスコアを返す
// hand の4枚が同一スートなら4点、starter も同じなら5点
// crib の場合は5枚全て同一スートでないと0点
func CribbageScoreFlush(hand []*Card, starter *Card, isCrib bool) int {
	if len(hand) < 4 {
		return 0
	}
	suit := hand[0].GetDesign()
	for i := 1; i < 4; i++ {
		if hand[i].GetDesign() != suit {
			return 0
		}
	}
	if starter != nil && starter.GetDesign() == suit {
		return 5
	}
	if isCrib {
		return 0 // crib は5枚全て同一でないと0
	}
	return 4
}

// CribbageScoreNobs ノブのスコアを返す (手札のJがstarterのスートと一致 → 1点)
func CribbageScoreNobs(hand []*Card, starter *Card) int {
	if starter == nil {
		return 0
	}
	for _, c := range hand {
		if c.GetValue() == CribbageJackValue && c.GetDesign() == starter.GetDesign() {
			return 1
		}
	}
	return 0
}

// CribbageScoreHand ハンド全体のスコアを計算する
func CribbageScoreHand(hand []*Card, starter *Card, isCrib bool) CribbageScoreDetail {
	allCards := make([]*Card, 0, 5)
	allCards = append(allCards, hand...)
	if starter != nil {
		allCards = append(allCards, starter)
	}

	detail := CribbageScoreDetail{
		Fifteens: CribbageScoreFifteens(allCards),
		Pairs:    CribbageScorePairs(allCards),
		Runs:     CribbageScoreRuns(allCards),
		Flush:    CribbageScoreFlush(hand, starter, isCrib),
		Nobs:     CribbageScoreNobs(hand, starter),
		Cards:    allCards,
	}
	detail.Total = detail.Fifteens + detail.Pairs + detail.Runs + detail.Flush + detail.Nobs
	return detail
}

// CribbageScorePegging ペギングのスコアを計算する
// playedCards は現在のペギングシーケンスで最後に追加されたカードがスコア対象
func CribbageScorePegging(playedCards []*Card, pegCount int) int {
	if len(playedCards) == 0 {
		return 0
	}
	score := 0

	// 15 or 31
	if pegCount == 15 {
		score += 2
	}
	if pegCount == 31 {
		score += 2
	}

	n := len(playedCards)

	// ペア/スリーカード/フォーカード (末尾から同じランクを数える)
	lastRank := playedCards[n-1].GetValue()
	pairCount := 0
	for i := n - 2; i >= 0; i-- {
		if playedCards[i].GetValue() == lastRank {
			pairCount++
		} else {
			break
		}
	}
	switch pairCount {
	case 1:
		score += 2 // pair
	case 2:
		score += 6 // three of a kind
	case 3:
		score += 12 // four of a kind
	}

	// ラン検出: 末尾からN枚を取り、ソートして連続かチェック (N=3,4,5,...,n)
	bestRun := 0
	for length := 3; length <= n; length++ {
		subset := make([]int, length)
		for i := 0; i < length; i++ {
			subset[i] = playedCards[n-length+i].GetValue()
		}
		sortInts(subset)
		isRun := true
		for i := 1; i < length; i++ {
			if subset[i] != subset[i-1]+1 {
				isRun = false
				break
			}
		}
		if isRun {
			bestRun = length
		}
	}
	if bestRun >= 3 {
		score += bestRun
	}

	return score
}

// sortInts 整数スライスをソート (sort パッケージを避けるシンプル実装)
func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}
