//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// レッドドッグフェーズ定数
const (
	RedDogPhaseBet            = 1 // ベットフェーズ
	RedDogPhaseInitialDealt   = 2 // 初手2枚配布済み（ResolveInitial 待ち）
	RedDogPhaseSpreadDecision = 3 // スプレッド判断フェーズ（レイズ可能）
	RedDogPhasePairThird      = 4 // ペア時の3枚目即時解決用の内部フェーズ（外部からは観測されない）
	RedDogPhaseEnd            = 5 // 終了フェーズ
)

// レッドドッグデフォルト値
const (
	RedDogDefaultChips = 1000  // デフォルトチップ
	RedDogMinBet       = 10    // 最低ベット額
	RedDogMaxBet       = 10000 // 最大ベット額
	RedDogInitialCards = 2     // 初手カード枚数
)

// レッドドッグ配当倍率（ボーナス分のみ。元ベットは別途返却）
const (
	RedDogPayPair    = 11 // 同位ペア＋3枚目同位 11:1
	RedDogPaySpread1 = 5  // スプレッド1 5:1
	RedDogPaySpread2 = 4  // スプレッド2 4:1
	RedDogPaySpread3 = 2  // スプレッド3 2:1
	RedDogPaySpread4 = 1  // スプレッド4以上 1:1
)

// RedDog レッドドッグクラス
type RedDog struct {
	trumpCards   *TrumpCards
	initialCards []*Card
	thirdCard    *Card
	chips        ChipHolder
	ante         int
	raise        int
	spread       int
	phase        int
	gameEndFlag  bool
	result       GameResult
	totalPayout  int
	actionLogBase
}

// NewRedDog コンストラクタ
func NewRedDog(trumpCards *TrumpCards) *RedDog {
	trumpCards.Shuffle()
	return &RedDog{
		trumpCards: trumpCards,
		phase:      RedDogPhaseBet,
	}
}

// NewDefaultRedDog デフォルト設定のレッドドッグを生成するファクトリ関数
func NewDefaultRedDog() *RedDog {
	rd := NewRedDog(NewTrumpCards(0))
	rd.chips.SetChips(RedDogDefaultChips)
	return rd
}

// Reset ゲーム初期化
func (rd *RedDog) Reset() {
	rd.gameEndFlag = false
	rd.phase = RedDogPhaseBet
	rd.initialCards = nil
	rd.thirdCard = nil
	rd.ante = 0
	rd.raise = 0
	rd.spread = 0
	rd.result = 0
	rd.totalPayout = 0
	rd.actionLog = nil
	if rd.chips.GetChips() < RedDogMinBet {
		rd.chips.SetChips(RedDogDefaultChips)
	}
	rd.trumpCards = NewTrumpCards(0)
	for range 10 {
		rd.trumpCards.Shuffle()
	}
}

// Bet 初回ベット＆カード配布。
func (rd *RedDog) Bet(amount int) error {
	if rd.phase != RedDogPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < RedDogMinBet || amount%RedDogMinBet != 0 || amount > RedDogMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !rd.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	rd.ante = amount
	rd.appendLog(0, "bet", fmt.Sprintf("ante=%d", amount), nil)

	rd.dealInitial()
	rd.phase = RedDogPhaseInitialDealt
	return nil
}

// dealInitial 初手2枚配布
func (rd *RedDog) dealInitial() {
	rd.initialCards = make([]*Card, 0, RedDogInitialCards)
	for range RedDogInitialCards {
		rd.initialCards = append(rd.initialCards, rd.trumpCards.DrawCard())
	}
	rd.appendLog(-1, "deal", "dealt 2 initial cards", rd.initialCards)
}

