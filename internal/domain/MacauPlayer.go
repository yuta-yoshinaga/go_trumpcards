//go:build !js || !wasm || solo

package domain

import "encoding/json"

// MacauPlayer マカオプレイヤークラス
type MacauPlayer struct {
	*GamePlayer
	RoundScoreHolder
	hasDeclared bool
}

// NewMacauPlayer コンストラクタ
func NewMacauPlayer(isHuman bool) *MacauPlayer {
	return &MacauPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態・宣言フラグを初期化）
func (p *MacauPlayer) ResetRound() {
	resetRoundScored(p)
	p.hasDeclared = false
}

// GetHasDeclared マカオ宣言済みかどうか
func (p *MacauPlayer) GetHasDeclared() bool { return p.hasDeclared }

// SetHasDeclared 宣言フラグを設定する
func (p *MacauPlayer) SetHasDeclared(v bool) { p.hasDeclared = v }

// macauPlayerJSON is the JSON wire format for MacauPlayer.
type macauPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	HasDeclared      bool              `json:"hd"`
}

// MarshalJSON implements json.Marshaler.
func (p *MacauPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(macauPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		HasDeclared:      p.hasDeclared,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MacauPlayer) UnmarshalJSON(data []byte) error {
	var j macauPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	p.hasDeclared = j.HasDeclared
	return nil
}
