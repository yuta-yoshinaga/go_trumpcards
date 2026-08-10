//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// バカラフェーズ定数
const (
	BaccaratPhaseBet = 1 // ベットフェーズ
	BaccaratPhaseEnd = 2 // 終了フェーズ
)

// バカラベットタイプ定数
const (
	BaccaratBetPlayer = 0 // プレイヤーベット
	BaccaratBetBanker = 1 // バンカーベット
	BaccaratBetTie    = 2 // タイベット
)

// バカラデフォルト値
const (
	BaccaratDefaultChips    = 1000  // デフォルトチップ
	BaccaratMinBet          = 10    // 最低ベット額
	BaccaratMaxBet          = 10000 // 最大ベット額
	BaccaratCommissionRate  = 5     // バンカーコミッション率(%)
	BaccaratTiePayoutRate   = 8     // タイ配当倍率
	BaccaratNaturalMinValue = 8     // ナチュラル最小値
	BaccaratDrawMaxValue    = 5     // プレイヤードロー閾値
	BaccaratStandMinValue   = 7     // スタンド最小値
)

// バカラ結果定数（罫線用）
const (
	BaccaratResultPlayer = 0 // プレイヤー勝利
	BaccaratResultBanker = 1 // バンカー勝利
	BaccaratResultTie    = 2 // タイ
)

// Baccarat バカラクラス
type Baccarat struct {
	trumpCards  *TrumpCards // トランプカード
	playerHand  []*Card     // プレイヤーハンド
	bankerHand  []*Card     // バンカーハンド
	chips       ChipHolder  // チップ
	betAmount   int         // ベット額
	betType     int         // ベットタイプ
	phase       int         // 現在のフェーズ
	gameEndFlag bool        // ゲーム終了フラグ
	result      GameResult  // ゲーム結果
	payout      int         // 配当金額
	actionLogBase
	history        []int               // 罫線（Big Road）履歴
	playerPairBet  int                 // プレイヤーペアベット額
	bankerPairBet  int                 // バンカーペアベット額
	sideBetResults []*BacSideBetResult // サイドベット結果
}

// NewBaccarat コンストラクタ
func NewBaccarat(trumpCards *TrumpCards) *Baccarat {
	trumpCards.Shuffle()
	b := &Baccarat{
		trumpCards: trumpCards,
		phase:      BaccaratPhaseBet,
	}
	return b
}

// NewDefaultBaccarat デフォルト設定のバカラを生成するファクトリ関数
func NewDefaultBaccarat() *Baccarat {
	b := NewBaccarat(NewTrumpCards(0))
	b.chips.SetChips(BaccaratDefaultChips)
	return b
}

// Reset ゲーム初期化（履歴は保持）
func (b *Baccarat) Reset() {
	b.gameEndFlag = false
	b.phase = BaccaratPhaseBet
	b.playerHand = nil
	b.bankerHand = nil
	b.betAmount = 0
	b.betType = 0
	b.result = 0
	b.payout = 0
	b.actionLog = nil
	b.playerPairBet = 0
	b.bankerPairBet = 0
	b.sideBetResults = nil
	if b.chips.GetChips() < BaccaratMinBet {
		b.chips.SetChips(BaccaratDefaultChips)
	}
	b.trumpCards = NewTrumpCards(0)
	for i := 0; i < 10; i++ {
		b.trumpCards.Shuffle()
	}
}

// ClearHistory 罫線履歴をクリアする
func (b *Baccarat) ClearHistory() {
	b.history = nil
}

// Bet ベット＆ゲーム実行（バカラはベット後に全自動進行）
func (b *Baccarat) Bet(amount, betType, ppBet, bpBet int) error {
	if b.phase != BaccaratPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if betType < BaccaratBetPlayer || betType > BaccaratBetTie {
		return NewDomainError(ErrInvalidPlay, "Invalid bet type.")
	}
	if amount < BaccaratMinBet || amount%BaccaratMinBet != 0 || amount > BaccaratMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if ppBet < 0 || bpBet < 0 {
		return NewDomainError(ErrInvalidAmount, "Side bet amount must not be negative.")
	}
	if ppBet > 0 && (ppBet < BaccaratMinBet || ppBet%BaccaratMinBet != 0 || ppBet > BaccaratMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid player pair bet amount.")
	}
	if bpBet > 0 && (bpBet < BaccaratMinBet || bpBet%BaccaratMinBet != 0 || bpBet > BaccaratMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid banker pair bet amount.")
	}
	totalCost := amount + ppBet + bpBet
	if !b.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	b.betAmount = amount
	b.betType = betType
	b.playerPairBet = ppBet
	b.bankerPairBet = bpBet
	b.appendLog(0, "bet", fmt.Sprintf("bet %d on %s", amount, betTypeName(betType)), nil)

	// ディール
	b.deal()
	// サードカードルール
	b.drawPhase()
	// 勝敗判定
	b.judge()
	return nil
}

