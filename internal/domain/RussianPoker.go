//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// ロシアンポーカーフェーズ定数
const (
	RussianPokerPhaseBet          = 1 // ベットフェーズ
	RussianPokerPhaseAction       = 2 // アクションフェーズ（Fold/Play/Exchange/Buy6th選択）
	RussianPokerPhaseSelect       = 3 // 6枚目購入後の選択フェーズ（5枚を選ぶ）
	RussianPokerPhasePostAction   = 4 // 交換/購入後のアクションフェーズ（Fold/Play選択）
	RussianPokerPhaseForceQualify = 5 // 強制クオリファイフェーズ
	RussianPokerPhaseEnd          = 6 // 終了フェーズ
)

// ロシアンポーカーデフォルト値
const (
	RussianPokerDefaultChips = 1000
	RussianPokerMinBet       = 10
	RussianPokerMaxBet       = 10000
	RussianPokerHandSize     = 5
	RussianPokerMaxExchange  = 5
)

// プレイベット配当倍率
const (
	RussianPokerPayRoyalFlush    = 100
	RussianPokerPayStraightFlush = 50
	RussianPokerPayFourOfAKind   = 20
	RussianPokerPayFullHouse     = 7
	RussianPokerPayFlush         = 5
	RussianPokerPayStraight      = 4
	RussianPokerPayThreeOfAKind  = 3
	RussianPokerPayTwoPair       = 2
	RussianPokerPayPair          = 1
)

// RussianPoker ロシアンポーカークラス
type RussianPoker struct {
	trumpCards       *TrumpCards
	playerHand       []*Card
	dealerHand       []*Card
	chips            ChipHolder
	anteBet          int
	exchangeCount    int
	exchangeFee      int
	bought6th        bool
	buy6thFee        int
	forceExchanged   bool
	forceExchangeFee int
	playBet          int
	phase            int
	gameEndFlag      bool
	result           GameResult
	antePayout       int
	playPayout       int
	dealerQualified  bool
	playerHandRank   int
	dealerHandRank   int
	actionLogBase
}

// NewRussianPoker コンストラクタ
func NewRussianPoker(trumpCards *TrumpCards) *RussianPoker {
	trumpCards.Shuffle()
	return &RussianPoker{
		trumpCards: trumpCards,
		phase:      RussianPokerPhaseBet,
	}
}

// NewDefaultRussianPoker デフォルト設定のロシアンポーカーを生成するファクトリ関数
func NewDefaultRussianPoker() *RussianPoker {
	rp := NewRussianPoker(NewTrumpCards(0))
	rp.chips.SetChips(RussianPokerDefaultChips)
	return rp
}

// Reset ゲーム初期化
func (rp *RussianPoker) Reset() {
	rp.gameEndFlag = false
	rp.phase = RussianPokerPhaseBet
	rp.playerHand = nil
	rp.dealerHand = nil
	rp.anteBet = 0
	rp.exchangeCount = 0
	rp.exchangeFee = 0
	rp.bought6th = false
	rp.buy6thFee = 0
	rp.forceExchanged = false
	rp.forceExchangeFee = 0
	rp.playBet = 0
	rp.result = 0
	rp.antePayout = 0
	rp.playPayout = 0
	rp.dealerQualified = false
	rp.playerHandRank = 0
	rp.dealerHandRank = 0
	rp.actionLog = nil
	if rp.chips.GetChips() < RussianPokerMinBet {
		rp.chips.SetChips(RussianPokerDefaultChips)
	}
	rp.trumpCards = NewTrumpCards(0)
	for range 10 {
		rp.trumpCards.Shuffle()
	}
}

// Bet アンテベット＆カード配布
func (rp *RussianPoker) Bet(ante int) error {
	if rp.phase != RussianPokerPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < RussianPokerMinBet || ante%RussianPokerMinBet != 0 || ante > RussianPokerMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if !rp.chips.SubtractChips(ante) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	rp.anteBet = ante
	rp.appendLog(0, "bet", fmt.Sprintf("ante=%d", ante), nil)
	rp.deal()
	rp.phase = RussianPokerPhaseAction
	return nil
}

