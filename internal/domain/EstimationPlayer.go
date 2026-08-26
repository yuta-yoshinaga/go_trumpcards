//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// EstimationCallType 宣言の種類。**種類で得点の振れ幅が変わる。**
type EstimationCallType int

// Estimation の宣言種別
const (
	// EstimationCallNormal 通常の宣言
	EstimationCallNormal EstimationCallType = iota
	// EstimationCallDash 0 宣言（Dash Call）。13 枚持って 1 つも取らない宣言。
	EstimationCallDash
	// EstimationCallRisk 最大宣言（Risk）。そのラウンドで最も高く宣言した人。
	EstimationCallRisk
)

// EstimationPlayer エスティメーションのプレイヤー
type EstimationPlayer struct {
	*GamePlayer
	TrickHolder
	// bid は宣言したトリック数 (-1: 未宣言)。
	bid int
	// callType は宣言の種類。Risk は全員の宣言が出そろってから確定する。
	callType EstimationCallType
	// roundScore は直前のラウンドの増減、totalScore は累計。
	roundScore int
	totalScore int
}

// NewEstimationPlayer コンストラクタ
func NewEstimationPlayer(isHuman bool) *EstimationPlayer {
	return &EstimationPlayer{GamePlayer: NewGamePlayer(isHuman), bid: -1}
}

// ResetGame ゲーム全体をリセットする
func (p *EstimationPlayer) ResetGame() {
	p.ResetRound()
	p.totalScore = 0
}

// ResetRound 1 ラウンド分の状態を初期化する
func (p *EstimationPlayer) ResetRound() {
	resetPlayerRound(p)
	p.bid = -1
	p.callType = EstimationCallNormal
	p.roundScore = 0
}

// GetBid 宣言したトリック数 (-1: 未宣言)
func (p *EstimationPlayer) GetBid() int { return p.bid }

// SetBid 宣言を設定する
func (p *EstimationPlayer) SetBid(n int) { p.bid = n }

// GetCallType 宣言の種類
func (p *EstimationPlayer) GetCallType() EstimationCallType { return p.callType }

// SetCallType 宣言の種類を設定する
func (p *EstimationPlayer) SetCallType(t EstimationCallType) { p.callType = t }

// GetRoundScore 直前のラウンドの増減
func (p *EstimationPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 直前のラウンドの増減を設定する
func (p *EstimationPlayer) SetRoundScore(n int) { p.roundScore = n }

// GetTotalScore 累計得点
func (p *EstimationPlayer) GetTotalScore() int { return p.totalScore }

// AddTotalScore 累計得点に加算する
func (p *EstimationPlayer) AddTotalScore(n int) { p.totalScore += n }

// SetTotalScore 累計得点を設定する（復元・テスト用）
func (p *EstimationPlayer) SetTotalScore(n int) { p.totalScore = n }

// estimationPlayerJSON is the JSON wire format for EstimationPlayer.
type estimationPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// 宣言・種別・得点は往復させる。Worker はリクエストごとに KV から作り直す
	// ので、抜けると宣言がやり直しになったり得点が消える (#4478)。
	Bid        int                `json:"bd"`
	CallType   EstimationCallType `json:"ct"`
	RoundScore int                `json:"rs"`
	TotalScore int                `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *EstimationPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(estimationPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Bid:         p.bid,
		CallType:    p.callType,
		RoundScore:  p.roundScore,
		TotalScore:  p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *EstimationPlayer) UnmarshalJSON(data []byte) error {
	var j estimationPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bid = j.Bid
	p.callType = j.CallType
	p.roundScore = j.RoundScore
	p.totalScore = j.TotalScore
	return nil
}