// deal 2枚ずつ配る
func (b *Baccarat) deal() {
	b.playerHand = make([]*Card, 0, 3)
	b.bankerHand = make([]*Card, 0, 3)
	// プレイヤー1枚目、バンカー1枚目、プレイヤー2枚目、バンカー2枚目
	b.playerHand = append(b.playerHand, b.trumpCards.DrawCard())
	b.bankerHand = append(b.bankerHand, b.trumpCards.DrawCard())
	b.playerHand = append(b.playerHand, b.trumpCards.DrawCard())
	b.bankerHand = append(b.bankerHand, b.trumpCards.DrawCard())
	b.appendLog(-1, "deal", fmt.Sprintf("player=%d banker=%d",
		b.CalculateHandValue(b.playerHand), b.CalculateHandValue(b.bankerHand)), nil)
}

// drawPhase サードカードルール適用
func (b *Baccarat) drawPhase() {
	playerTotal := b.CalculateHandValue(b.playerHand)
	bankerTotal := b.CalculateHandValue(b.bankerHand)

	// ナチュラル判定（どちらかが8か9ならドロー無し）
	if playerTotal >= BaccaratNaturalMinValue || bankerTotal >= BaccaratNaturalMinValue {
		b.appendLog(-1, "natural", fmt.Sprintf("player=%d banker=%d", playerTotal, bankerTotal), nil)
		return
	}

	// プレイヤーのサードカード
	playerDrew := false
	var playerThirdCardValue int
	if b.shouldPlayerDraw(playerTotal) {
		card := b.trumpCards.DrawCard()
		b.playerHand = append(b.playerHand, card)
		playerDrew = true
		playerThirdCardValue = b.cardPointValue(card)
		b.appendLog(-1, "draw", fmt.Sprintf("player draws %d", playerThirdCardValue), []*Card{card})
	}

	// バンカーのサードカード
	if b.shouldBankerDraw(bankerTotal, playerThirdCardValue, playerDrew) {
		card := b.trumpCards.DrawCard()
		b.bankerHand = append(b.bankerHand, card)
		b.appendLog(-1, "draw", fmt.Sprintf("banker draws %d", b.cardPointValue(card)), []*Card{card})
	}
}

// judge 勝敗判定＆配当計算
func (b *Baccarat) judge() {
	playerTotal := b.CalculateHandValue(b.playerHand)
	bankerTotal := b.CalculateHandValue(b.bankerHand)

	if playerTotal > bankerTotal {
		b.result = GameResultWin
		b.appendLog(-1, "result", "player wins", nil)
	} else if bankerTotal > playerTotal {
		b.result = GameResultLose
		b.appendLog(-1, "result", "banker wins", nil)
	} else {
		b.result = GameResultDraw
		b.appendLog(-1, "result", "tie", nil)
	}

	// 罫線に結果を追加
	switch b.result {
	case GameResultWin:
		b.history = append(b.history, BaccaratResultPlayer)
	case GameResultLose:
		b.history = append(b.history, BaccaratResultBanker)
	default:
		b.history = append(b.history, BaccaratResultTie)
	}

	// 配当計算
	b.payout = b.calculatePayout()
	b.chips.AddChips(b.payout)

	// サイドベット評価
	b.evaluateSideBets()

	b.gameEndFlag = true
	b.phase = BaccaratPhaseEnd
}

// calculatePayout 配当計算
func (b *Baccarat) calculatePayout() int {
	switch b.betType {
	case BaccaratBetPlayer:
		if b.result == GameResultWin {
			return b.betAmount * 2
		}
		if b.result == GameResultDraw {
			return b.betAmount // 引き分けは返却
		}
		return 0
	case BaccaratBetBanker:
		if b.result == GameResultLose { // GameResultLose = player lost = banker won
			// 5%コミッション
			winnings := b.betAmount - b.betAmount*BaccaratCommissionRate/100
			return b.betAmount + winnings
		}
		if b.result == GameResultDraw {
			return b.betAmount
		}
		return 0
	case BaccaratBetTie:
		if b.result == GameResultDraw {
			return b.betAmount + b.betAmount*BaccaratTiePayoutRate
		}
		return 0
	}
	return 0
}

// CalculateHandValue ハンドの合計値を計算（mod 10）
func (b *Baccarat) CalculateHandValue(cards []*Card) int {
	total := 0
	for _, card := range cards {
		total += b.cardPointValue(card)
	}
	return total % 10
}

// cardPointValue カードのポイント値（A=1, 2-9=額面, 10/J/Q/K=0）
func (b *Baccarat) cardPointValue(card *Card) int {
	v := card.GetValue()
	if v >= 10 {
		return 0
	}
	return v
}

