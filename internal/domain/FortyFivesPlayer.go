//go:build !js || !wasm || casino

package domain

import "encoding/json"

// FortyFivesPlayer オークション・フォーティファイブズのプレイヤークラス。手札
// （GamePlayer）と獲得トリック（TrickHolder）に加え、現ラウンドのトリック数を保持する。
// 点はチーム単位で FortyFives 本体が管理する。
type FortyFivesPlayer struct {
	*GamePlayer
	TrickHolder
	roundTricks int
}

// NewFortyFivesPlayer コンストラクタ
func NewFortyFivesPlayer(isHuman bool) *FortyFivesPlayer {
	return &FortyFivesPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・ラウンドトリック数を初期化）
func (p *FortyFivesPlayer) ResetRound() {
	resetPlayerRound(p)
	p.roundTricks = 0
}

// GetRoundTricks 現ラウンドの獲得トリック数を返す。
func (p *FortyFivesPlayer) GetRoundTricks() int { return p.roundTricks }

// IncRoundTricks 現ラウンドの獲得トリック数を 1 増やす。
func (p *FortyFivesPlayer) IncRoundTricks() { p.roundTricks++ }

// fortyFivesPlayerJSON is the JSON wire format for FortyFivesPlayer.
type fortyFivesPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	RoundTricks int          `json:"rt"`
}

// MarshalJSON implements json.Marshaler.
func (p *FortyFivesPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortyFivesPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder, RoundTricks: p.roundTricks})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *FortyFivesPlayer) UnmarshalJSON(data []byte) error {
	var j fortyFivesPlayerJSON
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
	p.roundTricks = j.RoundTricks
	return nil
}
