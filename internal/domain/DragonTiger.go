//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// ドラゴンタイガーフェーズ定数
const (
	DragonTigerPhaseBet = 1 // ベットフェーズ
	DragonTigerPhaseEnd = 2 // 終了フェーズ
)

// ドラゴンタイガーベットタイプ定数
const (
	DragonTigerBetDragon = 0 // ドラゴンベット
	DragonTigerBetTiger  = 1 // タイガーベット
	DragonTigerBetTie    = 2 // タイベット
)

// ドラゴンタイガーデフォルト値
const (
	DragonTigerDefaultChips  = 1000  // デフォルトチップ
	DragonTigerMinBet        = 10    // 最低ベット額
	DragonTigerMaxBet        = 10000 // 最大ベット額
	DragonTigerTiePayoutRate = 8     // タイ配当倍率（8:1 — 標準ルール）
)

// ドラゴンタイガー結果定数（罫線用）
const (
	DragonTigerResultDragon = 0 // ドラゴン勝利
	DragonTigerResultTiger  = 1 // タイガー勝利
	DragonTigerResultTie    = 2 // タイ
)

// DragonTiger ドラゴンタイガーゲーム本体。
// ルール: ドラゴン枠とタイガー枠に1枚ずつカードを配り、ランクの高い側が勝ち
// （A=1 が最弱、K=13 が最強）。タイベット未実施でタイになった場合、メイン
// ベット額の半分が返還される（業界標準の "tie returns half" ルール）。
type DragonTiger struct {
	trumpCards  *TrumpCards
	dragonCard  *Card
	tigerCard   *Card
	chips       ChipHolder
	betAmount   int
	betType     int
	phase       int
	gameEndFlag bool
	result      GameResult
	payout      int
	actionLogBase
	history []int // 罫線（Big Road）履歴
}

// NewDragonTiger コンストラクタ
func NewDragonTiger(trumpCards *TrumpCards) *DragonTiger {
	trumpCards.Shuffle()
	return &DragonTiger{
		trumpCards: trumpCards,
		phase:      DragonTigerPhaseBet,
	}
}

// NewDefaultDragonTiger デフォルト設定のドラゴンタイガーを生成するファクトリ関数
func NewDefaultDragonTiger() *DragonTiger {
	dt := NewDragonTiger(NewTrumpCards(0))
	dt.chips.SetChips(DragonTigerDefaultChips)
	return dt
}

// Reset ゲーム初期化（履歴は保持）
func (dt *DragonTiger) Reset() {
	dt.gameEndFlag = false
	dt.phase = DragonTigerPhaseBet
	dt.dragonCard = nil
	dt.tigerCard = nil
	dt.betAmount = 0
	dt.betType = 0
	dt.result = 0
	dt.payout = 0
	dt.actionLog = nil
	if dt.chips.GetChips() < DragonTigerMinBet {
		dt.chips.SetChips(DragonTigerDefaultChips)
	}
	dt.trumpCards = NewTrumpCards(0)
	dt.trumpCards.Shuffle()
}

// ClearHistory 罫線履歴をクリアする
func (dt *DragonTiger) ClearHistory() {
	dt.history = nil
}

// Bet ベット＆ゲーム実行（ベット後に全自動進行）
func (dt *DragonTiger) Bet(amount, betType int) error {
	if dt.phase != DragonTigerPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if betType < DragonTigerBetDragon || betType > DragonTigerBetTie {
		return NewDomainError(ErrInvalidPlay, "Invalid bet type.")
	}
	if amount < DragonTigerMinBet || amount%DragonTigerMinBet != 0 || amount > DragonTigerMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !dt.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	dt.betAmount = amount
	dt.betType = betType
	dt.appendLog(0, "bet", fmt.Sprintf("bet %d on %s", amount, dragonTigerBetTypeName(betType)), nil)

	dt.deal()
	dt.judge()
	return nil
}

