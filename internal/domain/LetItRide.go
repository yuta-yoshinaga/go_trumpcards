//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// レット・イット・ライドフェーズ定数
const (
	LetItRidePhaseBet            = 1 // ベットフェーズ
	LetItRidePhaseFirstDecision  = 2 // 第1判断フェーズ（コミュニティカード1枚目公開後）
	LetItRidePhaseSecondDecision = 3 // 第2判断フェーズ（コミュニティカード2枚目公開後）
	LetItRidePhaseEnd            = 4 // 終了フェーズ
)

// レット・イット・ライドデフォルト値
const (
	LetItRideDefaultChips   = 1000  // デフォルトチップ
	LetItRideMinBet         = 10    // 最低ベット額
	LetItRideMaxBet         = 10000 // 最大ベット額
	LetItRidePlayerHandSize = 3     // プレイヤーハンドサイズ
	LetItRideCommunitySize  = 2     // コミュニティカードサイズ
)

// レット・イット・ライド配当倍率
const (
	LetItRidePayRoyalFlush    = 1000 // ロイヤルフラッシュ 1000:1
	LetItRidePayStraightFlush = 200  // ストレートフラッシュ 200:1
	LetItRidePayFourOfAKind   = 50   // フォーカード 50:1
	LetItRidePayFullHouse     = 11   // フルハウス 11:1
	LetItRidePayFlush         = 8    // フラッシュ 8:1
	LetItRidePayStraight      = 5    // ストレート 5:1
	LetItRidePayThreeOfAKind  = 3    // スリーカード 3:1
	LetItRidePayTwoPair       = 2    // ツーペア 2:1
	LetItRidePayTensOrBetter  = 1    // 10以上のワンペア 1:1
)

// LetItRide レット・イット・ライドクラス
type LetItRide struct {
	trumpCards     *TrumpCards // トランプカード
	playerHand     []*Card     // プレイヤーハンド（3枚）
	communityCards []*Card     // コミュニティカード（2枚）
	chips          ChipHolder  // チップ
	betAmount      int         // 1口あたりのベット額
	bet1Active     bool        // ベット1（常にアクティブ）
	bet2Active     bool        // ベット2（第2判断で取り下げ可能）
	bet3Active     bool        // ベット3（第1判断で取り下げ可能）
	phase          int         // 現在のフェーズ
	gameEndFlag    bool        // ゲーム終了フラグ
	result         GameResult  // ゲーム結果（Win/Lose）
	handRank       int         // 最終ハンドランク
	bet1Payout     int         // ベット1配当
	bet2Payout     int         // ベット2配当
	bet3Payout     int         // ベット3配当
	totalPayout    int         // 合計配当
	actionLogBase
}

// NewLetItRide コンストラクタ
func NewLetItRide(trumpCards *TrumpCards) *LetItRide {
	trumpCards.Shuffle()
	return &LetItRide{
		trumpCards: trumpCards,
		phase:      LetItRidePhaseBet,
	}
}

// NewDefaultLetItRide デフォルト設定のレット・イット・ライドを生成するファクトリ関数
func NewDefaultLetItRide() *LetItRide {
	lir := NewLetItRide(NewTrumpCards(0))
	lir.chips.SetChips(LetItRideDefaultChips)
	return lir
}

// Reset ゲーム初期化
func (lir *LetItRide) Reset() {
	lir.gameEndFlag = false
	lir.phase = LetItRidePhaseBet
	lir.playerHand = nil
	lir.communityCards = nil
	lir.betAmount = 0
	lir.bet1Active = false
	lir.bet2Active = false
	lir.bet3Active = false
	lir.result = 0
	lir.handRank = 0
	lir.bet1Payout = 0
	lir.bet2Payout = 0
	lir.bet3Payout = 0
	lir.totalPayout = 0
	lir.actionLog = nil
	if lir.chips.GetChips() < LetItRideMinBet*3 {
		lir.chips.SetChips(LetItRideDefaultChips)
	}
	lir.trumpCards = NewTrumpCards(0)
	for range 10 {
		lir.trumpCards.Shuffle()
	}
}

