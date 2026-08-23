//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// オアシスポーカーフェーズ定数
const (
	OasisPokerPhaseBet      = 1 // ベットフェーズ
	OasisPokerPhaseExchange = 2 // カード交換フェーズ
	OasisPokerPhaseAction   = 3 // アクションフェーズ（Call/Fold選択）
	OasisPokerPhaseEnd      = 4 // 終了フェーズ
)

// オアシスポーカーデフォルト値
const (
	OasisPokerDefaultChips = 1000  // デフォルトチップ
	OasisPokerMinBet       = 10    // 最低ベット額
	OasisPokerMaxBet       = 10000 // 最大ベット額
	OasisPokerHandSize     = 5     // ハンドサイズ
	OasisPokerMaxExchange  = 5     // 1ハンドあたりの最大交換枚数
)

// プレイベット配当倍率（コール時）
// 配当は Caribbean Stud と同一の倍率を採用する。
const (
	OasisPokerPayRoyalFlush    = 100
	OasisPokerPayStraightFlush = 50
	OasisPokerPayFourOfAKind   = 20
	OasisPokerPayFullHouse     = 7
	OasisPokerPayFlush         = 5
	OasisPokerPayStraight      = 4
	OasisPokerPayThreeOfAKind  = 3
	OasisPokerPayTwoPair       = 2
	OasisPokerPayPair          = 1
)

// プログレッシブジャックポット（サイドベット）配当倍率
const (
	OasisPokerJackpotRoyalFlush    = 20000
	OasisPokerJackpotStraightFlush = 5000
	OasisPokerJackpotFourOfAKind   = 500
	OasisPokerJackpotFullHouse     = 100
	OasisPokerJackpotFlush         = 50
)

// OasisPoker オアシスポーカークラス
type OasisPoker struct {
	trumpCards      *TrumpCards
	playerHand      []*Card
	dealerHand      []*Card
	chips           ChipHolder
	anteBet         int
	jackpotBet      int
	exchangeCount   int // 交換した枚数（0..5）。手数料計算のSSoT
	exchangeFee     int // 徴収した交換手数料（ante * exchangeCount）
	playBet         int
	phase           int
	gameEndFlag     bool
	result          GameResult
	antePayout      int
	playPayout      int
	jackpotPayout   int
	dealerQualified bool
	playerHandRank  int
	dealerHandRank  int
	actionLogBase
}

// NewOasisPoker コンストラクタ
func NewOasisPoker(trumpCards *TrumpCards) *OasisPoker {
	trumpCards.Shuffle()
	return &OasisPoker{
		trumpCards: trumpCards,
		phase:      OasisPokerPhaseBet,
	}
}

// NewDefaultOasisPoker デフォルト設定のオアシスポーカーを生成するファクトリ関数
func NewDefaultOasisPoker() *OasisPoker {
	op := NewOasisPoker(NewTrumpCards(0))
	op.chips.SetChips(OasisPokerDefaultChips)
	return op
}

// Reset ゲーム初期化
func (op *OasisPoker) Reset() {
	op.gameEndFlag = false
	op.phase = OasisPokerPhaseBet
	op.playerHand = nil
	op.dealerHand = nil
	op.anteBet = 0
	op.jackpotBet = 0
	op.exchangeCount = 0
	op.exchangeFee = 0
	op.playBet = 0
	op.result = 0
	op.antePayout = 0
	op.playPayout = 0
	op.jackpotPayout = 0
	op.dealerQualified = false
	op.playerHandRank = 0
	op.dealerHandRank = 0
	op.actionLog = nil
	if op.chips.GetChips() < OasisPokerMinBet {
		op.chips.SetChips(OasisPokerDefaultChips)
	}
	op.trumpCards = NewTrumpCards(0)
	for range 10 {
		op.trumpCards.Shuffle()
	}
}

// Bet アンテベット＆カード配布。jackpot に正の値を渡すとジャックポットサイドベットを追加する。
func (op *OasisPoker) Bet(ante, jackpot int) error {
	if op.phase != OasisPokerPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < OasisPokerMinBet || ante%OasisPokerMinBet != 0 || ante > OasisPokerMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if jackpot < 0 {
		return NewDomainError(ErrInvalidAmount, "Jackpot bet must not be negative.")
	}
	if jackpot > 0 && (jackpot < OasisPokerMinBet || jackpot%OasisPokerMinBet != 0 || jackpot > OasisPokerMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid jackpot bet amount.")
	}
	totalCost := ante + jackpot
	if !op.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	op.anteBet = ante
	op.jackpotBet = jackpot
	op.appendLog(0, "bet", fmt.Sprintf("ante=%d jackpot=%d", ante, jackpot), nil)

	op.deal()
	op.phase = OasisPokerPhaseExchange
	return nil
}