// Exchange 指定インデックス（0..4）のカードを交換する。
// 交換枚数 1 枚あたり ante と同額の手数料を徴収する。
func (rp *RussianPoker) Exchange(indices []int) error {
	if rp.phase != RussianPokerPhaseAction {
		return NewDomainError(ErrWrongPhase, "Exchange is only allowed during the action phase.")
	}
	if len(indices) == 0 {
		return NewDomainError(ErrInvalidAmount, "Must exchange at least one card.")
	}
	if len(indices) > RussianPokerMaxExchange {
		return NewDomainError(ErrInvalidAmount, "Too many cards to exchange.")
	}
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= RussianPokerHandSize {
			return NewDomainError(ErrInvalidAmount, "Exchange index out of range.")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidAmount, "Duplicate exchange index.")
		}
		seen[idx] = true
	}

	fee := rp.anteBet * len(indices)
	if !rp.chips.SubtractChips(fee) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for exchange fee.")
	}
	for _, idx := range indices {
		newCard := rp.trumpCards.DrawCard()
		if newCard != nil {
			rp.playerHand[idx] = newCard
		}
	}
	rp.exchangeCount = len(indices)
	rp.exchangeFee = fee
	rp.appendLog(0, "exchange", fmt.Sprintf("exchange %d card(s) fee=%d", rp.exchangeCount, rp.exchangeFee), nil)
	rp.phase = RussianPokerPhasePostAction
	return nil
}

// Buy6th 6枚目のカードを購入する（手数料: アンテ同額）
func (rp *RussianPoker) Buy6th() error {
	if rp.phase != RussianPokerPhaseAction {
		return NewDomainError(ErrWrongPhase, "Buy6th is only allowed during the action phase.")
	}
	fee := rp.anteBet
	if !rp.chips.SubtractChips(fee) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips to buy 6th card.")
	}
	rp.bought6th = true
	rp.buy6thFee = fee
	newCard := rp.trumpCards.DrawCard()
	if newCard != nil {
		rp.playerHand = append(rp.playerHand, newCard)
	}
	rp.appendLog(0, "buy6th", fmt.Sprintf("bought 6th card fee=%d", fee), nil)
	rp.phase = RussianPokerPhaseSelect
	return nil
}

// Select 6枚の手札から1枚を捨てて5枚にする（discardIndex は 0..5）
func (rp *RussianPoker) Select(discardIndex int) error {
	if rp.phase != RussianPokerPhaseSelect {
		return NewDomainError(ErrWrongPhase, "Select is only allowed during the select phase.")
	}
	if discardIndex < 0 || discardIndex >= len(rp.playerHand) {
		return NewDomainError(ErrInvalidAmount, "Discard index out of range.")
	}
	copy(rp.playerHand[discardIndex:], rp.playerHand[discardIndex+1:])
	rp.playerHand[len(rp.playerHand)-1] = nil
	rp.playerHand = rp.playerHand[:len(rp.playerHand)-1]
	rp.appendLog(0, "select", fmt.Sprintf("discard index=%d", discardIndex), nil)
	rp.phase = RussianPokerPhasePostAction
	return nil
}

// Play コール（アンテの2倍のプレイベットを置いて勝負）
func (rp *RussianPoker) Play() error {
	if rp.phase != RussianPokerPhaseAction && rp.phase != RussianPokerPhasePostAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action or post-action phase.")
	}
	playBet := rp.anteBet * 2
	if !rp.chips.SubtractChips(playBet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	rp.playBet = playBet
	rp.appendLog(0, "play", fmt.Sprintf("play bet=%d", rp.playBet), nil)

	rp.resolve()
	return nil
}

// Fold フォールド（アンテと手数料は没収）
func (rp *RussianPoker) Fold() error {
	if rp.phase != RussianPokerPhaseAction && rp.phase != RussianPokerPhasePostAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action or post-action phase.")
	}
	rp.appendLog(0, "fold", "player folds", nil)
	rp.result = GameResultLose
	rp.playerHandRank = evalFiveCardHand(rp.playerHand)
	rp.dealerHandRank = evalFiveCardHand(rp.dealerHand)
	rp.gameEndFlag = true
	rp.phase = RussianPokerPhaseEnd
	rp.appendLog(-1, "result", "player folded", nil)
	return nil
}

// ForceExchange ディーラーの最高カードを交換させる（手数料: アンテ同額）
func (rp *RussianPoker) ForceExchange() error {
	if rp.phase != RussianPokerPhaseForceQualify {
		return NewDomainError(ErrWrongPhase, "ForceExchange is only allowed during the force qualify phase.")
	}
	fee := rp.anteBet
	if !rp.chips.SubtractChips(fee) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for force exchange.")
	}
	rp.forceExchanged = true
	rp.forceExchangeFee = fee
	rp.appendLog(0, "forceExchange", fmt.Sprintf("force exchange dealer's highest card fee=%d", fee), nil)

	highIdx := rp.findDealerHighestCardIndex()
	newCard := rp.trumpCards.DrawCard()
	if newCard != nil {
		rp.dealerHand[highIdx] = newCard
	}

	rp.resolveAfterForce()
	return nil
}

