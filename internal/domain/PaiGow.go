//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// パイガオポーカーフェーズ定数
const (
	PaiGowPhaseBet      = 1 // ベットフェーズ
	PaiGowPhaseSetHands = 2 // ハンド設定フェーズ（ロー/ハイ分割）
	PaiGowPhaseEnd      = 3 // 終了フェーズ
)

// パイガオポーカーデフォルト値
const (
	PaiGowDefaultChips   = 1000  // デフォルトチップ
	PaiGowMinBet         = 10    // 最低ベット額
	PaiGowMaxBet         = 10000 // 最大ベット額
	PaiGowHandSize       = 7     // 配布カード枚数
	PaiGowHighHandSize   = 5     // ハイハンド枚数
	PaiGowLowHandSize    = 2     // ローハンド枚数
	PaiGowCommissionRate = 5     // コミッション率（%）
	PaiGowJokerCount     = 1     // ジョーカー枚数
)

// パイガオ2枚ハンドランク定数
const (
	PaiGowLowHandHighCard = 0 // ハイカード
	PaiGowLowHandPair     = 1 // ペア
)

// PaiGowLowHandNames 2枚ハンド名
var PaiGowLowHandNames = []string{
	"High Card",
	"Pair",
}

// PaiGow パイガオポーカークラス
type PaiGow struct {
	trumpCards     *TrumpCards // トランプカード
	playerCards    []*Card     // プレイヤーの7枚
	dealerCards    []*Card     // ディーラーの7枚
	playerHighHand []*Card     // プレイヤーハイハンド（5枚）
	playerLowHand  []*Card     // プレイヤーローハンド（2枚）
	dealerHighHand []*Card     // ディーラーハイハンド（5枚）
	dealerLowHand  []*Card     // ディーラーローハンド（2枚）
	chips          ChipHolder  // チップ
	bet            int         // ベット額
	phase          int         // 現在のフェーズ
	gameEndFlag    bool        // ゲーム終了フラグ
	result         GameResult  // ゲーム結果
	highHandResult GameResult  // ハイハンド結果
	lowHandResult  GameResult  // ローハンド結果
	payout         int         // 配当
	commission     int         // コミッション
	playerHighRank int         // プレイヤーハイハンドランク
	playerLowRank  int         // プレイヤーローハンドランク
	dealerHighRank int         // ディーラーハイハンドランク
	dealerLowRank  int         // ディーラーローハンドランク
	actionLogBase
}

// NewPaiGow コンストラクタ
func NewPaiGow(trumpCards *TrumpCards) *PaiGow {
	trumpCards.Shuffle()
	return &PaiGow{
		trumpCards: trumpCards,
		phase:      PaiGowPhaseBet,
	}
}

// NewDefaultPaiGow デフォルト設定のパイガオポーカーを生成するファクトリ関数
func NewDefaultPaiGow() *PaiGow {
	pg := NewPaiGow(NewTrumpCards(PaiGowJokerCount))
	pg.chips.SetChips(PaiGowDefaultChips)
	return pg
}

// Reset ゲーム初期化
func (pg *PaiGow) Reset() {
	pg.gameEndFlag = false
	pg.phase = PaiGowPhaseBet
	pg.playerCards = nil
	pg.dealerCards = nil
	pg.playerHighHand = nil
	pg.playerLowHand = nil
	pg.dealerHighHand = nil
	pg.dealerLowHand = nil
	pg.bet = 0
	pg.result = 0
	pg.highHandResult = 0
	pg.lowHandResult = 0
	pg.payout = 0
	pg.commission = 0
	pg.playerHighRank = 0
	pg.playerLowRank = 0
	pg.dealerHighRank = 0
	pg.dealerLowRank = 0
	pg.actionLog = nil
	if pg.chips.GetChips() < PaiGowMinBet {
		pg.chips.SetChips(PaiGowDefaultChips)
	}
	pg.trumpCards = NewTrumpCards(PaiGowJokerCount)
	pg.trumpCards.Shuffle()
}

