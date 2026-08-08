//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// カリビアンスタッドポーカーフェーズ定数
const (
	CaribbeanStudPhaseBet    = 1 // ベットフェーズ
	CaribbeanStudPhaseAction = 2 // アクションフェーズ（Call/Fold選択）
	CaribbeanStudPhaseEnd    = 3 // 終了フェーズ
)

// カリビアンスタッドポーカーデフォルト値
const (
	CaribbeanStudDefaultChips = 1000  // デフォルトチップ
	CaribbeanStudMinBet       = 10    // 最低ベット額
	CaribbeanStudMaxBet       = 10000 // 最大ベット額
	CaribbeanStudHandSize     = 5     // ハンドサイズ
)

// プレイベット配当倍率（コール時）
const (
	CaribbeanStudPayRoyalFlush    = 100 // ロイヤルフラッシュ 100:1
	CaribbeanStudPayStraightFlush = 50  // ストレートフラッシュ 50:1
	CaribbeanStudPayFourOfAKind   = 20  // フォーカード 20:1
	CaribbeanStudPayFullHouse     = 7   // フルハウス 7:1
	CaribbeanStudPayFlush         = 5   // フラッシュ 5:1
	CaribbeanStudPayStraight      = 4   // ストレート 4:1
	CaribbeanStudPayThreeOfAKind  = 3   // スリーカード 3:1
	CaribbeanStudPayTwoPair       = 2   // ツーペア 2:1
	CaribbeanStudPayPair          = 1   // ワンペア以下 1:1
)

// プログレッシブジャックポット（サイドベット）配当倍率
const (
	CaribbeanStudJackpotRoyalFlush    = 20000 // ロイヤルフラッシュ
	CaribbeanStudJackpotStraightFlush = 5000  // ストレートフラッシュ
	CaribbeanStudJackpotFourOfAKind   = 500   // フォーカード
	CaribbeanStudJackpotFullHouse     = 100   // フルハウス
	CaribbeanStudJackpotFlush         = 50    // フラッシュ
)

// CaribbeanStud カリビアンスタッドポーカークラス
type CaribbeanStud struct {
	trumpCards      *TrumpCards // トランプカード
	playerHand      []*Card     // プレイヤーハンド
	dealerHand      []*Card     // ディーラーハンド
	chips           ChipHolder  // チップ
	anteBet         int         // アンテベット額
	jackpotBet      int         // ジャックポットサイドベット額
	playBet         int         // プレイ（コール）ベット額
	phase           int         // 現在のフェーズ
	gameEndFlag     bool        // ゲーム終了フラグ
	result          GameResult  // ゲーム結果
	antePayout      int         // アンテ配当
	playPayout      int         // プレイ配当
	jackpotPayout   int         // ジャックポット配当
	dealerQualified bool        // ディーラークオリファイフラグ
	playerHandRank  int         // プレイヤーハンドランク
	dealerHandRank  int         // ディーラーハンドランク
	actionLogBase
}

// NewCaribbeanStud コンストラクタ
func NewCaribbeanStud(trumpCards *TrumpCards) *CaribbeanStud {
	trumpCards.Shuffle()
	return &CaribbeanStud{
		trumpCards: trumpCards,
		phase:      CaribbeanStudPhaseBet,
	}
}

// NewDefaultCaribbeanStud デフォルト設定のカリビアンスタッドポーカーを生成するファクトリ関数
func NewDefaultCaribbeanStud() *CaribbeanStud {
	cs := NewCaribbeanStud(NewTrumpCards(0))
	cs.chips.SetChips(CaribbeanStudDefaultChips)
	return cs
}

// Reset ゲーム初期化
func (cs *CaribbeanStud) Reset() {
	cs.gameEndFlag = false
	cs.phase = CaribbeanStudPhaseBet
	cs.playerHand = nil
	cs.dealerHand = nil
	cs.anteBet = 0
	cs.jackpotBet = 0
	cs.playBet = 0
	cs.result = 0
	cs.antePayout = 0
	cs.playPayout = 0
	cs.jackpotPayout = 0
	cs.dealerQualified = false
	cs.playerHandRank = 0
	cs.dealerHandRank = 0
	cs.actionLog = nil
	if cs.chips.GetChips() < CaribbeanStudMinBet {
		cs.chips.SetChips(CaribbeanStudDefaultChips)
	}
	cs.trumpCards = NewTrumpCards(0)
	for range 10 {
		cs.trumpCards.Shuffle()
	}
}