// Bet ベット＆カード配布。amount は1口分のベット額（3口分差し引く）。
func (lir *LetItRide) Bet(amount int) error {
	if lir.phase != LetItRidePhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < LetItRideMinBet || amount%LetItRideMinBet != 0 || amount > LetItRideMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	totalCost := amount * 3
	if !lir.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	lir.betAmount = amount
	lir.bet1Active = true
	lir.bet2Active = true
	lir.bet3Active = true
	lir.appendLog(0, "bet", fmt.Sprintf("bet=%d (x3=%d)", amount, totalCost), nil)

	lir.deal()
	lir.phase = LetItRidePhaseFirstDecision
	return nil
}

// Pull ベットを取り下げる（第1判断ではベット3、第2判断ではベット2）
func (lir *LetItRide) Pull() error {
	switch lir.phase {
	case LetItRidePhaseFirstDecision:
		lir.bet3Active = false
		lir.chips.AddChips(lir.betAmount)
		lir.appendLog(0, "pull", "pull bet 3", nil)
		lir.phase = LetItRidePhaseSecondDecision
		return nil
	case LetItRidePhaseSecondDecision:
		lir.bet2Active = false
		lir.chips.AddChips(lir.betAmount)
		lir.appendLog(0, "pull", "pull bet 2", nil)
		lir.resolve()
		return nil
	default:
		return NewDomainError(ErrWrongPhase, "Pull is only allowed during decision phases.")
	}
}

// LetItRidePullPreview は Pull を実行したときの掛け金の動き。
type LetItRidePullPreview struct {
	// Returned は手元に戻る額 (1口分)。
	Returned int
	// RiskBefore / RiskAfter は場に残る総額の前後。
	RiskBefore int
	RiskAfter  int
}

// GetPullPreview は Pull を実行したときに戻る額とリスクの増減を返す。
// 判断フェーズ以外では nil。
//
// **Pull はリスクを「下げる」操作。**1口ぶん取り下げて手元に戻すので、
// 負ける額も勝てる額も減る。取り消せないのはそこで、危険だからではない。
func (lir *LetItRide) GetPullPreview() *LetItRidePullPreview {
	if lir.phase != LetItRidePhaseFirstDecision && lir.phase != LetItRidePhaseSecondDecision {
		return nil
	}
	active := 0
	for _, on := range []bool{lir.bet1Active, lir.bet2Active, lir.bet3Active} {
		if on {
			active++
		}
	}
	before := lir.betAmount * active
	after := before - lir.betAmount
	if after < 0 {
		after = 0
	}
	return &LetItRidePullPreview{Returned: lir.betAmount, RiskBefore: before, RiskAfter: after}
}

// LetItRideAction ベットをそのままにする（Let It Ride）。
// Named "Action" to avoid a name collision with the LetItRide type itself;
// the use-case layer exposes this as LetItRide() in its public interface.
func (lir *LetItRide) LetItRideAction() error {
	switch lir.phase {
	case LetItRidePhaseFirstDecision:
		lir.appendLog(0, "letitride", "let it ride (bet 3)", nil)
		lir.phase = LetItRidePhaseSecondDecision
		return nil
	case LetItRidePhaseSecondDecision:
		lir.appendLog(0, "letitride", "let it ride (bet 2)", nil)
		lir.resolve()
		return nil
	default:
		return NewDomainError(ErrWrongPhase, "Let It Ride is only allowed during decision phases.")
	}
}

// deal プレイヤーに3枚、コミュニティに2枚配る
func (lir *LetItRide) deal() {
	lir.playerHand = make([]*Card, 0, LetItRidePlayerHandSize)
	lir.communityCards = make([]*Card, 0, LetItRideCommunitySize)
	for range LetItRidePlayerHandSize {
		lir.playerHand = append(lir.playerHand, lir.trumpCards.DrawCard())
	}
	for range LetItRideCommunitySize {
		lir.communityCards = append(lir.communityCards, lir.trumpCards.DrawCard())
	}
	lir.appendLog(-1, "deal", "dealt 3 player cards + 2 community cards", nil)
}