// shouldPlayerDraw プレイヤーがサードカードを引くか（0-5なら引く）
func (b *Baccarat) shouldPlayerDraw(playerTotal int) bool {
	return playerTotal <= BaccaratDrawMaxValue
}

// shouldBankerDraw バンカーがサードカードを引くか
func (b *Baccarat) shouldBankerDraw(bankerTotal, playerThirdCardValue int, playerDrew bool) bool {
	// プレイヤーがドローしなかった場合、バンカーは0-5で引く
	if !playerDrew {
		return bankerTotal <= BaccaratDrawMaxValue
	}
	// プレイヤーがドローした場合のバンカールール
	switch bankerTotal {
	case 0, 1, 2:
		return true
	case 3:
		return playerThirdCardValue != 8
	case 4:
		return playerThirdCardValue >= 2 && playerThirdCardValue <= 7
	case 5:
		return playerThirdCardValue >= 4 && playerThirdCardValue <= 7
	case 6:
		return playerThirdCardValue == 6 || playerThirdCardValue == 7
	default: // 7+
		return false
	}
}

// betTypeName ベットタイプ名
func betTypeName(betType int) string {
	switch betType {
	case BaccaratBetPlayer:
		return "player"
	case BaccaratBetBanker:
		return "banker"
	case BaccaratBetTie:
		return "tie"
	default:
		return "unknown"
	}
}

// evaluateSideBets サイドベットを評価し、当選時はチップに加算する
func (b *Baccarat) evaluateSideBets() {
	b.sideBetResults = nil
	if b.playerPairBet > 0 {
		resultType, resultName := EvaluateBaccaratPair(b.playerHand[0], b.playerHand[1])
		payout := 0
		if resultType == BacPairMatch {
			payout = b.playerPairBet + b.playerPairBet*BacPairPayout(resultType)
			b.chips.AddChips(payout)
		}
		b.sideBetResults = append(b.sideBetResults, &BacSideBetResult{
			BetType:    BacSideBetPlayerPair,
			ResultType: resultType,
			ResultName: resultName,
			BetAmount:  b.playerPairBet,
			Payout:     payout,
		})
	}
	if b.bankerPairBet > 0 {
		resultType, resultName := EvaluateBaccaratPair(b.bankerHand[0], b.bankerHand[1])
		payout := 0
		if resultType == BacPairMatch {
			payout = b.bankerPairBet + b.bankerPairBet*BacPairPayout(resultType)
			b.chips.AddChips(payout)
		}
		b.sideBetResults = append(b.sideBetResults, &BacSideBetResult{
			BetType:    BacSideBetBankerPair,
			ResultType: resultType,
			ResultName: resultName,
			BetAmount:  b.bankerPairBet,
			Payout:     payout,
		})
	}
}

// --- Getters ---

// GetPlayerHand プレイヤーハンド取得
func (b *Baccarat) GetPlayerHand() []*Card { return b.playerHand }

// GetBankerHand バンカーハンド取得
func (b *Baccarat) GetBankerHand() []*Card { return b.bankerHand }

// GetPhase 現在のフェーズ
func (b *Baccarat) GetPhase() int { return b.phase }

// GetGameEndFlag ゲーム終了フラグ
func (b *Baccarat) GetGameEndFlag() bool { return b.gameEndFlag }

// GetBetAmount ベット額
func (b *Baccarat) GetBetAmount() int { return b.betAmount }

// GetBetType ベットタイプ
func (b *Baccarat) GetBetType() int { return b.betType }

// GetResult ゲーム結果
func (b *Baccarat) GetResult() GameResult { return b.result }

// GetPayout 配当金額
func (b *Baccarat) GetPayout() int { return b.payout }

// GetChips チップ
func (b *Baccarat) GetChips() int { return b.chips.GetChips() }

// GetPlayerHandValue プレイヤーハンド合計値
func (b *Baccarat) GetPlayerHandValue() int { return b.CalculateHandValue(b.playerHand) }

// GetBankerHandValue バンカーハンド合計値
func (b *Baccarat) GetBankerHandValue() int { return b.CalculateHandValue(b.bankerHand) }

// GetHistory 罫線履歴を取得する
func (b *Baccarat) GetHistory() []int { return b.history }

// GetPlayerPairBet プレイヤーペアベット額を取得する
func (b *Baccarat) GetPlayerPairBet() int { return b.playerPairBet }

// GetBankerPairBet バンカーペアベット額を取得する
func (b *Baccarat) GetBankerPairBet() int { return b.bankerPairBet }

// GetSideBetResults サイドベット結果を取得する
func (b *Baccarat) GetSideBetResults() []*BacSideBetResult { return b.sideBetResults }

// --- Test helpers ---

