//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// スリーカードポーカーフェーズ定数
const (
	ThreeCardPhaseBet    = 1 // ベットフェーズ
	ThreeCardPhaseAction = 2 // アクションフェーズ（Play/Fold選択）
	ThreeCardPhaseEnd    = 3 // 終了フェーズ
)

// スリーカードポーカーデフォルト値
const (
	ThreeCardDefaultChips = 1000  // デフォルトチップ
	ThreeCardMinBet       = 10    // 最低ベット額
	ThreeCardMaxBet       = 10000 // 最大ベット額
	ThreeCardHandSize     = 3     // ハンドサイズ
)

// アンテボーナス配当倍率
const (
	ThreeCardAnteBonusStraightFlush = 5 // ストレートフラッシュ 5:1
	ThreeCardAnteBonusThreeOfAKind  = 4 // スリーカード 4:1
	ThreeCardAnteBonusStraight      = 1 // ストレート 1:1
)

// ペアプラス配当倍率
const (
	ThreeCardPairPlusStraightFlush = 40 // ストレートフラッシュ 40:1
	ThreeCardPairPlusThreeOfAKind  = 30 // スリーカード 30:1
	ThreeCardPairPlusStraight      = 6  // ストレート 6:1
	ThreeCardPairPlusFlush         = 3  // フラッシュ 3:1
	ThreeCardPairPlusPair          = 1  // ペア 1:1
)

// ThreeCard スリーカードポーカークラス
type ThreeCard struct {
	trumpCards      *TrumpCards // トランプカード
	playerHand      []*Card     // プレイヤーハンド
	dealerHand      []*Card     // ディーラーハンド
	chips           ChipHolder  // チップ
	anteBet         int         // アンテベット額
	pairPlusBet     int         // ペアプラスベット額
	playBet         int         // プレイベット額
	phase           int         // 現在のフェーズ
	gameEndFlag     bool        // ゲーム終了フラグ
	result          GameResult  // ゲーム結果
	antePayout      int         // アンテ配当
	playPayout      int         // プレイ配当
	anteBonusPayout int         // アンテボーナス配当
	pairPlusPayout  int         // ペアプラス配当
	dealerQualified bool        // ディーラークオリファイフラグ
	playerHandRank  int         // プレイヤーハンドランク
	dealerHandRank  int         // ディーラーハンドランク
	actionLogBase
}

// NewThreeCard コンストラクタ
func NewThreeCard(trumpCards *TrumpCards) *ThreeCard {
	trumpCards.Shuffle()
	return &ThreeCard{
		trumpCards: trumpCards,
		phase:      ThreeCardPhaseBet,
	}
}

// NewDefaultThreeCard デフォルト設定のスリーカードポーカーを生成するファクトリ関数
func NewDefaultThreeCard() *ThreeCard {
	tc := NewThreeCard(NewTrumpCards(0))
	tc.chips.SetChips(ThreeCardDefaultChips)
	return tc
}

// Reset ゲーム初期化
func (tc *ThreeCard) Reset() {
	tc.gameEndFlag = false
	tc.phase = ThreeCardPhaseBet
	tc.playerHand = nil
	tc.dealerHand = nil
	tc.anteBet = 0
	tc.pairPlusBet = 0
	tc.playBet = 0
	tc.result = 0
	tc.antePayout = 0
	tc.playPayout = 0
	tc.anteBonusPayout = 0
	tc.pairPlusPayout = 0
	tc.dealerQualified = false
	tc.playerHandRank = 0
	tc.dealerHandRank = 0
	tc.actionLog = nil
	if tc.chips.GetChips() < ThreeCardMinBet {
		tc.chips.SetChips(ThreeCardDefaultChips)
	}
	tc.trumpCards = NewTrumpCards(0)
	for range 10 {
		tc.trumpCards.Shuffle()
	}
}

// Bet アンテベット＆カード配布
func (tc *ThreeCard) Bet(ante, pairPlus int) error {
	if tc.phase != ThreeCardPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < ThreeCardMinBet || ante%ThreeCardMinBet != 0 || ante > ThreeCardMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if pairPlus < 0 {
		return NewDomainError(ErrInvalidAmount, "Pair Plus bet must not be negative.")
	}
	if pairPlus > 0 && (pairPlus < ThreeCardMinBet || pairPlus%ThreeCardMinBet != 0 || pairPlus > ThreeCardMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid Pair Plus bet amount.")
	}
	totalCost := ante + pairPlus
	if !tc.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	tc.anteBet = ante
	tc.pairPlusBet = pairPlus
	tc.appendLog(0, "bet", fmt.Sprintf("ante=%d pairplus=%d", ante, pairPlus), nil)

	// ディール: 3枚ずつ配る
	tc.deal()
	tc.phase = ThreeCardPhaseAction
	return nil
}