// Decline 強制クオリファイを辞退する
func (rp *RussianPoker) Decline() error {
	if rp.phase != RussianPokerPhaseForceQualify {
		return NewDomainError(ErrWrongPhase, "Decline is only allowed during the force qualify phase.")
	}
	rp.appendLog(0, "decline", "declined force exchange", nil)
	rp.result = GameResultWin
	rp.antePayout = rp.anteBet * 2
	rp.playPayout = rp.playBet
	totalPayout := rp.antePayout + rp.playPayout
	if totalPayout > 0 {
		rp.chips.AddChips(totalPayout)
	}
	rp.gameEndFlag = true
	rp.phase = RussianPokerPhaseEnd
	rp.appendLog(-1, "result", "dealer does not qualify (declined)", nil)
	return nil
}

// deal 5枚ずつ配る
func (rp *RussianPoker) deal() {
	rp.playerHand = make([]*Card, 0, RussianPokerHandSize+1)
	rp.dealerHand = make([]*Card, 0, RussianPokerHandSize)
	for range RussianPokerHandSize {
		rp.playerHand = append(rp.playerHand, rp.trumpCards.DrawCard())
		rp.dealerHand = append(rp.dealerHand, rp.trumpCards.DrawCard())
	}
	rp.appendLog(-1, "deal", "dealt 5 cards each", nil)
}

// resolve ゲーム解決（Play後の処理）
func (rp *RussianPoker) resolve() {
	rp.playerHandRank = evalFiveCardHand(rp.playerHand)
	rp.dealerHandRank = evalFiveCardHand(rp.dealerHand)
	rp.dealerQualified = rp.checkDealerQualifies()

	if !rp.dealerQualified {
		rp.phase = RussianPokerPhaseForceQualify
		rp.appendLog(-1, "qualify", "dealer does not qualify", nil)
		return
	}

	rp.finalResolve()
}

// resolveAfterForce 強制クオリファイ後の解決処理
func (rp *RussianPoker) resolveAfterForce() {
	rp.dealerHandRank = evalFiveCardHand(rp.dealerHand)
	rp.dealerQualified = rp.checkDealerQualifies()

	if !rp.dealerQualified {
		rp.result = GameResultWin
		rp.antePayout = rp.anteBet * 2
		rp.playPayout = rp.playBet
		totalPayout := rp.antePayout + rp.playPayout
		if totalPayout > 0 {
			rp.chips.AddChips(totalPayout)
		}
		rp.gameEndFlag = true
		rp.phase = RussianPokerPhaseEnd
		rp.appendLog(-1, "result", "dealer still does not qualify after force exchange", nil)
		return
	}

	rp.finalResolve()
}

// finalResolve ディーラークオリファイ済みの最終解決
func (rp *RussianPoker) finalResolve() {
	cmp := rp.compareHands()
	switch {
	case cmp > 0:
		rp.result = GameResultWin
	case cmp < 0:
		rp.result = GameResultLose
	default:
		rp.result = GameResultDraw
	}

	rp.calculatePayouts()

	totalPayout := rp.antePayout + rp.playPayout
	if totalPayout > 0 {
		rp.chips.AddChips(totalPayout)
	}

	rp.gameEndFlag = true
	rp.phase = RussianPokerPhaseEnd

	var resultStr string
	switch rp.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	rp.appendLog(-1, "result", resultStr, nil)
}

// compareHands プレイヤーとディーラーのハンドを比較する
func (rp *RussianPoker) compareHands() int {
	if rp.playerHandRank > rp.dealerHandRank {
		return 1
	}
	if rp.playerHandRank < rp.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(rp.playerHand, rp.dealerHand)
}

// checkDealerQualifies ディーラークオリファイ条件: ペア以上、または A-K ハイ
func (rp *RussianPoker) checkDealerQualifies() bool {
	if rp.dealerHandRank >= PokerHandOnePair {
		return true
	}
	hasAce := false
	hasKing := false
	for _, c := range rp.dealerHand {
		switch c.GetValue() {
		case 1:
			hasAce = true
		case 13:
			hasKing = true
		}
	}
	return hasAce && hasKing
}