// deal ドラゴン枠とタイガー枠に1枚ずつ配る
func (dt *DragonTiger) deal() {
	if dt.dragonCard == nil {
		dt.dragonCard = dt.trumpCards.DrawCard()
	}
	if dt.tigerCard == nil {
		dt.tigerCard = dt.trumpCards.DrawCard()
	}
	dt.appendLog(-1, "deal", "dealt dragon and tiger cards", []*Card{dt.dragonCard, dt.tigerCard})
}

// judge 勝敗判定＆配当計算
func (dt *DragonTiger) judge() {
	dr, tr := dragonTigerRankOf(dt.dragonCard), dragonTigerRankOf(dt.tigerCard)
	switch {
	case dr > tr:
		dt.result = GameResultWin // Dragon wins
		dt.appendLog(-1, "result", "dragon wins", nil)
	case tr > dr:
		dt.result = GameResultLose // Tiger wins (player's "lose" semantics for the dragon-default frame)
		dt.appendLog(-1, "result", "tiger wins", nil)
	default:
		dt.result = GameResultDraw
		dt.appendLog(-1, "result", "tie", nil)
	}

	switch dt.result {
	case GameResultWin:
		dt.history = append(dt.history, DragonTigerResultDragon)
	case GameResultLose:
		dt.history = append(dt.history, DragonTigerResultTiger)
	default:
		dt.history = append(dt.history, DragonTigerResultTie)
	}

	dt.payout = dt.calculatePayout()
	dt.chips.AddChips(dt.payout)

	dt.gameEndFlag = true
	dt.phase = DragonTigerPhaseEnd
}

// calculatePayout 配当計算。Dragon/Tiger ベットは 1:1（賭け金返還込みで 2 倍）。
// Tie ベットは 8:1。Dragon/Tiger ベットでタイが出た場合は賭け金の半額が返還される。
func (dt *DragonTiger) calculatePayout() int {
	switch dt.betType {
	case DragonTigerBetDragon:
		switch dt.result {
		case GameResultWin:
			return dt.betAmount * 2
		case GameResultDraw:
			return dt.betAmount / 2 // half-refund on tie
		default:
			return 0
		}
	case DragonTigerBetTiger:
		switch dt.result {
		case GameResultLose: // tiger wins
			return dt.betAmount * 2
		case GameResultDraw:
			return dt.betAmount / 2
		default:
			return 0
		}
	case DragonTigerBetTie:
		if dt.result == GameResultDraw {
			return dt.betAmount + dt.betAmount*DragonTigerTiePayoutRate
		}
		return 0
	}
	return 0
}

// dragonTigerRankOf カード値をランク化（A=1, J=11, Q=12, K=13 — Aceが最弱）。
func dragonTigerRankOf(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}

// dragonTigerBetTypeName ベットタイプ名
func dragonTigerBetTypeName(betType int) string {
	switch betType {
	case DragonTigerBetDragon:
		return "dragon"
	case DragonTigerBetTiger:
		return "tiger"
	case DragonTigerBetTie:
		return "tie"
	default:
		return "unknown"
	}
}

// --- Getters ---

// GetDragonCard ドラゴン枠のカード
func (dt *DragonTiger) GetDragonCard() *Card { return dt.dragonCard }

// GetTigerCard タイガー枠のカード
func (dt *DragonTiger) GetTigerCard() *Card { return dt.tigerCard }

// GetPhase 現在のフェーズ
func (dt *DragonTiger) GetPhase() int { return dt.phase }

// GetGameEndFlag ゲーム終了フラグ
func (dt *DragonTiger) GetGameEndFlag() bool { return dt.gameEndFlag }

// GetBetAmount ベット額
func (dt *DragonTiger) GetBetAmount() int { return dt.betAmount }

// GetBetType ベットタイプ
func (dt *DragonTiger) GetBetType() int { return dt.betType }

// GetResult ゲーム結果
func (dt *DragonTiger) GetResult() GameResult { return dt.result }