// Bet ベット＆カード配布
func (pg *PaiGow) Bet(amount int) error {
	if pg.phase != PaiGowPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < PaiGowMinBet || amount%PaiGowMinBet != 0 || amount > PaiGowMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !pg.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	pg.bet = amount
	pg.appendLog(0, "bet", fmt.Sprintf("bet=%d", amount), nil)

	// ディール: 7枚ずつ配る
	pg.deal()
	pg.phase = PaiGowPhaseSetHands
	return nil
}

// SetHands プレイヤーがローハンドの2枚を指定する
func (pg *PaiGow) SetHands(lowIdx0, lowIdx1 int) error {
	if pg.phase != PaiGowPhaseSetHands {
		return NewDomainError(ErrWrongPhase, "SetHands is only allowed during the set hands phase.")
	}
	if lowIdx0 == lowIdx1 {
		return NewDomainError(ErrInvalidCard, "Low hand indices must be different.")
	}
	if lowIdx0 < 0 || lowIdx0 >= PaiGowHandSize || lowIdx1 < 0 || lowIdx1 >= PaiGowHandSize {
		return NewDomainError(ErrInvalidCard, "Card index out of range.")
	}

	// ローハンドとハイハンドに分割
	pg.playerLowHand = []*Card{pg.playerCards[lowIdx0], pg.playerCards[lowIdx1]}
	pg.playerHighHand = make([]*Card, 0, PaiGowHighHandSize)
	for i, c := range pg.playerCards {
		if i != lowIdx0 && i != lowIdx1 {
			pg.playerHighHand = append(pg.playerHighHand, c)
		}
	}

	// ハイハンドがローハンドより強いか検証
	highRank := evalPaiGowHighHand(pg.playerHighHand)
	lowRank := evalPaiGowLowHand(pg.playerLowHand)
	if !paiGowHighBeatsLow(highRank, pg.playerHighHand, lowRank, pg.playerLowHand) {
		// 分割を元に戻す
		pg.playerHighHand = nil
		pg.playerLowHand = nil
		// **キーで返す。**完成した英文を返すと、日本語でプレイしていても
		// 英語の一文だけが出て、なぜ弾かれたのか読めない (#5526)。
		return NewDomainErrorCode(ErrInvalidPlay, "paigow.foulHighMustBeat", nil)
	}

	pg.appendLog(0, "set", fmt.Sprintf("low=[%d,%d]", lowIdx0, lowIdx1), pg.playerLowHand)

	// ディーラーのハンドをハウスウェイで設定
	pg.dealerHighHand, pg.dealerLowHand = paiGowHouseWay(pg.dealerCards)

	pg.resolve()
	return nil
}

// PaiGowHint はセットハンドフェーズでの推奨分割。
type PaiGowHint struct {
	// LowIdx0 / LowIdx1 はローハンドに回す2枚の手札インデックス。
	LowIdx0 int
	LowIdx1 int
	// LowIsPair はローハンドがペアになるかどうか (推奨理由の出し分け用)。
	LowIsPair bool
	// Reason はヒント理由キー。
	Reason string
}

// GetHint はセットハンドフェーズでの推奨分割を返す。それ以外では nil。
//
// **ディーラーのハウスウェイと同じ経路を通す** (paiGowHouseWayIndices)。
// ハイハンドがローハンドを上回る分割しか候補にしないので、
// この通りに置けば必ず反則 (foul) を回避できる。
func (pg *PaiGow) GetHint() *PaiGowHint {
	if pg.phase != PaiGowPhaseSetHands {
		return nil
	}
	i, j, ok := paiGowHouseWayIndices(pg.playerCards)
	if !ok {
		return nil
	}
	lowIsPair := evalPaiGowLowHand([]*Card{pg.playerCards[i], pg.playerCards[j]}) == PaiGowLowHandPair
	reason := "house_way_high"
	if lowIsPair {
		reason = "house_way_pair"
	}
	return &PaiGowHint{LowIdx0: i, LowIdx1: j, LowIsPair: lowIsPair, Reason: reason}
}