// Bet アンテベット＆カード配布。jackpot に正の値を渡すとジャックポットサイドベットを追加する。
func (cs *CaribbeanStud) Bet(ante, jackpot int) error {
	if cs.phase != CaribbeanStudPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < CaribbeanStudMinBet || ante%CaribbeanStudMinBet != 0 || ante > CaribbeanStudMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if jackpot < 0 {
		return NewDomainError(ErrInvalidAmount, "Jackpot bet must not be negative.")
	}
	if jackpot > 0 && (jackpot < CaribbeanStudMinBet || jackpot%CaribbeanStudMinBet != 0 || jackpot > CaribbeanStudMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid jackpot bet amount.")
	}
	totalCost := ante + jackpot
	if !cs.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	cs.anteBet = ante
	cs.jackpotBet = jackpot
	cs.appendLog(0, "bet", fmt.Sprintf("ante=%d jackpot=%d", ante, jackpot), nil)

	cs.deal()
	cs.phase = CaribbeanStudPhaseAction
	return nil
}

// Play コール（アンテの2倍のプレイベットを置いて勝負）
func (cs *CaribbeanStud) Play() error {
	if cs.phase != CaribbeanStudPhaseAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action phase.")
	}
	playBet := cs.anteBet * 2
	if !cs.chips.SubtractChips(playBet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	cs.playBet = playBet
	cs.appendLog(0, "play", fmt.Sprintf("play bet=%d", cs.playBet), nil)

	cs.resolve()
	return nil
}

