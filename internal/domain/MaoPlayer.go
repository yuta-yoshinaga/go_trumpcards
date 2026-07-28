//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// MaoPlayer マオプレイヤークラス
type MaoPlayer struct {
	*GamePlayer
	RoundScoreHolder
	hasDeclared bool
}

// NewMaoPlayer コンストラクタ
func NewMaoPlayer(isHuman bool) *MaoPlayer {
	return &MaoPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態・宣言フラグを初期化）
func (p *MaoPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
	p.hasDeclared = false
}

// GetHasDeclared マオ宣言済みかどうか
func (p *MaoPlayer) GetHasDeclared() bool { return p.hasDeclared }

// SetHasDeclared 宣言フラグを設定する
func (p *MaoPlayer) SetHasDeclared(v bool) { p.hasDeclared = v }

// maoPlayerJSON is the JSON wire format for MaoPlayer.
type maoPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	HasDeclared      bool              `json:"hd"`
}

// MarshalJSON implements json.Marshaler.
func (p *MaoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(maoPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		HasDeclared:      p.hasDeclared,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MaoPlayer) UnmarshalJSON(data []byte) error {
	var j maoPlayerJSON
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
