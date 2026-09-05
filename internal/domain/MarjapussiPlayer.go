//go:build !js || !wasm || extra5

package domain

import "encoding/json"

// MarjapussiPlayer マルヤプッシ (Marjapussi) のプレイヤークラス。手札 (GamePlayer) と獲得トリック
// (TrickHolder) を保持する。点はゲーム本体が管理する。
type MarjapussiPlayer struct {
	*GamePlayer
	TrickHolder
}

// NewMarjapussiPlayer コンストラクタ
func NewMarjapussiPlayer(isHuman bool) *MarjapussiPlayer {
	return &MarjapussiPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *MarjapussiPlayer) ResetRound() {
	resetPlayerRound(p)
}

// marjapussiPlayerJSON is the JSON wire format for MarjapussiPlayer.
type marjapussiPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *MarjapussiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(marjapussiPlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MarjapussiPlayer) UnmarshalJSON(data []byte) error {
	var j marjapussiPlayerJSON
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
	return nil
}