// Fold フォールド（アンテ没収。ジャックポットは別途評価）
func (cs *CaribbeanStud) Fold() error {
	if cs.phase != CaribbeanStudPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	cs.appendLog(0, "fold", "player folds", nil)

	cs.result = GameResultLose
	cs.playerHandRank = evalFiveCardHand(cs.playerHand)
	cs.dealerHandRank = evalFiveCardHand(cs.dealerHand)

	cs.evaluateJackpot()
	if cs.jackpotPayout > 0 {
		cs.chips.AddChips(cs.jackpotPayout)
	}

	cs.gameEndFlag = true
	cs.phase = CaribbeanStudPhaseEnd
	cs.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal 5枚ずつ配る
func (cs *CaribbeanStud) deal() {
	cs.playerHand = make([]*Card, 0, CaribbeanStudHandSize)
	cs.dealerHand = make([]*Card, 0, CaribbeanStudHandSize)
	for range CaribbeanStudHandSize {
		cs.playerHand = append(cs.playerHand, cs.trumpCards.DrawCard())
		cs.dealerHand = append(cs.dealerHand, cs.trumpCards.DrawCard())
	}
	cs.appendLog(-1, "deal", "dealt 5 cards each", nil)
}

// resolve ゲーム解決（Play後の処理）
func (cs *CaribbeanStud) resolve() {
	cs.playerHandRank = evalFiveCardHand(cs.playerHand)
	cs.dealerHandRank = evalFiveCardHand(cs.dealerHand)
	cs.dealerQualified = cs.checkDealerQualifies()

	cmp := cs.compareHands()
	switch {
	case cmp > 0:
		cs.result = GameResultWin
	case cmp < 0:
		cs.result = GameResultLose
	default:
		cs.result = GameResultDraw
	}

	cs.calculatePayouts()
	cs.evaluateJackpot()

	totalPayout := cs.antePayout + cs.playPayout + cs.jackpotPayout
	if totalPayout > 0 {
		cs.chips.AddChips(totalPayout)
	}

	cs.gameEndFlag = true
	cs.phase = CaribbeanStudPhaseEnd

	var resultStr string
	switch cs.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	cs.appendLog(-1, "result", resultStr, nil)
}

// compareHands プレイヤーとディーラーのハンドを比較する
func (cs *CaribbeanStud) compareHands() int {
	if cs.playerHandRank > cs.dealerHandRank {
		return 1
	}
	if cs.playerHandRank < cs.dealerHandRank {
		return -1
	}
	return compareHighCardsSlice(cs.playerHand, cs.dealerHand)
}

// checkDealerQualifies ディーラークオリファイ条件: ペア以上、または A-K ハイ
func (cs *CaribbeanStud) checkDealerQualifies() bool {
	if cs.dealerHandRank >= PokerHandOnePair {
		return true
	}
	hasAce := false
	hasKing := false
	for _, c := range cs.dealerHand {
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
func (cs *CaribbeanStud) calculatePayouts() {
	if !cs.dealerQualified {
		// ディーラー未クオリファイ: アンテ1:1、プレイベットはプッシュ（返却のみ）
		cs.antePayout = cs.anteBet * 2
		cs.playPayout = cs.playBet
		return
	}
	switch cs.result {
	case GameResultWin:
		cs.antePayout = cs.anteBet * 2
		multiplier := cs.playMultiplier()
		cs.playPayout = cs.playBet + cs.playBet*multiplier
	case GameResultDraw:
		cs.antePayout = cs.anteBet
		cs.playPayout = cs.playBet
	case GameResultLose:
		cs.antePayout = 0
		cs.playPayout = 0
	}
}

// playMultiplier プレイベット配当倍率（プレイヤーハンドランクに基づく）
func (cs *CaribbeanStud) playMultiplier() int {
	switch cs.playerHandRank {
	case PokerHandRoyalFlush:
		return CaribbeanStudPayRoyalFlush
	case PokerHandStraightFlush:
		return CaribbeanStudPayStraightFlush
	case PokerHandFourOfAKind:
		return CaribbeanStudPayFourOfAKind
	case PokerHandFullHouse:
		return CaribbeanStudPayFullHouse
	case PokerHandFlush:
		return CaribbeanStudPayFlush
	case PokerHandStraight:
		return CaribbeanStudPayStraight
	case PokerHandThreeOfAKind:
		return CaribbeanStudPayThreeOfAKind
	case PokerHandTwoPair:
		return CaribbeanStudPayTwoPair
	default:
		return CaribbeanStudPayPair
	}
}

// evaluateJackpot ジャックポットサイドベット評価（独立）
func (cs *CaribbeanStud) evaluateJackpot() {
	if cs.jackpotBet <= 0 {
		return
	}
	switch cs.playerHandRank {
	case PokerHandRoyalFlush:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanStudJackpotRoyalFlush
	case PokerHandStraightFlush:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanStudJackpotStraightFlush
	case PokerHandFourOfAKind:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanStudJackpotFourOfAKind
	case PokerHandFullHouse:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanStudJackpotFullHouse
	case PokerHandFlush:
		cs.jackpotPayout = cs.jackpotBet * CaribbeanStudJackpotFlush
	}
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (cs *CaribbeanStud) GetPlayerHand() []*Card { return cs.playerHand }

// GetDealerHand ディーラーハンド取得
func (cs *CaribbeanStud) GetDealerHand() []*Card { return cs.dealerHand }

// GetPhase 現在のフェーズ
func (cs *CaribbeanStud) GetPhase() int { return cs.phase }

// GetGameEndFlag ゲーム終了フラグ
func (cs *CaribbeanStud) GetGameEndFlag() bool { return cs.gameEndFlag }

// GetAnteBet アンテベット額
func (cs *CaribbeanStud) GetAnteBet() int { return cs.anteBet }

// GetJackpotBet ジャックポットベット額
func (cs *CaribbeanStud) GetJackpotBet() int { return cs.jackpotBet }

// GetPlayBet プレイベット額
func (cs *CaribbeanStud) GetPlayBet() int { return cs.playBet }

// GetResult ゲーム結果
func (cs *CaribbeanStud) GetResult() GameResult { return cs.result }

// GetAntePayout アンテ配当
func (cs *CaribbeanStud) GetAntePayout() int { return cs.antePayout }

// GetPlayPayout プレイ配当
func (cs *CaribbeanStud) GetPlayPayout() int { return cs.playPayout }

// GetJackpotPayout ジャックポット配当
func (cs *CaribbeanStud) GetJackpotPayout() int { return cs.jackpotPayout }

// GetTotalPayout 合計配当
func (cs *CaribbeanStud) GetTotalPayout() int {
	return cs.antePayout + cs.playPayout + cs.jackpotPayout
}

// GetDealerQualified ディーラークオリファイ
func (cs *CaribbeanStud) GetDealerQualified() bool { return cs.dealerQualified }

// GetPlayerHandRank プレイヤーハンドランク
func (cs *CaribbeanStud) GetPlayerHandRank() int { return cs.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (cs *CaribbeanStud) GetDealerHandRank() int { return cs.dealerHandRank }

// GetChips チップ
func (cs *CaribbeanStud) GetChips() int { return cs.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (cs *CaribbeanStud) SetPhase(phase int) { cs.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (cs *CaribbeanStud) SetPlayerHand(cards []*Card) { cs.playerHand = cards }

// SetDealerHand ディーラーハンド設定（テスト用）
func (cs *CaribbeanStud) SetDealerHand(cards []*Card) { cs.dealerHand = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (cs *CaribbeanStud) SetAnteBet(amount int) { cs.anteBet = amount }

// SetJackpotBet ジャックポットベット額設定（テスト用）
func (cs *CaribbeanStud) SetJackpotBet(amount int) { cs.jackpotBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (cs *CaribbeanStud) SetPlayBet(amount int) { cs.playBet = amount }

// SetChips チップ設定（テスト用）
func (cs *CaribbeanStud) SetChips(chips int) { cs.chips.SetChips(chips) }

// caribbeanStudJSON は CaribbeanStud の JSON ワイヤーフォーマット
type caribbeanStudJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	JackpotBet      int               `json:"jb"`
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
func (cs *CaribbeanStud) MarshalJSON() ([]byte, error) {
	return json.Marshal(caribbeanStudJSON{
		TrumpCards:      cs.trumpCards,
		PlayerHand:      cs.playerHand,
		DealerHand:      cs.dealerHand,
		Chips:           &cs.chips,
		AnteBet:         cs.anteBet,
		JackpotBet:      cs.jackpotBet,
		PlayBet:         cs.playBet,
		Phase:           cs.phase,
		GameEndFlag:     cs.gameEndFlag,
		Result:          cs.result,
		AntePayout:      cs.antePayout,
		PlayPayout:      cs.playPayout,
		JackpotPayout:   cs.jackpotPayout,
		DealerQualified: cs.dealerQualified,
		PlayerHandRank:  cs.playerHandRank,
		DealerHandRank:  cs.dealerHandRank,
		ActionLog:       cs.actionLog,
	})
}

// caribbeanStudMaxSliceLen caps slice sizes during deserialisation.
const caribbeanStudMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (cs *CaribbeanStud) UnmarshalJSON(data []byte) error {
	var j caribbeanStudJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > caribbeanStudMaxSliceLen || len(j.DealerHand) > caribbeanStudMaxSliceLen ||
		len(j.ActionLog) > caribbeanStudMaxSliceLen {
		return fmt.Errorf("caribbeanstud: input array exceeds maximum allowed size")
	}

	cs.trumpCards = j.TrumpCards
	if cs.trumpCards == nil {
		cs.trumpCards = NewTrumpCards(0)
	}
	cs.playerHand = j.PlayerHand
	if cs.playerHand == nil {
		cs.playerHand = make([]*Card, 0)
	}
	cs.dealerHand = j.DealerHand
	if cs.dealerHand == nil {
		cs.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		cs.chips = *j.Chips
	}
	cs.anteBet = j.AnteBet
	cs.jackpotBet = j.JackpotBet
	cs.playBet = j.PlayBet
	cs.phase = j.Phase
	cs.gameEndFlag = j.GameEndFlag
	cs.result = j.Result
	cs.antePayout = j.AntePayout
	cs.playPayout = j.PlayPayout
	cs.jackpotPayout = j.JackpotPayout
	cs.dealerQualified = j.DealerQualified
	cs.playerHandRank = j.PlayerHandRank
	cs.dealerHandRank = j.DealerHandRank
	cs.actionLog = j.ActionLog
	if cs.actionLog == nil {
		cs.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