// AutoSetHands はハウスウェイの分割をそのまま適用する。
//
// 手作業の SetHands と同じ検証を通す (自前で分割を書き込まない) ので、
// **自動でも反則ハンドが成立する抜け道にはならない。**
func (pg *PaiGow) AutoSetHands() error {
	if pg.phase != PaiGowPhaseSetHands {
		return NewDomainError(ErrWrongPhase, "SetHands is only allowed during the set hands phase.")
	}
	i, j, ok := paiGowHouseWayIndices(pg.playerCards)
	if !ok {
		return NewDomainError(ErrInvalidPlay, "No legal split is available for this hand.")
	}
	return pg.SetHands(i, j)
}

// deal 7枚ずつ配る
func (pg *PaiGow) deal() {
	pg.playerCards = make([]*Card, 0, PaiGowHandSize)
	pg.dealerCards = make([]*Card, 0, PaiGowHandSize)
	for range PaiGowHandSize {
		pg.playerCards = append(pg.playerCards, pg.trumpCards.DrawCard())
		pg.dealerCards = append(pg.dealerCards, pg.trumpCards.DrawCard())
	}
	pg.appendLog(-1, "deal", "dealt 7 cards each", nil)
}

// resolve ゲーム解決
func (pg *PaiGow) resolve() {
	pg.playerHighRank = evalPaiGowHighHand(pg.playerHighHand)
	pg.playerLowRank = evalPaiGowLowHand(pg.playerLowHand)
	pg.dealerHighRank = evalPaiGowHighHand(pg.dealerHighHand)
	pg.dealerLowRank = evalPaiGowLowHand(pg.dealerLowHand)

	// ハイハンド比較
	highCmp := comparePaiGowHighHands(pg.playerHighHand, pg.dealerHighHand)
	if highCmp > 0 {
		pg.highHandResult = GameResultWin
	} else if highCmp < 0 {
		pg.highHandResult = GameResultLose
	} else {
		// タイはディーラーの勝ち
		pg.highHandResult = GameResultLose
	}

	// ローハンド比較
	lowCmp := comparePaiGowLowHands(pg.playerLowHand, pg.dealerLowHand)
	if lowCmp > 0 {
		pg.lowHandResult = GameResultWin
	} else if lowCmp < 0 {
		pg.lowHandResult = GameResultLose
	} else {
		// タイはディーラーの勝ち
		pg.lowHandResult = GameResultLose
	}

	// 総合結果判定
	if pg.highHandResult == GameResultWin && pg.lowHandResult == GameResultWin {
		// 両方勝ち: 1:1配当（勝ち分に5%コミッション）
		pg.result = GameResultWin
		pg.commission = pg.bet * PaiGowCommissionRate / 100
		pg.payout = pg.bet*2 - pg.commission
		pg.chips.AddChips(pg.payout)
	} else if pg.highHandResult == GameResultLose && pg.lowHandResult == GameResultLose {
		// 両方負け: ベット没収
		pg.result = GameResultLose
		pg.payout = 0
	} else {
		// 1勝1敗: プッシュ（ベット返却）
		pg.result = GameResultDraw
		pg.payout = pg.bet
		pg.chips.AddChips(pg.payout)
	}

	pg.gameEndFlag = true
	pg.phase = PaiGowPhaseEnd

	var resultStr string
	switch pg.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	pg.appendLog(-1, "result", resultStr, nil)
}

// --- ハンド評価関数 ---

// evalPaiGowHighHand パイガオのハイハンド（5枚）を評価する（ジョーカー対応）
func evalPaiGowHighHand(cards []*Card) int {
	if len(cards) != PaiGowHighHandSize {
		return PokerHandHighCard
	}
	// ジョーカーがあればバグジョーカーとして処理
	jokerIdx := -1
	for i, c := range cards {
		if IsJoker(c) {
			jokerIdx = i
			break
		}
	}
	if jokerIdx < 0 {
		return evalFiveCardHand(cards)
	}
	return evalPaiGowBugJokerHigh(cards, jokerIdx)
}

