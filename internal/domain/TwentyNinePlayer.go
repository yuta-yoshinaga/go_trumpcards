//go:build !js || !wasm || casino

package domain

import "encoding/json"

// TwentyNinePlayer トゥエンティナインのプレイヤークラス。手札（GamePlayer）と獲得
// トリック（TrickHolder）を保持する。点はチーム単位で TwentyNine 本体が管理する。
type TwentyNinePlayer struct {
	*GamePlayer
	TrickHolder
}

// NewTwentyNinePlayer コンストラクタ
func NewTwentyNinePlayer(isHuman bool) *TwentyNinePlayer {
	return &TwentyNinePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *TwentyNinePlayer) ResetRound() {
	resetPlayerRound(p)
}

// twentyNinePlayerJSON is the JSON wire format for TwentyNinePlayer.
type twentyNinePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
}

// MarshalJSON implements json.Marshaler.
func (p *TwentyNinePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(twentyNinePlayerJSON{GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TwentyNinePlayer) UnmarshalJSON(data []byte) error {
	var j twentyNinePlayerJSON
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