// resolve ゲーム解決（最終ハンド評価＆配当計算）
func (lir *LetItRide) resolve() {
	// 5枚ハンドを構成
	fullHand := make([]*Card, 0, 5)
	fullHand = append(fullHand, lir.playerHand...)
	fullHand = append(fullHand, lir.communityCards...)
	lir.handRank = evalFiveCardHand(fullHand)

	multiplier := lir.payoutMultiplier()
	if multiplier > 0 {
		lir.result = GameResultWin
		if lir.bet1Active {
			lir.bet1Payout = lir.betAmount + lir.betAmount*multiplier
		}
		if lir.bet2Active {
			lir.bet2Payout = lir.betAmount + lir.betAmount*multiplier
		}
		if lir.bet3Active {
			lir.bet3Payout = lir.betAmount + lir.betAmount*multiplier
		}
	} else {
		lir.result = GameResultLose
	}

	lir.totalPayout = lir.bet1Payout + lir.bet2Payout + lir.bet3Payout
	if lir.totalPayout > 0 {
		lir.chips.AddChips(lir.totalPayout)
	}

	lir.gameEndFlag = true
	lir.phase = LetItRidePhaseEnd

	var resultStr string
	if lir.result == GameResultWin {
		resultStr = "player wins"
	} else {
		resultStr = "player loses"
	}
	lir.appendLog(-1, "result", resultStr, nil)
}

// payoutMultiplier ハンドランクに基づく配当倍率（0 = 配当なし）
func (lir *LetItRide) payoutMultiplier() int {
	switch lir.handRank {
	case PokerHandRoyalFlush:
		return LetItRidePayRoyalFlush
	case PokerHandStraightFlush:
		return LetItRidePayStraightFlush
	case PokerHandFourOfAKind:
		return LetItRidePayFourOfAKind
	case PokerHandFullHouse:
		return LetItRidePayFullHouse
	case PokerHandFlush:
		return LetItRidePayFlush
	case PokerHandStraight:
		return LetItRidePayStraight
	case PokerHandThreeOfAKind:
		return LetItRidePayThreeOfAKind
	case PokerHandTwoPair:
		return LetItRidePayTwoPair
	case PokerHandOnePair:
		return lir.checkTensOrBetter()
	default:
		return 0
	}
}