// evalPaiGowBugJokerHigh バグジョーカー（エース or ストレート/フラッシュ補完）の5枚評価
func evalPaiGowBugJokerHigh(cards []*Card, jokerIdx int) int {
	original := cards[jokerIdx]
	bestRank := PokerHandHighCard
	var bestSub *Card

	// 1. エースとして評価
	aceSub := NewCard(CardDesignSpade, 1, false)
	cards[jokerIdx] = aceSub
	aceRank := evalFiveCardHand(cards)
	if aceRank > bestRank {
		bestRank = aceRank
		bestSub = aceSub
	}

	// 2. ストレート/フラッシュを完成させるカードを試す
	// 他のカードのスートを取得
	nonJokerDesigns := make(map[int]int)
	for i, c := range cards {
		if i == jokerIdx {
			continue
		}
		nonJokerDesigns[c.GetDesign()]++
	}

	// フラッシュ候補スート: 4枚以上同じスートがあればジョーカーを含めてフラッシュ可能
	flushDesign := -1
	for d, cnt := range nonJokerDesigns {
		if cnt >= 4 {
			flushDesign = d
			break
		}
	}

	// ストレート/フラッシュ候補値を試す
	for v := 1; v <= CardValueMax; v++ {
		design := CardDesignSpade
		if flushDesign > 0 {
			design = flushDesign
		}
		sub := NewCard(design, v, false)
		cards[jokerIdx] = sub
		rank := evalFiveCardHand(cards)
		if rank >= PokerHandStraight && rank > bestRank {
			bestRank = rank
			bestSub = sub
		}
	}

	if bestSub != nil {
		cards[jokerIdx] = bestSub
	} else {
		cards[jokerIdx] = aceSub
	}
	// 最終的なランクを返す（bestSubが設定された状態で）
	finalRank := evalFiveCardHand(cards)
	cards[jokerIdx] = original
	if finalRank > bestRank {
		return finalRank
	}
	return bestRank
}

// evalPaiGowLowHand パイガオのローハンド（2枚）を評価する
func evalPaiGowLowHand(cards []*Card) int {
	if len(cards) != PaiGowLowHandSize {
		return PaiGowLowHandHighCard
	}
	v0 := paiGowCardValue(cards[0])
	v1 := paiGowCardValue(cards[1])
	if v0 == v1 {
		return PaiGowLowHandPair
	}
	return PaiGowLowHandHighCard
}

// paiGowCardValue パイガオのカード値を返す（ジョーカー=エース=14, A=14）
func paiGowCardValue(c *Card) int {
	if IsJoker(c) {
		return 14 // ジョーカーはエースとして扱う
	}
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// comparePaiGowHighHands ハイハンド（5枚）を比較する
// Returns 1 if a wins, -1 if b wins, 0 if tie.
func comparePaiGowHighHands(a, b []*Card) int {
	rankA := evalPaiGowHighHand(a)
	rankB := evalPaiGowHighHand(b)
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}
	// 同ランク: カード値で比較
	return compareHighCardsSlice(paiGowResolveJokerCards(a), paiGowResolveJokerCards(b))
}

