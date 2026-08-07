//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

const (
	ChinesePokerPhaseBet      = 1
	ChinesePokerPhaseSetHands = 2
	ChinesePokerPhaseEnd      = 3
)

const (
	ChinesePokerDefaultChips = 1000
	ChinesePokerMinBet       = 10
	ChinesePokerMaxBet       = 10000
	ChinesePokerHandSize     = 13
	ChinesePokerFrontSize    = 3
	ChinesePokerMiddleSize   = 5
	ChinesePokerBackSize     = 5
)

// ChinesePoker チャイニーズポーカークラス
type ChinesePoker struct {
	trumpCards       *TrumpCards
	playerCards      []*Card
	dealerCards      []*Card
	playerFront      []*Card
	playerMiddle     []*Card
	playerBack       []*Card
	dealerFront      []*Card
	dealerMiddle     []*Card
	dealerBack       []*Card
	chips            ChipHolder
	bet              int
	phase            int
	gameEndFlag      bool
	result           GameResult
	frontResult      GameResult
	middleResult     GameResult
	backResult       GameResult
	payout           int
	playerFrontRank  int
	playerMiddleRank int
	playerBackRank   int
	dealerFrontRank  int
	dealerMiddleRank int
	dealerBackRank   int
	playerRoyalty    int
	dealerRoyalty    int
	scoop            bool
	actionLog        []*ActionLogEntry
}

// NewChinesePoker コンストラクタ
func NewChinesePoker(trumpCards *TrumpCards) *ChinesePoker {
	trumpCards.Shuffle()
	return &ChinesePoker{
		trumpCards: trumpCards,
		phase:      ChinesePokerPhaseBet,
	}
}

// NewDefaultChinesePoker デフォルト設定のチャイニーズポーカーを生成するファクトリ関数
func NewDefaultChinesePoker() *ChinesePoker {
	cp := NewChinesePoker(NewTrumpCards(0))
	cp.chips.SetChips(ChinesePokerDefaultChips)
	return cp
}

// Reset ゲーム初期化
func (cp *ChinesePoker) Reset() {
	cp.gameEndFlag = false
	cp.phase = ChinesePokerPhaseBet
	cp.playerCards = nil
	cp.dealerCards = nil
	cp.playerFront = nil
	cp.playerMiddle = nil
	cp.playerBack = nil
	cp.dealerFront = nil
	cp.dealerMiddle = nil
	cp.dealerBack = nil
	cp.bet = 0
	cp.result = 0
	cp.frontResult = 0
	cp.middleResult = 0
	cp.backResult = 0
	cp.payout = 0
	cp.playerFrontRank = 0
	cp.playerMiddleRank = 0
	cp.playerBackRank = 0
	cp.dealerFrontRank = 0
	cp.dealerMiddleRank = 0
	cp.dealerBackRank = 0
	cp.playerRoyalty = 0
	cp.dealerRoyalty = 0
	cp.scoop = false
	cp.actionLog = nil
	if cp.chips.GetChips() < ChinesePokerMinBet {
		cp.chips.SetChips(ChinesePokerDefaultChips)
	}
	cp.trumpCards = NewTrumpCards(0)
	cp.trumpCards.Shuffle()
}

// Bet ベット＆カード配布
func (cp *ChinesePoker) Bet(amount int) error {
	if cp.phase != ChinesePokerPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < ChinesePokerMinBet || amount%ChinesePokerMinBet != 0 || amount > ChinesePokerMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !cp.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	cp.bet = amount
	cp.cpAppendLog(0, "bet", fmt.Sprintf("bet=%d", amount), nil)

	cp.cpDeal()
	cp.phase = ChinesePokerPhaseSetHands
	return nil
}