// calculatePayouts アンテ／プレイの配当計算
func (rp *RussianPoker) calculatePayouts() {
	switch rp.result {
	case GameResultWin:
		rp.antePayout = rp.anteBet * 2
		multiplier := rp.playMultiplier()
		rp.playPayout = rp.playBet + rp.playBet*multiplier
	case GameResultDraw:
		rp.antePayout = rp.anteBet
		rp.playPayout = rp.playBet
	case GameResultLose:
		rp.antePayout = 0
		rp.playPayout = 0
	}
}

// playMultiplier プレイベット配当倍率（プレイヤーハンドランクに基づく）
func (rp *RussianPoker) playMultiplier() int {
	switch rp.playerHandRank {
	case PokerHandRoyalFlush:
		return RussianPokerPayRoyalFlush
	case PokerHandStraightFlush:
		return RussianPokerPayStraightFlush
	case PokerHandFourOfAKind:
		return RussianPokerPayFourOfAKind
	case PokerHandFullHouse:
		return RussianPokerPayFullHouse
	case PokerHandFlush:
		return RussianPokerPayFlush
	case PokerHandStraight:
		return RussianPokerPayStraight
	case PokerHandThreeOfAKind:
		return RussianPokerPayThreeOfAKind
	case PokerHandTwoPair:
		return RussianPokerPayTwoPair
	default:
		return RussianPokerPayPair
	}
}

// findDealerHighestCardIndex ディーラーの最高カードのインデックスを返す
func (rp *RussianPoker) findDealerHighestCardIndex() int {
	highIdx := 0
	highVal := cardSortValue(rp.dealerHand[0])
	for i := 1; i < len(rp.dealerHand); i++ {
		v := cardSortValue(rp.dealerHand[i])
		if v > highVal {
			highVal = v
			highIdx = i
		}
	}
	return highIdx
}