// Play プレイ（アンテと同額のプレイベットを置いて勝負）
func (tc *ThreeCard) Play() error {
	if tc.phase != ThreeCardPhaseAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action phase.")
	}
	// プレイベットはアンテと同額
	if !tc.chips.SubtractChips(tc.anteBet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	tc.playBet = tc.anteBet
	tc.appendLog(0, "play", fmt.Sprintf("play bet=%d", tc.playBet), nil)

	tc.resolve()
	return nil
}

// Fold フォールド（アンテ没収、ペアプラスは別途評価）
func (tc *ThreeCard) Fold() error {
	if tc.phase != ThreeCardPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	tc.appendLog(0, "fold", "player folds", nil)

	tc.result = GameResultLose
	tc.playerHandRank = evalThreeCardHand(tc.playerHand)
	tc.dealerHandRank = evalThreeCardHand(tc.dealerHand)

	// ペアプラスはフォールドしても評価される
	tc.evaluatePairPlus()

	tc.gameEndFlag = true
	tc.phase = ThreeCardPhaseEnd
	tc.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal 3枚ずつ配る
func (tc *ThreeCard) deal() {
	tc.playerHand = make([]*Card, 0, ThreeCardHandSize)
	tc.dealerHand = make([]*Card, 0, ThreeCardHandSize)
	for range ThreeCardHandSize {
		tc.playerHand = append(tc.playerHand, tc.trumpCards.DrawCard())
		tc.dealerHand = append(tc.dealerHand, tc.trumpCards.DrawCard())
	}
	tc.appendLog(-1, "deal", "dealt 3 cards each", nil)
}

// resolve ゲーム解決（Play後の処理）
func (tc *ThreeCard) resolve() {
	tc.playerHandRank = evalThreeCardHand(tc.playerHand)
	tc.dealerHandRank = evalThreeCardHand(tc.dealerHand)
	tc.dealerQualified = tc.checkDealerQualifies()

	// 勝敗判定
	cmp := compareThreeCardHands(tc.playerHand, tc.dealerHand)
	if cmp > 0 {
		tc.result = GameResultWin
	} else if cmp < 0 {
		tc.result = GameResultLose
	} else {
		tc.result = GameResultDraw
	}

	// 配当計算
	tc.calculatePayouts()
	// ペアプラス評価
	tc.evaluatePairPlus()
	// アンテボーナス評価
	tc.evaluateAnteBonus()

	// チップ加算
	totalPayout := tc.antePayout + tc.playPayout + tc.anteBonusPayout + tc.pairPlusPayout
	if totalPayout > 0 {
		tc.chips.AddChips(totalPayout)
	}

	tc.gameEndFlag = true
	tc.phase = ThreeCardPhaseEnd

	var resultStr string
	switch tc.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	tc.appendLog(-1, "result", resultStr, nil)
}

// checkDealerQualifies ディーラーがクオリファイするか（Qハイ以上、またはペア以上）
func (tc *ThreeCard) checkDealerQualifies() bool {
	// ペア以上は無条件でクオリファイ
	if tc.dealerHandRank >= ThreeCardHandPair {
		return true
	}
	// ハイカードの場合はQハイ以上が必要
	vals := threeCardHandHighValues(tc.dealerHand)
	return vals[0] >= 12 // Queen(12), King(13), Ace(14)
}

// calculatePayouts アンテ/プレイの配当計算
func (tc *ThreeCard) calculatePayouts() {
	if !tc.dealerQualified {
		// ディーラー未クオリファイ: アンテ1:1返却、プレイはプッシュ（返却）
		tc.antePayout = tc.anteBet * 2 // 元のベット + 1:1
		tc.playPayout = tc.playBet     // プッシュ（返却のみ）
		return
	}

	switch tc.result {
	case GameResultWin:
		tc.antePayout = tc.anteBet * 2 // 元のベット + 1:1
		tc.playPayout = tc.playBet * 2 // 元のベット + 1:1
	case GameResultDraw:
		tc.antePayout = tc.anteBet // プッシュ（返却のみ）
		tc.playPayout = tc.playBet // プッシュ（返却のみ）
	case GameResultLose:
		tc.antePayout = 0 // 没収
		tc.playPayout = 0 // 没収
	}
}

// evaluateAnteBonus アンテボーナス評価（ディーラーの結果に関係なく）
// アンテボーナスはアンテ配当とは別に支払われるボーナスのみ（元ベット返却なし）
func (tc *ThreeCard) evaluateAnteBonus() {
	switch tc.playerHandRank {
	case ThreeCardHandStraightFlush:
		tc.anteBonusPayout = tc.anteBet * ThreeCardAnteBonusStraightFlush
	case ThreeCardHandThreeOfAKind:
		tc.anteBonusPayout = tc.anteBet * ThreeCardAnteBonusThreeOfAKind
	case ThreeCardHandStraight:
		tc.anteBonusPayout = tc.anteBet * ThreeCardAnteBonusStraight
	}
}

// evaluatePairPlus ペアプラス評価（独立したサイドベット）
func (tc *ThreeCard) evaluatePairPlus() {
	if tc.pairPlusBet <= 0 {
		return
	}
	switch tc.playerHandRank {
	case ThreeCardHandStraightFlush:
		tc.pairPlusPayout = tc.pairPlusBet + tc.pairPlusBet*ThreeCardPairPlusStraightFlush
	case ThreeCardHandThreeOfAKind:
		tc.pairPlusPayout = tc.pairPlusBet + tc.pairPlusBet*ThreeCardPairPlusThreeOfAKind
	case ThreeCardHandStraight:
		tc.pairPlusPayout = tc.pairPlusBet + tc.pairPlusBet*ThreeCardPairPlusStraight
	case ThreeCardHandFlush:
		tc.pairPlusPayout = tc.pairPlusBet + tc.pairPlusBet*ThreeCardPairPlusFlush
	case ThreeCardHandPair:
		tc.pairPlusPayout = tc.pairPlusBet + tc.pairPlusBet*ThreeCardPairPlusPair
	}
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (tc *ThreeCard) GetPlayerHand() []*Card { return tc.playerHand }

// GetDealerHand ディーラーハンド取得
func (tc *ThreeCard) GetDealerHand() []*Card { return tc.dealerHand }

// GetPhase 現在のフェーズ
func (tc *ThreeCard) GetPhase() int { return tc.phase }

// GetGameEndFlag ゲーム終了フラグ
func (tc *ThreeCard) GetGameEndFlag() bool { return tc.gameEndFlag }

// GetAnteBet アンテベット額
func (tc *ThreeCard) GetAnteBet() int { return tc.anteBet }

// GetPairPlusBet ペアプラスベット額
func (tc *ThreeCard) GetPairPlusBet() int { return tc.pairPlusBet }

// GetPlayBet プレイベット額
func (tc *ThreeCard) GetPlayBet() int { return tc.playBet }

// GetResult ゲーム結果
func (tc *ThreeCard) GetResult() GameResult { return tc.result }

// GetAntePayout アンテ配当
func (tc *ThreeCard) GetAntePayout() int { return tc.antePayout }

// GetPlayPayout プレイ配当
func (tc *ThreeCard) GetPlayPayout() int { return tc.playPayout }

// GetAnteBonusPayout アンテボーナス配当
func (tc *ThreeCard) GetAnteBonusPayout() int { return tc.anteBonusPayout }

// GetPairPlusPayout ペアプラス配当
func (tc *ThreeCard) GetPairPlusPayout() int { return tc.pairPlusPayout }

// GetTotalPayout 合計配当
func (tc *ThreeCard) GetTotalPayout() int {
	return tc.antePayout + tc.playPayout + tc.anteBonusPayout + tc.pairPlusPayout
}

// GetDealerQualified ディーラークオリファイ
func (tc *ThreeCard) GetDealerQualified() bool { return tc.dealerQualified }

// GetPlayerHandRank プレイヤーハンドランク
func (tc *ThreeCard) GetPlayerHandRank() int { return tc.playerHandRank }

// GetDealerHandRank ディーラーハンドランク
func (tc *ThreeCard) GetDealerHandRank() int { return tc.dealerHandRank }

// GetChips チップ
func (tc *ThreeCard) GetChips() int { return tc.chips.GetChips() }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (tc *ThreeCard) SetPhase(phase int) { tc.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (tc *ThreeCard) SetPlayerHand(cards []*Card) { tc.playerHand = cards }

// SetDealerHand ディーラーハンド設定（テスト用）
func (tc *ThreeCard) SetDealerHand(cards []*Card) { tc.dealerHand = cards }

// SetAnteBet アンテベット額設定（テスト用）
func (tc *ThreeCard) SetAnteBet(amount int) { tc.anteBet = amount }

// SetPairPlusBet ペアプラスベット額設定（テスト用）
func (tc *ThreeCard) SetPairPlusBet(amount int) { tc.pairPlusBet = amount }

// SetPlayBet プレイベット額設定（テスト用）
func (tc *ThreeCard) SetPlayBet(amount int) { tc.playBet = amount }

// SetResult ゲーム結果設定（テスト用）
func (tc *ThreeCard) SetResult(result GameResult) { tc.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (tc *ThreeCard) SetGameEndFlag(flag bool) { tc.gameEndFlag = flag }

// SetChips チップ設定（テスト用）
func (tc *ThreeCard) SetChips(chips int) { tc.chips.SetChips(chips) }

// SetDealerQualified ディーラークオリファイ設定（テスト用）
func (tc *ThreeCard) SetDealerQualified(qualified bool) { tc.dealerQualified = qualified }

// SetPlayerHandRank プレイヤーハンドランク設定（テスト用）
func (tc *ThreeCard) SetPlayerHandRank(rank int) { tc.playerHandRank = rank }

// SetDealerHandRank ディーラーハンドランク設定（テスト用）
func (tc *ThreeCard) SetDealerHandRank(rank int) { tc.dealerHandRank = rank }

// SetAntePayout アンテ配当設定（テスト用）
func (tc *ThreeCard) SetAntePayout(payout int) { tc.antePayout = payout }

// SetPlayPayout プレイ配当設定（テスト用）
func (tc *ThreeCard) SetPlayPayout(payout int) { tc.playPayout = payout }

// SetAnteBonusPayout アンテボーナス配当設定（テスト用）
func (tc *ThreeCard) SetAnteBonusPayout(payout int) { tc.anteBonusPayout = payout }

// SetPairPlusPayout ペアプラス配当設定（テスト用）
func (tc *ThreeCard) SetPairPlusPayout(payout int) { tc.pairPlusPayout = payout }

// threeCardJSON is the JSON wire format for ThreeCard.
type threeCardJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	PairPlusBet     int               `json:"pp"`
	PlayBet         int               `json:"pb"`
	Phase           int               `json:"ps"`
	GameEndFlag     bool              `json:"ge"`
	Result          GameResult        `json:"rs"`
	AntePayout      int               `json:"ap"`
	PlayPayout      int               `json:"plp"`
	AnteBonusPayout int               `json:"abp"`
	PairPlusPayout  int               `json:"ppp"`
	DealerQualified bool              `json:"dq"`
	PlayerHandRank  int               `json:"pr"`
	DealerHandRank  int               `json:"dr"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (tc *ThreeCard) MarshalJSON() ([]byte, error) {
	return json.Marshal(threeCardJSON{
		TrumpCards:      tc.trumpCards,
		PlayerHand:      tc.playerHand,
		DealerHand:      tc.dealerHand,
		Chips:           &tc.chips,
		AnteBet:         tc.anteBet,
		PairPlusBet:     tc.pairPlusBet,
		PlayBet:         tc.playBet,
		Phase:           tc.phase,
		GameEndFlag:     tc.gameEndFlag,
		Result:          tc.result,
		AntePayout:      tc.antePayout,
		PlayPayout:      tc.playPayout,
		AnteBonusPayout: tc.anteBonusPayout,
		PairPlusPayout:  tc.pairPlusPayout,
		DealerQualified: tc.dealerQualified,
		PlayerHandRank:  tc.playerHandRank,
		DealerHandRank:  tc.dealerHandRank,
		ActionLog:       tc.actionLog,
	})
}

// threeCardMaxSliceLen caps slice sizes during deserialisation.
const threeCardMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (tc *ThreeCard) UnmarshalJSON(data []byte) error {
	var j threeCardJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > threeCardMaxSliceLen || len(j.DealerHand) > threeCardMaxSliceLen ||
		len(j.ActionLog) > threeCardMaxSliceLen {
		return fmt.Errorf("threecard: input array exceeds maximum allowed size")
	}

	tc.trumpCards = j.TrumpCards
	if tc.trumpCards == nil {
		tc.trumpCards = NewTrumpCards(0)
	}
	tc.playerHand = j.PlayerHand
	if tc.playerHand == nil {
		tc.playerHand = make([]*Card, 0)
	}
	tc.dealerHand = j.DealerHand
	if tc.dealerHand == nil {
		tc.dealerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		tc.chips = *j.Chips
	}
	tc.anteBet = j.AnteBet
	tc.pairPlusBet = j.PairPlusBet
	tc.playBet = j.PlayBet
	tc.phase = j.Phase
	tc.gameEndFlag = j.GameEndFlag
	tc.result = j.Result
	tc.antePayout = j.AntePayout
	tc.playPayout = j.PlayPayout
	tc.anteBonusPayout = j.AnteBonusPayout
	tc.pairPlusPayout = j.PairPlusPayout
	tc.dealerQualified = j.DealerQualified
	tc.playerHandRank = j.PlayerHandRank
	tc.dealerHandRank = j.DealerHandRank
	tc.actionLog = j.ActionLog
	if tc.actionLog == nil {
		tc.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