// SetHands プレイヤーがフロント3枚とミドル5枚を指定する
func (cp *ChinesePoker) SetHands(frontIndices []int, middleIndices []int) error {
	if cp.phase != ChinesePokerPhaseSetHands {
		return NewDomainError(ErrWrongPhase, "SetHands is only allowed during the set hands phase.")
	}
	if len(frontIndices) != ChinesePokerFrontSize {
		return NewDomainError(ErrInvalidCard, "Front hand must have exactly 3 cards.")
	}
	if len(middleIndices) != ChinesePokerMiddleSize {
		return NewDomainError(ErrInvalidCard, "Middle hand must have exactly 5 cards.")
	}
	if len(cp.playerCards) < ChinesePokerHandSize {
		return NewDomainError(ErrInvalidPlay, "Insufficient player cards.")
	}

	var used [ChinesePokerHandSize]bool
	for _, idx := range frontIndices {
		if idx < 0 || idx >= ChinesePokerHandSize {
			return NewDomainError(ErrInvalidCard, "Card index out of range.")
		}
		if used[idx] {
			return NewDomainError(ErrInvalidCard, "Duplicate card index.")
		}
		used[idx] = true
	}
	for _, idx := range middleIndices {
		if idx < 0 || idx >= ChinesePokerHandSize {
			return NewDomainError(ErrInvalidCard, "Card index out of range.")
		}
		if used[idx] {
			return NewDomainError(ErrInvalidCard, "Duplicate card index.")
		}
		used[idx] = true
	}

	front := make([]*Card, ChinesePokerFrontSize)
	for i, idx := range frontIndices {
		front[i] = cp.playerCards[idx]
	}
	middle := make([]*Card, ChinesePokerMiddleSize)
	for i, idx := range middleIndices {
		middle[i] = cp.playerCards[idx]
	}
	back := make([]*Card, 0, ChinesePokerBackSize)
	for i, c := range cp.playerCards {
		if !used[i] {
			back = append(back, c)
		}
	}

	if !cpValidateHands(front, middle, back) {
		return NewDomainError(ErrInvalidPlay, "Foul: Back hand must be stronger than or equal to Middle, and Middle must be stronger than or equal to Front.")
	}

	cp.playerFront = front
	cp.playerMiddle = middle
	cp.playerBack = back

	cp.cpAppendLog(0, "set", fmt.Sprintf("front=%v middle=%v", frontIndices, middleIndices), nil)

	cp.dealerFront, cp.dealerMiddle, cp.dealerBack = cpHouseWay(cp.dealerCards)

	cp.cpResolve()
	return nil
}

// cpDeal 13枚ずつ配る
func (cp *ChinesePoker) cpDeal() {
	cp.playerCards = make([]*Card, 0, ChinesePokerHandSize)
	cp.dealerCards = make([]*Card, 0, ChinesePokerHandSize)
	for range ChinesePokerHandSize {
		cp.playerCards = append(cp.playerCards, cp.trumpCards.DrawCard())
		cp.dealerCards = append(cp.dealerCards, cp.trumpCards.DrawCard())
	}
	cp.cpAppendLog(-1, "deal", "dealt 13 cards each", nil)
}

// cpResolve ゲーム解決
func (cp *ChinesePoker) cpResolve() {
	cp.playerFrontRank = evalThreeCardHand(cp.playerFront)
	cp.playerMiddleRank = evalFiveCardHand(cp.playerMiddle)
	cp.playerBackRank = evalFiveCardHand(cp.playerBack)
	cp.dealerFrontRank = evalThreeCardHand(cp.dealerFront)
	cp.dealerMiddleRank = evalFiveCardHand(cp.dealerMiddle)
	cp.dealerBackRank = evalFiveCardHand(cp.dealerBack)

	frontCmp := compareThreeCardHands(cp.playerFront, cp.dealerFront)
	if frontCmp > 0 {
		cp.frontResult = GameResultWin
	} else {
		cp.frontResult = GameResultLose
	}

	middleCmp := cpCompareFiveCardHands(cp.playerMiddle, cp.dealerMiddle)
	if middleCmp > 0 {
		cp.middleResult = GameResultWin
	} else {
		cp.middleResult = GameResultLose
	}

	backCmp := cpCompareFiveCardHands(cp.playerBack, cp.dealerBack)
	if backCmp > 0 {
		cp.backResult = GameResultWin
	} else {
		cp.backResult = GameResultLose
	}

	wins := 0
	if cp.frontResult == GameResultWin {
		wins++
	}
	if cp.middleResult == GameResultWin {
		wins++
	}
	if cp.backResult == GameResultWin {
		wins++
	}

	cp.playerRoyalty = cpCalcRoyalty(cp.playerFrontRank, cp.playerMiddleRank, cp.playerBackRank, cp.playerFront)
	cp.dealerRoyalty = cpCalcRoyalty(cp.dealerFrontRank, cp.dealerMiddleRank, cp.dealerBackRank, cp.dealerFront)
	royaltyDiff := cp.playerRoyalty - cp.dealerRoyalty

	switch wins {
	case 3:
		cp.result = GameResultWin
		cp.scoop = true
		cp.payout = cp.bet*4 + cp.bet*royaltyDiff
		if cp.payout < cp.bet {
			cp.payout = cp.bet
		}
		cp.chips.AddChips(cp.payout)
	case 2:
		cp.result = GameResultWin
		cp.payout = cp.bet*2 + cp.bet*royaltyDiff
		if cp.payout < cp.bet {
			cp.payout = cp.bet
		}
		cp.chips.AddChips(cp.payout)
	case 1:
		cp.result = GameResultLose
		cp.payout = 0
		if royaltyDiff > 0 {
			bonus := cp.bet * royaltyDiff
			cp.payout = bonus
			cp.chips.AddChips(bonus)
		}
	case 0:
		cp.result = GameResultLose
		cp.scoop = true
		cp.payout = 0
	}

	cp.gameEndFlag = true
	cp.phase = ChinesePokerPhaseEnd

	var resultStr string
	switch cp.result {
	case GameResultWin:
		if cp.scoop {
			resultStr = "player scoop"
		} else {
			resultStr = "player wins"
		}
	default:
		if cp.scoop {
			resultStr = "dealer scoop"
		} else {
			resultStr = "dealer wins"
		}
	}
	cp.cpAppendLog(-1, "result", resultStr, nil)
}

