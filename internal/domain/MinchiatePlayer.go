//go:build !js || !wasm || solo

package domain

import "encoding/json"

// MinchiatePlayer ミンキアーテのプレイヤークラス。手札 (GamePlayer) と獲得
// トリック (TrickHolder) を保持する。得点はチーム単位でゲーム本体が管理する。
type MinchiatePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewMinchiatePlayer コンストラクタ
func NewMinchiatePlayer(isHuman bool) *MinchiatePlayer {
	return &MinchiatePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット (トリック・手札・終了状態を初期化)
func (p *MinchiatePlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// minchiatePlayerJSON is the JSON wire format for MinchiatePlayer.
type minchiatePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *MinchiatePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(minchiatePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MinchiatePlayer) UnmarshalJSON(data []byte) error {
	var j minchiatePlayerJSON
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