// Exchange 指定インデックス（0..4）のカードを交換する。
// 交換枚数 1 枚あたり ante と同額の手数料を徴収する。indices が空の場合は手数料 0 で
// 交換フェーズをスキップする（Stand と等価）。
func (op *OasisPoker) Exchange(indices []int) error {
	if op.phase != OasisPokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Exchange is only allowed during the exchange phase.")
	}
	if len(indices) > OasisPokerMaxExchange {
		return NewDomainError(ErrInvalidAmount, "Too many cards to exchange.")
	}
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= OasisPokerHandSize {
			return NewDomainError(ErrInvalidAmount, "Exchange index out of range.")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidAmount, "Duplicate exchange index.")
		}
		seen[idx] = true
	}

	fee := op.anteBet * len(indices)
	if fee > 0 && !op.chips.SubtractChips(fee) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for exchange fee.")
	}
	for _, idx := range indices {
		newCard := op.trumpCards.DrawCard()
		if newCard != nil {
			op.playerHand[idx] = newCard
		}
	}
	op.exchangeCount = len(indices)
	op.exchangeFee = fee
	op.appendLog(0, "exchange", fmt.Sprintf("exchange %d card(s) fee=%d", op.exchangeCount, op.exchangeFee), nil)
	op.phase = OasisPokerPhaseAction
	return nil
}

// Stand 交換せずアクションフェーズへ進む（Exchange(nil) と等価）。
func (op *OasisPoker) Stand() error {
	return op.Exchange(nil)
}

// Play コール（アンテの2倍のプレイベットを置いて勝負）
func (op *OasisPoker) Play() error {
	if op.phase != OasisPokerPhaseAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action phase.")
	}
	playBet := op.anteBet * 2
	if !op.chips.SubtractChips(playBet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	op.playBet = playBet
	op.appendLog(0, "play", fmt.Sprintf("play bet=%d", op.playBet), nil)

	op.resolve()
	return nil
}