// GetPayout 配当金額
func (dt *DragonTiger) GetPayout() int { return dt.payout }

// GetChips チップ
func (dt *DragonTiger) GetChips() int { return dt.chips.GetChips() }

// GetHistory 罫線履歴を取得する
func (dt *DragonTiger) GetHistory() []int { return dt.history }

// --- Test helpers ---

// SetPhase テスト用
func (dt *DragonTiger) SetPhase(phase int) { dt.phase = phase }

// SetDragonCard テスト用
func (dt *DragonTiger) SetDragonCard(c *Card) { dt.dragonCard = c }

// SetTigerCard テスト用
func (dt *DragonTiger) SetTigerCard(c *Card) { dt.tigerCard = c }

// SetBetAmount テスト用
func (dt *DragonTiger) SetBetAmount(amount int) { dt.betAmount = amount }

// SetBetType テスト用
func (dt *DragonTiger) SetBetType(betType int) { dt.betType = betType }

// SetResult テスト用
func (dt *DragonTiger) SetResult(result GameResult) { dt.result = result }

// SetGameEndFlag テスト用
func (dt *DragonTiger) SetGameEndFlag(flag bool) { dt.gameEndFlag = flag }

// SetChips テスト用
func (dt *DragonTiger) SetChips(chips int) { dt.chips.SetChips(chips) }

// SetHistory テスト用
func (dt *DragonTiger) SetHistory(history []int) { dt.history = history }

// dragonTigerJSON は DragonTiger の JSON ワイヤーフォーマット
type dragonTigerJSON struct {
	TrumpCards  *TrumpCards       `json:"tc"`
	DragonCard  *Card             `json:"dc"`
	TigerCard   *Card             `json:"tg"`
	Chips       *ChipHolder       `json:"ch"`
	BetAmount   int               `json:"ba"`
	BetType     int               `json:"bt"`
	Phase       int               `json:"ps"`
	GameEndFlag bool              `json:"ge"`
	Result      GameResult        `json:"rs"`
	Payout      int               `json:"po"`
	ActionLog   []*ActionLogEntry `json:"al"`
	History     []int             `json:"hi"`
}

// MarshalJSON implements json.Marshaler.
func (dt *DragonTiger) MarshalJSON() ([]byte, error) {
	return json.Marshal(dragonTigerJSON{
		TrumpCards:  dt.trumpCards,
		DragonCard:  dt.dragonCard,
		TigerCard:   dt.tigerCard,
		Chips:       &dt.chips,
		BetAmount:   dt.betAmount,
		BetType:     dt.betType,
		Phase:       dt.phase,
		GameEndFlag: dt.gameEndFlag,
		Result:      dt.result,
		Payout:      dt.payout,
		ActionLog:   dt.actionLog,
		History:     dt.history,
	})
}

// dragonTigerMaxSliceLen はデシリアライズ時のスライス長の上限
const dragonTigerMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (dt *DragonTiger) UnmarshalJSON(data []byte) error {
	var j dragonTigerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > dragonTigerMaxSliceLen || len(j.History) > dragonTigerMaxSliceLen {
		return fmt.Errorf("dragontiger: input array exceeds maximum allowed size")
	}
	dt.trumpCards = j.TrumpCards
	if dt.trumpCards == nil {
		dt.trumpCards = NewTrumpCards(0)
	}
	dt.dragonCard = j.DragonCard
	dt.tigerCard = j.TigerCard
	if j.Chips != nil {
		dt.chips = *j.Chips
	}
	dt.betAmount = j.BetAmount
	dt.betType = j.BetType
	dt.phase = j.Phase
	dt.gameEndFlag = j.GameEndFlag
	dt.result = j.Result
	dt.payout = j.Payout
	dt.actionLog = j.ActionLog
	if dt.actionLog == nil {
		dt.actionLog = make([]*ActionLogEntry, 0)
	}
	dt.history = j.History
	if dt.history == nil {
		dt.history = make([]int, 0)
	}
	return nil
}