// comparePaiGowLowHands ローハンド（2枚）を比較する
// Returns 1 if a wins, -1 if b wins, 0 if tie.
func comparePaiGowLowHands(a, b []*Card) int {
	rankA := evalPaiGowLowHand(a)
	rankB := evalPaiGowLowHand(b)
	if rankA > rankB {
		return 1
	}
	if rankA < rankB {
		return -1
	}
	// 同ランク: カード値で比較（高い方から）
	aVals := paiGowLowHandValues(a)
	bVals := paiGowLowHandValues(b)
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

// paiGowLowHandValues 2枚の値を降順で返す
func paiGowLowHandValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		vals[i] = paiGowCardValue(c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	return vals
}

// paiGowResolveJokerCards ジョーカーをエースに置換したカードスライスを返す（比較用）
func paiGowResolveJokerCards(cards []*Card) []*Card {
	result := make([]*Card, len(cards))
	for i, c := range cards {
		if IsJoker(c) {
			result[i] = NewCard(CardDesignSpade, 1, false)
		} else {
			result[i] = c
		}
	}
	return result
}

// paiGowHighBeatsLow ハイハンドがローハンドより強いか検証
func paiGowHighBeatsLow(highRank int, highHand []*Card, lowRank int, lowHand []*Card) bool {
	// ハイハンドがペア以上なら必ずローハンド（2枚）より強い
	if highRank >= PokerHandOnePair {
		return true
	}
	// ハイハンドがハイカードの場合: ハイハンドの最高値 >= ローハンドの最高値
	if lowRank >= PaiGowLowHandPair {
		// ローハンドがペアでハイハンドがハイカード: 無効
		return false
	}
	// 両方ハイカード: ハイハンドの最高カード >= ローハンドの最高カード
	highVals := make([]int, len(highHand))
	for i, c := range highHand {
		highVals[i] = paiGowCardValue(c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(highVals)))
	lowVals := paiGowLowHandValues(lowHand)
	return highVals[0] >= lowVals[0]
}

// --- ハウスウェイ ---

// paiGowHouseWay ディーラーのハウスウェイ（自動カード分割）
func paiGowHouseWay(cards []*Card) (high []*Card, low []*Card) {
	i, j, ok := paiGowHouseWayIndices(cards)
	if !ok {
		return cards, nil
	}
	low = []*Card{cards[i], cards[j]}
	high = make([]*Card, 0, PaiGowHighHandSize)
	for k, c := range cards {
		if k != i && k != j {
			high = append(high, c)
		}
	}
	return high, low
}

// paiGowHouseWayIndices はハウスウェイが選ぶローハンド2枚の**インデックス**を返す。
//
// ディーラーの分割 (paiGowHouseWay) と人間への推奨 (GetHint / AutoSetHands) は
// ここを共有する。**別実装にすると、ディーラー自身がやらない分割を人間に
// 勧めることになる。**
//
// ok=false は7枚でないとき、および合法な分割が1つも無いとき。
func paiGowHouseWayIndices(cards []*Card) (lowIdx0, lowIdx1 int, ok bool) {
	if len(cards) != PaiGowHandSize {
		return 0, 0, false
	}

	// 全組み合わせから最適な分割を探す（2枚選択 = C(7,2) = 21通り）
	bestI, bestJ := 0, 0
	found := false
	bestHighRank := -1
	bestLowRank := -1
	bestHighVals := []int{}
	bestLowVals := []int{}

	for i := range len(cards) {
		for j := i + 1; j < len(cards); j++ {
			lowCandidate := []*Card{cards[i], cards[j]}
			highCandidate := make([]*Card, 0, PaiGowHighHandSize)
			for k, c := range cards {
				if k != i && k != j {
					highCandidate = append(highCandidate, c)
				}
			}

			highRank := evalPaiGowHighHand(highCandidate)
			lowRank := evalPaiGowLowHand(lowCandidate)

			// ハイハンドがローハンドより弱い分割は無効
			if !paiGowHighBeatsLow(highRank, highCandidate, lowRank, lowCandidate) {
				continue
			}

			highVals := paiGowHighHandSortedValues(highCandidate)
			lowVals := paiGowLowHandValues(lowCandidate)

			// ハウスウェイ: ローハンドを最大化しつつハイハンドを維持
			if paiGowBetterSplit(highRank, highVals, lowRank, lowVals,
				bestHighRank, bestHighVals, bestLowRank, bestLowVals) {
				bestI, bestJ = i, j
				found = true
				bestHighRank = highRank
				bestLowRank = lowRank
				bestHighVals = highVals
				bestLowVals = lowVals
			}
		}
	}

	return bestI, bestJ, found
}

// paiGowBetterSplit 新しい分割が現在のベストより良いか判定（ハウスウェイ基準）
func paiGowBetterSplit(
	highRank int, highVals []int, lowRank int, lowVals []int,
	bestHighRank int, bestHighVals []int, bestLowRank int, bestLowVals []int,
) bool {
	if bestHighRank < 0 {
		return true
	}

	// 1. ローハンドのランクが高い方を優先
	if lowRank > bestLowRank {
		return true
	}
	if lowRank < bestLowRank {
		return false
	}

	// 2. ローハンドのランクが同じ場合: ローハンドの値が高い方を優先
	for i := 0; i < len(lowVals) && i < len(bestLowVals); i++ {
		if lowVals[i] > bestLowVals[i] {
			return true
		}
		if lowVals[i] < bestLowVals[i] {
			return false
		}
	}

	// 3. ローハンドが同じ場合: ハイハンドのランクが高い方を優先
	if highRank > bestHighRank {
		return true
	}
	if highRank < bestHighRank {
		return false
	}

	// 4. ハイハンドのランクも同じ場合: ハイハンドの値が高い方を優先
	for i := 0; i < len(highVals) && i < len(bestHighVals); i++ {
		if highVals[i] > bestHighVals[i] {
			return true
		}
		if highVals[i] < bestHighVals[i] {
			return false
		}
	}

	return false
}

// paiGowHighHandSortedValues ハイハンドの値を比較用にソート
func paiGowHighHandSortedValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		vals[i] = paiGowCardValue(c)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	return vals
}

