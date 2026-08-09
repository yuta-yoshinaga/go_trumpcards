package domain

import "encoding/json"

// PageOnePlayer ページワンプレイヤークラス
type PageOnePlayer struct {
	*GamePlayer
	RoundScoreHolder
	hasDeclared bool
}

// NewPageOnePlayer コンストラクタ
func NewPageOnePlayer(isHuman bool) *PageOnePlayer {
	return &PageOnePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態・宣言フラグを初期化）
func (p *PageOnePlayer) ResetRound() {
	resetRoundScored(p)
	p.hasDeclared = false
}

// GetHasDeclared 宣言済みかどうか
func (p *PageOnePlayer) GetHasDeclared() bool { return p.hasDeclared }

// SetHasDeclared 宣言フラグを設定する
func (p *PageOnePlayer) SetHasDeclared(v bool) { p.hasDeclared = v }

// pageOnePlayerJSON is the JSON wire format for PageOnePlayer.
type pageOnePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	HasDeclared      bool              `json:"hd"`
}

// MarshalJSON implements json.Marshaler.
func (p *PageOnePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pageOnePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		HasDeclared:      p.hasDeclared,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PageOnePlayer) UnmarshalJSON(data []byte) error {
	var j pageOnePlayerJSON
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