// --- ハンド比較 ---

// cpCompareFiveCardHands 5枚ハンドを比較する（タイ=ディーラー勝ち → 0以下を返す）
func cpCompareFiveCardHands(a, b []*Card) int {
	rankA := evalFiveCardHand(a)
	rankB := evalFiveCardHand(b)
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}
	return compareHighCardsSlice(a, b)
}

// --- ファウルチェック ---

// ChinesePokerSuggestedArrangement は推奨する13枚の分け方。
type ChinesePokerSuggestedArrangement struct {
	// Front / Middle / Back は手札のインデックス (前列3枚・中列5枚・後列5枚)。
	Front  []int
	Middle []int
	Back   []int
	// Foul はこの分け方がファウル (前列 > 中列、または中列 > 後列) になるか。
	Foul bool
}

// GetSuggestedArrangement は13枚をランク順に「後列に強い5枚・中列に次の5枚・
// 前列に残り3枚」で分けた案を返す (#4717)。セットハンドフェーズで13枚
// そろっていないときは nil。
//
// **ランク順に切るだけでは合法とは限らない。**役のカテゴリは高い札の順に
// 従わないので、前列に低いスリーカードが入ると中列のハイカードより強くなって
// ファウルする。判定は実際の検証 cpValidateHands を通し、ファウルするなら
// そう伝える。**総当たりで直さない**のは 13C3 × 10C5 = 72,072 通りになるため
// で、フロントの getChinesePokerHint も同じ理由で同じ挙動にしている。
func (c *ChinesePoker) GetSuggestedArrangement() *ChinesePokerSuggestedArrangement {
	if c.phase != ChinesePokerPhaseSetHands || len(c.playerCards) != ChinesePokerHandSize {
		return nil
	}
	idx := make([]int, len(c.playerCards))
	for i := range idx {
		idx[i] = i
	}
	// **エースは 14。**value は 1 なので、そのまま並べると最強の札が前列に落ちる。
	rank := func(i int) int {
		if v := c.playerCards[i].GetValue(); v == 1 {
			return 14
		}
		return c.playerCards[i].GetValue()
	}
	sort.SliceStable(idx, func(a, b int) bool { return rank(idx[a]) > rank(idx[b]) })

	back := idx[:ChinesePokerBackSize]
	middle := idx[ChinesePokerBackSize : ChinesePokerBackSize+ChinesePokerMiddleSize]
	front := idx[ChinesePokerBackSize+ChinesePokerMiddleSize:]

	pick := func(indices []int) []*Card {
		out := make([]*Card, len(indices))
		for i, j := range indices {
			out[i] = c.playerCards[j]
		}
		return out
	}
	return &ChinesePokerSuggestedArrangement{
		Front:  front,
		Middle: middle,
		Back:   back,
		Foul:   !cpValidateHands(pick(front), pick(middle), pick(back)),
	}
}