// checkTensOrBetter ワンペアが10以上かチェック
func (lir *LetItRide) checkTensOrBetter() int {
	fullHand := make([]*Card, 0, 5)
	fullHand = append(fullHand, lir.playerHand...)
	fullHand = append(fullHand, lir.communityCards...)

	counts := make(map[int]int)
	for _, c := range fullHand {
		counts[c.GetValue()]++
	}
	for val, cnt := range counts {
		if cnt >= 2 {
			// 10, J(11), Q(12), K(13), A(1) がペア
			if val >= 10 || val == 1 {
				return LetItRidePayTensOrBetter
			}
		}
	}
	return 0
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (lir *LetItRide) GetPlayerHand() []*Card { return lir.playerHand }

// GetCommunityCards コミュニティカード取得
func (lir *LetItRide) GetCommunityCards() []*Card { return lir.communityCards }

// GetPhase 現在のフェーズ
func (lir *LetItRide) GetPhase() int { return lir.phase }

// GetGameEndFlag ゲーム終了フラグ
func (lir *LetItRide) GetGameEndFlag() bool { return lir.gameEndFlag }

// GetBetAmount 1口あたりのベット額
func (lir *LetItRide) GetBetAmount() int { return lir.betAmount }

// GetBet1Active ベット1アクティブ状態
func (lir *LetItRide) GetBet1Active() bool { return lir.bet1Active }

// GetBet2Active ベット2アクティブ状態
func (lir *LetItRide) GetBet2Active() bool { return lir.bet2Active }

// GetBet3Active ベット3アクティブ状態
func (lir *LetItRide) GetBet3Active() bool { return lir.bet3Active }

// GetResult ゲーム結果
func (lir *LetItRide) GetResult() GameResult { return lir.result }

// GetHandRank ハンドランク
func (lir *LetItRide) GetHandRank() int { return lir.handRank }

// GetBet1Payout ベット1配当
func (lir *LetItRide) GetBet1Payout() int { return lir.bet1Payout }

// GetBet2Payout ベット2配当
func (lir *LetItRide) GetBet2Payout() int { return lir.bet2Payout }

// GetBet3Payout ベット3配当
func (lir *LetItRide) GetBet3Payout() int { return lir.bet3Payout }

// GetTotalPayout 合計配当
func (lir *LetItRide) GetTotalPayout() int { return lir.totalPayout }

// GetChips チップ
func (lir *LetItRide) GetChips() int { return lir.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (lir *LetItRide) SetPhase(phase int) { lir.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (lir *LetItRide) SetPlayerHand(cards []*Card) { lir.playerHand = cards }

// SetCommunityCards コミュニティカード設定（テスト用）
func (lir *LetItRide) SetCommunityCards(cards []*Card) { lir.communityCards = cards }

// SetBetAmount ベット額設定（テスト用）
func (lir *LetItRide) SetBetAmount(amount int) { lir.betAmount = amount }

// SetBet1Active ベット1アクティブ設定（テスト用）
func (lir *LetItRide) SetBet1Active(active bool) { lir.bet1Active = active }

// SetBet2Active ベット2アクティブ設定（テスト用）
func (lir *LetItRide) SetBet2Active(active bool) { lir.bet2Active = active }

// SetBet3Active ベット3アクティブ設定（テスト用）
func (lir *LetItRide) SetBet3Active(active bool) { lir.bet3Active = active }

// SetChips チップ設定（テスト用）
func (lir *LetItRide) SetChips(chips int) { lir.chips.SetChips(chips) }

// letItRideJSON は LetItRide の JSON ワイヤーフォーマット
type letItRideJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	PlayerHand     []*Card           `json:"ph"`
	CommunityCards []*Card           `json:"cc"`
	Chips          *ChipHolder       `json:"ch"`
	BetAmount      int               `json:"ba"`
	Bet1Active     bool              `json:"b1"`
	Bet2Active     bool              `json:"b2"`
	Bet3Active     bool              `json:"b3"`
	Phase          int               `json:"ps"`
	GameEndFlag    bool              `json:"ge"`
	Result         GameResult        `json:"rs"`
	HandRank       int               `json:"hr"`
	Bet1Payout     int               `json:"p1"`
	Bet2Payout     int               `json:"p2"`
	Bet3Payout     int               `json:"p3"`
	TotalPayout    int               `json:"tp"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (lir *LetItRide) MarshalJSON() ([]byte, error) {
	return json.Marshal(letItRideJSON{
		TrumpCards:     lir.trumpCards,
		PlayerHand:     lir.playerHand,
		CommunityCards: lir.communityCards,
		Chips:          &lir.chips,
		BetAmount:      lir.betAmount,
		Bet1Active:     lir.bet1Active,
		Bet2Active:     lir.bet2Active,
		Bet3Active:     lir.bet3Active,
		Phase:          lir.phase,
		GameEndFlag:    lir.gameEndFlag,
		Result:         lir.result,
		HandRank:       lir.handRank,
		Bet1Payout:     lir.bet1Payout,
		Bet2Payout:     lir.bet2Payout,
		Bet3Payout:     lir.bet3Payout,
		TotalPayout:    lir.totalPayout,
		ActionLog:      lir.actionLog,
	})
}

// letItRideMaxSliceLen caps slice sizes during deserialisation.
const letItRideMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (lir *LetItRide) UnmarshalJSON(data []byte) error {
	var j letItRideJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > letItRideMaxSliceLen || len(j.CommunityCards) > letItRideMaxSliceLen ||
		len(j.ActionLog) > letItRideMaxSliceLen {
		return fmt.Errorf("letitride: input array exceeds maximum allowed size")
	}

	lir.trumpCards = j.TrumpCards
	if lir.trumpCards == nil {
		lir.trumpCards = NewTrumpCards(0)
	}
	lir.playerHand = j.PlayerHand
	if lir.playerHand == nil {
		lir.playerHand = make([]*Card, 0)
	}
	lir.communityCards = j.CommunityCards
	if lir.communityCards == nil {
		lir.communityCards = make([]*Card, 0)
	}
	if j.Chips != nil {
		lir.chips = *j.Chips
	}
	lir.betAmount = j.BetAmount
	lir.bet1Active = j.Bet1Active
	lir.bet2Active = j.Bet2Active
	lir.bet3Active = j.Bet3Active
	lir.phase = j.Phase
	lir.gameEndFlag = j.GameEndFlag
	lir.result = j.Result
	lir.handRank = j.HandRank
	lir.bet1Payout = j.Bet1Payout
	lir.bet2Payout = j.Bet2Payout
	lir.bet3Payout = j.Bet3Payout
	lir.totalPayout = j.TotalPayout
	lir.actionLog = j.ActionLog
	if lir.actionLog == nil {
		lir.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