// rankOf カード値をランク化（A=14, K=13, …, 2=2）
func rankOf(c *Card) int {
	v := c.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// ResolveInitial 初手2枚を評価しフェーズ遷移
func (rd *RedDog) ResolveInitial() {
	if len(rd.initialCards) != 2 {
		return
	}
	r1, r2 := rankOf(rd.initialCards[0]), rankOf(rd.initialCards[1])
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	diff := r2 - r1
	switch diff {
	case 0:
		// ペア → プレイヤーの判断は介在しないため、3枚目を即引いて決着させる
		// （PairThird で止めるとどのコマンドも受け付けずデッドエンドになる）
		rd.spread = 0
		rd.phase = RedDogPhasePairThird
		rd.appendLog(-1, "pair", "initial pair, drawing third card", nil)
		rd.dealThird()
		rd.ResolveThird()
	case 1:
		// 連続 → プッシュ（即終了、アンテ返却）
		rd.spread = 0
		rd.result = GameResultDraw
		rd.totalPayout = rd.ante
		rd.chips.AddChips(rd.ante)
		rd.gameEndFlag = true
		rd.phase = RedDogPhaseEnd
		rd.appendLog(-1, "push", "consecutive ranks → push", nil)
	default:
		rd.spread = diff - 1
		rd.phase = RedDogPhaseSpreadDecision
		rd.appendLog(-1, "spread", fmt.Sprintf("spread=%d", rd.spread), nil)
	}
}

// Raise レイズ（最大アンテと同額まで）
func (rd *RedDog) Raise(amount int) error {
	if rd.phase != RedDogPhaseSpreadDecision {
		return NewDomainError(ErrWrongPhase, "Raise is only allowed during the spread decision phase.")
	}
	if amount < RedDogMinBet || amount%RedDogMinBet != 0 || amount > rd.ante {
		return NewDomainError(ErrInvalidAmount, "Invalid raise amount.")
	}
	if !rd.chips.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	rd.raise = amount
	rd.appendLog(0, "raise", fmt.Sprintf("raise=%d", amount), nil)
	rd.dealThird()
	rd.ResolveThird()
	return nil
}

// Stay レイズせず3枚目を引く
func (rd *RedDog) Stay() error {
	if rd.phase != RedDogPhaseSpreadDecision {
		return NewDomainError(ErrWrongPhase, "Stay is only allowed during the spread decision phase.")
	}
	rd.appendLog(0, "stay", "stay (no raise)", nil)
	rd.dealThird()
	rd.ResolveThird()
	return nil
}

// dealThird 3枚目を1枚引く（既にセット済みなら何もしない）
func (rd *RedDog) dealThird() {
	if rd.thirdCard != nil {
		return
	}
	rd.thirdCard = rd.trumpCards.DrawCard()
	rd.appendLog(-1, "deal", "dealt third card", []*Card{rd.thirdCard})
}

// ResolveThird 3枚目評価＆ペイアウト
func (rd *RedDog) ResolveThird() {
	if rd.thirdCard == nil || rd.gameEndFlag {
		return
	}
	switch rd.phase {
	case RedDogPhasePairThird:
		if rankOf(rd.thirdCard) == rankOf(rd.initialCards[0]) {
			rd.result = GameResultWin
			rd.totalPayout = rd.ante + rd.ante*RedDogPayPair
		} else {
			rd.result = GameResultDraw
			rd.totalPayout = rd.ante // 返金
		}
	case RedDogPhaseSpreadDecision:
		r1, r2 := rankOf(rd.initialCards[0]), rankOf(rd.initialCards[1])
		if r1 > r2 {
			r1, r2 = r2, r1
		}
		r3 := rankOf(rd.thirdCard)
		if r3 > r1 && r3 < r2 {
			rd.result = GameResultWin
			multiplier := rd.payoutMultiplier()
			totalBet := rd.ante + rd.raise
			rd.totalPayout = totalBet + totalBet*multiplier
		} else {
			rd.result = GameResultLose
			rd.totalPayout = 0
		}
	}
	if rd.totalPayout > 0 {
		rd.chips.AddChips(rd.totalPayout)
	}
	rd.gameEndFlag = true
	rd.phase = RedDogPhaseEnd

	var s string
	switch rd.result {
	case GameResultWin:
		s = "player wins"
	case GameResultLose:
		s = "player loses"
	default:
		s = "push"
	}
	rd.appendLog(-1, "result", s, nil)
}

// payoutMultiplier スプレッドに基づく配当倍率
func (rd *RedDog) payoutMultiplier() int {
	switch rd.spread {
	case 1:
		return RedDogPaySpread1
	case 2:
		return RedDogPaySpread2
	case 3:
		return RedDogPaySpread3
	default:
		return RedDogPaySpread4
	}
}

// --- Getters ---

// GetInitialCards 初手2枚
func (rd *RedDog) GetInitialCards() []*Card { return rd.initialCards }

// GetThirdCard 3枚目
func (rd *RedDog) GetThirdCard() *Card { return rd.thirdCard }

// GetPhase フェーズ
func (rd *RedDog) GetPhase() int { return rd.phase }

// GetGameEndFlag 終了フラグ
func (rd *RedDog) GetGameEndFlag() bool { return rd.gameEndFlag }

// GetAnte アンテ額
func (rd *RedDog) GetAnte() int { return rd.ante }

// GetRaise レイズ額
func (rd *RedDog) GetRaise() int { return rd.raise }

// GetSpread スプレッド
func (rd *RedDog) GetSpread() int { return rd.spread }

// GetResult 結果
func (rd *RedDog) GetResult() GameResult { return rd.result }

// GetTotalPayout 合計配当
func (rd *RedDog) GetTotalPayout() int { return rd.totalPayout }

// GetChips チップ
func (rd *RedDog) GetChips() int { return rd.chips.GetChips() }

// --- Test helpers ---

// SetPhase テスト用
func (rd *RedDog) SetPhase(phase int) { rd.phase = phase }

// SetInitialCards テスト用
func (rd *RedDog) SetInitialCards(cards []*Card) { rd.initialCards = cards }

// SetThirdCard テスト用
func (rd *RedDog) SetThirdCard(c *Card) { rd.thirdCard = c }

// SetChips テスト用
func (rd *RedDog) SetChips(chips int) { rd.chips.SetChips(chips) }

// redDogJSON は RedDog の JSON ワイヤーフォーマット
type redDogJSON struct {
	TrumpCards   *TrumpCards       `json:"tc"`
	InitialCards []*Card           `json:"ic"`
	ThirdCard    *Card             `json:"tr"`
	Chips        *ChipHolder       `json:"ch"`
	Ante         int               `json:"an"`
	Raise        int               `json:"rs"`
	Spread       int               `json:"sp"`
	Phase        int               `json:"ps"`
	GameEndFlag  bool              `json:"ge"`
	Result       GameResult        `json:"gr"`
	TotalPayout  int               `json:"tp"`
	ActionLog    []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (rd *RedDog) MarshalJSON() ([]byte, error) {
	return json.Marshal(redDogJSON{
		TrumpCards:   rd.trumpCards,
		InitialCards: rd.initialCards,
		ThirdCard:    rd.thirdCard,
		Chips:        &rd.chips,
		Ante:         rd.ante,
		Raise:        rd.raise,
		Spread:       rd.spread,
		Phase:        rd.phase,
		GameEndFlag:  rd.gameEndFlag,
		Result:       rd.result,
		TotalPayout:  rd.totalPayout,
		ActionLog:    rd.actionLog,
	})
}

// redDogMaxSliceLen caps slice sizes during deserialisation.
const redDogMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (rd *RedDog) UnmarshalJSON(data []byte) error {
	var j redDogJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.InitialCards) > redDogMaxSliceLen || len(j.ActionLog) > redDogMaxSliceLen {
		return fmt.Errorf("reddog: input array exceeds maximum allowed size")
	}
	rd.trumpCards = j.TrumpCards
	if rd.trumpCards == nil {
		rd.trumpCards = NewTrumpCards(0)
	}
	rd.initialCards = j.InitialCards
	if rd.initialCards == nil {
		rd.initialCards = make([]*Card, 0)
	}
	rd.thirdCard = j.ThirdCard
	if j.Chips != nil {
		rd.chips = *j.Chips
	}
	rd.ante = j.Ante
	rd.raise = j.Raise
	rd.spread = j.Spread
	rd.phase = j.Phase
	rd.gameEndFlag = j.GameEndFlag
	rd.result = j.Result
	rd.totalPayout = j.TotalPayout
	rd.actionLog = j.ActionLog
	if rd.actionLog == nil {
		rd.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