// --- Getters ---

// GetPlayerCards プレイヤーの7枚を取得
func (pg *PaiGow) GetPlayerCards() []*Card { return pg.playerCards }

// GetDealerCards ディーラーの7枚を取得
func (pg *PaiGow) GetDealerCards() []*Card { return pg.dealerCards }

// GetPlayerHighHand プレイヤーハイハンドを取得
func (pg *PaiGow) GetPlayerHighHand() []*Card { return pg.playerHighHand }

// GetPlayerLowHand プレイヤーローハンドを取得
func (pg *PaiGow) GetPlayerLowHand() []*Card { return pg.playerLowHand }

// GetDealerHighHand ディーラーハイハンドを取得
func (pg *PaiGow) GetDealerHighHand() []*Card { return pg.dealerHighHand }

// GetDealerLowHand ディーラーローハンドを取得
func (pg *PaiGow) GetDealerLowHand() []*Card { return pg.dealerLowHand }

// GetPhase 現在のフェーズ
func (pg *PaiGow) GetPhase() int { return pg.phase }

// GetGameEndFlag ゲーム終了フラグ
func (pg *PaiGow) GetGameEndFlag() bool { return pg.gameEndFlag }

// GetBet ベット額
func (pg *PaiGow) GetBet() int { return pg.bet }

// GetResult ゲーム結果
func (pg *PaiGow) GetResult() GameResult { return pg.result }

// GetHighHandResult ハイハンド結果
func (pg *PaiGow) GetHighHandResult() GameResult { return pg.highHandResult }

// GetLowHandResult ローハンド結果
func (pg *PaiGow) GetLowHandResult() GameResult { return pg.lowHandResult }

// GetPayout 配当
func (pg *PaiGow) GetPayout() int { return pg.payout }

// GetCommission コミッション
func (pg *PaiGow) GetCommission() int { return pg.commission }

// GetPlayerHighRank プレイヤーハイハンドランク
func (pg *PaiGow) GetPlayerHighRank() int { return pg.playerHighRank }

// GetPlayerLowRank プレイヤーローハンドランク
func (pg *PaiGow) GetPlayerLowRank() int { return pg.playerLowRank }