// Fold フォールド（アンテと交換手数料は没収。ジャックポットは別途評価）
func (op *OasisPoker) Fold() error {
	if op.phase != OasisPokerPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	op.appendLog(0, "fold", "player folds", nil)

	op.result = GameResultLose
	op.playerHandRank = evalFiveCardHand(op.playerHand)
	op.dealerHandRank = evalFiveCardHand(op.dealerHand)

	op.evaluateJackpot()
	if op.jackpotPayout > 0 {
		op.chips.AddChips(op.jackpotPayout)
	}

	op.gameEndFlag = true
	op.phase = OasisPokerPhaseEnd
	op.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal 5枚ずつ配る
func (op *OasisPoker) deal() {
	op.playerHand = make([]*Card, 0, OasisPokerHandSize)
	op.dealerHand = make([]*Card, 0, OasisPokerHandSize)
	for range OasisPokerHandSize {
		op.playerHand = append(op.playerHand, op.trumpCards.DrawCard())
		op.dealerHand = append(op.dealerHand, op.trumpCards.DrawCard())
	}
	op.appendLog(-1, "deal", "dealt 5 cards each", nil)
}

// resolve ゲーム解決（Play後の処理）
func (op *OasisPoker) resolve() {
	op.playerHandRank = evalFiveCardHand(op.playerHand)
	op.dealerHandRank = evalFiveCardHand(op.dealerHand)
	op.dealerQualified = op.checkDealerQualifies()

	// **勝敗は精算に合わせる。** ディーラーが qualify しなかったときの精算には
	// 特例（アンテ 1:1、もう一方のベットは返却）があり、手の強弱に関わらず
	// プレイヤーはアンテぶん増える。手を比べただけで result を決めると、
	// チップが増えているのに画面が「負け」と言う (#6213)。
	cmp := op.compareHands()
	switch {
	case !op.dealerQualified:
		op.result = GameResultWin
	case cmp > 0:
		op.result = GameResultWin
	case cmp < 0:
		op.result = GameResultLose
	default:
		op.result = GameResultDraw
	}

	op.calculatePayouts()
	op.evaluateJackpot()

	totalPayout := op.antePayout + op.playPayout + op.jackpotPayout
	if totalPayout > 0 {
		op.chips.AddChips(totalPayout)
	}

	op.gameEndFlag = true
	op.phase = OasisPokerPhaseEnd

	var resultStr string
	switch op.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	op.appendLog(-1, "result", resultStr, nil)
}

// compareHands プレイヤーとディーラーのハンドを比較する
func (op *OasisPoker) compareHands() int {
	if op.playerHandRank > op.dealerHandRank {
		return 1
	}
	if op.playerHandRank < op.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(op.playerHand, op.dealerHand)
}

// checkDealerQualifies ディーラークオリファイ条件: ペア以上、または A-K ハイ
func (op *OasisPoker) checkDealerQualifies() bool {
	return dealerQualifies(op.dealerHandRank, op.dealerHand)
}

// calculatePayouts アンテ／プレイの配当計算
func (op *OasisPoker) calculatePayouts() {
	if !op.dealerQualified {
		// ディーラー未クオリファイ: アンテ1:1、プレイベットはプッシュ（返却のみ）
		op.antePayout = op.anteBet * 2
		op.playPayout = op.playBet
		return
	}
	switch op.result {
	case GameResultWin:
		op.antePayout = op.anteBet * 2
		multiplier := op.playMultiplier()
		op.playPayout = op.playBet + op.playBet*multiplier
	case GameResultDraw:
		op.antePayout = op.anteBet
		op.playPayout = op.playBet
	case GameResultLose:
		op.antePayout = 0
		op.playPayout = 0
	}
}

// playMultiplier プレイベット配当倍率（プレイヤーハンドランクに基づく）
func (op *OasisPoker) playMultiplier() int {
	switch op.playerHandRank {
	case PokerHandRoyalFlush:
		return OasisPokerPayRoyalFlush
	case PokerHandStraightFlush:
		return OasisPokerPayStraightFlush
	case PokerHandFourOfAKind:
		return OasisPokerPayFourOfAKind
	case PokerHandFullHouse:
		return OasisPokerPayFullHouse
	case PokerHandFlush:
		return OasisPokerPayFlush
	case PokerHandStraight:
		return OasisPokerPayStraight
	case PokerHandThreeOfAKind:
		return OasisPokerPayThreeOfAKind
	case PokerHandTwoPair:
		return OasisPokerPayTwoPair
	default:
		return OasisPokerPayPair
	}
}

// evaluateJackpot ジャックポットサイドベット評価（独立）
func (op *OasisPoker) evaluateJackpot() {
	if op.jackpotBet <= 0 {
		return
	}
	switch op.playerHandRank {
	case PokerHandRoyalFlush:
		op.jackpotPayout = op.jackpotBet * OasisPokerJackpotRoyalFlush
	case PokerHandStraightFlush:
		op.jackpotPayout = op.jackpotBet * OasisPokerJackpotStraightFlush
	case PokerHandFourOfAKind:
		op.jackpotPayout = op.jackpotBet * OasisPokerJackpotFourOfAKind
	case PokerHandFullHouse:
		op.jackpotPayout = op.jackpotBet * OasisPokerJackpotFullHouse
	case PokerHandFlush:
		op.jackpotPayout = op.jackpotBet * OasisPokerJackpotFlush
	}
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (op *OasisPoker) GetPlayerHand() []*Card { return op.playerHand }

// RecommendPlay はアクションフェーズでプレイ (続行) を推奨するかを返す。基本戦略として
// ワンペア以上、または A/K を含むときプレイ、それ以外はフォールドを推奨する。
func (op *OasisPoker) RecommendPlay() bool {
	if len(op.playerHand) < 5 {
		return true
	}
	if evalFiveCardHand(op.playerHand) >= PokerHandOnePair {
		return true
	}
	for _, c := range op.playerHand {
		if v := c.GetValue(); v == 1 || v == 13 { // Ace or King
			return true
		}
	}
	return false
}

// GetDealerHand ディーラーハンド取得
func (op *OasisPoker) GetDealerHand() []*Card { return op.dealerHand }

// GetPhase 現在のフェーズ
func (op *OasisPoker) GetPhase() int { return op.phase }

// GetGameEndFlag ゲーム終了フラグ
func (op *OasisPoker) GetGameEndFlag() bool { return op.gameEndFlag }

// GetAnteBet アンテベット額
func (op *OasisPoker) GetAnteBet() int { return op.anteBet }

// GetJackpotBet ジャックポットベット額
func (op *OasisPoker) GetJackpotBet() int { return op.jackpotBet }

// GetPlayBet プレイベット額
func (op *OasisPoker) GetPlayBet() int { return op.playBet }

// GetExchangeCount 交換した枚数
func (op *OasisPoker) GetExchangeCount() int { return op.exchangeCount }

// GetExchangeFee 徴収した交換手数料
func (op *OasisPoker) GetExchangeFee() int { return op.exchangeFee }

// GetResult ゲーム結果
func (op *OasisPoker) GetResult() GameResult { return op.result }

// GetAntePayout アンテ配当
func (op *OasisPoker) GetAntePayout() int { return op.antePayout }

// GetPlayPayout プレイ配当
func (op *OasisPoker) GetPlayPayout() int { return op.playPayout }

// GetJackpotPayout ジャックポット配当
func (op *OasisPoker) GetJackpotPayout() int { return op.jackpotPayout }

// GetTotalPayout 合計配当
func (op *OasisPoker) GetTotalPayout() int {
	return op.antePayout + op.playPayout + op.jackpotPayout
}

// GetDealerQualified ディーラークオリファイ
func (op *OasisPoker) GetDealerQualified() bool { return op.dealerQualified }

// GetPlayerHandRank プレイヤーハンドランク
func (op *OasisPoker) GetPlayerHandRank() int { return op.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (op *OasisPoker) GetDealerHandRank() int { return op.dealerHandRank }

// GetChips チップ
func (op *OasisPoker) GetChips() int { return op.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (op *OasisPoker) SetPhase(phase int) { op.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (op *OasisPoker) SetPlayerHand(cards []*Card) { op.playerHand = cards }

// SetDealerHand ディーラーハンド設定（テスト用）
func (op *OasisPoker) SetDealerHand(cards []*Card) { op.dealerHand = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (op *OasisPoker) SetAnteBet(amount int) { op.anteBet = amount }

// SetJackpotBet ジャックポットベット額設定（テスト用）
func (op *OasisPoker) SetJackpotBet(amount int) { op.jackpotBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (op *OasisPoker) SetPlayBet(amount int) { op.playBet = amount }

// SetChips チップ設定（テスト用）
func (op *OasisPoker) SetChips(chips int) { op.chips.SetChips(chips) }

// oasisPokerJSON は OasisPoker の JSON ワイヤーフォーマット
type oasisPokerJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	JackpotBet      int               `json:"jb"`
	ExchangeCount   int               `json:"ec"`
	ExchangeFee     int               `json:"ef"`
	PlayBet         int               `json:"pb"`
	Phase           int               `json:"ps"`
	GameEndFlag     bool              `json:"ge"`
	Result          GameResult        `json:"rs"`
	AntePayout      int               `json:"ap"`
	PlayPayout      int               `json:"plp"`
	JackpotPayout   int               `json:"jp"`
	DealerQualified bool              `json:"dq"`
	PlayerHandRank  int               `json:"pr"`
	DealerHandRank  int               `json:"dr"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (op *OasisPoker) MarshalJSON() ([]byte, error) {
	return json.Marshal(oasisPokerJSON{
		TrumpCards:      op.trumpCards,
		PlayerHand:      op.playerHand,
		DealerHand:      op.dealerHand,
		Chips:           &op.chips,
		AnteBet:         op.anteBet,
		JackpotBet:      op.jackpotBet,
		ExchangeCount:   op.exchangeCount,
		ExchangeFee:     op.exchangeFee,
		PlayBet:         op.playBet,
		Phase:           op.phase,
		GameEndFlag:     op.gameEndFlag,
		Result:          op.result,
		AntePayout:      op.antePayout,
		PlayPayout:      op.playPayout,
		JackpotPayout:   op.jackpotPayout,
		DealerQualified: op.dealerQualified,
		PlayerHandRank:  op.playerHandRank,
		DealerHandRank:  op.dealerHandRank,
		ActionLog:       op.actionLog,
	})
}

// oasisPokerMaxSliceLen caps slice sizes during deserialisation.
const oasisPokerMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (op *OasisPoker) UnmarshalJSON(data []byte) error {
	var j oasisPokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > oasisPokerMaxSliceLen || len(j.DealerHand) > oasisPokerMaxSliceLen ||
		len(j.ActionLog) > oasisPokerMaxSliceLen {
		return fmt.Errorf("oasispoker: input array exceeds maximum allowed size")
	}

	op.trumpCards = j.TrumpCards
	if op.trumpCards == nil {
		op.trumpCards = NewTrumpCards(0)
	}
	op.playerHand = j.PlayerHand
	if op.playerHand == nil {
		op.playerHand = make([]*Card, 0)
	}
	op.dealerHand = j.DealerHand
	if op.dealerHand == nil {
		op.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		op.chips = *j.Chips
	}
	op.anteBet = j.AnteBet
	op.jackpotBet = j.JackpotBet
	op.exchangeCount = j.ExchangeCount
	op.exchangeFee = j.ExchangeFee
	op.playBet = j.PlayBet
	op.phase = j.Phase
	op.gameEndFlag = j.GameEndFlag
	op.result = j.Result
	op.antePayout = j.AntePayout
	op.playPayout = j.PlayPayout
	op.jackpotPayout = j.JackpotPayout
	op.dealerQualified = j.DealerQualified
	op.playerHandRank = j.PlayerHandRank
	op.dealerHandRank = j.DealerHandRank
	op.actionLog = j.ActionLog
	if op.actionLog == nil {
		op.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
