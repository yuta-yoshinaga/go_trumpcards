package domain

import "sort"

// HoldemPlayer テキサスホールデムプレイヤークラス
type HoldemPlayer struct {
	Player                     // 親クラス
	isHuman    bool            // 人間フラグ
	chips      int             // チップ
	handRank   int             // ベストハンドランク
	bestHand   []*Card         // ベスト5枚
	folded     bool            // フォールド済
	allIn      bool            // オールイン済
	currentBet int             // 現ラウンドベット額
	playStyle  HoldemPlayStyle // CPUプレイスタイル
}

// NewHoldemPlayer コンストラクタ
func NewHoldemPlayer(isHuman bool, style HoldemPlayStyle) *HoldemPlayer {
	return &HoldemPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (hp *HoldemPlayer) GetIsHuman() bool { return hp.isHuman }

// GetChips チップ取得
func (hp *HoldemPlayer) GetChips() int { return hp.chips }

// SetChips チップ設定
func (hp *HoldemPlayer) SetChips(chips int) { hp.chips = chips }

// AddChips チップ追加
func (hp *HoldemPlayer) AddChips(amount int) { hp.chips += amount }

// SubtractChips チップ減算 (不足時はfalseを返す)
func (hp *HoldemPlayer) SubtractChips(amount int) bool {
	if hp.chips < amount {
		return false
	}
	hp.chips -= amount
	return true
}

// GetHandRank ハンドランク取得
func (hp *HoldemPlayer) GetHandRank() int { return hp.handRank }

// GetBestHand ベストハンド取得
func (hp *HoldemPlayer) GetBestHand() []*Card { return hp.bestHand }

// GetFolded フォールド状態取得
func (hp *HoldemPlayer) GetFolded() bool { return hp.folded }

// SetFolded フォールド状態設定
func (hp *HoldemPlayer) SetFolded(folded bool) { hp.folded = folded }

// GetAllIn オールイン状態取得
func (hp *HoldemPlayer) GetAllIn() bool { return hp.allIn }

// SetAllIn オールイン状態設定
func (hp *HoldemPlayer) SetAllIn(allIn bool) { hp.allIn = allIn }

// GetCurrentBet 現ラウンドベット取得
func (hp *HoldemPlayer) GetCurrentBet() int { return hp.currentBet }

// SetCurrentBet 現ラウンドベット設定
func (hp *HoldemPlayer) SetCurrentBet(bet int) { hp.currentBet = bet }

// GetPlayStyle プレイスタイル取得
func (hp *HoldemPlayer) GetPlayStyle() HoldemPlayStyle { return hp.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (hp *HoldemPlayer) GetPlayStyleName() string {
	if int(hp.playStyle) < len(HoldemPlayStyleNames) {
		return HoldemPlayStyleNames[hp.playStyle]
	}
	return "Unknown"
}

// SetHandRank ハンドランク設定（テスト用）
func (hp *HoldemPlayer) SetHandRank(rank int) { hp.handRank = rank }

// SetBestHand ベストハンド設定（テスト用）
func (hp *HoldemPlayer) SetBestHand(hand []*Card) { hp.bestHand = hand }

// EvalBestHand コミュニティカードとホールカードからベスト5枚を評価
func (hp *HoldemPlayer) EvalBestHand(communityCards []*Card) int {
	all := make([]*Card, 0, len(hp.cards)+len(communityCards))
	all = append(all, hp.cards...)
	all = append(all, communityCards...)

	if len(all) < 5 {
		hp.handRank = PokerHandHighCard
		hp.bestHand = nil
		return hp.handRank
	}

	combos := combinations(all, 5)
	bestRank := -1
	var bestCards []*Card

	for _, combo := range combos {
		rank := evalFiveCards(combo)
		if rank > bestRank || (rank == bestRank && compareHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}

	hp.handRank = bestRank
	hp.bestHand = bestCards
	return hp.handRank
}

// evalFiveCards 5枚のカードからハンドランクを評価
func evalFiveCards(cards []*Card) int {
	if len(cards) != 5 {
		return PokerHandHighCard
	}

	values := make([]int, 5)
	designs := make([]int, 5)
	for i, c := range cards {
		values[i] = c.GetValue()
		designs[i] = c.GetDesign()
	}
	sort.Ints(values)

	// フラッシュチェック
	isFlush := true
	for i := 1; i < 5; i++ {
		if designs[i] != designs[0] {
			isFlush = false
			break
		}
	}

	// ストレートチェック
	isStraight := checkStraightValues(values)

	// カード値の出現回数カウント
	valueCounts := make(map[int]int)
	for _, v := range values {
		valueCounts[v]++
	}
	counts := make([]int, 0, len(valueCounts))
	for _, c := range valueCounts {
		counts = append(counts, c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))

	if isFlush && isStraight {
		if checkRoyalStraightValues(values) {
			return PokerHandRoyalFlush
		}
		return PokerHandStraightFlush
	}
	if counts[0] == 4 {
		return PokerHandFourOfAKind
	}
	if len(counts) >= 2 && counts[0] == 3 && counts[1] == 2 {
		return PokerHandFullHouse
	}
	if isFlush {
		return PokerHandFlush
	}
	if isStraight {
		return PokerHandStraight
	}
	if counts[0] == 3 {
		return PokerHandThreeOfAKind
	}
	if len(counts) >= 2 && counts[0] == 2 && counts[1] == 2 {
		return PokerHandTwoPair
	}
	if counts[0] == 2 {
		return PokerHandOnePair
	}
	return PokerHandHighCard
}

// checkStraightValues ストレートチェック (ソート済み値スライス)
func checkStraightValues(sortedValues []int) bool {
	// A-2-3-4-5 (ホイール)
	if len(sortedValues) == 5 &&
		sortedValues[0] == 1 && sortedValues[1] == 2 &&
		sortedValues[2] == 3 && sortedValues[3] == 4 && sortedValues[4] == 5 {
		return true
	}
	// A-10-J-Q-K (ブロードウェイ)
	if len(sortedValues) == 5 &&
		sortedValues[0] == 1 && sortedValues[1] == 10 &&
		sortedValues[2] == 11 && sortedValues[3] == 12 && sortedValues[4] == 13 {
		return true
	}
	// 通常のストレート
	for i := 1; i < len(sortedValues); i++ {
		if sortedValues[i] != sortedValues[i-1]+1 {
			return false
		}
	}
	return true
}

// checkRoyalStraightValues ロイヤルフラッシュかチェック
func checkRoyalStraightValues(sortedValues []int) bool {
	return len(sortedValues) == 5 &&
		sortedValues[0] == 1 &&
		sortedValues[1] == 10 &&
		sortedValues[2] == 11 &&
		sortedValues[3] == 12 &&
		sortedValues[4] == 13
}

// combinations n枚からk枚を選ぶ全組み合わせを返す
func combinations(cards []*Card, k int) [][]*Card {
	var result [][]*Card
	n := len(cards)
	if k > n {
		return result
	}
	combo := make([]int, k)
	var generate func(start, idx int)
	generate = func(start, idx int) {
		if idx == k {
			hand := make([]*Card, k)
			for i, ci := range combo {
				hand[i] = cards[ci]
			}
			result = append(result, hand)
			return
		}
		for i := start; i <= n-(k-idx); i++ {
			combo[idx] = i
			generate(i+1, idx+1)
		}
	}
	generate(0, 0)
	return result
}

// isWheelHand ホイール (A-2-3-4-5) かどうか判定
func isWheelHand(cards []*Card) bool {
	if len(cards) != 5 {
		return false
	}
	vals := make([]int, 5)
	for i, c := range cards {
		vals[i] = c.GetValue()
	}
	sort.Ints(vals)
	return vals[0] == 1 && vals[1] == 2 && vals[2] == 3 && vals[3] == 4 && vals[4] == 5
}

// compareHighCardsSlice 2つの5枚ハンドのハイカード比較 (a > b: 1, a < b: -1, a == b: 0)
func compareHighCardsSlice(a, b []*Card) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	aWheel := isWheelHand(a)
	bWheel := isWheelHand(b)
	aVals := make([]int, len(a))
	bVals := make([]int, len(b))
	for i, c := range a {
		v := c.GetValue()
		if v == 1 && !aWheel {
			v = 14
		}
		aVals[i] = v
	}
	for i, c := range b {
		v := c.GetValue()
		if v == 1 && !bWheel {
			v = 14
		}
		bVals[i] = v
	}
	sort.Sort(sort.Reverse(sort.IntSlice(aVals)))
	sort.Sort(sort.Reverse(sort.IntSlice(bVals)))
	for i := 0; i < len(aVals) && i < len(bVals); i++ {
		if aVals[i] > bVals[i] {
			return 1
		}
		if aVals[i] < bVals[i] {
			return -1
		}
	}
	return 0
}