// cardSortValue はカードのソート値を返す（A=14 として扱う）
func cardSortValue(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (rp *RussianPoker) GetPlayerHand() []*Card { return rp.playerHand }

// GetDealerHand ディーラーハンド取得
func (rp *RussianPoker) GetDealerHand() []*Card { return rp.dealerHand }

// GetPhase 現在のフェーズ
func (rp *RussianPoker) GetPhase() int { return rp.phase }

// GetGameEndFlag ゲーム終了フラグ
func (rp *RussianPoker) GetGameEndFlag() bool { return rp.gameEndFlag }

// GetAnteBet アンテベット額
func (rp *RussianPoker) GetAnteBet() int { return rp.anteBet }

// GetPlayBet プレイベット額
func (rp *RussianPoker) GetPlayBet() int { return rp.playBet }

// GetExchangeCount 交換した枚数
func (rp *RussianPoker) GetExchangeCount() int { return rp.exchangeCount }

// GetExchangeFee 徴収した交換手数料
func (rp *RussianPoker) GetExchangeFee() int { return rp.exchangeFee }

// GetBought6th 6枚目を購入したかどうか
func (rp *RussianPoker) GetBought6th() bool { return rp.bought6th }

// GetBuy6thFee 6枚目購入手数料
func (rp *RussianPoker) GetBuy6thFee() int { return rp.buy6thFee }

// GetForceExchanged 強制クオリファイを実行したかどうか
func (rp *RussianPoker) GetForceExchanged() bool { return rp.forceExchanged }

// GetForceExchangeFee 強制クオリファイ手数料
func (rp *RussianPoker) GetForceExchangeFee() int { return rp.forceExchangeFee }

// GetResult ゲーム結果
func (rp *RussianPoker) GetResult() GameResult { return rp.result }

// GetAntePayout アンテ配当
func (rp *RussianPoker) GetAntePayout() int { return rp.antePayout }

// GetPlayPayout プレイ配当
func (rp *RussianPoker) GetPlayPayout() int { return rp.playPayout }

// GetTotalPayout 合計配当
func (rp *RussianPoker) GetTotalPayout() int {
	return rp.antePayout + rp.playPayout
}

// GetDealerQualified ディーラークオリファイ
func (rp *RussianPoker) GetDealerQualified() bool { return rp.dealerQualified }

// GetPlayerHandRank プレイヤーハンドランク
func (rp *RussianPoker) GetPlayerHandRank() int { return rp.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (rp *RussianPoker) GetDealerHandRank() int { return rp.dealerHandRank }

// GetChips チップ
func (rp *RussianPoker) GetChips() int { return rp.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (rp *RussianPoker) SetPhase(phase int) { rp.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (rp *RussianPoker) SetPlayerHand(cards []*Card) { rp.playerHand = cards }

// SetDealerHand ディーラーハンド設定（テスト用）
func (rp *RussianPoker) SetDealerHand(cards []*Card) { rp.dealerHand = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (rp *RussianPoker) SetAnteBet(amount int) { rp.anteBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (rp *RussianPoker) SetPlayBet(amount int) { rp.playBet = amount }

// SetChips チップ設定（テスト用）
func (rp *RussianPoker) SetChips(chips int) { rp.chips.SetChips(chips) }

// russianPokerJSON は RussianPoker の JSON ワイヤーフォーマット
type russianPokerJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	PlayerHand       []*Card           `json:"ph"`
	DealerHand       []*Card           `json:"dh"`
	Chips            *ChipHolder       `json:"ch"`
	AnteBet          int               `json:"ab"`
	ExchangeCount    int               `json:"ec"`
	ExchangeFee      int               `json:"ef"`
	Bought6th        bool              `json:"b6"`
	Buy6thFee        int               `json:"bf"`
	ForceExchanged   bool              `json:"fe"`
	ForceExchangeFee int               `json:"ff"`
	PlayBet          int               `json:"pb"`
	Phase            int               `json:"ps"`
	GameEndFlag      bool              `json:"ge"`
	Result           GameResult        `json:"rs"`
	AntePayout       int               `json:"ap"`
	PlayPayout       int               `json:"plp"`
	DealerQualified  bool              `json:"dq"`
	PlayerHandRank   int               `json:"pr"`
	DealerHandRank   int               `json:"dr"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (rp *RussianPoker) MarshalJSON() ([]byte, error) {
	return json.Marshal(russianPokerJSON{
		TrumpCards:       rp.trumpCards,
		PlayerHand:       rp.playerHand,
		DealerHand:       rp.dealerHand,
		Chips:            &rp.chips,
		AnteBet:          rp.anteBet,
		ExchangeCount:    rp.exchangeCount,
		ExchangeFee:      rp.exchangeFee,
		Bought6th:        rp.bought6th,
		Buy6thFee:        rp.buy6thFee,
		ForceExchanged:   rp.forceExchanged,
		ForceExchangeFee: rp.forceExchangeFee,
		PlayBet:          rp.playBet,
		Phase:            rp.phase,
		GameEndFlag:      rp.gameEndFlag,
		Result:           rp.result,
		AntePayout:       rp.antePayout,
		PlayPayout:       rp.playPayout,
		DealerQualified:  rp.dealerQualified,
		PlayerHandRank:   rp.playerHandRank,
		DealerHandRank:   rp.dealerHandRank,
		ActionLog:        rp.actionLog,
	})
}

// russianPokerMaxHandLen caps hand sizes during deserialisation (max 6 = 5 + buy6th).
const russianPokerMaxHandLen = 10

// russianPokerMaxLogLen caps action log size during deserialisation.
const russianPokerMaxLogLen = 200

// UnmarshalJSON implements json.Unmarshaler.
func (rp *RussianPoker) UnmarshalJSON(data []byte) error {
	var j russianPokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > russianPokerMaxHandLen || len(j.DealerHand) > russianPokerMaxHandLen ||
		len(j.ActionLog) > russianPokerMaxLogLen {
		return fmt.Errorf("russianpoker: input array exceeds maximum allowed size")
	}

	rp.trumpCards = j.TrumpCards
	if rp.trumpCards == nil {
		rp.trumpCards = NewTrumpCards(0)
	}
	rp.playerHand = j.PlayerHand
	if rp.playerHand == nil {
		rp.playerHand = make([]*Card, 0)
	}
	rp.dealerHand = j.DealerHand
	if rp.dealerHand == nil {
		rp.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		rp.chips = *j.Chips
	}
	rp.anteBet = j.AnteBet
	rp.exchangeCount = j.ExchangeCount
	rp.exchangeFee = j.ExchangeFee
	rp.bought6th = j.Bought6th
	rp.buy6thFee = j.Buy6thFee
	rp.forceExchanged = j.ForceExchanged
	rp.forceExchangeFee = j.ForceExchangeFee
	rp.playBet = j.PlayBet
	rp.phase = j.Phase
	rp.gameEndFlag = j.GameEndFlag
	rp.result = j.Result
	rp.antePayout = j.AntePayout
	rp.playPayout = j.PlayPayout
	rp.dealerQualified = j.DealerQualified
	rp.playerHandRank = j.PlayerHandRank
	rp.dealerHandRank = j.DealerHandRank
	rp.actionLog = j.ActionLog
	if rp.actionLog == nil {
		rp.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