// GetDealerHighRank ディーラーハイハンドランク
func (pg *PaiGow) GetDealerHighRank() int { return pg.dealerHighRank }

// GetDealerLowRank ディーラーローハンドランク
func (pg *PaiGow) GetDealerLowRank() int { return pg.dealerLowRank }

// GetChips チップ
func (pg *PaiGow) GetChips() int { return pg.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (pg *PaiGow) SetPhase(phase int) { pg.phase = phase }

// SetPlayerCards プレイヤーカード設定（テスト用）
func (pg *PaiGow) SetPlayerCards(cards []*Card) { pg.playerCards = cards }

// SetDealerCards ディーラーカード設定（テスト用）
func (pg *PaiGow) SetDealerCards(cards []*Card) { pg.dealerCards = cards }

// SetPlayerHighHand プレイヤーハイハンド設定（テスト用）
func (pg *PaiGow) SetPlayerHighHand(cards []*Card) { pg.playerHighHand = cards }

// SetPlayerLowHand プレイヤーローハンド設定（テスト用）
func (pg *PaiGow) SetPlayerLowHand(cards []*Card) { pg.playerLowHand = cards }

// SetDealerHighHand ディーラーハイハンド設定（テスト用）
func (pg *PaiGow) SetDealerHighHand(cards []*Card) { pg.dealerHighHand = cards }

// SetDealerLowHand ディーラーローハンド設定（テスト用）
func (pg *PaiGow) SetDealerLowHand(cards []*Card) { pg.dealerLowHand = cards }

// SetBet ベット額設定（テスト用）
func (pg *PaiGow) SetBet(amount int) { pg.bet = amount }

// SetChips チップ設定（テスト用）
func (pg *PaiGow) SetChips(chips int) { pg.chips.SetChips(chips) }

// SetResult ゲーム結果設定（テスト用）
func (pg *PaiGow) SetResult(result GameResult) { pg.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (pg *PaiGow) SetGameEndFlag(flag bool) { pg.gameEndFlag = flag }

// SetHighHandResult ハイハンド結果設定（テスト用）
func (pg *PaiGow) SetHighHandResult(result GameResult) { pg.highHandResult = result }

// SetLowHandResult ローハンド結果設定（テスト用）
func (pg *PaiGow) SetLowHandResult(result GameResult) { pg.lowHandResult = result }

// SetPayout 配当設定（テスト用）
func (pg *PaiGow) SetPayout(payout int) { pg.payout = payout }

// SetCommission コミッション設定（テスト用）
func (pg *PaiGow) SetCommission(commission int) { pg.commission = commission }

// SetPlayerHighRank プレイヤーハイハンドランク設定（テスト用）
func (pg *PaiGow) SetPlayerHighRank(rank int) { pg.playerHighRank = rank }

// SetPlayerLowRank プレイヤーローハンドランク設定（テスト用）
func (pg *PaiGow) SetPlayerLowRank(rank int) { pg.playerLowRank = rank }

// SetDealerHighRank ディーラーハイハンドランク設定（テスト用）
func (pg *PaiGow) SetDealerHighRank(rank int) { pg.dealerHighRank = rank }

// SetDealerLowRank ディーラーローハンドランク設定（テスト用）
func (pg *PaiGow) SetDealerLowRank(rank int) { pg.dealerLowRank = rank }

// paiGowJSON is the JSON wire format for PaiGow.
type paiGowJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	PlayerCards    []*Card           `json:"pc"`
	DealerCards    []*Card           `json:"dc"`
	PlayerHighHand []*Card           `json:"phh"`
	PlayerLowHand  []*Card           `json:"plh"`
	DealerHighHand []*Card           `json:"dhh"`
	DealerLowHand  []*Card           `json:"dlh"`
	Chips          *ChipHolder       `json:"ch"`
	Bet            int               `json:"bt"`
	Phase          int               `json:"ps"`
	GameEndFlag    bool              `json:"ge"`
	Result         GameResult        `json:"rs"`
	HighHandResult GameResult        `json:"hhr"`
	LowHandResult  GameResult        `json:"lhr"`
	Payout         int               `json:"po"`
	Commission     int               `json:"cm"`
	PlayerHighRank int               `json:"phr"`
	PlayerLowRank  int               `json:"plr"`
	DealerHighRank int               `json:"dhr"`
	DealerLowRank  int               `json:"dlr"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (pg *PaiGow) MarshalJSON() ([]byte, error) {
	return json.Marshal(paiGowJSON{
		TrumpCards:     pg.trumpCards,
		PlayerCards:    pg.playerCards,
		DealerCards:    pg.dealerCards,
		PlayerHighHand: pg.playerHighHand,
		PlayerLowHand:  pg.playerLowHand,
		DealerHighHand: pg.dealerHighHand,
		DealerLowHand:  pg.dealerLowHand,
		Chips:          &pg.chips,
		Bet:            pg.bet,
		Phase:          pg.phase,
		GameEndFlag:    pg.gameEndFlag,
		Result:         pg.result,
		HighHandResult: pg.highHandResult,
		LowHandResult:  pg.lowHandResult,
		Payout:         pg.payout,
		Commission:     pg.commission,
		PlayerHighRank: pg.playerHighRank,
		PlayerLowRank:  pg.playerLowRank,
		DealerHighRank: pg.dealerHighRank,
		DealerLowRank:  pg.dealerLowRank,
		ActionLog:      pg.actionLog,
	})
}

// paiGowMaxSliceLen caps slice sizes during deserialisation.
const paiGowMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (pg *PaiGow) UnmarshalJSON(data []byte) error {
	var j paiGowJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerCards) > paiGowMaxSliceLen || len(j.DealerCards) > paiGowMaxSliceLen ||
		len(j.PlayerHighHand) > paiGowMaxSliceLen || len(j.PlayerLowHand) > paiGowMaxSliceLen ||
		len(j.DealerHighHand) > paiGowMaxSliceLen || len(j.DealerLowHand) > paiGowMaxSliceLen ||
		len(j.ActionLog) > paiGowMaxSliceLen {
		return fmt.Errorf("paigow: input array exceeds maximum allowed size")
	}

	pg.trumpCards = j.TrumpCards
	if pg.trumpCards == nil {
		pg.trumpCards = NewTrumpCards(PaiGowJokerCount)
	}
	pg.playerCards = j.PlayerCards
	if pg.playerCards == nil {
		pg.playerCards = make([]*Card, 0)
	}
	pg.dealerCards = j.DealerCards
	if pg.dealerCards == nil {
		pg.dealerCards = make([]*Card, 0)
	}
	pg.playerHighHand = j.PlayerHighHand
	if pg.playerHighHand == nil {
		pg.playerHighHand = make([]*Card, 0)
	}
	pg.playerLowHand = j.PlayerLowHand
	if pg.playerLowHand == nil {
		pg.playerLowHand = make([]*Card, 0)
	}
	pg.dealerHighHand = j.DealerHighHand
	if pg.dealerHighHand == nil {
		pg.dealerHighHand = make([]*Card, 0)
	}
	pg.dealerLowHand = j.DealerLowHand
	if pg.dealerLowHand == nil {
		pg.dealerLowHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		pg.chips = *j.Chips
	}
	pg.bet = j.Bet
	pg.phase = j.Phase
	pg.gameEndFlag = j.GameEndFlag
	pg.result = j.Result
	pg.highHandResult = j.HighHandResult
	pg.lowHandResult = j.LowHandResult
	pg.payout = j.Payout
	pg.commission = j.Commission
	pg.playerHighRank = j.PlayerHighRank
	pg.playerLowRank = j.PlayerLowRank
	pg.dealerHighRank = j.DealerHighRank
	pg.dealerLowRank = j.DealerLowRank
	pg.actionLog = j.ActionLog
	if pg.actionLog == nil {
		pg.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