// cpValidateHands Back ≥ Middle ≥ Front を検証する
func cpValidateHands(front, middle, back []*Card) bool {
	middleRank := evalFiveCardHand(middle)
	backRank := evalFiveCardHand(back)

	if backRank < middleRank {
		return false
	}
	if backRank == middleRank {
		cmp := compareHighCardsSlice(back, middle)
		if cmp < 0 {
			return false
		}
	}

	frontRank := evalThreeCardHand(front)
	return cpFrontNotStrongerThanMiddle(frontRank, front, middleRank, middle)
}

// cpFrontNotStrongerThanMiddle フロント（3枚）がミドル（5枚）より強くないか検証する
func cpFrontNotStrongerThanMiddle(frontRank int, front []*Card, middleRank int, middle []*Card) bool {
	cpFront := cpMapThreeCardToFiveCardRank(frontRank)
	if cpFront < middleRank {
		return true
	}
	if cpFront > middleRank {
		return false
	}
	frontVals := threeCardHandHighValues(front)
	middleVals := cpFiveCardHighValues(middle)
	for i := 0; i < len(frontVals) && i < len(middleVals); i++ {
		if frontVals[i] > middleVals[i] {
			return false
		}
		if frontVals[i] < middleVals[i] {
			return true
		}
	}
	return true
}

// cpMapThreeCardToFiveCardRank 3枚ランクを5枚ランク空間にマッピング
// Chinese Poker foul check: front must not be stronger than middle.
// 3-card straights/flushes are real hands — mapping them to HighCard
// would let a front straight-flush bypass the foul check.
func cpMapThreeCardToFiveCardRank(threeCardRank int) int {
	switch threeCardRank {
	case ThreeCardHandHighCard:
		return PokerHandHighCard
	case ThreeCardHandPair:
		return PokerHandOnePair
	case ThreeCardHandFlush:
		return PokerHandTwoPair
	case ThreeCardHandStraight:
		return PokerHandTwoPair
	case ThreeCardHandThreeOfAKind:
		return PokerHandThreeOfAKind
	case ThreeCardHandStraightFlush:
		return PokerHandStraight
	default:
		return PokerHandHighCard
	}
}

// cpFiveCardHighValues 5枚ハンドのカード値を降順で返す（比較用）
func cpFiveCardHighValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		vals[i] = v
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	return vals
}

// --- ロイヤリティ ---

// cpCalcRoyalty 3つのハンドのロイヤリティ合計を計算する
func cpCalcRoyalty(frontRank, middleRank, backRank int, front []*Card) int {
	return cpFrontRoyalty(frontRank, front) + cpMiddleRoyalty(middleRank) + cpBackRoyalty(backRank)
}

// cpBackRoyalty バックハンドのロイヤリティ
func cpBackRoyalty(rank int) int {
	switch rank {
	case PokerHandStraight:
		return 2
	case PokerHandFlush:
		return 4
	case PokerHandFullHouse:
		return 6
	case PokerHandFourOfAKind:
		return 10
	case PokerHandStraightFlush:
		return 15
	case PokerHandRoyalFlush:
		return 25
	default:
		return 0
	}
}

// cpMiddleRoyalty ミドルハンドのロイヤリティ
func cpMiddleRoyalty(rank int) int {
	switch rank {
	case PokerHandThreeOfAKind:
		return 2
	case PokerHandStraight:
		return 4
	case PokerHandFlush:
		return 8
	case PokerHandFullHouse:
		return 12
	case PokerHandFourOfAKind:
		return 20
	case PokerHandStraightFlush:
		return 30
	case PokerHandRoyalFlush:
		return 50
	default:
		return 0
	}
}

// cpFrontRoyalty フロントハンドのロイヤリティ
func cpFrontRoyalty(rank int, front []*Card) int {
	if rank == ThreeCardHandThreeOfAKind {
		v := front[0].GetValue()
		if v == 1 {
			return 22
		}
		if v >= 10 {
			return 12 + (v - 10)
		}
		return 10
	}
	if rank != ThreeCardHandPair {
		return 0
	}
	pairVal := cpPairValue(front)
	if pairVal < 6 {
		return 0
	}
	if pairVal == 14 {
		return 9
	}
	return 1 + (pairVal - 6)
}