// SetPhase フェーズ設定（テスト用）
func (b *Baccarat) SetPhase(phase int) { b.phase = phase }

// SetPlayerHand プレイヤーハンド設定（テスト用）
func (b *Baccarat) SetPlayerHand(cards []*Card) { b.playerHand = cards }

// SetBankerHand バンカーハンド設定（テスト用）
func (b *Baccarat) SetBankerHand(cards []*Card) { b.bankerHand = cards }

// SetBetAmount ベット額設定（テスト用）
func (b *Baccarat) SetBetAmount(amount int) { b.betAmount = amount }

// SetBetType ベットタイプ設定（テスト用）
func (b *Baccarat) SetBetType(betType int) { b.betType = betType }

// SetResult ゲーム結果設定（テスト用）
func (b *Baccarat) SetResult(result GameResult) { b.result = result }

// SetGameEndFlag ゲーム終了フラグ設定（テスト用）
func (b *Baccarat) SetGameEndFlag(flag bool) { b.gameEndFlag = flag }

// SetChips チップ設定（テスト用）
func (b *Baccarat) SetChips(chips int) { b.chips.SetChips(chips) }

// SetHistory 罫線履歴設定（テスト用）
func (b *Baccarat) SetHistory(history []int) { b.history = history }

// SetPlayerPairBet プレイヤーペアベット額設定（テスト用）
func (b *Baccarat) SetPlayerPairBet(amount int) { b.playerPairBet = amount }

// SetBankerPairBet バンカーペアベット額設定（テスト用）
func (b *Baccarat) SetBankerPairBet(amount int) { b.bankerPairBet = amount }

// baccaratJSON is the JSON wire format for Baccarat.
type baccaratJSON struct {
	TrumpCards     *TrumpCards         `json:"tc"`
	PlayerHand     []*Card             `json:"ph"`
	BankerHand     []*Card             `json:"bh"`
	Chips          *ChipHolder         `json:"ch"`
	BetAmount      int                 `json:"ba"`
	BetType        int                 `json:"bt"`
	Phase          int                 `json:"ps"`
	GameEndFlag    bool                `json:"ge"`
	Result         GameResult          `json:"rs"`
	Payout         int                 `json:"po"`
	ActionLog      []*ActionLogEntry   `json:"al"`
	History        []int               `json:"hi"`
	PlayerPairBet  int                 `json:"pp"`
	BankerPairBet  int                 `json:"bp"`
	SideBetResults []*BacSideBetResult `json:"sb"`
}

// MarshalJSON implements json.Marshaler.
func (b *Baccarat) MarshalJSON() ([]byte, error) {
	return json.Marshal(baccaratJSON{
		TrumpCards:     b.trumpCards,
		PlayerHand:     b.playerHand,
		BankerHand:     b.bankerHand,
		Chips:          &b.chips,
		BetAmount:      b.betAmount,
		BetType:        b.betType,
		Phase:          b.phase,
		GameEndFlag:    b.gameEndFlag,
		Result:         b.result,
		Payout:         b.payout,
		ActionLog:      b.actionLog,
		History:        b.history,
		PlayerPairBet:  b.playerPairBet,
		BankerPairBet:  b.bankerPairBet,
		SideBetResults: b.sideBetResults,
	})
}

// baccaratMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const baccaratMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (b *Baccarat) UnmarshalJSON(data []byte) error {
	var j baccaratJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// Validate slice sizes to mitigate DoS from oversized arrays.
	if len(j.PlayerHand) > baccaratMaxSliceLen || len(j.BankerHand) > baccaratMaxSliceLen ||
		len(j.ActionLog) > baccaratMaxSliceLen || len(j.History) > baccaratMaxSliceLen ||
		len(j.SideBetResults) > baccaratMaxSliceLen {
		return fmt.Errorf("baccarat: input array exceeds maximum allowed size")
	}

	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCards(0)
	}
	b.playerHand = j.PlayerHand
	if b.playerHand == nil {
		b.playerHand = make([]*Card, 0)
	}
	b.bankerHand = j.BankerHand
	if b.bankerHand == nil {
		b.bankerHand = make([]*Card, 0)
	}
	if j.Chips != nil {
		b.chips = *j.Chips
	}
	b.betAmount = j.BetAmount
	b.betType = j.BetType
	b.phase = j.Phase
	b.gameEndFlag = j.GameEndFlag
	b.result = j.Result
	b.payout = j.Payout
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	b.history = j.History
	if b.history == nil {
		b.history = make([]int, 0)
	}
	b.playerPairBet = j.PlayerPairBet
	b.bankerPairBet = j.BankerPairBet
	b.sideBetResults = j.SideBetResults
	if b.sideBetResults == nil {
		b.sideBetResults = make([]*BacSideBetResult, 0)
	}
	return nil
}