// cpPairValue 3枚からペア値を抽出する（A=14）
func cpPairValue(cards []*Card) int {
	var freq [15]int
	for _, c := range cards {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	for v := 14; v >= 2; v-- {
		if freq[v] == 2 {
			return v
		}
	}
	return 0
}

// --- ハウスウェイ ---

// cpHouseWay ディーラーのハウスウェイ（自動カード分割）
func cpHouseWay(cards []*Card) (front, middle, back []*Card) {
	if len(cards) != ChinesePokerHandSize {
		return cards[:3], cards[3:8], cards[8:]
	}

	var bestFront, bestMiddle, bestBack []*Card
	bestScore := -1

	var frontBuf [3]*Card
	var middleBuf [5]*Card
	var backBuf [5]*Card
	var remaining [10]int

	for i := 0; i < ChinesePokerHandSize-2; i++ {
		for j := i + 1; j < ChinesePokerHandSize-1; j++ {
			for k := j + 1; k < ChinesePokerHandSize; k++ {
				frontBuf[0], frontBuf[1], frontBuf[2] = cards[i], cards[j], cards[k]
				frontSlice := frontBuf[:]

				ri := 0
				for idx := range ChinesePokerHandSize {
					if idx != i && idx != j && idx != k {
						remaining[ri] = idx
						ri++
					}
				}

				for mi0 := 0; mi0 < 6; mi0++ {
					for mi1 := mi0 + 1; mi1 < 7; mi1++ {
						for mi2 := mi1 + 1; mi2 < 8; mi2++ {
							for mi3 := mi2 + 1; mi3 < 9; mi3++ {
								for mi4 := mi3 + 1; mi4 < 10; mi4++ {
									middleBuf[0] = cards[remaining[mi0]]
									middleBuf[1] = cards[remaining[mi1]]
									middleBuf[2] = cards[remaining[mi2]]
									middleBuf[3] = cards[remaining[mi3]]
									middleBuf[4] = cards[remaining[mi4]]
									middleSlice := middleBuf[:]

									var bi int
									var midSet [10]bool
									midSet[mi0] = true
									midSet[mi1] = true
									midSet[mi2] = true
									midSet[mi3] = true
									midSet[mi4] = true
									for ri2 := range 10 {
										if !midSet[ri2] {
											backBuf[bi] = cards[remaining[ri2]]
											bi++
										}
									}
									backSlice := backBuf[:]

									if !cpValidateHands(frontSlice, middleSlice, backSlice) {
										continue
									}

									score := cpHouseWayScore(frontSlice, middleSlice, backSlice)
									if score > bestScore {
										bestScore = score
										bestFront = make([]*Card, 3)
										copy(bestFront, frontSlice)
										bestMiddle = make([]*Card, 5)
										copy(bestMiddle, middleSlice)
										bestBack = make([]*Card, 5)
										copy(bestBack, backSlice)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if bestScore < 0 {
		return cards[:3], cards[3:8], cards[8:]
	}
	return bestFront, bestMiddle, bestBack
}

// cpHouseWayScore ハウスウェイ用スコア（最も弱いハンドを最大化）
func cpHouseWayScore(front, middle, back []*Card) int {
	fr := evalThreeCardHand(front)
	mr := evalFiveCardHand(middle)
	br := evalFiveCardHand(back)

	fMapped := cpMapThreeCardToFiveCardRank(fr)

	minRank := fMapped
	if mr < minRank {
		minRank = mr
	}
	if br < minRank {
		minRank = br
	}

	return minRank*10000 + fMapped*100 + mr*10 + br +
		cpCalcRoyalty(fr, mr, br, front)
}

// cpAppendLog 棋譜にエントリを追加する
func (cp *ChinesePoker) cpAppendLog(playerIdx int, actionType, detail string, cards []*Card) {
	cp.actionLog = append(cp.actionLog, &ActionLogEntry{
		TurnNumber: len(cp.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPlayerCards プレイヤーの13枚を取得
func (cp *ChinesePoker) GetPlayerCards() []*Card { return cp.playerCards }

// GetDealerCards ディーラーの13枚を取得
func (cp *ChinesePoker) GetDealerCards() []*Card { return cp.dealerCards }

// GetPlayerFront プレイヤーフロントハンドを取得
func (cp *ChinesePoker) GetPlayerFront() []*Card { return cp.playerFront }

// GetPlayerMiddle プレイヤーミドルハンドを取得
func (cp *ChinesePoker) GetPlayerMiddle() []*Card { return cp.playerMiddle }

// GetPlayerBack プレイヤーバックハンドを取得
func (cp *ChinesePoker) GetPlayerBack() []*Card { return cp.playerBack }

// GetDealerFront ディーラーフロントハンドを取得
func (cp *ChinesePoker) GetDealerFront() []*Card { return cp.dealerFront }

// GetDealerMiddle ディーラーミドルハンドを取得
func (cp *ChinesePoker) GetDealerMiddle() []*Card { return cp.dealerMiddle }

// GetDealerBack ディーラーバックハンドを取得
func (cp *ChinesePoker) GetDealerBack() []*Card { return cp.dealerBack }

// GetPhase 現在のフェーズ
func (cp *ChinesePoker) GetPhase() int { return cp.phase }

// GetGameEndFlag ゲーム終了フラグ
func (cp *ChinesePoker) GetGameEndFlag() bool { return cp.gameEndFlag }

// GetBet ベット額
func (cp *ChinesePoker) GetBet() int { return cp.bet }

// GetResult ゲーム結果
func (cp *ChinesePoker) GetResult() GameResult { return cp.result }

// GetFrontResult フロントハンド結果
func (cp *ChinesePoker) GetFrontResult() GameResult { return cp.frontResult }

// GetMiddleResult ミドルハンド結果
func (cp *ChinesePoker) GetMiddleResult() GameResult { return cp.middleResult }

// GetBackResult バックハンド結果
func (cp *ChinesePoker) GetBackResult() GameResult { return cp.backResult }

// GetPayout 配当
func (cp *ChinesePoker) GetPayout() int { return cp.payout }

// GetPlayerFrontRank プレイヤーフロントハンドランク
func (cp *ChinesePoker) GetPlayerFrontRank() int { return cp.playerFrontRank }

// GetPlayerMiddleRank プレイヤーミドルハンドランク
func (cp *ChinesePoker) GetPlayerMiddleRank() int { return cp.playerMiddleRank }

// GetPlayerBackRank プレイヤーバックハンドランク
func (cp *ChinesePoker) GetPlayerBackRank() int { return cp.playerBackRank }

// GetDealerFrontRank ディーラーフロントハンドランク
func (cp *ChinesePoker) GetDealerFrontRank() int { return cp.dealerFrontRank }

// GetDealerMiddleRank ディーラーミドルハンドランク
func (cp *ChinesePoker) GetDealerMiddleRank() int { return cp.dealerMiddleRank }

// GetDealerBackRank ディーラーバックハンドランク
func (cp *ChinesePoker) GetDealerBackRank() int { return cp.dealerBackRank }

// GetPlayerRoyalty プレイヤーロイヤリティ
func (cp *ChinesePoker) GetPlayerRoyalty() int { return cp.playerRoyalty }

// GetDealerRoyalty ディーラーロイヤリティ
func (cp *ChinesePoker) GetDealerRoyalty() int { return cp.dealerRoyalty }

// GetScoop スクープフラグ
func (cp *ChinesePoker) GetScoop() bool { return cp.scoop }

// GetChips チップ
func (cp *ChinesePoker) GetChips() int { return cp.chips.GetChips() }

// GetActionLog 棋譜を取得する
func (cp *ChinesePoker) GetActionLog() []*ActionLogEntry { return cp.actionLog }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (cp *ChinesePoker) SetPhase(phase int) { cp.phase = phase }

// SetPlayerCards プレイヤーカード設定（テスト用）
func (cp *ChinesePoker) SetPlayerCards(cards []*Card) { cp.playerCards = cards }

// SetDealerCards ディーラーカード設定（テスト用）
func (cp *ChinesePoker) SetDealerCards(cards []*Card) { cp.dealerCards = cards }

// SetPlayerFront プレイヤーフロント設定（テスト用）
func (cp *ChinesePoker) SetPlayerFront(cards []*Card) { cp.playerFront = cards }

// SetPlayerMiddle プレイヤーミドル設定（テスト用）
func (cp *ChinesePoker) SetPlayerMiddle(cards []*Card) { cp.playerMiddle = cards }

// SetPlayerBack プレイヤーバック設定（テスト用）
func (cp *ChinesePoker) SetPlayerBack(cards []*Card) { cp.playerBack = cards }

// SetDealerFront ディーラーフロント設定（テスト用）
func (cp *ChinesePoker) SetDealerFront(cards []*Card) { cp.dealerFront = cards }

// SetDealerMiddle ディーラーミドル設定（テスト用）
func (cp *ChinesePoker) SetDealerMiddle(cards []*Card) { cp.dealerMiddle = cards }

// SetDealerBack ディーラーバック設定（テスト用）
func (cp *ChinesePoker) SetDealerBack(cards []*Card) { cp.dealerBack = cards }

// SetBet ベット額設定（テスト用）
func (cp *ChinesePoker) SetBet(amount int) { cp.bet = amount }

// SetChips チップ設定（テスト用）
func (cp *ChinesePoker) SetChips(chips int) { cp.chips.SetChips(chips) }

// SetResult ゲーム結果設定（テスト用）
func (cp *ChinesePoker) SetResult(result GameResult) { cp.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (cp *ChinesePoker) SetGameEndFlag(flag bool) { cp.gameEndFlag = flag }

// SetFrontResult フロントハンド結果設定（テスト用）
func (cp *ChinesePoker) SetFrontResult(result GameResult) { cp.frontResult = result }

// SetMiddleResult ミドルハンド結果設定（テスト用）
func (cp *ChinesePoker) SetMiddleResult(result GameResult) { cp.middleResult = result }

// SetBackResult バックハンド結果設定（テスト用）
func (cp *ChinesePoker) SetBackResult(result GameResult) { cp.backResult = result }

// SetPayout 配当設定（テスト用）
func (cp *ChinesePoker) SetPayout(payout int) { cp.payout = payout }

// SetPlayerFrontRank プレイヤーフロントハンドランク設定（テスト用）
func (cp *ChinesePoker) SetPlayerFrontRank(rank int) { cp.playerFrontRank = rank }

// SetPlayerMiddleRank プレイヤーミドルハンドランク設定（テスト用）
func (cp *ChinesePoker) SetPlayerMiddleRank(rank int) { cp.playerMiddleRank = rank }

// SetPlayerBackRank プレイヤーバックハンドランク設定（テスト用）
func (cp *ChinesePoker) SetPlayerBackRank(rank int) { cp.playerBackRank = rank }

// SetDealerFrontRank ディーラーフロントハンドランク設定（テスト用）
func (cp *ChinesePoker) SetDealerFrontRank(rank int) { cp.dealerFrontRank = rank }

// SetDealerMiddleRank ディーラーミドルハンドランク設定（テスト用）
func (cp *ChinesePoker) SetDealerMiddleRank(rank int) { cp.dealerMiddleRank = rank }

// SetDealerBackRank ディーラーバックハンドランク設定（テスト用）
func (cp *ChinesePoker) SetDealerBackRank(rank int) { cp.dealerBackRank = rank }

// SetPlayerRoyalty プレイヤーロイヤリティ設定（テスト用）
func (cp *ChinesePoker) SetPlayerRoyalty(royalty int) { cp.playerRoyalty = royalty }

// SetDealerRoyalty ディーラーロイヤリティ設定（テスト用）
func (cp *ChinesePoker) SetDealerRoyalty(royalty int) { cp.dealerRoyalty = royalty }

// SetScoop スクープフラグ設定（テスト用）
func (cp *ChinesePoker) SetScoop(scoop bool) { cp.scoop = scoop }

// chinesePokerJSON is the JSON wire format for ChinesePoker.
type chinesePokerJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	PlayerCards      []*Card           `json:"pc"`
	DealerCards      []*Card           `json:"dc"`
	PlayerFront      []*Card           `json:"pf"`
	PlayerMiddle     []*Card           `json:"pm"`
	PlayerBack       []*Card           `json:"pb"`
	DealerFront      []*Card           `json:"df"`
	DealerMiddle     []*Card           `json:"dm"`
	DealerBack       []*Card           `json:"db"`
	Chips            *ChipHolder       `json:"ch"`
	Bet              int               `json:"bt"`
	Phase            int               `json:"ps"`
	GameEndFlag      bool              `json:"ge"`
	Result           GameResult        `json:"rs"`
	FrontResult      GameResult        `json:"fr"`
	MiddleResult     GameResult        `json:"mr"`
	BackResult       GameResult        `json:"br"`
	Payout           int               `json:"po"`
	PlayerFrontRank  int               `json:"pfr"`
	PlayerMiddleRank int               `json:"pmr"`
	PlayerBackRank   int               `json:"pbr"`
	DealerFrontRank  int               `json:"dfr"`
	DealerMiddleRank int               `json:"dmr"`
	DealerBackRank   int               `json:"dbr"`
	PlayerRoyalty    int               `json:"pry"`
	DealerRoyalty    int               `json:"dry"`
	Scoop            bool              `json:"sc"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (cp *ChinesePoker) MarshalJSON() ([]byte, error) {
	return json.Marshal(chinesePokerJSON{
		TrumpCards:       cp.trumpCards,
		PlayerCards:      cp.playerCards,
		DealerCards:      cp.dealerCards,
		PlayerFront:      cp.playerFront,
		PlayerMiddle:     cp.playerMiddle,
		PlayerBack:       cp.playerBack,
		DealerFront:      cp.dealerFront,
		DealerMiddle:     cp.dealerMiddle,
		DealerBack:       cp.dealerBack,
		Chips:            &cp.chips,
		Bet:              cp.bet,
		Phase:            cp.phase,
		GameEndFlag:      cp.gameEndFlag,
		Result:           cp.result,
		FrontResult:      cp.frontResult,
		MiddleResult:     cp.middleResult,
		BackResult:       cp.backResult,
		Payout:           cp.payout,
		PlayerFrontRank:  cp.playerFrontRank,
		PlayerMiddleRank: cp.playerMiddleRank,
		PlayerBackRank:   cp.playerBackRank,
		DealerFrontRank:  cp.dealerFrontRank,
		DealerMiddleRank: cp.dealerMiddleRank,
		DealerBackRank:   cp.dealerBackRank,
		PlayerRoyalty:    cp.playerRoyalty,
		DealerRoyalty:    cp.dealerRoyalty,
		Scoop:            cp.scoop,
		ActionLog:        cp.actionLog,
	})
}

const cpMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (cp *ChinesePoker) UnmarshalJSON(data []byte) error {
	var j chinesePokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerCards) > cpMaxSliceLen || len(j.DealerCards) > cpMaxSliceLen ||
		len(j.PlayerFront) > cpMaxSliceLen || len(j.PlayerMiddle) > cpMaxSliceLen ||
		len(j.PlayerBack) > cpMaxSliceLen || len(j.DealerFront) > cpMaxSliceLen ||
		len(j.DealerMiddle) > cpMaxSliceLen || len(j.DealerBack) > cpMaxSliceLen ||
		len(j.ActionLog) > cpMaxSliceLen {
		return fmt.Errorf("chinesepoker: input array exceeds maximum allowed size")
	}

	cp.trumpCards = j.TrumpCards
	if cp.trumpCards == nil {
		cp.trumpCards = NewTrumpCards(0)
	}
	cp.playerCards = j.PlayerCards
	if cp.playerCards == nil {
		cp.playerCards = make([]*Card, 0)
	}
	cp.dealerCards = j.DealerCards
	if cp.dealerCards == nil {
		cp.dealerCards = make([]*Card, 0)
	}
	cp.playerFront = j.PlayerFront
	if cp.playerFront == nil {
		cp.playerFront = make([]*Card, 0)
	}
	cp.playerMiddle = j.PlayerMiddle
	if cp.playerMiddle == nil {
		cp.playerMiddle = make([]*Card, 0)
	}
	cp.playerBack = j.PlayerBack
	if cp.playerBack == nil {
		cp.playerBack = make([]*Card, 0)
	}
	cp.dealerFront = j.DealerFront
	if cp.dealerFront == nil {
		cp.dealerFront = make([]*Card, 0)
	}
	cp.dealerMiddle = j.DealerMiddle
	if cp.dealerMiddle == nil {
		cp.dealerMiddle = make([]*Card, 0)
	}
	cp.dealerBack = j.DealerBack
	if cp.dealerBack == nil {
		cp.dealerBack = make([]*Card, 0)
	}
	if j.Chips != nil {
		cp.chips = *j.Chips
	}
	cp.bet = j.Bet
	cp.phase = j.Phase
	cp.gameEndFlag = j.GameEndFlag
	cp.result = j.Result
	cp.frontResult = j.FrontResult
	cp.middleResult = j.MiddleResult
	cp.backResult = j.BackResult
	cp.payout = j.Payout
	cp.playerFrontRank = j.PlayerFrontRank
	cp.playerMiddleRank = j.PlayerMiddleRank
	cp.playerBackRank = j.PlayerBackRank
	cp.dealerFrontRank = j.DealerFrontRank
	cp.dealerMiddleRank = j.DealerMiddleRank
	cp.dealerBackRank = j.DealerBackRank
	cp.playerRoyalty = j.PlayerRoyalty
	cp.dealerRoyalty = j.DealerRoyalty
	cp.scoop = j.Scoop
	cp.actionLog = j.ActionLog
	if cp.actionLog == nil {
		cp.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
